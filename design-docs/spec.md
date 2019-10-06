# Functional Spec and Requirements

## Contract Invariants and "Always"

Any property listed here that must be true "always", "at all times", or "whenever" is a _contract invariant_, and must be true:
- After its constructor is called.
- After every contract method call.
- At any point where a contract may call a contract method outside this system, including any time it transfers ETH to an external account.

(i.e., the property can be transiently violated during a transaction, nso long as control cannot pass to any other contract while that property is false, and the property is true again before the end of the transaction.)

## Correct Deployment
This spec is over the entire system of contracts. Except where otherwise stated, it presumes a correct deployment, satisfying the following conditions:

- Manager, Vault, and RSV are all deployed Ethereum smart contracts.
- Vault.manager and RSV.minter is The Manager.
- Manager.owner, Vault.owner, and RSV.owner are all one `owner` account.

## Emergency Stops (Pausing)

RSV and Manager are both owner-pausable. That is, for each of these contracts:

- The contract has a `paused` boolean state.
- Only the owner can modify the `paused` state.
- When the contract is paused, the owner can always issue a txn to pause the contract.
- When the contract is unpaused, the owner can always issue a txn to unpause the contract.
- While the contract is paused, no one except `owner` can change its state in any way.
    - This is actually mechanically checkable: we require that every function in RSV or Manager to have at least one of the following modifiers: `notPaused`, `onlyOwner`, `view`, or `pure`.

## Semantic State

(By "semantic state", I'm laying out this functional spec's model of each contract's state. Each contract may implement this model with details unspecified here.)

Other than auth state, contracts have the following semantic state:

- RSV.balanceOf: mapping(address => uint256)
- RSV.allowance: mapping(address => mapping(address => uint256))
- Manager.makers: set(address)
- Manager.weights: mapping(address => rational)

The semantic state of the Vault is its balances of ERC-20 tokens, which are not directly represented by storage of the Vault itself.

## Invariants of Semantic State
- RSV is always fully backed by assets held by the Vault.
    - That is, at all times, for every pair (`addr -> weight`) in the mapping `Manager.weights`, `addr.balanceOf(Vault) >= RSV.totalSupply * weight`

- There are never more than 10 entries in `Manager.weights`.
    (Why 10? This has to be some constant in order to stay under the gas block limit. 20 might be fine; it really depends on the current block limit, gas fees, and how hard it is to get a transaction into a block.)

- For every pair `addr -> weight` in `Manager.weights`:
    - `weight > 0`
    - `addr` is a contract, and satisfies the ERC-20 interface for `balanceOf`, `transfer`, `allowance`, and `transferFrom`.

- The sum of weights in `Manager.weights` is 1.
    - Note that this cannot easily be required in the contracts themselves, as we want the contracts to generalize to cases where prices are unknown. However, we can build a `check-proposal` off-chain script to check the sum-of-weights, and ensure that our proposal-acceptance playbook entry requires it to pass.

## Auth

- Only the owner can modify `Manager.weights`.
- Only the owner can modify `Manager.makers`.
- The owner can arbitrarily add and remove addresses in `Manager.makers`.

## Economic Properties
- RSV is always fully backed by assets held by the Vault.
    - That is, at all times, for every pair (`addr -> weight`) in the mapping `Manager.weights`, `addr.balanceOf(Vault) >= RSV.totalSupply * weight`


- If the owner has large enough balances in the named assets, the owner can always (through some series of transactions) unilaterally modify `Manager.weights` to any other mapping, 1-10 entries,

- Issuance works. That is, when `maker` calls `Manager.issue(amount)`:
    - Preconditions:
        - `Manager` is unpaused.
        - `maker` is in `Manager.makers`.
        - For every pair (`addr -> weight`) in `Manager.weights`, `addr.allowed[maker] >= amount * weight`
    - Changes:
        - For every pair (`addr -> weight`) in `Manager.weights`:
            - `addr.allowed[maker][Vault]` decreases by `amount * weight`
            - `addr.balanceOf[maker]` decreases by `amount * weight`
            - `addr.balanceOf[Vault]` increases by `amount * weight`
        - `RSV.balanceOf[maker]` increases by `amount`
        - `RSV.totalBalance` increases by `amount`

- Redemption works. That is, when `maker` calls `Manager.redeem(amount)`:
    - Preconditions:
        - `Manager` is unpaused.
        - `maker` is in `Manager.makers`
        - `RSV.allowed[maker] >= amount`
    - Changes:
        - For every pair (`addr -> weight`) in `Manager.weights`:
            - `addr.balanceOf[maker]` increases by `amount * weight`
            - `addr.balanceOf[Vault]` decreases by `amount * weight`
        - `RSV.balanceOf[maker]` decreases by `amount`
        - `RSV.totalBalance` decreases by `amount`

## RSV
RSV functions strictly per ERC-20:

### ERC-20 Interface

    # view functions.
    function name() public view returns (string)
    function symbol() public view returns (string)
    function decimals() public view returns (uint8)
    function totalSupply() public view returns (uint256)
    function balanceOf(address owner) public view returns (uint256 balance)

    # transfer transfers `value` tokens from msg.sender to the address `to`.
    # Must emit a Transfer event, return true on success, and revert on failure.
    # Must allow zero-value transfers.
    function transfer(address to, uint256 value) public returns (bool success)

    # Returns the amount which `_spender` is still allowed to withdraw from `_owner`.
    function allowance(address owner, address spender) public view returns (uint256 remaining)

    # approve: allowance[msg.sender][spender] = value
    function approve(address spender, uint256 value) public returns (bool success)

    # transferFrom: If allowance[msg.sender][from] >= value, transfer `value` from address `from` to
    # `to`, reduce allowance[msg.sender][from] by `value`, and emit a Transfer event. If not, revert.
    function transferFrom(address from, address to, uint256 value) public returns (bool success)

    # Must trigger whenever tokens are transferred, including zero-value transfers.
    event Transfer(address indexed from, address indexed to, uint256 value)

    # Must trigger on any successful call to approve()
    event Approval(address indexed owner, address indexed spender, uint256 value)
