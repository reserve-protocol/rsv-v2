// +build all

package tests

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
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
	weights         []*big.Int
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
}

// ========================= WeightProposal Tests =================================

// BeforeTest runs before each test in the suite.
func (s *WeightProposalSuite) BeforeTest(suiteName, testName string) {
	s.owner = s.account[0]
	s.proposer = s.account[1]

	s.logParsers = map[common.Address]logParser{}

	// Deploy collateral ERC20s for a basket.
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

	// Make a non-empty basket
	basketAddress, tx, basket, err := abi.DeployBasket(
		s.signer,
		s.node,
		zeroAddress(),
		s.erc20Addresses,
		s.weights,
	)

	s.requireTxWithStrictEvents(tx, err)()
	s.basketAddress = basketAddress
	s.basket = basket

	// Deploy a Weight Proposal.
	proposalAddress, tx, proposal, err := abi.DeployWeightProposal(s.signer, s.node, s.proposer.address(), s.basketAddress)

	s.logParsers[proposalAddress] = proposal

	s.requireTxWithStrictEvents(tx, err)(
		abi.WeightProposalOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
		abi.WeightProposalProposalCreated{
			Proposer: s.proposer.address(),
		},
	)

	// Check that proposer was set.
	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.proposer.address(), proposer)

	// Check that futureTime was not set.
	time, err := proposal.Time(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), time.String())

	// Check that the initial state is Created.
	state, err := proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(0), state)

	// Check that the basket was set.
	foundBasketAddress, err := proposal.TrustedBasket(nil)
	s.Require().NoError(err)
	s.Equal(s.basketAddress, foundBasketAddress)

	s.proposal = proposal
	s.proposalAddress = proposalAddress

	// Set an arbitrary address for rsv.
	s.reserveAddress = s.account[3].address()
}

func (s *WeightProposalSuite) TestDeploy() {

}

// TestBadConstruction tests that the requires in the constructor exist.
func (s *WeightProposalSuite) TestBadConstruction() {
	// A basket can't even be created if it is empty, so we should fail here:
	basketAddress, tx, _, err := abi.DeployBasket(
		s.signer,
		s.node,
		s.account[5].address(),
		[]common.Address{},
		[]*big.Int{},
	)
	s.requireTxFails(tx, err)

	// However, if we get here, we should fail anyway.
	_, tx, _, err = abi.DeployWeightProposal(s.signer, s.node, s.proposer.address(), basketAddress)
	s.requireTxFails(tx, err)
}

// TestAccept tests that `accept` changes state as expected.
func (s *WeightProposalSuite) TestAccept() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.WeightProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Cancel(s.signer))(
		abi.WeightProposalProposalCancelled{Proposer: s.proposer.address()},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Cancel(s.signer))(
		abi.WeightProposalProposalCancelled{Proposer: s.proposer.address()},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.WeightProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Complete the proposal and check the right basket is returned.
	s.requireTxWithStrictEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalProposalCompleted{Proposer: s.proposer.address(), Basket: s.basketAddress},
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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.WeightProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Complete the proposal and check the right basket is returned.
	s.requireTxWithStrictEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalProposalCompleted{Proposer: s.proposer.address(), Basket: s.basketAddress},
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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.WeightProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), s.reserveAddress, s.basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteRequiresAccepted() {

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Cancel the proposal.
	s.requireTxWithStrictEvents(s.proposal.Cancel(s.signer))(
		abi.WeightProposalProposalCancelled{Proposer: s.proposer.address()},
	)

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// TestCompleteCanOnlyHappenOnce tests that `complete` changes the state as expected.
func (s *WeightProposalSuite) TestCompleteCanOnlyHappenOnce() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.WeightProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Complete the proposal and check the right basket is returned.
	s.requireTxWithStrictEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalProposalCompleted{Proposer: s.proposer.address(), Basket: s.basketAddress},
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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.WeightProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Check that calling `complete` reverts.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Advance the time.
	s.node.(backend).AdjustTime(100 * time.Second)

	// Now the proposal can be completed.
	s.requireTxWithStrictEvents(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))(
		abi.WeightProposalProposalCompleted{Proposer: s.proposer.address(), Basket: s.basketAddress},
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

	// Deploy a Swap Proposal.
	proposalAddress, tx, proposal, err := abi.DeploySwapProposal(
		s.signer, s.node, s.proposer.address(), s.tokens, s.amounts, s.toVault)

	s.logParsers = map[common.Address]logParser{
		proposalAddress: proposal,
	}

	s.requireTxWithStrictEvents(tx, err)(
		abi.SwapProposalOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
		abi.SwapProposalProposalCreated{
			Proposer: s.proposer.address(),
		},
	)

	// Check that proposer was set.
	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.proposer.address(), proposer)

	// Check that futureTime was not set.
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

	// Deploy PreviousReserve to set up for upgrade.
	oldReserveAddress, tx, oldReserve, err := abi.DeployPreviousReserve(s.signer, s.node)

	s.logParsers[oldReserveAddress] = oldReserve

	s.requireTx(tx, err)(
		abi.PreviousReserveOwnershipTransferred{PreviousOwner: zeroAddress(), NewOwner: s.owner.address()},
	)

	oldMaxSupply, err := oldReserve.MaxSupply(nil)
	s.Require().NoError(err)

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = oldReserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	// Deploy Reserve and store a handle to the Go binding and the contract address.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers[reserveAddress] = reserve

	s.requireTx(tx, err)
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Confirm it begins paused.
	paused, err := reserve.Paused(nil)
	s.Require().NoError(err)
	s.Equal(true, paused)

	// Upgrade PreviousReserve to Reserve.
	s.requireTxWithStrictEvents(oldReserve.NominateNewOwner(s.signer, reserveAddress))(
		abi.PreviousReserveNewOwnerNominated{
			PreviousOwner: s.owner.address(), Nominee: reserveAddress,
		},
	)
	s.requireTxWithStrictEvents(s.reserve.AcceptUpgrade(s.signer, oldReserveAddress))(
		abi.ReserveMaxSupplyChanged{NewMaxSupply: oldMaxSupply},
		abi.ReserveUnpaused{Account: s.owner.address()},
		abi.PreviousReserveOwnershipTransferred{
			PreviousOwner: s.owner.address(), NewOwner: reserveAddress,
		},
		abi.PreviousReservePauserChanged{NewPauser: reserveAddress},
		abi.PreviousReservePaused{Account: reserveAddress},
		abi.PreviousReserveEternalStorageTransferred{NewReserveAddress: reserveAddress},
		abi.ReserveEternalStorageReserveAddressTransferred{
			OldReserveAddress: oldReserveAddress,
			NewReserveAddress: reserveAddress,
		},
		abi.PreviousReserveMinterChanged{NewMinter: zeroAddress()},
		abi.PreviousReservePauserChanged{NewPauser: zeroAddress()},
		abi.PreviousReserveOwnershipTransferred{
			PreviousOwner: reserveAddress, NewOwner: zeroAddress(),
		},
	)

	// Accept ownership.
	s.requireTxWithStrictEvents(s.eternalStorage.AcceptOwnership(s.signer))(
		abi.ReserveEternalStorageOwnershipTransferred{
			PreviousOwner: oldReserveAddress, NewOwner: s.owner.address(),
		},
	)

	// Make RSV supply nonzero so weights can be calculated.
	s.requireTxWithStrictEvents(s.reserve.ChangeMinter(s.signer, s.owner.address()))(
		abi.ReserveMinterChanged{NewMinter: s.owner.address()},
	)
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, s.proposer.address(), bigInt(1)))(
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

	s.requireTx(tx, err)()
	s.basketAddress = basketAddress
	s.basket = basket
}

