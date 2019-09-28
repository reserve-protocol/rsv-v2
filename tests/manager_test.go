package tests

import (
	"fmt"
	"math/big"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
	"github.com/reserve-protocol/rsv-beta/soltools"
)

func TestManager(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

type ManagerSuite struct {
	TestSuite

	operator account
	proposer account
}

var (
	// Compile-time check that ManagerSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &ManagerSuite{}
	_ suite.SetupAllSuite    = &ManagerSuite{}
	_ suite.TearDownAllSuite = &ManagerSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *ManagerSuite) SetupSuite() {
	s.setup()
}

// TearDownSuite runs once, after all of the tests in the suite.
func (s *ManagerSuite) TearDownSuite() {
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

// BeforeTest runs before each test in the suite.
func (s *ManagerSuite) BeforeTest(suiteName, testName string) {
	s.owner = s.account[0]
	s.operator = s.account[1]
	s.proposer = s.account[5]

	// Deploy Reserve and store a handle to the Go binding and the contract address.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		reserveAddress: reserve,
	}

	s.requireTxWithEvents(tx, err)(abi.ReserveOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Unpause Reserve.
	s.requireTxWithEvents(s.reserve.Unpause(s.signer))(
		abi.ReserveUnpaused{Account: s.owner.address()},
	)

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = s.reserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	// Vault.
	vaultAddress, tx, vault, err := abi.DeployVault(s.signer, s.node)

	s.logParsers[vaultAddress] = vault
	s.requireTxWithEvents(tx, err)(
		abi.VaultOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
		abi.VaultManagerTransferred{
			PreviousManager: zeroAddress(), NewManager: s.owner.address(),
		},
	)
	s.vault = vault
	s.vaultAddress = vaultAddress

	// Manager.
	managerAddress, tx, manager, err := abi.DeployManager(
		s.signer, s.node, vaultAddress, reserveAddress, bigInt(0),
	)

	s.logParsers[managerAddress] = manager
	s.requireTxWithEvents(tx, err)(abi.ManagerOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})
	s.manager = manager
	s.managerAddress = managerAddress

	// Set all auths to Manager.
	s.requireTxWithEvents(s.reserve.ChangeMinter(s.signer, managerAddress))(
		abi.ReserveMinterChanged{NewMinter: managerAddress},
	)
	s.requireTxWithEvents(s.reserve.ChangePauser(s.signer, managerAddress))(
		abi.ReservePauserChanged{NewPauser: managerAddress},
	)
	s.requireTxWithEvents(s.vault.ChangeManager(s.signer, managerAddress))(
		abi.VaultManagerTransferred{PreviousManager: s.owner.address(), NewManager: managerAddress},
	)

	// Set the operator.
	s.requireTxWithEvents(s.manager.SetOperator(s.signer, s.operator.address()))(
		abi.ManagerOperatorChanged{
			OldAccount: zeroAddress(), NewAccount: s.operator.address(),
		},
	)

	// Set the basket.
	basketAddress, err := s.manager.Basket(nil)
	s.Require().NoError(err)
	s.NotEqual(zeroAddress(), basketAddress)

	basket, err := abi.NewBasket(basketAddress, s.node)
	s.Require().NoError(err)

	s.basketAddress = basketAddress
	s.basket = basket

	// Deploy collateral ERC20s and set allowances
	s.erc20s = make([]*abi.BasicERC20, 3)
	s.erc20Addresses = make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)

		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
		s.logParsers[erc20Address] = erc20

		// Transfer all of the ERC20 tokens to `proposer`.
		s.requireTxWithEvents(erc20.Transfer(s.signer, s.proposer.address(), shiftRight(1, 36)))(
			abi.BasicERC20Transfer{
				From: s.owner.address(), To: s.proposer.address(), Value: shiftRight(1, 36),
			},
		)
		// Have `proposer` approve the Manager to spend its funds.
		s.requireTxWithEvents(erc20.Approve(signer(s.proposer), s.managerAddress, shiftRight(1, 36)))(
			abi.BasicERC20Approval{
				Owner: s.proposer.address(), Spender: s.managerAddress, Value: shiftRight(1, 36),
			},
		)

	}
}

func (s *ManagerSuite) TestDeploy() {}

// TestConstructor tests that the constructor sets initial state appropriately.
func (s *ManagerSuite) TestConstructor() {
	vaultAddr, err := s.manager.Vault(nil)
	s.Require().NoError(err)
	s.Equal(s.vaultAddress, vaultAddr)

	rsvAddr, err := s.manager.Rsv(nil)
	s.Require().NoError(err)
	s.Equal(s.reserveAddress, rsvAddr)

	seigniorage, err := s.manager.Seigniorage(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), seigniorage.String())

	paused, err := s.manager.Paused(nil)
	s.Require().NoError(err)
	s.Equal(true, paused)
}

