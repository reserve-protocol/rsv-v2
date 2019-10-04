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
	weights  []*big.Int
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

	s.requireTx(tx, err)
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Unpause Reserve.
	s.requireTxWithStrictEvents(s.reserve.Unpause(s.signer))(
		abi.ReserveUnpaused{Account: s.owner.address()},
	)

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = s.reserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	// Accept ownership of eternal storage.
	s.requireTxWithStrictEvents(s.eternalStorage.AcceptOwnership(s.signer))(
		abi.ReserveEternalStorageOwnershipTransferred{
			PreviousOwner: s.reserveAddress, NewOwner: s.account[0].address(),
		},
	)

	// Vault.
	vaultAddress, tx, vault, err := abi.DeployVault(s.signer, s.node)

	s.logParsers[vaultAddress] = vault
	s.requireTxWithStrictEvents(tx, err)(
		abi.VaultOwnershipTransferred{
			PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
		},
		abi.VaultManagerTransferred{
			PreviousManager: zeroAddress(), NewManager: s.owner.address(),
		},
	)
	s.vault = vault
	s.vaultAddress = vaultAddress

	// ProposalFactory.
	propFactoryAddress, tx, propFactory, err := abi.DeployProposalFactory(s.signer, s.node)
	s.logParsers[propFactoryAddress] = propFactory
	s.requireTx(tx, err)

	s.proposalFactory = propFactory
	s.proposalFactoryAddress = propFactoryAddress

	// Manager.
	managerAddress, tx, manager, err := abi.DeployManager(
		s.signer, s.node, vaultAddress, reserveAddress, propFactoryAddress, s.operator.address(), bigInt(0),
	)

	s.logParsers[managerAddress] = manager
	s.requireTx(tx, err)(abi.ManagerOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})
	s.manager = manager
	s.managerAddress = managerAddress

	// Confirm we start in emergency state.
	emergency, err := s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Equal(true, emergency)

	// Unpause from emergency.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, false))(
		abi.ManagerEmergencyChanged{OldVal: true, NewVal: false},
	)

	// Confirm we are unpaused from emergency.
	emergency, err = s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Equal(false, emergency)

	// Add the proposer to the list of approved issuers.
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), true))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: false, NewVal: true},
	)

	// Set all auths to Manager.
	s.requireTxWithStrictEvents(s.reserve.ChangeMinter(s.signer, managerAddress))(
		abi.ReserveMinterChanged{NewMinter: managerAddress},
	)
	s.requireTxWithStrictEvents(s.reserve.ChangePauser(s.signer, managerAddress))(
		abi.ReservePauserChanged{NewPauser: managerAddress},
	)
	s.requireTxWithStrictEvents(s.vault.ChangeManager(s.signer, managerAddress))(
		abi.VaultManagerTransferred{PreviousManager: s.owner.address(), NewManager: managerAddress},
	)

	// Set the basket.
	basketAddress, err := s.manager.TrustedBasket(nil)
	s.Require().NoError(err)
	s.NotEqual(zeroAddress(), basketAddress)

	basket, err := abi.NewBasket(basketAddress, s.node)
	s.Require().NoError(err)

	s.basketAddress = basketAddress
	s.basket = basket

	// Deploy collateral ERC20s.
	s.erc20s = make([]*abi.BasicERC20, 3)
	s.erc20Addresses = make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)

		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
		s.logParsers[erc20Address] = erc20
	}

	// Fund and set allowances.
	amounts := []*big.Int{shiftLeft(1, 46), shiftLeft(1, 46), shiftLeft(1, 46)}
	s.fundAccountWithErc20sAndApprove(s.proposer, amounts)

	// Pass a WeightProposal so we are able to Issue/Redeem.
	s.weights = []*big.Int{shiftLeft(1, 35), shiftLeft(3, 35), shiftLeft(6, 35)}
	s.changeBasketUsingWeightProposal(s.erc20Addresses, s.weights)
}

func (s *ManagerSuite) TestDeploy() {}

