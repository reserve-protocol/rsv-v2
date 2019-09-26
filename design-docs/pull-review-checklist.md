# Code Review Checklist

Assume that the author has made mistakes that you need to catch.

Save our users from those bugs.

Check an item when you, the reviewer, can assert that item is true in the pull request -- or when you've left a comment everywhere you think it isn't.


## Think For Yourself
 - [ ] I've thought about how I'd solve this problem, and I don't think it's materially better than the PR.

## Purpose
 - [ ] I know the overall purpose of this PR.
 - [ ] I know the relevant canonical specification of this purpose, if any.
 - [ ] All new code serves that purpose.
 - [ ] That purpose is fully served by this code.

## Style
 - [ ] There are no typos.
 - [ ] There is no commented-out code.
 - [ ] This code follows existing patterns and conventions.
 - [ ] Names are everywhere pretty good, consistent, and clear.
 
## Specs and Docs
 - [ ] All TODOs in this code refer to relevant entries in our ticket system.
 - [ ] Comments, design docs, and user docs are all up-to-date with this change.
 - [ ] Comments, developer docs, and commit messages are reasonably concise and clear enough for a future developer to understand.

## Tests
 - [ ] The tests cover any important ways this code could break (including integration or system tests, if necessary).
 - [ ] There is adequate test coverage for changed lines and critical code paths.
 - [ ] Each test is reasonably complete, asserts valuable behavior, and non-trivially passes
 - [ ] Where possible, tests do not test implementation details.

## Code Quality
 - [ ] Program flow is everywhere consistent and clear.
 - [ ] No shortcuts, missing cases, or unusual assumptions can be broken by some possible input.
 - [ ] Any reimplementation of functionality has a good reason.
 - [ ] No duplication of code could be cleanly unified by abstraction.
 - [ ] Our frameworks and libraries are used idiomatically.
 - [ ] This code introduces no substantial risks of breaking code.
 - [ ] Any added compile-time or run-time dependencies have an extremely good reason.
 
