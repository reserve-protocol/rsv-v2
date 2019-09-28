package tests

import (
	"fmt"
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
}

type SwapProposalSuite struct {
	TestSuite

	proposer        account
	proposal        *abi.SwapProposal
	proposalAddress common.Address
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
	basketAddress := s.account[2].address()

	// Deploy a Weight Proposal.
	proposalAddress, tx, proposal, err := abi.DeployWeightProposal(s.signer, s.node, s.proposer.address(), basketAddress)

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
	s.Equal(basketAddress, foundBasketAddress)

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
	basketAddress := s.account[3].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Complete the proposal.
	s.requireTxWeakly(s.proposal.Complete(s.signer, rsvAddress, basketAddress))(
		abi.WeightProposalCompletedProposalWithBasket{BasketAddress: basketAddress},
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
	basketAddress := s.account[3].address()

	// Accept proposal.
	s.requireTxStrongly(s.proposal.Accept(s.signer, time))()

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), rsvAddress, basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteRequiresAccepted() {
	rsvAddress := s.account[2].address()
	basketAddress := s.account[3].address()

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, basketAddress))

	// Cancel the proposal.
	s.requireTxStrongly(s.proposal.Cancel(s.signer))()

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, rsvAddress, basketAddress))
}

// ========================= SwapProposal Tests =================================

// BeforeTest runs before each test in the suite.
func (s *SwapProposalSuite) BeforeTest(suiteName, testName string) {
	// s.owner = s.account[0]
	// s.proposer = s.account[1]
	// basketAddress := s.account[2].address()

	// // Deploy a Weight Proposal.
	// proposalAddress, tx, proposal, err := abi.DeployWeightProposal(s.signer, proposer.address(), basketAddress)
	// s.requireTxStrongly(tx, err)(
	//     abi.ProposalOwnershipTransferred{
	//         PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	//     },
	// )

	// // Check that state was constructed correctly
	// proposer, err := proposal.Proposer(nil)
	// s.Require().NoError(err)
	// s.Equal(s.proposer, proposer)

	// time, err := proposal.Time(nil)
	// s.Require().NoError(err)
	// s.Equal(bigInt(0).String, time)

	// state, err := proposal.State(nil)
	// s.Require().NoError(err)
	// s.Equal(0, state)

	// foundBasketAddress, err := proposal.Basket(nil)
	// s.Require().NoError(err)
	// s.Equal(basketAddress, foundBasketAddress)

	// s.proposal = proposal
	// s.proposalAddress = proposalAddress

}