// TestConstructor tests that the constructor sets initial state appropriately.
func (s *ManagerSuite) TestConstructor() {
	vaultAddr, err := s.manager.TrustedVault(nil)
	s.Require().NoError(err)
	s.Equal(s.vaultAddress, vaultAddr)

	rsvAddr, err := s.manager.TrustedRSV(nil)
	s.Require().NoError(err)
	s.Equal(s.reserveAddress, rsvAddr)

	proposalFactory, err := s.manager.TrustedProposalFactory(nil)
	s.Require().NoError(err)
	s.Equal(s.proposalFactoryAddress, proposalFactory)

	operator, err := s.manager.Operator(nil)
	s.Require().NoError(err)
	s.Equal(s.operator.address(), operator)

	seigniorage, err := s.manager.Seigniorage(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), seigniorage.String())

	// `emergency` is tested in `BeforeTest`

	filterOn, err := s.manager.FilterIssuers(nil)
	s.Require().NoError(err)
	s.Equal(true, filterOn)

	isIssuer, err := s.manager.Issuers(nil, s.owner.address())
	s.Require().NoError(err)
	s.Equal(true, isIssuer)
}

// TestSetFilterIssuers tests that `setFilterIssuers` changes `filterIssuers`
func (s *ManagerSuite) TestSetFilterIssuers() {
	// Confirm it begins true
	filter, err := s.manager.FilterIssuers(nil)
	s.Require().NoError(err)
	s.Equal(true, filter)

	// Take the proposer off the list.
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), false))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: true, NewVal: false},
	)

	// Turn filter off.
	s.requireTxWithStrictEvents(s.manager.SetFilterIssuers(s.signer, false))(
		abi.ManagerFilterIssuersChanged{OldVal: true, NewVal: false},
	)

	// Confirm filter is off.
	filter, err = s.manager.FilterIssuers(nil)
	s.Require().NoError(err)
	s.Equal(false, filter)

	// Confirm the proposer can issue.
	s.requireTx(s.manager.Issue(signer(s.proposer), bigInt(1)))

	// Turn it back on.
	s.requireTxWithStrictEvents(s.manager.SetFilterIssuers(s.signer, true))(
		abi.ManagerFilterIssuersChanged{OldVal: false, NewVal: true},
	)

	// Confirm filter is on.
	filter, err = s.manager.FilterIssuers(nil)
	s.Require().NoError(err)
	s.Equal(true, filter)

	// Confirm the proposer cannot issue.
	s.requireTxFails(s.manager.Issue(signer(s.proposer), bigInt(1)))
}

// TestSetFilterIssuanceIsProtected tests that `setFilterIssuers` can only be called by owner.
func (s *ManagerSuite) TestSetFilterIssuanceIsProtected() {
	s.requireTxFails(s.manager.SetFilterIssuers(signer(s.account[2]), true))
	s.requireTxFails(s.manager.SetFilterIssuers(signer(s.operator), true))
}

// TestSetIssuerStatus tests that `setIssuerStatus` changes the list of issuers.
func (s *ManagerSuite) TestSetIssuerStatus() {
	// Confirm it begins true for the proposer.
	isIssuer, err := s.manager.Issuers(nil, s.proposer.address())
	s.Require().NoError(err)
	s.Equal(true, isIssuer)

	// Take proposer off the list
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), false))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: true, NewVal: false},
	)

	// Confirm they are off the list.
	isIssuer, err = s.manager.Issuers(nil, s.proposer.address())
	s.Require().NoError(err)
	s.Equal(false, isIssuer)

	// Put them back on the list.
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), true))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: false, NewVal: true},
	)

	// Confirm they are on the list.
	isIssuer, err = s.manager.Issuers(nil, s.proposer.address())
	s.Require().NoError(err)
	s.Equal(true, isIssuer)
}

// TestSetIssuerStatusIsProtected tests that `setFilterIssuers` can only be called by owner.
func (s *ManagerSuite) TestSetIssuerStatusIsProtected() {
	s.requireTxFails(s.manager.SetIssuerStatus(signer(s.account[2]), s.proposer.address(), true))
	s.requireTxFails(s.manager.SetIssuerStatus(signer(s.operator), s.proposer.address(), true))
}

