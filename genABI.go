package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func outputGoFile(contractName string) string {
	return fmt.Sprintf("abi/%v.go", contractName)
}

const combinedJsonDir = "evm"

func combinedJsonFilename(contractName string) string {
	return fmt.Sprintf("%v/%v.json", combinedJsonDir, contractName)
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("genABI: requires at least one argument, got \"%v\"", os.Args[1:])
	}

	for _, contractName := range os.Args[1:] {
		type compiledOutput struct {
			ABI           string
			Bin           string
			BinRuntime    string `json:"bin-runtime"`
			Srcmap        string
			SrcmapRuntime string `json:"srcmap-runtime"`
		}
		jsonName := combinedJsonFilename(contractName)
		combinedJson, err := os.Open(jsonName)
		check(err, "Opening combined json file")

		// Parse the combined-json outputs from solc.
		var compilationResult struct {
			Contracts  map[string]compiledOutput
			SourceList []string
		}

		check(
			json.NewDecoder(combinedJson).Decode(&compilationResult),
			"parsing solc output in "+jsonName)
		combinedJson.Close()

		// The compilationResult keys have the format <.sol filename>.sol:<contract name>
		// We're treating all the contracts as if they're in one namespace, so we're just
		// referring to them by their contract name (and we only know the contract name)
		// Given that, find the contract key from contractName.
		contractKey := ""
		for k := range compilationResult.Contracts {
			index := strings.LastIndex(k, ":")
			tail := k[index+1:]
			if tail == contractName {
				if contractKey != "" {
					log.Fatal("multiple $v instances in evm/%v.json", contractName, contractName)
				}
				contractKey = k
			}
		}
		if contractKey == "" {
			log.Fatal("no $v instances in evm/%s.json.", contractName, contractName)
		}
		output := compilationResult.Contracts[contractKey]

		// Generate bindings.
		check(os.MkdirAll("abi", 0755), "creating abi directory")
		code, err := bind.Bind(
			[]string{contractName},
			[]string{output.ABI},
			[]string{output.Bin},
			"abi",
			bind.LangGo)
		check(err, "generating Go bindings")

		// Write to .go file.
		name := outputGoFile(contractName)
		check(ioutil.WriteFile(name, []byte(code), 0644), "writing "+name)

		// Generate event bindings.
		//
		// We generate a String() function for each event and a
		// Parse<ContractName>Log(*types.Log) function for each contract.
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
			"Contract": contractName,
			"Events":   parsedABI.Events,
		}), "generating event bindings")
		eventCode, err := format.Source(buf.Bytes())
		check(err, "running gofmt")
		check(
			ioutil.WriteFile(
				filepath.Join("abi", contractName+"Events.go"),
				eventCode,
				0644,
			),
			"writing event bindings to disk",
		)
	}
}

func check(err error, msg string) {
	if err != nil {
		log.Fatal(msg, ": ", err)
	}
}
