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

const (
	combinedJsonDir = "evm"
	solCovDir       = "sol-coverage-evm"
)

func combinedJsonFilename(contractName string) string {
	return fmt.Sprintf("%v/%v.json", combinedJsonDir, contractName)
}

func solCovFilename(contractName string) string {
	return fmt.Sprintf("%v/%v.json", solCovDir, contractName)
}

type target struct {
	Filename     string
	ContractName string
	SolcVersion  string
	OptimizeRuns string
}

var targets = []target{
	target{Filename: "contracts/Basket.sol", ContractName: "Basket", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Manager.sol", ContractName: "Manager", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	//	target{Filename: "contracts/Proposal.sol", ContractName: "Proposal", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Proposal.sol", ContractName: "AdjustQuantities", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Proposal.sol", ContractName: "SetWeights", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/Vault.sol", ContractName: "Vault", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/rsv/Reserve.sol", ContractName: "Reserve", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
	target{Filename: "contracts/rsv/ReserveEternalStorage.sol", ContractName: "ReserveEternalStorage", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
	target{Filename: "contracts/test/BasicOwnable.sol", ContractName: "BasicOwnable", SolcVersion: "0.5.8", OptimizeRuns: "1"},
	target{Filename: "contracts/test/ReserveV2.sol", ContractName: "ReserveV2", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
	target{Filename: "contracts/test/BasicERC20.sol", ContractName: "BasicERC20", SolcVersion: "0.5.8", OptimizeRuns: "1000000"},
}

/* TODO: This would be cleaner as a few separate tools, tied together by Make:
   - compile combined-json files from .sol files. (That's just solc with the right args)
   - produce Go bindings from combined-json files.
   - convert combined-json files into sol-coverage format json files. (That's just JSON rewriting.)
*/

func main() {
	// Compile.
	type compiledOutput struct {
		ABI           string
		Bin           string
		BinRuntime    string `json:"bin-runtime"`
		Srcmap        string
		SrcmapRuntime string `json:"srcmap-runtime"`
	}
	base := os.ExpandEnv(os.Getenv("REPO_DIR"))

	for _, t := range targets {
		var compilationResult struct {
			Contracts  map[string]compiledOutput
			SourceList []string
		}

		// Output combined-json to evm/ directory.
		check(os.MkdirAll(combinedJsonDir, 0755), "creating combined-json evm directory")
		jsonName := combinedJsonFilename(t.ContractName)
		combinedJson, err := os.Create(jsonName)
		if err != nil {
			log.Fatalf("generate.go: %v", err)
		}

		// Run solc, generate combined-json files.
		cmd := exec.Command(
			"solc",
			append(
				[]string{
					"--allow-paths " + base + "/contracts",
					"--optimize",
					"--optimize-runs=" + t.OptimizeRuns,
					"--combined-json=abi,bin,bin-runtime,srcmap,srcmap-runtime,userdoc,devdoc",
				},
				t.Filename,
			)...,
		)
		cmd.Env = append(cmd.Env, "SOLC_VERSION="+t.SolcVersion)
		cmd.Stdin = os.Stdin // solc doesn't need stdin to be set, but trailofbits' solc-select does
		cmd.Stdout = combinedJson
		cmd.Stderr = os.Stderr
		log.Printf("%s: Compiling into combined-json with solc", t.Filename)

		err = cmd.Run()
		if err != nil {
			log.Fatalf("solc failed to build %s: %v\n", jsonName, err)
		}

		_, err = combinedJson.Seek(0, 0)
		if err != nil {
			log.Fatalf("Failed to return to beginning of %v: %v", jsonName, err)
		}

		// Parse the combined-json outputs from solc.
		check(json.NewDecoder(combinedJson).Decode(&compilationResult), "parsing solc output in "+jsonName)
		combinedJson.Close()

		output := compilationResult.Contracts[t.Filename+":"+t.ContractName]

		// Generate bindings.
		check(os.MkdirAll("abi", 0755), "creating abi directory")
		code, err := bind.Bind([]string{t.ContractName}, []string{output.ABI}, []string{output.Bin}, "abi", bind.LangGo)
		check(err, "generating Go bindings")

		// Write to .go file.

		name := outputGoFile(t.ContractName)
		check(ioutil.WriteFile(name, []byte(code), 0644), "writing "+name)
		// Generate event bindings.
		//
		// We generate a String() function for each event and a
		// Parse<ContractName>Log(*types.Log) function for each contract.
		{
			log.Printf("%s: Generating Go bindings", t.Filename)

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
			log.Printf("%s: Reformatting build for sol-coverage", t.Filename)
			// sources records an ordering on source files.
			// sol-compiler uses a different format from solc for this,
			// so here we're converting from the former to the latter.
			check(os.MkdirAll(solCovDir, 0755), "creating sol-coverage evm directory")
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
				ioutil.WriteFile(solCovFilename(t.ContractName), b, 0644),
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