// TestProposeWeightsUseCase sets a basket, issues RSV, changes the basket, and redeems RSV.
func (s *ManagerSuite) TestProposeWeightsFullUsecase() {
	tokens := s.erc20Addresses
	weights := []*big.Int{shiftRight(1, 36), shiftRight(2, 36), shiftRight(3, 36)}

	// Change the basket using a weight proposal.
	s.changeBasketUsingWeightProposal(tokens, weights)

	// Now we should be able to unpause.
	s.requireTxWithEvents(s.manager.Unpause(s.signer))(
		abi.ManagerUnpaused{
			Account: s.owner.address(),
		},
	)

	// Issue a billion RSV.
	rsvToIssue := shiftRight(1, 9) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// Change to a new basket.
	newWeights := []*big.Int{shiftRight(2, 36), shiftRight(3, 36), shiftRight(1, 36)}
	s.changeBasketUsingWeightProposal(tokens, newWeights)

	// Approve the manager to spend a billion RSV.
	s.requireTxWithEvents(s.reserve.Approve(signer(s.proposer), s.managerAddress, rsvToIssue))(
		abi.ReserveApproval{Owner: s.proposer.address(), Spender: s.managerAddress, Value: rsvToIssue},
	)

	// Redeem a billion RSV.
	s.requireTx(s.manager.Redeem(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// We should be back to zero RSV supply.
	s.assertRSVTotalSupply(bigInt(0))

}

// TestProposeSwapFullUsecase sets up a basket with a WeightProposal, issues RSV,
// changes the basket using a SwapProposal, and redeems the RSV.
func (s *ManagerSuite) TestProposeSwapFullUsecase() {
	tokens := s.erc20Addresses
	weights := []*big.Int{shiftRight(1, 48), shiftRight(2, 48), shiftRight(3, 48)}

	// Change the basket using a weight proposal.
	s.changeBasketUsingWeightProposal(tokens, weights)

	// Now we should be able to unpause.
	s.requireTxWithEvents(s.manager.Unpause(s.signer))(
		abi.ManagerUnpaused{
			Account: s.owner.address(),
		},
	)

	// Issue a billion RSV.
	rsvToIssue := shiftRight(1, 9) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// Change to a new basket using a SwapProposal
	amounts := []*big.Int{shiftRight(2, 18), shiftRight(3, 18), shiftRight(1, 18)}
	toVault := []bool{true, false, true}
	s.changeBasketUsingSwapProposal(tokens, amounts, toVault)

	// Approve the manager to spend a billion RSV.
	s.requireTxWithEvents(s.reserve.Approve(signer(s.proposer), s.managerAddress, rsvToIssue))(
		abi.ReserveApproval{Owner: s.proposer.address(), Spender: s.managerAddress, Value: rsvToIssue},
	)

	// Redeem a billion RSV.
	s.requireTx(s.manager.Redeem(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// We should be back to zero RSV supply.
	s.assertRSVTotalSupply(bigInt(0))
}

// TestPauseIsProtected tests that pause can only be called by owner.
func (s *ManagerSuite) TestPauseIsProtected() {
	s.requireTxFails(s.manager.Pause(signer(s.account[2])))
	s.requireTxFails(s.manager.Pause(signer(s.operator)))
}

// Helpers

func (s *ManagerSuite) changeBasketUsingWeightProposal(tokens []common.Address, weights []*big.Int) {
	// Propose the new basket.
	s.requireTx(s.manager.ProposeWeights(signer(s.proposer), tokens, weights))

	// Confirm proposals length increments.
	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	proposalID := bigInt(0).Sub(proposalsLength, bigInt(1))

	// Construct Proposal binding.
	proposalAddress, err := s.manager.Proposals(nil, proposalID)
	s.Require().NoError(err)
	proposal, err := abi.NewWeightProposal(proposalAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[proposalAddress] = proposal

	// Get Proposal Basket.
	proposalBasketAddress, err := proposal.Basket(nil)
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
	s.requireTxWithEvents(s.manager.AcceptProposal(signer(s.operator), proposalID))(
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

func (s *ManagerSuite) changeBasketUsingSwapProposal(tokens []common.Address, amounts []*big.Int, toVault []bool) {
	// Propose the new basket.
	s.requireTx(s.manager.ProposeSwap(signer(s.proposer), tokens, amounts, toVault))

	// Confirm proposals length increments.
	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	proposalID := bigInt(0).Sub(proposalsLength, bigInt(1))

	// Construct Proposal binding.
	proposalAddress, err := s.manager.Proposals(nil, proposalID)
	s.Require().NoError(err)
	proposal, err := abi.NewSwapProposal(proposalAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[proposalAddress] = proposal

	// Accept the Proposal.
	s.requireTxWithEvents(s.manager.AcceptProposal(signer(s.operator), proposalID))(
		abi.ManagerProposalAccepted{
			Id: proposalID, Proposer: s.proposer.address(),
		},
	)

	// Confirm we cannot execute the proposal yet.
	s.requireTxFails(s.manager.ExecuteProposal(signer(s.operator), proposalID))

	// Advance 24h.
	s.Require().NoError(s.node.(backend).AdjustTime(24 * time.Hour))

	// DEBUGGING
	oldBasketAddress, err := s.manager.Basket(nil)
	s.Require().NoError(err)
	oldBasket, err := abi.NewBasket(oldBasketAddress, s.node)
	s.Require().NoError(err)
	for i, erc20Address := range s.erc20Addresses {
		weight, err := oldBasket.Weights(nil, erc20Address)
		s.Require().NoError(err)
		fmt.Println(weight)

		amount, err := proposal.Amounts(nil, bigInt(uint32(i)))
		s.Require().NoError(err)
		fmt.Println(amount)
	}

	// Execute Proposal.
	s.requireTx(s.manager.ExecuteProposal(signer(s.operator), proposalID))

	// Gets the current basket and makes sure it is correct.
	// s.assertBasket(basket, tokens, weights)

	// Assert that the vault is still collateralized.
	s.assertManagerCollateralized()
}

func (s *ManagerSuite) newWeights(
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
