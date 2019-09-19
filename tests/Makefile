DOCKER?=docker
DOCKER_RUN_ECHIDNA=$(DOCKER) run --rm -t -v `pwd`/echidna:/src -w /src reserveprotocol/echidna echidna-test

fuzz: fuzz-total-supply \
	fuzz-ownership \
	fuzz-constant-owner-under-free-money \
	fuzz-freezing


fuzz-total-supply: setup
	$(DOCKER_RUN_ECHIDNA) EchidnaTransferTests.sol EchidnaTotalSupplyTests --config multisender.yaml

fuzz-ownership: setup
	$(DOCKER_RUN_ECHIDNA) EchidnaOwnershipTests.sol EchidnaConstantOwnerSimple --config multisender_with_owner.yaml


fuzz-constant-owner-under-free-money: setup 
	$(DOCKER_RUN_ECHIDNA) EchidnaOwnershipTests.sol EchidnaConstantOwnerFreeMoney --config multisender.yaml

fuzz-freezing: setup
	$(DOCKER_RUN_ECHIDNA) EchidnaFreezeTests.sol EchidnaFrozenAccount --config multisender.yaml

setup: echidna/ReserveDollar.sol \
       echidna/ReserveDollarEternalStorage.sol \
       echidna/zeppelin/SafeMath.sol

echidna/ReserveDollar.sol: ../contracts/ReserveDollar.sol
	rm $@
#echo "\033[31;1m[DANGER] Inserting internal minting function into ReserveDollar contract.\n    This is the naive approach to give users a balance in the constructor of a\nfuzz test contract. The contract being fuzzed is no longer exactly the\nReserveDollar contract!\033[0m"
#sed -e 's|\(function mint(address account, uint256 value) external\)|function testFixtureMint(address account, uint256 value) internal notPaused only(minter) {\n        require(account != address(0), "cant mint to address zero");\n\n        _totalSupply = _totalSupply.add(value);\n        data.addBalance(account, value);\n        emit Transfer(address(0), account, value); \n    }\n\n    \1|' $< > $@
# Use hard links so that changes to files copied into the tests/ dir will show
# up in version control. This reduces odds of "worked on my machine". We use
# hard links because solc refuses to accept soft links
	ln $< $@

echidna/ReserveDollarEternalStorage.sol: ../contracts/ReserveDollarEternalStorage.sol
	rm $<
# echo "\033[31;1m[WARNING] Removing onlyOwner modifier from eternal storage.\nechidna\033[0m"
# sed -e 's|require(msg.sender == owner, "onlyOwner");||' $< > $@
	ln $< $@

echidna/zeppelin/SafeMath.sol: ../contracts/zeppelin/SafeMath.sol
	mkdir -p echidna/zeppelin
	rm $@ || true
# Use hard links so that changes to files copied into the tests/ dir will show
# up in version control. This reduces odds of "worked on my machine". We use
# hard links because solc refuses to accept soft links
	ln $< $@

clean:
	rm echidna/ReserveDollar.sol \
       echidna/ReserveDollarEternalStorage.sol \
       echidna/zeppelin/SafeMath.sol

.PHONY: fuzz fuzz-total-supply fuzz-ownership fuzz-constant-owner-under-free-money \
	fuzz-freezing
