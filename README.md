# RSV V2

The RSV v2 is a stable token, implemented as a series of [Ethereum][] smart contracts, designed to maintain a stable price on the open market.


Links:

-   Our announcement [blog post][] describes the system at some length, in relatively friendly terms. This readme will be terser.
-   This system is now deployed on Ethereum! Some key addresses are:
    -   Reserve token: [0x196f4727526eA7FB1e17b2071B3d8eAA38486988](https://etherscan.io/token/0x196f4727526eA7FB1e17b2071B3d8eAA38486988)
    -   Manager: [0x4B481872f31bab47C6780D5488c84D309b1B8Bb6](https://etherscan.io/address/0x4B481872f31bab47C6780D5488c84D309b1B8Bb6)
    -   Vault: [0xAeDCFcdD80573c2a312d15d6Bb9d921a01E4FB0f](https://etherscan.io/address/0xAeDCFcdD80573c2a312d15d6Bb9d921a01E4FB0f)


Stale V1 Addresses:

    -   Reserve token: [0x1C5857e110CD8411054660F60B5De6a6958CfAE2](https://etherscan.io/address/0x1c5857e110cd8411054660f60b5de6a6958cfae2)
    -   Manager: [0x5BA9d812f5533F7Cf2854963f7A9d212f8f28673](https://etherscan.io/address/0x5BA9d812f5533F7Cf2854963f7A9d212f8f28673)

## What does it do?

RSV v2 is a standard [ERC-20][] token with support for upgrades and emergency pausing. Beyond this, the system ensures that it always maintains full backing for the outstanding supply of RSV, in terms of its current basket. Moreover, the system can issue new RSV to a user that provides matching basket assets, allow a user to redeem RSV for basket assets, and process rebalancing proposals that update the assets and weightings of assets in the basket.

RSV v2 is _not_ the version of RSV described in our [whitepaper][]. RSV v2 supports:

-   RSV issuance
-   RSV redemption
-   Vault rebalancing

Key features of Reserve not included in RSV v2 include:

-   Linking RSV to RSR
-   A price feed, which is necessary for RSV to peg to traditional currencies while holding more-volatile assets in its vault
-   Decentralized governance

## How does it fit together?

The center of this system are the smart contracts in `contracts/` and `contracts/rsv`.

-   `Manager.sol`: Handles issuance and redemption of RSV, and vault-rebalancing proposals. `Manager` is the root of this system's automated permissions; it holds the `manager` role on `Vault` and the `minter` role on `Reserve`.
-   `rsv/Reserve.sol`: The actual RSV token.
-   `rsv/ReserveEternalStorage.sol`: The backing store for RSV, implementing the [eternal storage pattern][].
-   `Vault.sol`: The RSV Vault. This contract is very simple; it just allows some manager address make withdrawals. (In the deployed system, that manager is the `Manager` contract.) Having the Vault contract, instead of just letting the `Reserve` or `Manager` contracts store the backing assets, lets us leave the collateral assets at the same address if we upgrade the manager, which is good both for auditing transparency and minimizing transaction overhead.
-   `Basket.sol`: Essentially just the data structure that represents a set of vault assets, and their weighting per RSV. There is always a current basket, and rebalancing proposals make new potential baskets.
-   `Proposal.sol`: Actually contains quite a few contracts:
    -   `Proposal`: The base proposal class. A proposal has a state machine describing its current state in the proposal acceptance-or-rejection process, and must implement a function that yields a basket at completion time.
    -   `WeightProposal`: A proposal that yields a static, proposed basket at completion time.
    -   `SwapProposal`: A proposal to exchange specific quantities of specific tokens, and which will compute its precise basket at completion time.
    -   `ProposalFactory`: A factory for new `SwapProposal`s and `WeightProposal`s. This exists instead of the equivalent `new` statements in `Manager`, because `new` in `Manager` would force `Manager` over the 24-KB contract bytecode limit due to [EIP 170][].

For greater technical detail, see the source code itself -- each of these contracts' interfaces are generally documented in detail there.

[eip 170]: https://eips.ethereum.org/EIPS/eip-170
[whitepaper]: https://reserve.org/whitepaper
[ethereum]: https://www.ethereum.org/
[blog post]: https://medium.com/reserve-currency/reserve-beta-launch-86855468d506
[erc-20]: https://en.wikipedia.org/wiki/ERC-20
[eternal storage pattern]: https://fravoll.github.io/solidity-patterns/eternal_storage.html

# Environment Setup

To build and test these contracts, your development environment will need:

-   Make
-   Go 1.12 or later
-   Either [solc-select][], or a manual installation of [solc][] version 0.5.7
-   [slither][], for basic source analysis.

Specific further makefile targets assume some other tools:

-   The `mythril` target, in order to perform security analyses using symbolic execution, requires a [mythril][] installation. This installation usually takes some patience and fiddling; and it's not critical for working with these contracts.
-   The `run-geth` target launches a local ethereum chain suitable for testing. It requires a working [docker][] installation on your development machine.
-   The `sizes` target assumes that you have [jq][], and a bunch of standard Unix utilities (`sed`, `awk`, `tr`, and `sort`) installed.

[docker]: https://docs.docker.com/v17.09/engine/installation/
[mythril]: https://github.com/ConsenSys/mythril
[solc]: https://solidity.readthedocs.io/en/v0.5.7/installing-solidity.html
[solc-select]: https://github.com/crytic/solc-select
[slither]: https://github.com/crytic/slither
[jq]: https://stedolan.github.io/jq/

# Building and Testing

The whole build-and-test workflow is automated in the makefile. Just running `make` will build everything and run basic tests; the default `make` target is a good default, in-development, build-and-test feedback loop.

-   `make json`: Build just the smart contracts, outputs in `evm/`
-   `make abi`: Build the smart-contract Go bindings, outputs in `abi/`
-   `make test`: Build contract, run normal tests.
-   `make clean`: Clean up built artifacts in this directory.
-   `make fuzz`: Run a short round of fuzz testing. (Tinker with the command this target invokes for larger or different fuzz-test runs.
-   `make sizes`: Output the current sizes of each contract's bytecode, in bytes. (Useful when you're trying out bytecode-size optimizations, which is important for staying under the 24KB bytecode size limit.)
-   `make flat`: Produce flattened Solidity files, as is useful for getting that deployed code verified on [Etherscan][], or playing with it inside [Remix][].
-   `make check`: Do analysis of smart contracts with slither.
-   `make triage-check`: Like `make check`, but runs slither in [triage mode][], which you can use to suppress specific reports in future runs.
-   `make run-geth`: Launch a local Ethereum chain for smart contract tinkering. Tools for that interaction are not included here; we use [poke][] for this.
-   `make -j1 mythril`: Run [mythril][] on these smart contracts. The `-j1` flag is necessary if you have make set up to run in [parallel by default][] (do this!), because mythril does not really support being run in parallel. This is sort of fine, because a single instance of mythril will eat all your cores and still be hungry, but it is something extra to remember when you call it.

[triage mode]: https://github.com/crytic/slither/wiki/Usage#triage-mode
[parallel by default]: https://stackoverflow.com/questions/10567890/parallel-make-set-j8-as-the-default-option
[etherscan]: https://etherscan.io
[remix]: https://remix.ethereum.org
[poke]: https://github.com/reserve-protocol/poke

# Directory Layout

Contents of this repository:

-   `contracts/`: Actual smart contract source; the point of this repo.
-   `tests/`: Set of tests, in Go, exercising our smart contracts.
-   `soltools/`: Contains some test dependencies (that we haven't moved into `tests/`).
-   `design-docs/`: Documentation and scratch notes. Most of this is really drafty notes from our team to our team. It's not really intended to be comprehensible to passersby. but it might be useful for understanding some of the considerations behind the design of these contracts.
-   `go.mod`, `go.sum`: Files for using this directory as a [Go module][].
-   `genABI.go`: A Go script for generating Go bindings for Solidity smart contracts.
-   `scripts/sizes`: The shell script to compute bytecode sizes, run by `make sizes`.
-   `slither.db.json`: The Slither [triage][triage mode] file.
-   `Makefile`: The makefile; automates workflow steps.
-   `README.md`: The file you're reading now.
-   `LICENSE`: The license file. (We're using the [Blue Oak Model License][], and it's quite possible that you should, too!)

[blue oak model license]: https://blueoakcouncil.org/2019/03/06/model.html
[go module]: https://blog.golang.org/using-go-modules

# Brownie

## Getting brownie

`python3 -m pip install --user pipx`
`python3 -m pipx ensurepath`
`pipx install eth-brownie`

## Setting up the required brownie network

`brownie networks add Development short-lived gas_limit=12000000 mnemonic=brownie chainid=17`