// TestSetIssuancePaused tests that `setIssuancePaused` changes the state as expected.
func (s *ManagerSuite) TestSetIssuancePaused() {
	// Confirm Issuance is Unpaused.
	paused, err := s.manager.IssuancePaused(nil)
	s.Require().NoError(err)
	s.Equal(false, paused)

	// Pause.
	s.requireTxWithStrictEvents(s.manager.SetIssuancePaused(s.signer, true))(
		abi.ManagerIssuancePausedChanged{OldVal: false, NewVal: true},
	)

	// Confirm Issuance is Paused.
	paused, err = s.manager.IssuancePaused(nil)
	s.Require().NoError(err)
	s.Equal(true, paused)

	// Unpause.
	s.requireTxWithStrictEvents(s.manager.SetIssuancePaused(s.signer, false))(
		abi.ManagerIssuancePausedChanged{OldVal: true, NewVal: false},
	)

	// Confirm Issuance is Unpaused.
	paused, err = s.manager.IssuancePaused(nil)
	s.Require().NoError(err)
	s.Equal(false, paused)
}

// TestSetIssuancePausedIsProtected tests that `setIssuancePaused` can only be called by owner.
func (s *ManagerSuite) TestSetIssuancePausedIsProtected() {
	s.requireTxFails(s.manager.SetIssuancePaused(signer(s.account[2]), true))
	s.requireTxFails(s.manager.SetIssuancePaused(signer(s.operator), true))
}

// TestSetEmergency tests that `setEmergency` changes the state as expected.
func (s *ManagerSuite) TestSetEmergency() {
	// Confirm we being not in an emergency.
	emergency, err := s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Equal(false, emergency)

	// Pause for emergency.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, true))(
		abi.ManagerEmergencyChanged{OldVal: false, NewVal: true},
	)

	// Confirm we are in an emergency.
	emergency, err = s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Equal(true, emergency)

	// Unpause for emergency.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, false))(
		abi.ManagerEmergencyChanged{OldVal: true, NewVal: false},
	)

	// Confirm we are not in an emergency.
	emergency, err = s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Equal(false, emergency)
}

// TestSetEmergencyIsProtected tests that `setEmergency` can only be called by owner.
func (s *ManagerSuite) TestSetEmergencyIsProtected() {
	s.requireTxFails(s.manager.SetEmergency(signer(s.account[2]), true))
	s.requireTxFails(s.manager.SetEmergency(signer(s.operator), true))
}

// TestSetOperator tests that `setOperator` manipulates state correctly.
func (s *ManagerSuite) TestSetOperator() {
	newOperator := s.account[5].address()
	s.requireTxWithStrictEvents(s.manager.SetOperator(s.signer, newOperator))(
		abi.ManagerOperatorChanged{
			OldAccount: s.operator.address(), NewAccount: newOperator,
		},
	)

	// Check that state is correct.
	foundOperator, err := s.manager.Operator(nil)
	s.Require().NoError(err)
	s.Equal(newOperator, foundOperator)
}

// TestSetOperatorIsProtected tests that `setOperator` can only be called by owner.
func (s *ManagerSuite) TestSetOperatorIsProtected() {
	s.requireTxFails(s.manager.SetOperator(signer(s.account[2]), s.account[5].address()))
	s.requireTxFails(s.manager.SetOperator(signer(s.operator), s.account[5].address()))
}

// TestSetSeigniorage tests that `setSeigniorage` manipulates state correctly.
func (s *ManagerSuite) TestSetSeigniorage() {
	seigniorage := bigInt(1)
	s.requireTxWithStrictEvents(s.manager.SetSeigniorage(s.signer, seigniorage))(
		abi.ManagerSeigniorageChanged{
			OldVal: bigInt(0), NewVal: seigniorage,
		},
	)

	// Check that state is correct.
	foundSeigniorage, err := s.manager.Seigniorage(nil)
	s.Require().NoError(err)
	s.Equal(seigniorage.String(), foundSeigniorage.String())
}

// TestSetSeigniorageIsProtected tests that `setSeigniorage` can only be called by owner.
func (s *ManagerSuite) TestSetSeigniorageIsProtected() {
	seigniorage := bigInt(1)
	s.requireTxFails(s.manager.SetSeigniorage(signer(s.account[2]), seigniorage))
	s.requireTxFails(s.manager.SetSeigniorage(signer(s.operator), seigniorage))
}

// TestSetDelay tests that `setDelay` manipulates state correctly.
func (s *ManagerSuite) TestSetDelay() {
	delay := bigInt(172800) // 48 hours
	s.requireTxWithStrictEvents(s.manager.SetDelay(s.signer, delay))(
		abi.ManagerDelayChanged{
			OldVal: bigInt(86400), NewVal: delay,
		},
	)

	// Check that state is correct.
	foundDelay, err := s.manager.Delay(nil)
	s.Require().NoError(err)
	s.Equal(delay.String(), foundDelay.String())
}

