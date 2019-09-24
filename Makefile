export REPO_DIR = $(shell pwd)

clean:
	rm -rf abi evm sol-coverage-evm

contracts: generate.go contracts/*.sol
	go run generate.go

test: 
	go test ./tests

fmt:
	npx solium -d contracts/ --fix
	npx solium -d tests/echidna/ --fix

run-geth:
	docker run -it --rm -p 8545:8501 0xorg/devnet
