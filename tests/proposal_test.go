package tests

import (
	"fmt"
	"math/big"
	"os/exec"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
	"github.com/reserve-protocol/rsv-beta/soltools"
)

func TestProposal(t *testing.T) {
	suite.Run(t, new(WeightProposalSuite))
	suite.Run(t, new(SwapProposalSuite))
}

type WeightProposalSuite struct {
	TestSuite

	proposer        account
	proposal        *abi.WeightProposal
	proposalAddress common.Address
	basketAddress   common.Address
}

type SwapProposalSuite struct {
	TestSuite

	proposer        account
	proposal        *abi.SwapProposal
	proposalAddress common.Address
	tokens          []common.Address
	weights         []*big.Int
	amounts         []*big.Int
	toVault         []bool
}

var (
	// Compile-futureTimecheck that WeightProposalSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &WeightProposalSuite{}
	_ suite.SetupAllSuite    = &WeightProposalSuite{}
	_ suite.TearDownAllSuite = &WeightProposalSuite{}
	_ suite.BeforeTest       = &SwapProposalSuite{}
	_ suite.SetupAllSuite    = &SwapProposalSuite{}
	_ suite.TearDownAllSuite = &SwapProposalSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *WeightProposalSuite) SetupSuite() {
	s.setup()
}

// SetupSuite runs once, before all of the tests in the suite.
func (s *SwapProposalSuite) SetupSuite() {
	s.setup()
}

// TearDownSuite runs once, after all of the tests in the suite.
func (s *WeightProposalSuite) TearDownSuite() {
	if coverageEnabled {
		// Write coverage profile to disk.
		s.Assert().NoError(s.node.(*soltools.Backend).WriteCoverage())

		// Close the node.js process.
		s.Assert().NoError(s.node.(*soltools.Backend).Close())

		// Process coverage profile into an HTML report.
		if out, err := exec.Command("npx", "istanbul", "report", "html").CombinedOutput(); err != nil {
			fmt.Println()
			fmt.Println("I generated coverage information in coverage/coverage.json.")
			fmt.Println("I tried to process it with `istanbul` to turn it into a readable report, but failed.")
			fmt.Println("The error I got when running istanbul was:", err)
			fmt.Println("Istanbul's output was:\n" + string(out))
		}
	}
}

// TearDownSuite runs once, after all of the tests in the suite.
func (s *SwapProposalSuite) TearDownSuite() {
	if coverageEnabled {
		// Write coverage profile to disk.
		s.Assert().NoError(s.node.(*soltools.Backend).WriteCoverage())

		// Close the node.js process.
		s.Assert().NoError(s.node.(*soltools.Backend).Close())

		// Process coverage profile into an HTML report.
		if out, err := exec.Command("npx", "istanbul", "report", "html").CombinedOutput(); err != nil {
			fmt.Println()
			fmt.Println("I generated coverage information in coverage/coverage.json.")
			fmt.Println("I tried to process it with `istanbul` to turn it into a readable report, but failed.")
			fmt.Println("The error I got when running istanbul was:", err)
			fmt.Println("Istanbul's output was:\n" + string(out))
		}
	}
}

// ========================= WeightProposal Tests =================================

// BeforeTest runs before each test in the suite.
func (s *WeightProposalSuite) BeforeTest(suiteName, testName string) {
	s.owner = s.account[0]
	s.proposer = s.account[1]
	s.basketAddress = s.account[2].address()

	// Deploy a Weight Proposal.
	proposalAddress, tx, proposal, err := abi.DeployWeightProposal(s.signer, s.node, s.proposer.address(), s.basketAddress)

	s.logParsers = map[common.Address]logParser{
		proposalAddress: proposal,
	}

	s.requireTxWithEvents(tx, err)(
		abi.WeightProposalOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
	)

	// Check that proposer was set.
	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.proposer.address(), proposer)

	// Check that futureTimewas not set.
	time, err := proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), time.String())

	// Check that the initial state is Created.
	state, err := proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(0), state)

	// Check that the basket was set.
	foundBasketAddress, err := proposal.Basket(nil)
	s.Require().NoError(err)
	s.Equal(s.basketAddress, foundBasketAddress)

	s.proposal = proposal
	s.proposalAddress = proposalAddress

	// Set an arbitrary address for rsv.
	s.reserveAddress = s.account[3].address()
}

func (s *WeightProposalSuite) TestDeploy() {

}

