soltools
---

Go wrappers for 0x's suite of solidity tools.

We use [sol-coverage](https://sol-coverage.com/) to get coverage reports for our Solidity contracts. sol-coverage is written in JavaScript, and our tests are written in Go, so we need a way to bridge between the two languages. This package provides that bridge.

The bridge works by running the relevant 0x libraries in a node.js process, and communicating with the process using HTTP requests over localhost.
