tests
---
Tests for the Reserve Dollar smart contracts.

Unit tests for the central smart contracts are in `reserve_test.go` and `mint_and_burn_admin_test.go`.  These are built on `base.go`, and are run from `make test` in the repository root.

`make fuzz` from the repo root will attempt fuzz testing with echidna as set up in the `echidna/` directory. The tests there should currently pass, but don't yet have the reach we've intended.

`make coverage` from the repo root can prdouce coverage output in the `coverage/` directory, here.