// TestSetDelayIsProtected tests that `setDelay` can only be called by owner.
func (s *ManagerSuite) TestSetDelayIsProtected() {
	delay := bigInt(1)
	s.requireTxFails(s.manager.SetDelay(signer(s.account[2]), delay))
	s.requireTxFails(s.manager.SetDelay(signer(s.operator), delay))
}

// TestClearProposals tests that `clearProposals` manipulates state correctly.
func (s *ManagerSuite) TestClearProposals() {
	// ProposalsLength should start at 1.
	proposalsLength, err := s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(1).String(), proposalsLength.String())

	// Clear it.
	s.requireTxWithStrictEvents(s.manager.ClearProposals(s.signer))(
		abi.ManagerProposalsCleared{},
	)

	// Check that the length is now 0.
	proposalsLength, err = s.manager.ProposalsLength(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(0).String(), proposalsLength.String())
}

// TestClearProposalsIsProtected tests that `clearProposals` can only be called by owner.
func (s *ManagerSuite) TestClearProposalsIsProtected() {
	s.requireTxFails(s.manager.ClearProposals(signer(s.account[2])))
	s.requireTxFails(s.manager.ClearProposals(signer(s.operator)))
}

// TestIssue tests that `issue` costs the correct amounts given basket + seigniorage.
func (s *ManagerSuite) TestIssue() {
	buyer := s.account[4]

	//First set seigniorage, in BPS
	seigniorage := bigInt(10) // 0.1%
	s.requireTxWithStrictEvents(s.manager.SetSeigniorage(s.signer, seigniorage))(
		abi.ManagerSeigniorageChanged{
			OldVal: bigInt(0), NewVal: seigniorage,
		},
	)

	rsvAmount := shiftLeft(1, 27) // 1 billion
	expectedAmounts := s.computeExpectedIssueAmounts(seigniorage, rsvAmount)
	s.fundAccountWithErc20sAndApprove(buyer, expectedAmounts)

	// Add buyer to `issuers`
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, buyer.address(), true))(
		abi.ManagerIssuersChanged{Issuer: buyer.address(), OldVal: false, NewVal: true},
	)

	// Issue.
	s.requireTx(s.manager.Issue(signer(buyer), rsvAmount))

	// Expect RSV balance.
	balance, err := s.reserve.BalanceOf(nil, buyer.address())
	s.Require().NoError(err)
	s.Equal(rsvAmount.String(), balance.String())

	for i, erc20 := range s.erc20s {
		// Expect no leftover tokens.
		balance, err = erc20.BalanceOf(nil, buyer.address())
		s.Require().NoError(err)
		s.Equal(bigInt(0).String(), balance.String())

		// Expect tokens are all in the vault.
		balance, err = erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)
		s.Equal(expectedAmounts[i].String(), balance.String())
	}

	s.assertManagerCollateralized()
}

// TestIssueIsProtected tests that `issue` reverts when in an emergency or it is paused.
func (s *ManagerSuite) TestIssueIsProtected() {
	amount := bigInt(1)

	// We should be able to issue initially.
	s.requireTx(s.manager.Issue(signer(s.proposer), amount))

	// Set `emergency` to true.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, true))(
		abi.ManagerEmergencyChanged{OldVal: false, NewVal: true},
	)

	// Confirm `emergency` is true.
	emergency, err := s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Equal(true, emergency)

	// Issue should fail.
	s.requireTxFails(s.manager.Issue(signer(s.proposer), amount))

	// Set `emergency` to false.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, false))(
		abi.ManagerEmergencyChanged{OldVal: true, NewVal: false},
	)

	// Now we should be able to issue.
	s.requireTx(s.manager.Issue(signer(s.proposer), amount))

	// Pause just issuance.
	s.requireTxWithStrictEvents(s.manager.SetIssuancePaused(s.signer, true))(
		abi.ManagerIssuancePausedChanged{OldVal: false, NewVal: true},
	)

	// Confirm we are Paused.
	paused, err := s.manager.IssuancePaused(nil)
	s.Require().NoError(err)
	s.Equal(true, paused)

	// Issue should fail now.
	s.requireTxFails(s.manager.Issue(signer(s.proposer), amount))

	// Unpause issuance.
	s.requireTxWithStrictEvents(s.manager.SetIssuancePaused(s.signer, false))(
		abi.ManagerIssuancePausedChanged{OldVal: true, NewVal: false},
	)

	// Now we should be able to issue.
	s.requireTx(s.manager.Issue(signer(s.proposer), amount))

	// Take proposer off the list
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), false))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: true, NewVal: false},
	)

	// Issue should fail.
	s.requireTxFails(s.manager.Issue(signer(s.proposer), amount))

	// Turn off the filter.
	s.requireTxWithStrictEvents(s.manager.SetFilterIssuers(s.signer, false))(
		abi.ManagerFilterIssuersChanged{OldVal: true, NewVal: false},
	)

	// Should be able to issue.
	s.requireTx(s.manager.Issue(signer(s.proposer), bigInt(1)))

	// Turn the filter back on.
	s.requireTxWithStrictEvents(s.manager.SetFilterIssuers(s.signer, true))(
		abi.ManagerFilterIssuersChanged{OldVal: false, NewVal: true},
	)

	// Issue should fail.
	s.requireTxFails(s.manager.Issue(signer(s.proposer), amount))

	// Put them back on the list.
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), true))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: false, NewVal: true},
	)

	// Issue should succeed
	s.requireTx(s.manager.Issue(signer(s.proposer), amount))

}