func (s *SwapProposalSuite) TestDeploy() {

}

// TestBadConstruction tests that the requires in the SwapProposal constructor exist.
func (s *SwapProposalSuite) TestBadConstruction() {
	// Test that we cannot create a SwapProposal with tokens.length equal to 0.
	_, tx, _, err := abi.DeploySwapProposal(
		s.signer, s.node, s.proposer.address(), []common.Address{}, []*big.Int{}, []bool{})
	s.requireTxFails(tx, err)

	// Test that we cannot create a SwapProposal with bad token length.
	_, tx, _, err = abi.DeploySwapProposal(
		s.signer, s.node, s.proposer.address(), []common.Address{s.account[4].address()}, []*big.Int{}, []bool{})
	s.requireTxFails(tx, err)

	// Test that we cannot create a SwapProposal with bad amounts length.
	_, tx, _, err = abi.DeploySwapProposal(
		s.signer, s.node, s.proposer.address(), []common.Address{}, []*big.Int{bigInt(0)}, []bool{})
	s.requireTxFails(tx, err)

	// Test that we cannot create a SwapProposal with bad toVault length.
	_, tx, _, err = abi.DeploySwapProposal(
		s.signer, s.node, s.proposer.address(), []common.Address{}, []*big.Int{}, []bool{true})
	s.requireTxFails(tx, err)
}

// TestAccept tests that `accept` changes state as expected.
func (s *SwapProposalSuite) TestAccept() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.SwapProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Check that the futureTime was set.
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
	s.requireTxWithStrictEvents(s.proposal.Cancel(s.signer))(
		abi.SwapProposalProposalCancelled{Proposer: s.proposer.address()},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Cancel(s.signer))(
		abi.SwapProposalProposalCancelled{Proposer: s.proposer.address()},
	)

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

// Test that `cancel` reverts if the state is completed.
func (s *SwapProposalSuite) TestCancelRequiresNotCompleted() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.SwapProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.SwapProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.SwapProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Try to complete the proposal with the wrong signer..
	s.requireTxFails(s.proposal.Complete(signer(s.proposer), s.reserveAddress, s.basketAddress))
}

// TestCompleteRequiresAccepted tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteRequiresAccepted() {

	// Try to complete the proposal without accepting it first.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Cancel the proposal.
	s.requireTxWithStrictEvents(s.proposal.Cancel(s.signer))(
		abi.SwapProposalProposalCancelled{Proposer: s.proposer.address()},
	)

	// Try to complete the proposal after it has been cancelled.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// TestCompleteCanOnlyHappenOnce tests that `complete` changes the state as expected.
func (s *SwapProposalSuite) TestCompleteCanOnlyHappenOnce() {
	futureTime := bigInt(0)

	// Accept proposal.
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.SwapProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

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
	s.requireTxWithStrictEvents(s.proposal.Accept(s.signer, futureTime))(
		abi.SwapProposalProposalAccepted{Proposer: s.proposer.address(), Time: futureTime},
	)

	// Check that calling `complete` reverts.
	s.requireTxFails(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))

	// Advance the time.
	s.node.(backend).AdjustTime(100 * time.Second)

	// Now the proposal can be completed.
	s.requireTx(s.proposal.Complete(s.signer, s.reserveAddress, s.basketAddress))
}

// On proposal upgrades:
//   There is no reason to upgrade SwapProposal or WeightProposal. They can just be replaced by changes to
//   ProposalFactory.
//   There is no state in ProposalFactory, so upgrading it just amounts to either
//   (1) deploying a new ProposalFactory contract and changing the pointer in the Manager or
//   (2) upgrading the Manager contract, if the ProposalFactory and related code are changed enough
//   that it would require a Manager upgrade to handle proposals in the new way.
//   TODO: Move these thoughts to the design documentation and change these comments to point to that.
