# Deployment Sketch

1. OWNER :: Deploy Vault()
2. OWNER :: Deploy Reserve()
3. OWNER :: ReserveES.acceptNomination()
4. OWNER :: Deploy ProposalFactory()
5. OWNER :: Deploy Basket({ PAX: .333333e18, TUSD: .333334e18, USDC: .333333e6 })
6. OWNER :: construct Manager(Vault, Reserve, ProposalFactory, Basket, DAILY, 0)

7. Update a bunch of auth roles and contract backlinks
   - OWNER :: Vault.changeManager(Manager)
   - OWNER :: Reserve.changeMinter(Manager)
   - OWNER :: Reserve.changePauser(DAILY)
   - OWNER :: Reserve.changeFeeRecipient(DAILY) // or whatever should be the fee recipient, if not DAILY

8. Check that all the deployed state is quite correct

9. Unpause contracts
   - DAILY :: Reserve.unpause()
   - DAILY :: Manager.setEmergency(false)

- Pause the old RSV

# Pre-deployment checklist

- [x] Double-check github issues. Anything outstanding that we really should deal with?
  (Any other todo lists, or places where notes would've gotten stashed?)
- [x] Double-check `make sizes`. Everything below 24K?

- [x] In `git status` output, we are on branch `production`, and nothing is modified or untracked.
- [x] In `git log -1` output, the commit-hash prefix matches current hash prefix on Github

- [x] Double check the state of RSV-alpha: https://etherscan.io/token/0x1dcac83e90775b5f4bc2ffac5a5749e25acc610d?a=0x04a1cd180c1414629ce4512da70d9c71d79771a2
      Ensure that there's only the one address with 10,000 RSV. (That's River@eth0/0)

- [x] DAILY and OWNER hardware keys are at hand.
- [x] DAILY and OWNER start out with at least 0.225 ETH (Owner needs most of this)

- [x] Triple-check PAX, TUSD, and USDC addresses
- [x] Pull up https://ethgasstation.info for continuous monitoring

POKE NODES:
- Set in deployment script: https://mainnet.infura.io/v3/d884cdc2e05b4f0897f6dffd0bdc1821
- Alternate node for checker: https://eth-mainnet.alchemyapi.io/jsonrpc/-vPGIFwUyjlMRF9beTLXiGQUK6Nf3k8z

- [x] Tools on system:
    - [x] freshly-pulled poke. Reinstall with `go install`
    - [x] solc-select (for SOLC_VERSION), or your local `solc --version` yields 0.5.7
    - [x] freshly-pulled rsv-beta, from production. Do `make test`.

# Deployment Checklist

CLI-level instructions are in `deployment-script.sh`. As you go:

- [ ] After each `deploy` command:

  - [x] Vault
      - [x] Slack: Post the address in Slack
      - [x] Env Var: record the address as a variable for further use in that session
      - [x] Alt Node: we can poke that contract, from another machine, through the alt node
      - [x] Etherscan: etherscan says that we've deployed something to that address
      - [ ] Keybase: Save in our keybase deployment record

  - [x] ReserveEternalStorage
      - [x] Slack: Post the address in Slack
      - [x] Env Var: record the address as a variable for further use in that session
      - [x] Alt Node: we can poke that contract, from another machine, through the alt node
      - [x] Etherscan: etherscan says that we've deployed something to that address
      - [x] Keybase: Save in our keybase deployment record

  - [x] Reserve
      - [x] Slack: Post the address in Slack
      - [x] Env Var: record the address as a variable for further use in that session
      - [x] Alt Node: we can poke that contract, from another machine, through the alt node
      - [x] Etherscan: etherscan says that we've deployed something to that address
      - [x] Keybase: Save in our keybase deployment record

  - [x] ProposalFactory
      - [x] Slack: Post the address in Slack
      - [x] Env Var: record the address as a variable for further use in that session
      - [x] Alt Node: we can poke that contract, from another machine, through the alt node
      - [x] Etherscan: etherscan says that we've deployed something to that address
      - [x] Keybase: Save in our keybase deployment record

  - [x] Basket
      - [x] Slack: Post the address in Slack
      - [x] Env Var: record the address as a variable for further use in that session
      - [x] Alt Node: we can poke that contract, from another machine, through the alt node
      - [x] Etherscan: etherscan says that we've deployed something to that address
      - [x] Keybase: Save in our keybase deployment record

  - [x] Manager
      - [x] Slack: Post the address in Slack
      - [x] Env Var: record the address as a variable for further use in that session
      - [x] Alt Node: we can poke that contract, from another machine, through the alt node
      - [x] Etherscan: etherscan says that we've deployed something to that address
      - [x] Keybase: Save in our keybase deployment record


- [x] At "Check the states of things using poke" make `poke` calls to check:
    - [x] All of the state of the contracts matches the "initial target system"
          (Except that Reserve.paused and Manager.emergency are both true)
        - [x] Vault
        - [x] Basket
        - [x] ReserveEternalStorage
        - [x] Manager
        - [x] Reserve
    - [x] Manager.toIssue(1e18) is sensible
    - [x] Manager.toRedeem(1e18) is also sensible


# Post-Deployment checklist

- [x] Get source code verified on etherscan! This might help: https://github.com/BlockCatIO/solidity-flattener
    - [x] Vault
    - [x] Basket
    - [x] ReserveEternalStorage
    - [x] Manager
    - [x] Reserve

- [x] Per etherscan views, all of the state of the contracts matches the "initial target system"
    - [x] Vault
    - [x] Basket
    - [x] ReserveEternalStorage
    - [x] Manager
    - [x] Reserve

- [x] Issue 10,000 RSV to River @ eth0/1

# Initial target system

- External Roles:
  - OWNER
  - DAILY

- Contracts and their critical initial state
  - Vault
    - owner = OWNER
    - manager = Manager  [critical!]

  - Basket
    - tokens: [PAX, TUSD, USDC]
    - weights PAX:  0.333333e36
    - weights TUSD: 0.333333e36
    - weights USDC:   333334e18

  - Manager
      - owner: OWNER
      - vault: Vault
      - rsv: Reserve
      - operator: DAILY
      - proposalFactory: ProposalFactory
      - seigniorage: 0
      - proposals/length: {}
      - emergency: True
      - issuancePaused: False
      - delay: 24 hours
      - basket: Basket

  - Reserve
    - owner: OWNER
    - data: RSV_ES
    - txFee: 0
    - totalSupply: 0
    - maxSupply: MAX_INT
    - paused: True

    - minter: Manager  [critical!]

    - pauser: DAILY
    - feeRecipient: DAILY

  - ReserveEternalStorage
    - reserveAddress: Reserve
    - balance + allowed: as in current token

  - ProposalFactory must be deployed but has no state
