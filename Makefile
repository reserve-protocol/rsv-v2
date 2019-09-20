export REPO_DIR = $(shell pwd)

abi: abi/generate.go
	go run abi/generate.go

test: 
	go test ./tests

fmt:
	npx solium -d contracts/ --fix
	npx solium -d tests/echidna/ --fix

run-geth:
	docker run -it --rm -p 8545:8501 0xorg/devnet
