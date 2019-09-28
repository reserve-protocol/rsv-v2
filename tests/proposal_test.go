package tests

import (
	"fmt"
	"math/big"
	"os/exec"
	"testing"

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
	amounts         []*big.Int
	toVault         []bool
}

var (
	// Compile-time check that WeightProposalSuite implements the interfaces we think it does.
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
	if coverageEnabled {
		s.createSlowCoverageNode()
	} else {
		s.createFastNode()
	}
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

	s.requireTxStrongly(tx, err)(
		abi.WeightProposalOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
	)

	// Check that proposer was set.
	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.proposer.address(), proposer)

	// Check that time was not set.
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

}

func (s *WeightProposalSuite) TestDeploy() {

}

// TestAccept tests that `accept` changes state as expected.
func (s *WeightProposalSuite) TestAccept() {
	time := bigInt(100)

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Check that the time was set.
	foundTime, err := s.proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(time.String(), foundTime.String())

	// Check that the state is Accepted.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(1), state)
}

// TestAcceptIsProtected tests that `accept` is protected.
func (s *WeightProposalSuite) TestAcceptIsProtected() {
	time := bigInt(100)

	// Accept proposal.
	s.requireTxFails(s.proposal.Accept(signer(s.proposer), time))
}

// TestAcceptRequiresCreated tests that `accept` reverts in the require case.
func (s *WeightProposalSuite) TestAcceptRequiresCreated() {
	time := bigInt(100)

	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

	// Confirm it's cancelled.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(2), state)

	// Now try to accept.
	s.requireTxFails(s.proposal.Accept(s.signer, time))
}

// TestCancel tests that `cancel` changes the state as expected.
func (s *WeightProposalSuite) TestCancel() {
	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

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

// // Test that `cancel` reverts if the state is completed..
// func (s *WeightProposalSuite) TestCancelRequiresNotCompleted() {
// 	// Cancel the proposal.
// 	s.requireTxFails(s.proposal.Cancel(s.signer))
// }

// TestComplete tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestComplete() {
	time := bigInt(100)
	rsvAddress := s.account[2].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxStrongly(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

}

// TestCompleteIsProtected tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteIsProtected() {
	time := bigInt(100)
	rsvAddress := s.account[2].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), rsvAddress, s.basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteRequiresAccepted() {
	rsvAddress := s.account[2].address()

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))

	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))
}

// TestCompleteCanOnlyHappenOnce tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteCanOnlyHappenOnce() {
	time := bigInt(100)
	rsvAddress := s.account[2].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxStrongly(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

	// Check that the proposal can't be completed a second time.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))
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

	s.requireTxStrongly(tx, err)(
		abi.SwapProposalOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
	)

	// Check that proposer was set.
	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.proposer.address(), proposer)

	// Check that time was not set.
	time, err := proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), time.String())

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

}

func (s *SwapProposalSuite) TestDeploy() {

}

// TestAccept tests that `accept` changes state as expected.
func (s *SwapProposalSuite) TestAccept() {
	time := bigInt(100)

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Check that the time was set.
	foundTime, err := s.proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(time.String(), foundTime.String())

	// Check that the state is Accepted.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(1), state)
}

// TestAcceptIsProtected tests that `accept` is protected.
func (s *SwapProposalSuite) TestAcceptIsProtected() {
	time := bigInt(100)

	// Accept proposal.
	s.requireTxFails(s.proposal.Accept(signer(s.proposer), time))
}

// TestAcceptRequiresCreated tests that `accept` reverts in the require case.
func (s *SwapProposalSuite) TestAcceptRequiresCreated() {
	time := bigInt(100)

	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

	// Confirm it's cancelled.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(2), state)

	// Now try to accept.
	s.requireTxFails(s.proposal.Accept(s.signer, time))
}

// TestCancel tests that `cancel` changes the state as expected.
func (s *SwapProposalSuite) TestCancel() {
	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

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

// // Test that `cancel` reverts if the state is completed..
// func (s *SwapProposalSuite) TestCancelRequiresNotCompleted() {
// 	// Cancel the proposal.
// 	s.requireTxFails(s.proposal.Cancel(s.signer))
// }

// TestComplete tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestComplete() {
	time := bigInt(100)
	rsvAddress := s.account[2].address()

	// // Set up basket that the Swap Proposal will be extended from.
	// // Deploy collateral ERC20s
	// s.erc20s = make([]*abi.BasicERC20, 3)
	// s.erc20Addresses = make([]common.Address, 3)
	// for i := 0; i < 3; i++ {
	// 	erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
	// 	s.Require().NoError(err)
	// 	s.erc20s[i] = erc20
	// 	s.erc20Addresses[i] = erc20Address
	// }

	// s.weights = makeLinearWeights(bigInt(1), len(s.erc20s))

	// // Make a simple basket
	// basketAddress, tx, basket, err := abi.DeployBasket(
	// 	s.signer,
	// 	s.node,
	// 	zeroAddress(),
	// 	s.erc20Addresses,
	// 	s.weights,
	// )

	// s.requireTxStrongly(tx, err)()
	// s.basketAddress = basketAddress
	// s.basket = basket

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxStrongly(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

}

// TestCompleteIsProtected tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteIsProtected() {
	time := bigInt(100)
	rsvAddress := s.account[2].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), rsvAddress, s.basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteRequiresAccepted() {
	rsvAddress := s.account[2].address()

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))

	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))
}

// TestCompleteCanOnlyHappenOnce tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteCanOnlyHappenOnce() {
	time := bigInt(100)
	rsvAddress := s.account[2].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Complete the proposal and check the right basket is returned.
	s.requireTxStrongly(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: s.basketAddress},
	)

	// Check that State is Completed.
	state, err := s.proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(3), state)

	// Check that the proposal can't be completed a second time.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, s.basketAddress))
}
