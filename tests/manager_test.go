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

func TestManager(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

type ManagerSuite struct {
	TestSuite
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
	s.owner = s.account[0].address()

	// Re-deploy Reserve and store a handle to the Go binding and the contract address.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		reserveAddress: reserve,
	}

	s.requireTxStrongly(tx, err)(abi.ReserveOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner,
	})
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = s.reserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	// Vault
	vaultAddress, tx, vault, err := abi.DeployVault(s.signer, s.node)

	s.logParsers[vaultAddress] = vault
	s.requireTxStrongly(tx, err)(abi.VaultOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner,
	})
	s.vault = vault
	s.vaultAddress = vaultAddress

	// Manager
	managerAddress, tx, manager, err := abi.DeployManager(
		s.signer, s.node, vaultAddress, reserveAddress, bigInt(0),
	)

	s.logParsers[managerAddress] = manager
	s.requireTxStrongly(tx, err)(abi.ManagerOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner,
	})
	s.manager = manager
	s.managerAddress = managerAddress

	// Set all auths to Manager
	s.requireTxStrongly(s.reserve.ChangeMinter(s.signer, managerAddress))(
		abi.ReserveMinterChanged{NewMinter: managerAddress},
	)
	s.requireTxStrongly(s.reserve.ChangePauser(s.signer, managerAddress))(
		abi.ReservePauserChanged{NewPauser: managerAddress},
	)
	s.requireTxStrongly(s.reserve.ChangeFreezer(s.signer, managerAddress))(
		abi.ReserveFreezerChanged{NewFreezer: managerAddress},
	)
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

	whitelisted, err := s.manager.Whitelist(nil, s.owner)
	s.Require().NoError(err)
	s.Equal(true, whitelisted)

	seigniorage, err := s.manager.Seigniorage(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), seigniorage.String())

	paused, err := s.manager.Paused(nil)
	s.Require().NoError(err)
	s.Equal(true, paused)

	useWhitelist, err := s.manager.UseWhitelist(nil)
	s.Require().NoError(err)
	s.Equal(true, useWhitelist)
}

func (s *ManagerSuite) TestProposeNewBasket() {
	tokens := []common.Address{s.account[5].address()}
	backing := []*big.Int{bigInt(1000)}
	s.requireTxWeakly(s.manager.ProposeNewBasket(signer(s.account[1]), tokens, backing))(
		abi.ManagerNewBasketProposalCreated{
			Id: bigInt(0), Proposer: s.account[1].address(), Tokens: tokens, Backing: backing,
		},
		abi.ProposalOwnershipTransferred{PreviousOwner: zeroAddress(), NewOwner: s.managerAddress},
	)

	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	s.Equal(proposalsLength, bigInt(1))

	proposalAddress, err := s.manager.Proposals(nil, bigInt(0))
	s.Require().NoError(err)

	proposal, err := abi.NewProposal(proposalAddress, s.node)
	s.Require().NoError(err)

	id, err := proposal.Id(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), id.String())

	proposer, err := proposal.Proposer(nil)
	s.Require().NoError(err)
	s.Equal(s.account[1].address(), proposer)

	token, err := proposal.Tokens(nil, bigInt(0))
	s.Require().NoError(err)
	s.Equal(tokens[0], token)

	token, err = proposal.Tokens(nil, bigInt(1))
	s.Require().Error(err)

	status, err := proposal.GetStatus(nil)
	s.Require().NoError(err)
	s.Equal(uint8(0), status) // Statuses.Created should have value 0

	proposalBasketAddress, err := proposal.Basket(nil)
	s.Require().NoError(err)
	s.NotEqual(proposalBasketAddress, zeroAddress())

	// basket, err := abi.NewBasket(basketAddress, s.node)
	// s.Require().NoError(err)

}

// func (s *ManagerSuite) TestProposeQuantitiesAdjustment() {
// 	in := []*big.Int{bigInt(1), bigInt(2)}
// 	out := []*big.Int{bigInt(2), bigInt(1)}
// 	s.requireTx(s.manager.ProposeQuantitiesAdjustment(signer(s.account[1]), in, out))(
// 		abi.ProposalProposalCreated{Id: bigInt(0), Proposer: s.account[1].address()},
// 	)
// }
