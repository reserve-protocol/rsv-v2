package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func outputGoFile(contractName string) string {
	return fmt.Sprintf("abi/%v.go", contractName)
}

type target struct {
	Filename     string
	ContractName string
	SolcVersion  string
	OptimizeRuns string
}

var targets = []target{
	target{Filename: "contracts/Reserve.sol", ContractName: "Reserve", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
	target{Filename: "contracts/ReserveV2.sol", ContractName: "ReserveV2", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
	target{Filename: "contracts/ReserveEternalStorage.sol", ContractName: "ReserveEternalStorage", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
	target{Filename: "contracts/Basket.sol", ContractName: "Basket", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Manager.sol", ContractName: "Manager", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Proposal.sol", ContractName: "Proposal", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Vault.sol", ContractName: "Vault", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Ownable.sol", ContractName: "Ownable", SolcVersion: "0.5.8", OptimizeRuns: "1"},
}

func main() {
	// Compile.
	type compiledOutput struct {
		ABI           string
		Bin           string
		BinRuntime    string `json:"bin-runtime"`
		Srcmap        string
		SrcmapRuntime string `json:"srcmap-runtime"`
	}
	for _, t := range targets {
		var compilationResult struct {
			Contracts  map[string]compiledOutput
			SourceList []string
		}

		// Run solc.
		stdout := new(bytes.Buffer)
		cmd := exec.Command(
			"solc",
			append(
				[]string{
					"--allow-paths contracts",
					"--optimize",
					"--optimize-runs=" + t.OptimizeRuns,
					"--combined-json=abi,bin,bin-runtime,srcmap,srcmap-runtime",
				},
				t.Filename,
			)...,
		)
		cmd.Env = append(cmd.Env, "SOLC_VERSION="+t.SolcVersion)
		cmd.Stdin = os.Stdin // solc doesn't need stdin to be set, but trailofbits' solc-select does
		cmd.Stdout = stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("%s\nsolc failed: %v\n", stdout.Bytes(), err)
		}

		// Parse the JSON output from solc.
		check(json.NewDecoder(bytes.NewReader(stdout.Bytes())).Decode(&compilationResult), "parsing solc output\n"+stdout.String())

		output := compilationResult.Contracts[t.Filename+":"+t.ContractName]
		// Generate bindings.
		code, err := bind.Bind([]string{t.ContractName}, []string{output.ABI}, []string{output.Bin}, "abi", bind.LangGo)
		check(err, "generating Go bindings")

		// Write to .go file.
		check(os.MkdirAll("abi", 0755), "creating abi directory")
		name := outputGoFile(t.ContractName)
		check(ioutil.WriteFile(name, []byte(code), 0644), "writing "+name)

		// Generate event bindings.
		//
		// We generate a String() function for each event and a
		// Parse<ContractName>Log(*types.Log) function for each contract.
		{
			buf := new(bytes.Buffer)
			parsedABI, err := abi.JSON(bytes.NewReader([]byte(output.ABI)))
			check(err, "parsing ABI JSON")
			check(template.Must(template.New("").Funcs(
				template.FuncMap{
					"flags": func(inputs abi.Arguments) string {
						result := make([]string, len(inputs))
						for i := range result {
							switch inputs[i].Type.String() {
							case "string":
								result[i] = "%q"
							default:
								result[i] = "%v"
							}
						}
						return strings.Join(result, ", ")
					},
					"format": func(inputs abi.Arguments) string {
						result := make([]string, len(inputs))
						for i := range result {
							arg := "e." + abi.ToCamelCase(inputs[i].Name)
							switch inputs[i].Type.String() {
							case "address":
								arg = arg + ".Hex()"
							}
							result[i] = arg
						}
						return strings.Join(result, ",")
					},
				},
			).Parse(`
                // This file is auto-generated. Do not edit.

                package abi

                import (
                    "fmt"

                    "github.com/ethereum/go-ethereum/core/types"
                )

                {{$contract := .Contract}}

                func (c *{{$contract}}Filterer) ParseLog(log *types.Log) (fmt.Stringer, error) {
                    var event fmt.Stringer
                    var eventName string
                    switch log.Topics[0].Hex() {
                    {{- range .Events}}
                    case {{with .Id}}{{printf "%q" .Hex}}{{end}}: // {{.Name}}
                        event = new({{$contract}}{{.Name}})
                        eventName = "{{.Name}}"
                    {{- end}}
                    default:
                        return nil, fmt.Errorf("no such event hash for {{$contract}}: %v", log.Topics[0])
                    }

                    err := c.contract.UnpackLog(event, eventName, *log)
                    if err != nil {
                        return nil, err
                    }
                    return event, nil
                }

                {{range .Events}}
                func (e {{$contract}}{{.Name}}) String() string {
                    return fmt.Sprintf("{{$contract}}.{{.Name}}({{flags .Inputs}})",{{format .Inputs}})
                }
                {{end}}
            `)).Execute(buf, map[string]interface{}{
				"Contract": t.ContractName,
				"Events":   parsedABI.Events,
			}), "generating event bindings")
			eventCode, err := format.Source(buf.Bytes())
			check(err, "running gofmt")
			check(
				ioutil.WriteFile(
					filepath.Join("abi", t.ContractName+"Events.go"),
					eventCode,
					0644,
				),
				"writing event bindings to disk",
			)
		}

		// Write JSON artifacts for sol-coverage.
		{
			// sources records an ordering on source files.
			// sol-compiler uses a different format from solc for this,
			// so here we're converting from the former to the latter.

			check(os.MkdirAll("artifacts", 0755), "creating artifacts directory")
			// m is a helper for writing succinct json object literals
			type m map[string]interface{}

			sources := make(m)
			for i, source := range compilationResult.SourceList {
				sources[source] = m{"id": i}
			}

			compiledOutput := compilationResult.Contracts[t.Filename+":"+t.ContractName]
			b, err := json.Marshal(m{
				"schemaVersion": "2.0.0",
				"ContractName":  t.ContractName,
				"compilerOutput": m{
					"abi": json.RawMessage(compiledOutput.ABI),
					"evm": m{
						"bytecode": m{
							"object":    compiledOutput.Bin,
							"sourceMap": compiledOutput.Srcmap,
						},
						"deployedBytecode": m{
							"object":    compiledOutput.BinRuntime,
							"sourceMap": compiledOutput.SrcmapRuntime,
						},
					},
				},
				"sources": sources,
			})
			check(err, "json-encoding sol-compiler-style artifact")
			check(
				ioutil.WriteFile("artifacts/"+t.ContractName+".json", b, 0644),
				"writing sol-compiler-style artifact",
			)
		}

	}
}

func check(err error, msg string) {
	if err != nil {
		log.Fatal(msg, ": ", err)
	}
}
