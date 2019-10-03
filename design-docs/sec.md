title: Security Analysis and Policy

In general, all of this documentation should be viewed as a place to explicate and record thoughts-in-progress. Even if it looks done, it is probably incomplete. (In particular, each of these categories is deliberately set up so that when we learn more things or do more analysis, we can just add more to these items.)

Many scratch notes in our shared Keybase. I started feeling a twinge of discomfort at putting them here; moving them away is easier than dealing with the discomfort.

# Security Policy
What properties must this system have in order to be satisfactorily secure? Expect that these are each necessary elements of the plan, but possibly incomplete.

This security policy is actually complete when we're convinced by an _explicit_, careful argument that we have a complete account of all relevant risks (in the [sec](sec.md) file), and that we have adequate measures to respond or prevent each risk. (To be honest, that's probably an unrealistically high bar, but it's the principled, aspirational goal.) Like software, in practice, a security policy is never finished.


- Always, if the manager is unpaused, the total weight in the basket is > 0.

- The owner of any contract can only be changed by owner.
- The owner of any contract can only be changed to an address that has sent an explicit "acceptOwnership" message.
- The owner of any contract cannot be changed with only short-lived, temporary access to the set of keys needed to make other changes. (i.e., "brief, customary" access)
  (Roughly: if we did some normal operations and left the keys on a desk for 20 minutes, a sneaky attacker shouldn't be able to change who the owner is. This could be ensured by a slow wallet, a multi-sig wallet, or some other scheme.)

- The owner cannot accept an arbitrary basket proposal with only brief, customary key access.
- The owner can accept a previously-approved alternative basket proposal with only brief, customary key access.
    - For each asset in the current basket, there is almost always an available, alternative basket proposal that does not contain that asset. ("Almost always" because this can't really be ensured through each basket transition; it's fine, though, for that change to be unavailable for a day or less after each basket transition.)


- We never have access to third-party private keys.
