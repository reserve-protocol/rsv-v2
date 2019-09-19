package tests

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsd/abi"
	"github.com/reserve-protocol/rsd/soltools"
)

type logParser interface {
	ParseLog(*types.Log) (fmt.Stringer, error)
}

// TestSuite holds functionality common between our two test suites.
//
// It knows how to create a connection to an Ethereum node, it holds a list of accounts
// to use with that node, and it implements common test assertions.
type TestSuite struct {
	suite.Suite

	account []account
	signer  *bind.TransactOpts
	node    interface {
		bind.ContractBackend
		TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	}
	reserve               *abi.ReserveDollar
	reserveAddress        common.Address
	eternalStorage        *abi.ReserveDollarEternalStorage
	eternalStorageAddress common.Address

	logParsers map[common.Address]logParser
}

// requireTx requires that a transaction is successfully mined and does
// not revert. It also takes an extra error argument, and checks that the
// error is nil. This signature allows the function to directly wrap our
// abigen'd mutator calls.
//
// requireTx returns a closure that can be used to assert the list of events
// that were emitted during the transaction. This API is a bit weird -- it would
// be more natural to pass the events in to the `requireTx` call itself -- but
// this is the cleanest way that is compatible with directly wrapping the abigen'd
// calls, without using intermediate placeholder variables in calling code.
func (s *TestSuite) requireTx(tx *types.Transaction, err error) func(assertEvent ...fmt.Stringer) {
	receipt := s._requireTxStatus(tx, err, types.ReceiptStatusSuccessful)

	// return a closure that can take a varargs list of events,
	// and assert that the transaction generates those events.
	return func(assertEvent ...fmt.Stringer) {
		if s.Equal(len(assertEvent), len(receipt.Logs), "did not get the expected number of events") {
			for i, wantEvent := range assertEvent {
				parser := s.logParsers[receipt.Logs[i].Address]
				if s.NotNil(parser, "got an event from an unexpected contract address: "+receipt.Logs[i].Address.Hex()) {
					gotEvent, err := parser.ParseLog(receipt.Logs[i])
					if s.NoErrorf(err, "parsing event %v", i) {
						s.Equal(wantEvent.String(), gotEvent.String())
					}
				}
			}
		}
	}
}

// requireTxFails is like requireTx, but it requires that the transaction either
// reverts or is not successfully made in the first place due to gas estimation
// failing.
func (s *TestSuite) requireTxFails(tx *types.Transaction, err error) {
	if err != nil && err.Error() ==
		"failed to estimate gas needed: gas required exceeds allowance or always failing transaction" {
		return
	}

	receipt := s._requireTxStatus(tx, err, types.ReceiptStatusFailed)
	s.Equal(0, len(receipt.Logs), "Zero logs should be generated for a failed transaction")
}

func (s *TestSuite) _requireTxStatus(tx *types.Transaction, err error, status uint64) *types.Receipt {
	s.Require().NoError(err)
	s.Require().NotNil(tx)
	receipt, err := bind.WaitMined(context.Background(), s.node, tx)
	s.Require().NoError(err)
	s.Require().Equal(status, receipt.Status)
	return receipt
}

// assertBalance asserts that the Reserve Dollar balance of `address` is `amount`.
func (s *TestSuite) assertBalance(address common.Address, amount *big.Int) {
	balance, err := s.reserve.BalanceOf(nil, address)
	s.NoError(err)
	s.Equal(amount.String(), balance.String()) // assert.Equal can mis-compare big.Ints, so compare strings instead
}

// assertAllowance asserts that the allowance of Reserve Dollars that `owner` has given `spender` is `amount`.
func (s *TestSuite) assertAllowance(owner, spender common.Address, amount *big.Int) {
	allowance, err := s.reserve.Allowance(nil, owner, spender)
	s.NoError(err)
	s.Equal(amount.String(), allowance.String())
}

// assertTotalSupply asserts that the total supply of Reserve Dollars is `amount`.
func (s *TestSuite) assertTotalSupply(amount *big.Int) {
	totalSupply, err := s.reserve.TotalSupply(nil)
	s.NoError(err)
	s.Equal(amount.String(), totalSupply.String())
}