// TestIssueRequireStatements tests that `issue` reverts when Paused.
func (s *ManagerSuite) TestIssueRequireStatements() {
	amount := bigInt(1)

	// Issue should succeed first.
	s.requireTx(s.manager.Issue(signer(s.proposer), amount))
	s.assertManagerCollateralized()

	// Issue should fail now.
	s.requireTxFails(s.manager.Issue(signer(s.proposer), bigInt(0)))
	s.assertManagerCollateralized()
}

// TestRedeem tests that `redeem` compensates the person with the correct amounts.
func (s *ManagerSuite) TestRedeem() {
	// Issue.
	rsvAmount := shiftLeft(1, 27) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvAmount))

	redeemer := s.account[4]

	// Add redeemer to the list of approved issuers.
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, redeemer.address(), true))(
		abi.ManagerIssuersChanged{Issuer: redeemer.address(), OldVal: false, NewVal: true},
	)

	// Send the RSV to someone else who doesn't have any Erc20s.
	s.requireTx(s.reserve.Transfer(signer(s.proposer), redeemer.address(), rsvAmount))

	// Redeem that RSV.
	s.requireTx(s.reserve.Approve(signer(redeemer), s.managerAddress, rsvAmount))
	s.requireTx(s.manager.Redeem(signer(redeemer), rsvAmount))

	// Figure out what to expect back.
	amounts := s.computeExpectedRedeemAmounts(rsvAmount)

	// Assert our balances are what we expected.
	for i, erc20 := range s.erc20s {
		// Expect no leftover tokens.
		balance, err := erc20.BalanceOf(nil, redeemer.address())
		s.Require().NoError(err)
		s.Equal(amounts[i].String(), balance.String())
	}

	s.assertManagerCollateralized()
}

