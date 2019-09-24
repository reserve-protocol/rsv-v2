title: The Testing Standard for the Basket Protocol

tl;dr: Testing is complete when each property in our contract specs is well-tested. Specification is complete when the specs contain all properties our contracts must satisfy in order to serve their use cases, satisfy their needed economic properties, and be secure against attack.

# Necessary Documents

To state the testing standard, we need to talk about a bunch of other pieces:

- We've made explicit all important use cases, economic properties, and critical security goals. How do we know?
    - The use cases are straightforward; they're part of the design.
    - In this case, the economic properties are also relatively straightforward; the basket demands only simple invariants.
    - Critical security goals are harder; here, we're done when we've captured our security measures against all known high- and medium-importance risks, and we've hit diminishing marginal returns in expressing threats, risks, and attacks.

- The functional spec, if satisfied, describes a system that covers all important use cases.
- The security policy, if satisfied, describes a system that prevents all known high-and medium- importance business risks.
- The functional spec and security policy, if satisfied, describes a system that has the required economic properties.

# What to Test

Our tests reach the Testing Standard when:

- Each function call in the system is well-tested.
- Each property in the functional spec and security policy is individually well-tested.
- Every contract invariant that we can derive from the functional spec and security policy is manifested as test-time checks, which can be reached by randomized testing -- _and_ we've subjected the system to substantial randomized testing (e.g., with Echidna).

# "Well-Tested" Properties

This is not quite a _rigorous_ definition of The Standard, as there's a lot of necessary judgment entailed in whether a particular component or property is well-tested. An unrigorous gesture at "well-tested" is this: if the code were changed in any way to introduce a new bug, we expect that either the new code should fail a test that already exists, or fail a quick code review as being obviously too complicated or incorrect.

Another way to state this is perhaps more evocative. Imagine we're playing a programming game.

- You get to write the tests.
- Then, I, your opponent, get to look at your tests and the code and modify it in some subtly incorrect way.
- Then, someone else on the team, believing my code to be normal, well-meaning code, will do a normal code review.

I win, and the testing team loses, if I can find some way to introduce a bug that both passes all the tests and a code review.
