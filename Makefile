export REPO_DIR = $(shell pwd)

abi/bindings: contracts/*.sol abi/generate.go compiler.json
	npx sol-compiler
	go run abi/generate.go
	@echo "placeholder output file for 'make abi/bindings'" > abi/bindings

test: abi/bindings
	go test ./tests

fmt:
	npx solium -d contracts/ --fix
	npx solium -d tests/echidna/ --fix

run-geth:
	docker run -it --rm -p 8545:8501 0xorg/devnet
