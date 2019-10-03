package tests

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"reflect"
	"strings"
	"time"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
	"github.com/reserve-protocol/rsv-beta/soltools"
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
	owner                 account
	reserve               *abi.Reserve
	reserveAddress        common.Address
	eternalStorage        *abi.ReserveEternalStorage
	eternalStorageAddress common.Address
	manager               *abi.Manager
	managerAddress        common.Address
	vault                 *abi.Vault
	vaultAddress          common.Address
	basket                *abi.Basket
	basketAddress         common.Address
	erc20s                []*abi.BasicERC20
	erc20Addresses        []common.Address

	utilContract *bind.BoundContract

	logParsers map[common.Address]logParser

	operator account
	proposer account
	weights  []*big.Int
}

var coverageEnabled = os.Getenv("COVERAGE_ENABLED") != ""

// requireTxWithStrictEvents(tx, err)(events...) requires that a transaction is successfully mined,
// does not revert, and that err is nil. The result of requireTxWithStrictEvents takes a
// variable-length list error arguments, and requires that exactly that set of events was thrown
// while processing tx. This signature allows the function to directly wrap our abigen'd mutator
// calls.
//
// requireTxWithStrictEvents(tx, err) returns a closure that can be used to assert the list of
// events that were emitted during the transaction. This API is a bit weird -- it would be more
// natural to pass the events in to the requireTxWithStrictEvents call itself -- but this is the
// cleanest way that is compatible with directly wrapping the abigen'd calls, without using
// intermediate placeholder variables in calling code.
//
// Note: This closure asserts exactly the set of expected events and no more. It is strict.
func (s *TestSuite) requireTxWithStrictEvents(tx *types.Transaction, err error) func(assertEvent ...fmt.Stringer) {
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

// requireTx(tx, err)(events...) requires that a transaction is successfully mined, does not
// revert, and that err is nil. The result of requireTx takes a variable-length
// list error arguments, and requires that exactly that set of events was thrown while processing
// tx. This signature allows the function to directly wrap our abigen'd mutator calls.
//
// requireTx(tx, err) returns a closure that can be used to assert the list of events that were
// emitted during the transaction. This API is a bit weird -- it would be more natural to pass the
// events in to the requireTx call itself -- but this is the cleanest way that is compatible with
// directly wrapping the abigen'd calls, without using intermediate placeholder variables in
// calling code.
//
// Note: This closure asserts that each expected event was emitted, but not that all emitted events
// match the given list. It is less strict than requir
func (s *TestSuite) requireTx(tx *types.Transaction, err error) func(assertEvent ...fmt.Stringer) {
	receipt := s._requireTxStatus(tx, err, types.ReceiptStatusSuccessful)

	// return a closure that can take a varargs list of events,
	// and assert that the transaction generates at least that set of events.
	return func(assertEvent ...fmt.Stringer) {
		for _, wantEvent := range assertEvent {
			found := false
			for _, log := range receipt.Logs {
				parser := s.logParsers[log.Address]
				if parser != nil {
					gotEvent, err := parser.ParseLog(log)
					if err == nil && wantEvent.String() == gotEvent.String() {
						found = true
					}
				}
			}
			s.Truef(found, "event not found: %v", wantEvent)
		}
	}
}

// requireTxFails is like requireTxWithEvents, but it requires that the transaction either
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

// assertRSVBalance asserts that the Reserve Dollar balance of `address` is `amount`.
func (s *TestSuite) assertRSVBalance(address common.Address, amount *big.Int) {
	balance, err := s.reserve.BalanceOf(nil, address)
	s.NoError(err)
	s.Equal(amount.String(), balance.String()) // assert.Equal can mis-compare big.Ints, so compare strings instead
}

// assertRSVAllowance asserts that the allowance of Reserve Dollars that `owner` has given `spender` is `amount`.
func (s *TestSuite) assertRSVAllowance(owner, spender common.Address, amount *big.Int) {
	allowance, err := s.reserve.Allowance(nil, owner, spender)
	s.NoError(err)
	s.Equal(amount.String(), allowance.String())
}

// assertRSVTotalSupply asserts that the total supply of Reserve Dollars is `amount`.
func (s *TestSuite) assertRSVTotalSupply(amount *big.Int) {
	totalSupply, err := s.reserve.TotalSupply(nil)
	s.NoError(err)
	s.Equal(amount.String(), totalSupply.String())
}

// assertManagerCollateralized asserts that the Manager is collateralized.
func (s *TestSuite) assertManagerCollateralized() {
	collateralized, err := s.manager.IsFullyCollateralized(nil)
	s.Require().NoError(err)
	s.True(collateralized)
}

// assertBasket asserts that the current manager basket matches expectations.
func (s *TestSuite) assertBasket(basket *abi.Basket, tokens []common.Address, weights []*big.Int) {
	// Get tokens
	basketTokens, err := basket.GetTokens(nil)
	s.Require().NoError(err)
	s.Equal(len(tokens), len(basketTokens))

	// Check that tokens and backings are as expected.
	for i := 0; i < len(tokens); i++ {
		s.Equal(tokens[i], basketTokens[i])
		weight, err := basket.Weights(nil, tokens[i])
		s.Require().NoError(err)
		s.Equal(weights[i].String(), weight.String())
	}
}

// currentTimestamp retrieves the current block time.
func (s *TestSuite) currentTimestamp() *big.Int {
	result := new(big.Int)
	s.NoError(s.utilContract.Call(nil, &result, "time"))
	return result
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
			// of the ReserveV2 constructor. But we still want it about the
			// same order of magnitude as mainnet.
			//
			// The Reserve constructor is edging close to the mainnet block limit.
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

	s.createFastNode()

	// Deploy utility contract just for reading block time
	bytecode := "0x6080604052348015600f57600080fd5b5060918061001e6000396000f3fe6080604052348015600f57600080fd5b50600436106044577c0100000000000000000000000000000000000000000000000000000000600035046316ada54781146049575b600080fd5b604f6061565b60408051918252519081900360200190f35b429056fea165627a7a723058205524d6a0c4d80ea5535c2ea64615c2619a21518e242cb929275cbd678b04468f0029"
	utilABI, err := ethabi.JSON(strings.NewReader(`
	[{"constant":true,"inputs":[],"name":"time","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]
	`))
	s.Require().NoError(err)

	code, err := hex.DecodeString(strings.TrimPrefix(bytecode, "0x"))
	s.Require().NoError(err)

	_, tx, utilContract, err := bind.DeployContract(s.signer, utilABI, code, s.node)
	s.requireTx(tx, err)( /* assert zero events */ )
	s.utilContract = utilContract
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

//////////////// Utility

func maxUint256() *big.Int {
	z := bigInt(1)
	z = z.Lsh(z, 256)
	z = z.Sub(z, bigInt(1))
	return z
}

func maxUint160() *big.Int {
	z := bigInt(1)
	z = z.Lsh(z, 160)
	z = z.Sub(z, bigInt(1))
	return z
}

func minInt160AsUint160() *big.Int {
	z := bigInt(1)
	z = z.Lsh(z, 159)
	return z
}

func bigInt(n uint32) *big.Int {
	return big.NewInt(int64(n))
}

func zeroAddress() common.Address {
	return common.BigToAddress(bigInt(0))
}

func mintingTransfer(to common.Address, value *big.Int) abi.ReserveTransfer {
	return abi.ReserveTransfer{
		From:  common.BigToAddress(bigInt(0)),
		To:    to,
		Value: value,
	}
}

func burningTransfer(from common.Address, value *big.Int) abi.ReserveTransfer {
	return abi.ReserveTransfer{
		From:  from,
		To:    common.BigToAddress(bigInt(0)),
		Value: value,
	}
}

// shiftLeft returns `n`, shifted left by `decimals` zeroes.
func shiftLeft(n uint32, decimals uint32) *big.Int {
	attoBase := big.NewInt(0).Exp(bigInt(10), bigInt(decimals), nil)
	return big.NewInt(0).Mul(bigInt(n), attoBase)
}

// containsAddress tells whether `a` contains `x`.
func containsAddress(a []common.Address, x common.Address) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// TestSuite Helpers

func (s *TestSuite) fundAccountWithErc20sAndApprove(acc account, amounts []*big.Int) {
	// Transfer all of the ERC20 tokens to `proposer`.
	for i, amount := range amounts {
		s.requireTxWithStrictEvents(s.erc20s[i].Transfer(s.signer, acc.address(), amount))(
			abi.BasicERC20Transfer{
				From: s.owner.address(), To: acc.address(), Value: amount,
			},
		)
		// Have `proposer` approve the Manager to spend its funds.
		s.requireTxWithStrictEvents(s.erc20s[i].Approve(signer(acc), s.managerAddress, amount))(
			abi.BasicERC20Approval{
				Owner: acc.address(), Spender: s.managerAddress, Value: amount,
			},
		)
	}
}

func (s *TestSuite) changeBasketUsingWeightProposal(tokens []common.Address, weights []*big.Int) {
	// Propose the new basket.
	s.requireTx(s.manager.ProposeWeights(signer(s.proposer), tokens, weights))

	// Confirm proposals length increments.
	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	proposalID := bigInt(0).Sub(proposalsLength, bigInt(1))

	// Construct Proposal binding.
	proposalAddress, err := s.manager.TrustedProposals(nil, proposalID)
	s.Require().NoError(err)
	proposal, err := abi.NewWeightProposal(proposalAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[proposalAddress] = proposal

	// Get Proposal Basket.
	proposalBasketAddress, err := proposal.TrustedBasket(nil)
	s.Require().NoError(err)
	s.NotEqual(zeroAddress(), proposalBasketAddress)

	basket, err := abi.NewBasket(proposalBasketAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[proposalBasketAddress] = basket

	// Check Basket has correct fields
	// Tokens
	basketTokens, err := basket.GetTokens(nil)
	s.Require().NoError(err)
	s.True(reflect.DeepEqual(basketTokens, tokens))

	// Size
	basketSize, err := basket.Size(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(uint32(len(tokens))).String(), basketSize.String())

	// Weights
	for i := 0; i < len(weights); i++ {
		foundBacking, err := basket.Weights(nil, tokens[i])
		s.Require().NoError(err)
		s.Equal(weights[i], foundBacking)
	}

	// Accept the Proposal.
	s.requireTx(s.manager.AcceptProposal(signer(s.operator), proposalID))(
		abi.ManagerProposalAccepted{
			Id: proposalID, Proposer: s.proposer.address(),
		},
	)

	// Confirm we cannot execute the proposal yet.
	s.requireTxFails(s.manager.ExecuteProposal(signer(s.operator), proposalID))

	// Advance 24h.
	s.Require().NoError(s.node.(backend).AdjustTime(24 * time.Hour))

	// Execute Proposal.
	s.requireTx(s.manager.ExecuteProposal(signer(s.operator), proposalID))

	// Gets the current basket and makes sure it is correct.
	s.assertBasket(basket, tokens, weights)

	// Assert that the vault is still collateralized.
	s.assertManagerCollateralized()
}

func (s *TestSuite) changeBasketUsingSwapProposal(tokens []common.Address, amounts []*big.Int, toVault []bool) {
	// Propose the new basket.
	s.requireTx(s.manager.ProposeSwap(signer(s.proposer), tokens, amounts, toVault))

	// Confirm proposals length increments.
	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	proposalID := bigInt(0).Sub(proposalsLength, bigInt(1))

	// Construct Proposal binding.
	proposalAddress, err := s.manager.TrustedProposals(nil, proposalID)
	s.Require().NoError(err)
	proposal, err := abi.NewSwapProposal(proposalAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[proposalAddress] = proposal

	// Accept the Proposal.
	s.requireTx(s.manager.AcceptProposal(signer(s.operator), proposalID))(
		abi.ManagerProposalAccepted{
			Id: proposalID, Proposer: s.proposer.address(),
		},
	)

	// Confirm we cannot execute the proposal yet.
	s.requireTxFails(s.manager.ExecuteProposal(signer(s.operator), proposalID))

	// Advance 24h.
	s.Require().NoError(s.node.(backend).AdjustTime(24 * time.Hour))

	// Execute Proposal.
	s.requireTx(s.manager.ExecuteProposal(signer(s.operator), proposalID))

	// Gets the current basket and makes sure it is correct.
	// s.assertBasket(basket, tokens, weights)

	// Assert that the vault is still collateralized.
	s.assertManagerCollateralized()
}

func (s *TestSuite) computeExpectedIssueAmounts(
	seigniorage *big.Int, rsvSupply *big.Int,
) []*big.Int {
	BPS_FACTOR := bigInt(10000)

	// Get current basket.
	basketAddress, err := s.manager.TrustedBasket(nil)
	s.Require().NoError(err)
	basket, err := abi.NewBasket(basketAddress, s.node)
	s.Require().NoError(err)
	size, err := basket.Size(nil)
	s.Require().NoError(err)

	// Compute expected amounts.
	var expectedAmounts []*big.Int
	for i := bigInt(0); i.Cmp(size) == -1; i.Add(i, bigInt(1)) {
		token, err := basket.Tokens(nil, i)
		s.Require().NoError(err)
		weight, err := basket.Weights(nil, token)
		s.Require().NoError(err)

		// Compute expectedAmount.
		sum := bigInt(0).Add(BPS_FACTOR, seigniorage)
		effectiveAmount := bigInt(0).Div(bigInt(0).Mul(rsvSupply, sum), BPS_FACTOR)
		expectedAmount := bigInt(0).Div(bigInt(0).Mul(effectiveAmount, weight), shiftLeft(1, 36))
		expectedAmounts = append(expectedAmounts, expectedAmount)
	}

	return expectedAmounts
}

func (s *TestSuite) computeExpectedRedeemAmounts(rsvSupply *big.Int) []*big.Int {
	// Get current basket.
	basketAddress, err := s.manager.TrustedBasket(nil)
	s.Require().NoError(err)
	basket, err := abi.NewBasket(basketAddress, s.node)
	s.Require().NoError(err)
	size, err := basket.Size(nil)
	s.Require().NoError(err)

	// Compute expected amounts.
	var expectedAmounts []*big.Int
	for i := bigInt(0); i.Cmp(size) == -1; i.Add(i, bigInt(1)) {
		token, err := basket.Tokens(nil, i)
		s.Require().NoError(err)
		weight, err := basket.Weights(nil, token)
		s.Require().NoError(err)

		// Compute expectedAmount.
		expectedAmount := bigInt(0).Div(bigInt(0).Mul(rsvSupply, weight), shiftLeft(1, 36))
		expectedAmounts = append(expectedAmounts, expectedAmount)
	}

	return expectedAmounts
}

func (s *TestSuite) newWeights(
	oldWeights []*big.Int, amounts []*big.Int, toVault []bool,
) []*big.Int {
	// Find rsv supply
	rsvSupply, err := s.reserve.TotalSupply(nil)
	s.Require().NoError(err)

	// Compute newWeights.
	var newWeights []*big.Int
	for i, _ := range s.erc20s {
		weight := oldWeights[i]
		oldAmount := bigInt(0).Mul(weight, rsvSupply)

		var newAmount *big.Int
		if toVault[i] {
			newAmount = bigInt(0).Add(oldAmount, amounts[i])
		} else {
			newAmount = bigInt(0).Sub(oldAmount, amounts[i])
		}

		// TODO: Rounding?
		if rsvSupply.Cmp(newAmount) == 1 {
			newWeights[i] = bigInt(0).Div(newAmount, rsvSupply)
		} else {
			newWeights[i] = bigInt(0)
		}
	}

	return newWeights
}
