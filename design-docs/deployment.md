Deployment
==========

Initial target system:

- External Roles:
  - MASTER (probably Malcolm)
  - DAILY  (probably Kaylee)

- Contracts and their critical initial state
  - Vault
    - owner = DAILY
    - manager = Manager
    
  - Basket
    - tokens: [PAX, TUSD, USDC]
    - backing: [.333e18, .333e18, .333e18]
    - frontTokenDecimals: 18
    
  - Manager
    - owner: MASTER
    - operator: DAILY
    - basket: Basket
    - vault: Vault
    - rsv: Reserve
    - proposals: {}
    - whitelist: {MASTER: True} [TODO: MASTER, or DAILY?]
    - useWhitelist: True
    - paused: True
    - seigniorage: ???
    
  - Reserve
    - owner: MASTER
    - minter: Manager
    - pauser: DAILY
    - freezer: DAILY
    - feeRecipient: (DAILY | any money-management account)
    - data: Fresh ReserveEternalStorage address
    - txFee: 0
    - name, symbol: "Reserve", "RSV"
    - totalSupply: 0
    - maxSupply: MAX_INT
    - paused: True
    
  - ReserveEternalStorage
    - Shouldn't this continue to be the already-existing eternal storage contract? Where is that guy?

What happens for deployment?

- DAILY :: construct Vault
   - vault.owner: DAILY
   - vault.manager: DAILY

- DAILY :: construct Reserve
   - data: fresh RES.
   - txFee: 0
   - pauser: DAILY
   - feeRecipient: DAILY
   - owner: DAILY
   - maxSupply: MAX_INT
   - totalSupply: 0

- DAILY :: Reserve.pause() (TODO: set in constructor!)
- DAILY :: Reserve.changeName("Reserve", "RSV") (TODO: also set in constructor!)
   
- MASTER :: construct Manager(Vault, Reserve, ???)
   - seigniorage: ???
   - vault: Vault
   - rsv: Reserve
   - basket: 0
   - proposals, whitelist: {}, {MASTER: True}
   - paused: True
   - useWhitelist: True

- MASTER :: setOperator(DAILY) (Also set in constructor?)

- DAILY :: Reserve.changeMinter(Manager)
- DAILY :: Reserve.changeOwner(MASTER)
- DAILY :: Vault.changeManager(Manager)

- MASTER :: Manager.proposeNewBasket([PAX, TUSD, USDC], [1/3, 1/3, 1/3]*1e18)
- DAILY :: Manager.acceptProposal(P) 
    - where P = the proposal number for the above proposal. Maybe not 0, if someone else slipped another proposal in there. Actually check!
- Wait 24+ hours
- ANY :: Manager.executeProposal(P)

- DAILY :: Reserve.unpause()
- DAILY :: Manager.unpause()