// createSlowCoverageNode creates a connection to a local geth node that passes through
// sol-coverage instrumentation. This mode is significantly slower than running against
// the in-process node created by `createFastNode`.
//
// This connection is then available as `s.node`.
func (s *TestSuite) createSlowCoverageNode() {
	fmt.Fprintln(os.Stderr, "\nA local geth node must be running for coverage to work.")
	fmt.Fprintln(os.Stderr, "If one is not already running, start one in a new terminal with:")
	fmt.Fprintln(os.Stderr, "\n\tmake run-geth")

	var err error
	s.node, err = soltools.NewBackend("http://localhost:8545")
	s.Require().NoError(err)

	// Throwaway initial transaction.
	// The tests fail if running against a newly-initialized 0xorg/devnet container.
	// I (jeremy) suspect that this is because the node is configured to move through
	// the historical Ethereum hard forks over the course of the first few blocks, rather
	// than all at once in the first block. Meaning the first transactions run against different
	// versions of Etherum than the rest of the transactions:
	//
	//   https://github.com/0xProject/0x-monorepo/blob/e909faa3ef9cea5d9b4044b993251e98afdb0d19/packages/devnet/genesis.json#L4-L9
	//
	// To work around this issue, we try to send a throwaway transaction at the beginning with a
	// Homestead-style signature. This will fail if it is not the first transaction on the chain,
	// but that's ok. If it is the first transaction on the chain, it succeeds and causes the chain
	// to advance by one block, upgrading the Ethereum version and allowing the rest of the tests
	// to pass.
	tx, _ := types.SignTx(
		types.NewTransaction(0, common.Address{100}, bigInt(0), 21000, bigInt(1), nil),
		types.HomesteadSigner{},
		s.account[0].key,
	)
	s.node.SendTransaction(context.Background(), tx)
}

// createFastNode creates a fast in-process Ethereum node. It is then available as `s.node`.
func (s *TestSuite) createFastNode() {
	genesisAlloc := core.GenesisAlloc{}
	for _, account := range s.account {
		genesisAlloc[account.address()] = core.GenesisAccount{
			Balance: big.NewInt(math.MaxInt64),
		}
	}
	s.node = backend{
		backends.NewSimulatedBackend(
			genesisAlloc,
			// Block gas limit. Needs to be more than 7e6, which is about the cost
			// of the ReserveDollarV2 constructor. But we still want it about the
			// same order of magnitude as mainnet.
			//
			// The ReserveDollar constructor is edging close to the mainnet block limit.
			// We'll probably stay under it without any problem. If not, we can split
			// the Eternal Storage contract deployment into a different transaction.
			8e6,
		),
	}
}

// setup sets up the TestSuite. It must be called before using s.account or s.signer.
func (s *TestSuite) setup() {
	// The first few keys from the following well-known mnemonic used by 0x:
	//	concert load couple harbor equip island argue ramp clarify fence smart topic
	keys := []string{
		"f2f48ee19680706196e2e339e5da3491186e0c4c5030670656b0e0164837257d",
		"5d862464fe9303452126c8bc94274b8c5f9874cbd219789b3eb2128075a76f72",
		"df02719c4df8b9b8ac7f551fcb5d9ef48fa27eef7a66453879f4d8fdc6e78fb1",
		"ff12e391b79415e941a94de3bf3a9aee577aed0731e297d5cfa0b8a1e02fa1d0",
		"752dd9cf65e68cfaba7d60225cbdbc1f4729dd5e5507def72815ed0d8abc6249",
		"efb595a0178eb79a8df953f87c5148402a224cdf725e88c0146727c6aceadccd",
	}
	s.account = make([]account, len(keys))
	for i, key := range keys {
		b, err := hex.DecodeString(key)
		s.Require().NoError(err)
		s.account[i].key, err = crypto.ToECDSA(b)
		s.Require().NoError(err)
	}
	s.signer = signer(s.account[0])
}

// backend is a wrapper around *backends.SimulatedBackend.
//
// *backends.SimulatedBackend requires blocks to be mined manually -- they are not automatically
// mined on every transaction. We want them to be automatically mined on every transaction, though,
// so we use this wrapper to do so.
type backend struct {
	*backends.SimulatedBackend
}

// SendTransaction overrides the function by the same name in *backends.SimulatedBackend,
// adding auto-mining for each transaction.
func (b backend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	defer b.Commit()
	return b.SimulatedBackend.SendTransaction(ctx, tx)
}

// AdjustTime overrides the function by the same name in *backends.SimulatedBackend,
// adding auto-committing.
func (b backend) AdjustTime(delta time.Duration) error {
	defer b.Commit()
	return b.SimulatedBackend.AdjustTime(delta)
}

// signer returns a *bind.TransactOpts that uses a's private key to sign transactions.
func signer(a account) *bind.TransactOpts {
	return bind.NewKeyedTransactor(a.key)
}

// account is a utility type to make it easier to convert from a private key to an address.
type account struct {
	key *ecdsa.PrivateKey
}

// address returns the address corresponding to `a.key`.
func (a account) address() common.Address {
	return crypto.PubkeyToAddress(a.key.PublicKey)
}
