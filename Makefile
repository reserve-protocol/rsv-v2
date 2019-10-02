export REPO_DIR = $(shell pwd)
export SOLC_VERSION = 0.5.7

root_contracts := Basket Manager SwapProposal WeightProposal Vault ProposalFactory ManagerTest
rsv_contracts := Reserve ReserveEternalStorage
test_contracts := BasicOwnable ReserveV2 BasicERC20
contracts := $(root_contracts) $(rsv_contracts) $(test_contracts) ## All contract names

sol := $(shell find contracts -name '*.sol' -not -name '.*' ) ## All Solidity files
json := $(foreach contract,$(contracts),evm/$(contract).json) ## All JSON files
abi := $(foreach contract,$(contracts),abi/$(contract).go) ## All ABI files

all: test json abi

abi: $(abi)
json: $(json)

test: abi
	go test ./tests

clean:
	rm -rf abi evm sol-coverage-evm

sizes: json
	scripts/sizes $(json)

fmt:
	npx solium -d contracts/ --fix
	npx solium -d tests/echidna/ --fix

run-geth:
	docker run -it --rm -p 8545:8501 0xorg/devnet

# Pattern rule: generate ABI files
abi/%.go: evm/%.json genABI.go
	go run genABI.go $*

# solc recipe template for building all the JSON outputs.
# To use as a build recipe, optimized for (e.g.) 1000 runs,
# use "$(call solc,1000)" in your recipe.
define solc
@mkdir -p evm
solc --allow-paths $(REPO_DIR)/contracts --optimize --optimize-runs $1 \
     --combined-json=abi,bin,bin-runtime,srcmap,srcmap-runtime,userdoc,devdoc \
     $< > $@
endef

evm/ManagerTest.json : contracts/Echidna.sol $(sol)
	$(call solc,1)

evm/Basket.json : contracts/Basket.sol $(sol)
	$(call solc,1)

evm/Manager.json: contracts/Manager.sol $(sol)
	$(call solc,1)

evm/ProposalFactory.json: contracts/Proposal.sol $(sol)
	$(call solc,1)

evm/SwapProposal.json: contracts/Proposal.sol $(sol)
	$(call solc,1)

evm/WeightProposal.json: contracts/Proposal.sol $(sol)
	$(call solc,1)

evm/Vault.json: contracts/Vault.sol $(sol)
	$(call solc,1)

evm/Reserve.json: contracts/rsv/Reserve.sol $(sol)
	$(call solc,1000000)

evm/ReserveEternalStorage.json: contracts/rsv/ReserveEternalStorage.sol $(sol)
	$(call solc,1000000)

evm/BasicOwnable.json: contracts/test/BasicOwnable.sol $(sol)
	$(call solc,1)

evm/ReserveV2.json: contracts/test/ReserveV2.sol $(sol)
	$(call solc,1000000)

evm/BasicERC20.json: contracts/test/BasicERC20.sol $(sol)
	$(call solc,1000000)


# Mark "action" targets PHONY, to save occasional headaches.
.PHONY: all clean json abi test fmt run-geth sizes
