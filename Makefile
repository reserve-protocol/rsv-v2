export REPO_DIR = $(shell pwd)
export SOLC_VERSION = 0.5.7
export OPT_RUNS = 1000

# All the contracts we're building.
root_contracts := Basket Manager SwapProposal WeightProposal Vault 
rsv_contracts := Reserve ReserveEternalStorage
test_contracts := BasicOwnable ReserveV2 BasicERC20
contracts := $(root_contracts) $(rsv_contracts) $(test_contracts)

sol := $(shell find contracts -name '*.sol')
json := $(foreach contract,$(contracts),evm/$(contract).json)
abi := $(foreach contract,$(contracts),abi/$(contract).go)

.PHONY: clean contracts test fmt run-geth

clean:
	rm -rf abi evm sol-coverage-evm

# previously:
# contracts: generate.go contracts/*.sol
# 	go run generate.go

# debug:
# 	echo $(json)
# 	echo $(abi)
json: $(json)
abi: $(abi)

test: 
	go test ./tests

fmt:
	npx solium -d contracts/ --fix
	npx solium -d tests/echidna/ --fix

run-geth:
	docker run -it --rm -p 8545:8501 0xorg/devnet


abi/%.go: evm/%.json genABI.go
	go run genABI.go $*

# solc recipe. 
define run_solc
@mkdir -p evm
solc --allow-paths $(REPO_DIR)/contracts --optimize --optimize-runs $(OPT_RUNS) \
     --combined-json=abi,bin,bin-runtime,srcmap,srcmap-runtime,userdoc,devdoc \
     $< > $@
endef

evm/Basket.json : contracts/Basket.sol $(sol)
	$(run_solc)

evm/Manager.json: contracts/Manager.sol $(sol)
	$(run_solc)

evm/SwapProposal.json: contracts/Proposal.sol $(sol)
	$(run_solc)

evm/WeightProposal.json: contracts/Proposal.sol $(sol)
	$(run_solc)

evm/Vault.json: contracts/Vault.sol $(sol)
	$(run_solc)

evm/Reserve.json: contracts/rsv/Reserve.sol $(sol)
	$(run_solc)

evm/ReserveEternalStorage.json: contracts/rsv/ReserveEternalStorage.sol $(sol)
	$(run_solc)

evm/BasicOwnable.json: contracts/test/BasicOwnable.sol $(sol)
	$(run_solc)

evm/ReserveV2.json: contracts/test/ReserveV2.sol $(sol)
	$(run_solc)

evm/BasicERC20.json: contracts/test/BasicERC20.sol $(sol)
	$(run_solc)