// TestRedeemIsProtected tests that `redeem` compensates the person with the correct amounts.
func (s *ManagerSuite) TestRedeemIsProtected() {
	// Issue.
	rsvAmount := shiftLeft(1, 27) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvAmount))

	// Make sure we have the balance we expect to have.
	rsvBalance, err := s.reserve.BalanceOf(nil, s.proposer.address())
	s.Require().NoError(err)
	s.Equal(rsvAmount.String(), rsvBalance.String())

	// Approve the manager to spend our RSV.
	s.requireTx(s.reserve.Approve(signer(s.proposer), s.managerAddress, rsvAmount))(
		abi.ReserveApproval{
			Owner:   s.proposer.address(),
			Spender: s.managerAddress,
			Value:   rsvAmount,
		},
	)

	// Redeem a tiny amount first to make sure it works.
	s.requireTx(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Emergency Pause.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, true))(
		abi.ManagerEmergencyChanged{OldVal: false, NewVal: true},
	)

	// Confirm the same redemption now fails.
	s.requireTxFails(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Unpause from emergency.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, false))(
		abi.ManagerEmergencyChanged{OldVal: true, NewVal: false},
	)

	// Should be able to Redeem.
	s.requireTx(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Take proposer off the list
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), false))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: true, NewVal: false},
	)

	// Redeem should fail.
	s.requireTxFails(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Turn off the filter.
	s.requireTxWithStrictEvents(s.manager.SetFilterIssuers(s.signer, false))(
		abi.ManagerFilterIssuersChanged{OldVal: true, NewVal: false},
	)

	// Should be able to Redeem.
	s.requireTx(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Turn the filter back on.
	s.requireTxWithStrictEvents(s.manager.SetFilterIssuers(s.signer, true))(
		abi.ManagerFilterIssuersChanged{OldVal: false, NewVal: true},
	)

	// Redeem should fail.
	s.requireTxFails(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Put them back on the list.
	s.requireTxWithStrictEvents(s.manager.SetIssuerStatus(s.signer, s.proposer.address(), true))(
		abi.ManagerIssuersChanged{Issuer: s.proposer.address(), OldVal: false, NewVal: true},
	)

	// Redeem should succeed
	s.requireTx(s.manager.Redeem(signer(s.proposer), bigInt(1)))
}

// TestRedeemRequireStatements tests that `redeem` reverts for 0 RSV.
func (s *ManagerSuite) TestRedeemRequireStatements() {
	// Issue.
	rsvAmount := shiftLeft(1, 27) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvAmount))

	// Make sure we have the balance we expect to have.
	rsvBalance, err := s.reserve.BalanceOf(nil, s.proposer.address())
	s.Require().NoError(err)
	s.Equal(rsvAmount.String(), rsvBalance.String())

	// Approve the manager to spend our RSV.
	s.requireTx(s.reserve.Approve(signer(s.proposer), s.managerAddress, rsvAmount))(
		abi.ReserveApproval{
			Owner:   s.proposer.address(),
			Spender: s.managerAddress,
			Value:   rsvAmount,
		},
	)

	// Redeem a tiny amount first to make sure it works.
	s.requireTx(s.manager.Redeem(signer(s.proposer), bigInt(1)))

	// Confirm redeeming for 0 fails.
	s.requireTxFails(s.manager.Redeem(signer(s.proposer), bigInt(0)))

	s.assertManagerCollateralized()
}

// TestProposeWeightsUseCase sets a basket, issues RSV, changes the basket, and redeems RSV.
func (s *ManagerSuite) TestProposeWeightsFullUsecase() {
	// Issue a billion RSV.
	rsvToIssue := shiftLeft(1, 27) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// Change to a new basket.
	newWeights := []*big.Int{shiftLeft(2, 48), shiftLeft(3, 48), shiftLeft(1, 48)}
	s.changeBasketUsingWeightProposal(s.erc20Addresses, newWeights)

	// Approve the manager to spend a billion RSV.
	s.requireTx(s.reserve.Approve(signer(s.proposer), s.managerAddress, rsvToIssue))(
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
	// Issue a billion RSV.
	rsvToIssue := shiftLeft(1, 27) // 1 billion
	s.requireTx(s.manager.Issue(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// Change to a new basket using a SwapProposal
	amounts := []*big.Int{shiftLeft(2, 18), shiftLeft(3, 18), shiftLeft(1, 18)}
	toVault := []bool{true, false, true}
	s.changeBasketUsingSwapProposal(s.erc20Addresses, amounts, toVault)

	// Approve the manager to spend a billion RSV.
	s.requireTx(s.reserve.Approve(signer(s.proposer), s.managerAddress, rsvToIssue))(
		abi.ReserveApproval{Owner: s.proposer.address(), Spender: s.managerAddress, Value: rsvToIssue},
	)

	// Redeem a billion RSV.
	s.requireTx(s.manager.Redeem(signer(s.proposer), rsvToIssue))
	s.assertManagerCollateralized()

	// We should be back to zero RSV supply.
	s.assertRSVTotalSupply(bigInt(0))
}

// ===================================== Helpers ===========================================

func (s *ManagerSuite) changeBasketUsingWeightProposal(tokens []common.Address, weights []*big.Int) {
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

func (s *ManagerSuite) changeBasketUsingSwapProposal(tokens []common.Address, amounts []*big.Int, toVault []bool) {
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

func (s *ManagerSuite) computeExpectedIssueAmounts(
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

func (s *ManagerSuite) computeExpectedRedeemAmounts(rsvSupply *big.Int) []*big.Int {
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

func (s *ManagerSuite) fundAccountWithErc20sAndApprove(acc account, amounts []*big.Int) {
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
