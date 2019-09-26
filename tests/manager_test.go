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
	if coverageEnabled {
		s.createSlowCoverageNode()
	} else {
		s.createFastNode()
	}
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

	// Re-deploy Reserve and store a handle to the Go binding and the contract address.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		reserveAddress: reserve,
	}

	s.requireTxStrongly(tx, err)(abi.ReserveOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = s.reserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	// Vault.
	vaultAddress, tx, vault, err := abi.DeployVault(s.signer, s.node)

	s.logParsers[vaultAddress] = vault
	s.requireTxStrongly(tx, err)(
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
	s.requireTxStrongly(tx, err)(abi.ManagerOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})
	s.manager = manager
	s.managerAddress = managerAddress

	// Set all auths to Manager.
	s.requireTxStrongly(s.reserve.ChangeMinter(s.signer, managerAddress))(
		abi.ReserveMinterChanged{NewMinter: managerAddress},
	)
	s.requireTxStrongly(s.reserve.ChangePauser(s.signer, managerAddress))(
		abi.ReservePauserChanged{NewPauser: managerAddress},
	)

	// Set the operator.
	s.requireTxStrongly(s.manager.SetOperator(s.signer, s.operator.address()))(
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

	// Deploy collateral ERC20s
	s.erc20s = make([]*abi.BasicERC20, 3)
	s.erc20Addresses = make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)
		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
		s.logParsers[erc20Address] = erc20
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

func (s *ManagerSuite) TestProposeWeightsHappyPath() {
	tokens := s.erc20Addresses
	weights := generateBackings(len(tokens))
	proposer := s.account[2]

	// Propose the basket.
	s.requireTxWeakly(s.manager.ProposeWeights(signer(proposer), tokens, weights))(
		abi.ManagerWeightsProposed{
			Id: bigInt(0), Proposer: proposer.address(), Tokens: tokens, Weights: weights,
		},
		abi.ProposalOwnershipTransferred{PreviousOwner: zeroAddress(), NewOwner: s.managerAddress},
	)

	// Confirm proposals length increments.
	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	s.Equal(proposalsLength, bigInt(1))

	// Construct Proposal binding.
	proposalAddress, err := s.manager.Proposals(nil, bigInt(0))
	s.Require().NoError(err)
	proposal, err := abi.NewWeightProposal(proposalAddress, s.node)
	s.Require().NoError(err)

	// Check Proposal has correct fields
	foundProposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(proposer.address(), foundProposer)

	state, err := proposal.State(nil)
	s.Require().NoError(err)
	s.Equal(uint8(0), state) // State.Created should have value 0

	proposalBasketAddress, err := proposal.Basket(nil)
	s.Require().NoError(err)
	s.NotEqual(zeroAddress(), proposalBasketAddress)

	basket, err := abi.NewBasket(proposalBasketAddress, s.node)
	s.Require().NoError(err)

	// Check Basket has correct fields
	basketTokens, err := basket.GetTokens(nil)
	s.Require().NoError(err)
	s.True(reflect.DeepEqual(basketTokens, tokens))

	basketSize, err := basket.Size(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(uint32(len(tokens))).String(), basketSize.String())

	for i := 0; i < len(weights); i++ {
		foundBacking, err := basket.Weights(nil, tokens[i])
		s.Require().NoError(err)
		s.Equal(weights[i], foundBacking)
	}

	// Accept the Proposal.
	s.requireTxStrongly(s.manager.AcceptProposal(signer(s.operator), bigInt(0)))(
		abi.ManagerProposalAccepted{
			Id: bigInt(0), Proposer: proposer.address(),
		},
	)

	// Advance 24h.
	s.Require().NoError(s.node.(backend).AdjustTime(24 * time.Hour))

	// Execute Proposal.
	s.requireTxStrongly(s.manager.ExecuteProposal(signer(proposer), bigInt(0)))(
		abi.ManagerProposalExecuted{
			Id:        bigInt(0),
			Proposer:  proposer.address(),
			Executor:  proposer.address(),
			OldBasket: s.basketAddress,
			NewBasket: proposalBasketAddress,
		},
	)

	// Gets the current basket and makes sure it is the same as `tokens` + `weights`
	s.assertCurrentBasketMirrorsTargets(tokens, weights)

	// Are we collateralized?
	s.assertManagerCollateralized()

	// Now we should be able to unpause.
	s.requireTxStrongly(s.manager.Unpause(s.signer))(
		abi.ManagerUnpaused{
			Account: s.owner.address(),
		},
	)

	// // Fund proposer account with ERC20s.
	// s.fundAccountWithERC20sAndApprove(proposer, toAtto(1000))

}

func (s *ManagerSuite) TestProposeQuantitiesAdjustment() {
	// in := []*big.Int{bigInt(1), bigInt(2)}
	// out := []*big.Int{bigInt(2), bigInt(1)}
	// s.requireTx(s.manager.ProposeQuantitiesAdjustment(signer(s.account[1]), in, out))(
	// 	abi.ProposalProposalCreated{Id: bigInt(0), Proposer: s.account[1].address()},
	// )
}

func (s *ManagerSuite) TestPauseAuth() {
	s.requireTxFails(s.manager.Pause(signer(s.account[2])))
	s.requireTxFails(s.manager.Pause(signer(s.operator)))
	s.requireTxStrongly(s.manager.Pause(s.signer))(
		abi.ManagerPaused{
			Account: s.owner.address(),
		},
	)
}

// Helpers

func (s *ManagerSuite) fundAccountWithERC20sAndApprove(acc account, val *big.Int) {
	for _, erc20 := range s.erc20s {
		s.requireTxStrongly(erc20.Transfer(s.signer, acc.address(), val))(
			abi.BasicERC20Transfer{
				From: s.owner.address(), To: acc.address(), Value: val,
			},
		)
		s.requireTxStrongly(erc20.Approve(signer(acc), s.managerAddress, val))(
			abi.BasicERC20Approval{
				Owner: acc.address(), Spender: s.managerAddress, Value: val,
			},
		)
	}
}