// TestAccept tests that `accept` changes state as expected.
func (s *WeightProposalSuite) TestAccept() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Check that the futureTimewas set.
	foundTime, err := s.proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(futureTime.String(), foundTime.String())

	// Check that the state is Accepted.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(1), state)
}

// TestAcceptIsProtected tests that `accept` is protected.
func (s *WeightProposalSuite) TestAcceptIsProtected() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxFails(s.proposal.Accept(signer(s.proposer), futureTime))
}

// TestAcceptRequiresCreated tests that `accept` reverts in the require case.
func (s *WeightProposalSuite) TestAcceptRequiresCreated() {
	futureTime := bigInt(0)

	// Cancel the proposal.
	s.requireTxWithEvents(s.proposal.Cancel(s.signer))()

	// Confirm it's cancelled.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(2), state)

	// Now try to accept.
	s.requireTxFails(s.proposal.Accept(s.signer, futureTime))
}

// TestCancel tests that `cancel` changes the state as expected.
func (s *WeightProposalSuite) TestCancel() {
	// Cancel the proposal.
	s.requireTxWithEvents(s.proposal.Cancel(s.signer))()

	// Confirm it's cancelled.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(2), state)
}

// TestCancelIsProtected tests that `cancel` is protected.
func (s *WeightProposalSuite) TestCancelIsProtected() {
	// Cancel the proposal.
	s.requireTxFails(s.proposal.Cancel(signer(s.proposer)))
}

// Test that `cancel` reverts if the state is completed..
func (s *WeightProposalSuite) TestCancelRequiresNotCompleted() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxWithEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

	// Check that cancelling at this point reverts.
	s.requireTxFails(s.proposal.Cancel(s.signer))
}

// TestComplete tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestComplete() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxWithEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

}

// TestCompleteIsProtected tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteIsProtected() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), s.reserveAddress, s.basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteRequiresAccepted() {

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Cancel the proposal.
	s.requireTxWithEvents(s.proposal.Cancel(s.signer))()

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// TestCompleteCanOnlyHappenOnce tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteCanOnlyHappenOnce() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxWithEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

	// Check that the proposal can't be completed a second time.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// TestCompleteCannotHappenBeforeNow tests that `complete` reverts when it is not yet `now`.
func (s *WeightProposalSuite) TestCompleteCannotHappenBeforeNow() {

	currentTime := s.currentTimestamp()
	futureTime := bigInt(0).Add(currentTime, bigInt(100))

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Check that calling `complete` reverts.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Advance the time.
	s.node.(backend).AdjustTime(100 * time.Second)

	// Now the proposal can be completed.
	s.requireTxWithEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)
}

// ========================= SwapProposal Tests =================================

// BeforeTest runs before each test in the suite.
func (s *SwapProposalSuite) BeforeTest(suiteName, testName string) {
	s.owner = s.account[0]
	s.proposer = s.account[1]

	s.tokens = []common.Address{s.account[3].address(), s.account[4].address()}
	s.amounts = []*big.Int{bigInt(1), bigInt(2)}
	s.toVault = []bool{true, true}

	// Deploy a Weight Proposal.
	proposalAddress, tx, proposal, err := abi.DeploySwapProposal(
		s.signer, s.node, s.proposer.address(), s.tokens, s.amounts, s.toVault)

	s.logParsers = map[common.Address]logParser{
		proposalAddress: proposal,
	}

	s.requireTxWithEvents(tx, err)(
		abi.SwapProposalOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
	)

	// Check that proposer was set.
	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.proposer.address(), proposer)

	// Check that futureTimewas not set.
	futureTime, err := proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), futureTime.String())

	// Check that the initial state is Created.
	state, err := proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(0), state)

	for i := 0; i < len(s.tokens); i++ {
		index := bigInt(uint32(i))
		// Check that tokens was set.
		token, err := proposal.Tokens(nil, index)
		s.Require().NoError(err)
		s.Equal(s.tokens[i], token)

		// Check that amounts was set.
		amount, err := proposal.Amounts(nil, index)
		s.Require().NoError(err)
		s.Equal(s.amounts[i], amount)

		// Check that toVault was set.
		toVault, err := proposal.ToVault(nil, index)
		s.Require().NoError(err)
		s.Equal(s.toVault[i], toVault)
	}

	s.proposal = proposal
	s.proposalAddress = proposalAddress

	// Deploy a Reserve instance as well.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		reserveAddress: reserve,
	}

	s.requireTxWithEvents(tx, err)(abi.ReserveOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Unpause.
	s.requireTxWithEvents(s.reserve.Unpause(s.signer))(
		abi.ReserveUnpaused{Account: s.owner.address()},
	)

	// Make RSV supply nonzero so weights can be calculated.
	s.requireTxWithEvents(s.reserve.ChangeMinter(s.signer, s.owner.address()))(
		abi.ReserveMinterChanged{NewMinter: s.owner.address()},
	)
	s.requireTxWithEvents(s.reserve.Mint(s.signer, s.proposer.address(), bigInt(1)))(
		mintingTransfer(s.proposer.address(), bigInt(1)),
	)

	// Deploy collateral ERC20s
	s.erc20s = make([]*abi.BasicERC20, 3)
	s.erc20Addresses = make([]common.Address, 3)
	s.weights = make([]*big.Int, 3)
	for i := 0; i < 3; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)
		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
		s.weights[i] = bigInt(uint32(i + 1))
		s.logParsers[erc20Address] = erc20
	}

	// Finally, deploy a basket.
	basketAddress, tx, basket, err := abi.DeployBasket(
		s.signer,
		s.node,
		zeroAddress(),
		s.erc20Addresses,
		s.weights,
	)

	s.requireTxWithEvents(tx, err)()
	s.basketAddress = basketAddress
	s.basket = basket
}

