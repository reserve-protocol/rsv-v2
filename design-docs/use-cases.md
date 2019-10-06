In general, all of this documentation should be viewed as a place to explicate and record thoughts-in-progress. Even if it looks done, it is probably incomplete.

We're in the ontology of Ethereum, so, for our purposes, agents are identified with their addresses, and the ability to sign a transaction is assumed to be adequate authentication. These are _not_ "users" in the usual sense of "User Stories" or "Use Cases"; that's a layer of abstraction farther out. These agents may be humans, but they might also be companies, automated systems, pieces of some other interface, etc, etc. (Why not do user stories? They honestly might be pretty useful to do -- but this is infrastructure, rather than a real product, and so it takes substantial work to turn use cases into these functionality paths.)

# System Agents
Agents named in this document. Kinds of authenticated users, and their degrees of authorization.

## User
A **user**, without further modification, is assumed to be a normal user of the token, with no other special privileges, and who hasn't been explicitly frozen.

## Maker
A **maker** is a user with the further authority to issue and redeem RSV tokens.

"Maker" here is short for "Market Maker." Imagine an agent of a OTC desk, a staffer for an exchange, an operator for Reserve, or an automated system doing market making on our behalf.

## Owner
The system administrator. Probably these credentials are held by someone on our engineering team (in the short run), or maybe a DAO (in the very long run).

# Use Cases and General System Requirements
When in doubt, state each of these as: "A {user} can {act} in order to {satisfy a purpose}", though "A {user} can {act}" is often good enough for a stub.

## Basics
1. A user can transfer their RSV tokens to any other account, in order to pay for goods and services.
    - In general, RSV serves the purposes of an ERC-20 token. We ... don't need to belabor this further.

2. A user can buy or sell RSV on the open market at 1 USD, in order to:
    - stably store value
    - exchange goods or services
    - transmit value across jurisdictions

## Auth
1. The owner can give any user the privileges of a maker.

2. The owner can transfer ownership to a different user.
    - But not without a signed messagae from that user! (Avoid losing auth to typos or miscommunication.)

3. The owner can disown ownership.
    - But not easily or accidentally! This function should not be simple to invoke, or invokable without delay by a "fast" key.

## Economics of Issuance and Redemption
1. The manager maintains which tokens are in the basket, and the basket weight for each.
    - These values can only be changed with permission from the owner.

2. At all times, for every basket asset, (the vault's balance of that asset) >= (RSV supply) times (that asset's basket weight).

3. A maker can redeem RSV for the basket-equivalent tokens, in order to:
    - earn arbitrage if the RSV market price is low, and/or
    - increase the RSV market price (by buying up RSV from the market) without substantial loss

4. A maker can be issued RSV for the basket-equivalent tokens, in order to:
    - earn arbitrage if the RSV market price is high, and/or
    - increase the RSV supply in periods of high demand

5. When a maker redeems an amount `amt` of RSV:
    - The manager burns that RSV from the maker.
    - For each basket asset, the maker receives `amt` times the asset's basket weight.

6. For any amount `amt` of RSV to be issued, if a maker has enough of each basket asset, the maker can set the manager's allowance on that asset to at least `amt` times that asset's basket weight, and then call `Manager.issue(amt)`. Then:
    - The manager transfers `amt * weight` of each basket asset, from the maker, to the vault.
    - The manager mints and issues `amt` RSV to the maker.

## Changing the Basket

1. The owner can modify the basket weights, in order to respond to economic changes.

2. The owner can remove a basket asset, in order to vacate a fiatcoin expected to fail.
    - This requires that we can remove the basket asset pretty quickly, and possibly at a loss.

3. The owner can update a basket asset's address, in order to respond to a token upgrade.

4. The owner can add a basket asset, better to diversify the basket.

5. A user may propose any of the above, and the owner can enact that proposal, using only the user's capital for asset rebalancing, in order that the owner can incentivize users to trustlessly provide that capital.

## Deployment

Upon deployment:

- The deploying account is Owner.
- RSV and the Manager are paused.
- RSV balance is zero.

## Pausing and Upgrading

1. The owner can pause and unpause any state-changing activities of all users and makers, in order to stop all economic behavior long enough to understand and fix bugs in emergencies.
    - No one but owner can effect pausing and unpausing.

2. The owner can upgrade the Manager contract while the system is paused, in order to fix bugs.
    - The owner can do this without changing the Vault address.

3. The owner can change how fees are computed on RSV transactions, issuance, and redemption.

## ... ?

[TODO: Probably need more entries here to finish a first pass.]

## System Requirements

- Follow https://consensys.github.io/smart-contract-best-practices/recommendations/