func (s *SwapProposalSuite) TestDeploy() {

}

// TestAccept tests that `accept` changes state as expected.
func (s *SwapProposalSuite) TestAccept() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Check that the futureTimewas set.
	foundTime, err := s.proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(futureTime.String(), foundTime.String())

	// Check that the state is Accepted.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(1), state)
}

// TestAcceptIsProtected tests that `accept` is protected.
func (s *SwapProposalSuite) TestAcceptIsProtected() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxFails(s.proposal.Accept(signer(s.proposer), futureTime))
}

// TestAcceptRequiresCreated tests that `accept` reverts in the require case.
func (s *SwapProposalSuite) TestAcceptRequiresCreated() {
	futureTime := bigInt(0)

	// Cancel the proposal.
	s.requireTxWithEvents(s.proposal.Cancel(s.signer))()

	// Confirm it's cancelled.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(2), state)

	// Now try to accept.
	s.requireTxFails(s.proposal.Accept(s.signer, futureTime))
}

// TestCancel tests that `cancel` changes the state as expected.
func (s *SwapProposalSuite) TestCancel() {
	// Cancel the proposal.
	s.requireTxWithEvents(s.proposal.Cancel(s.signer))()

	// Confirm it's cancelled.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(2), state)
}

// TestCancelIsProtected tests that `cancel` is protected.
func (s *SwapProposalSuite) TestCancelIsProtected() {
	// Cancel the proposal.
	s.requireTxFails(s.proposal.Cancel(signer(s.proposer)))
}

// Test that `cancel` reverts if the state is completed..
func (s *SwapProposalSuite) TestCancelRequiresNotCompleted() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Complete the proposal and check the right basket is returned.
	s.requireTx(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

	// Check that cancelling at this point reverts.
	s.requireTxFails(s.proposal.Cancel(s.signer))
}

// TestComplete tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestComplete() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Complete the proposal and check the right basket is returned.
	s.requireTx(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

}

// TestCompleteIsProtected tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteIsProtected() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), s.reserveAddress, s.basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteRequiresAccepted() {

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Cancel the proposal.
	s.requireTxWithEvents(s.proposal.Cancel(s.signer))()

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// TestCompleteCanOnlyHappenOnce tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteCanOnlyHappenOnce() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Complete the proposal and check the right basket is returned.
	s.requireTx(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

	// Check that the proposal can't be completed a second time.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// TestCompleteCannotHappenBeforeNow tests that `complete` reverts when it is not yet `now`.
func (s *SwapProposalSuite) TestCompleteCannotHappenBeforeNow() {

	currentTime := s.currentTimestamp()
	futureTime := bigInt(0).Add(currentTime, bigInt(100))

	// Accept proposal.
	s.requireTxWithEvents(s.proposal.Accept(s.signer, futureTime))()

	// Check that calling `complete` reverts.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Advance the time.
	s.node.(backend).AdjustTime(100 * time.Second)

	// Now the proposal can be completed.
	s.requireTx(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}
