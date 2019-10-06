// +build all

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

func TestVault(t *testing.T) {
	suite.Run(t, new(VaultSuite))
}

type VaultSuite struct {
	TestSuite
}

var (
	// Compile-time check that VaultSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &VaultSuite{}
	_ suite.SetupAllSuite    = &VaultSuite{}
	_ suite.TearDownAllSuite = &VaultSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *VaultSuite) SetupSuite() {
	s.setup()
}

// TearDownSuite runs once, after all of the tests in the suite.
func (s *VaultSuite) TearDownSuite() {
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
func (s *VaultSuite) BeforeTest(suiteName, testName string) {
	s.owner = s.account[0]

	// Vault
	vaultAddress, tx, vault, err := abi.DeployVault(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		vaultAddress: vault,
	}
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

	// Deploy collateral ERC20s
	s.erc20s = make([]*abi.BasicERC20, 3)
	s.erc20Addresses = make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)
		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
		s.logParsers[erc20Address] = erc20

		val := bigInt(1000)
		s.requireTxWithStrictEvents(erc20.Transfer(s.signer, vaultAddress, val))(
			abi.BasicERC20Transfer{
				From: s.owner.address(), To: vaultAddress, Value: val,
			},
		)
	}
}

func (s *VaultSuite) TestDeploy() {}

// TestConstructor tests that the constructor sets initial state appropriately.
func (s *VaultSuite) TestConstructor() {
	// Initial owner should be deployer.
	ownerAddress, err := s.vault.Owner(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), ownerAddress)

	// Initial manager should be deployer.
	managerAddress, err := s.vault.Manager(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), managerAddress)
}

// TestChangeManager unit tests the changeManager function.
func (s *VaultSuite) TestChangeManager() {
	// Change the Manager address.
	manager := s.account[1]
	s.requireTxWithStrictEvents(s.vault.ChangeManager(s.signer, manager.address()))(
		abi.VaultManagerTransferred{
			PreviousManager: s.owner.address(), NewManager: manager.address(),
		},
	)

	// Confirm the Manager address is changed.
	managerAddress, err := s.vault.Manager(nil)
	s.Require().NoError(err)
	s.Equal(manager.address(), managerAddress)
}

// TestChangeManagerRequires tests that the requires in `changeManager` work.
func (s *VaultSuite) TestChangeManagerRequires() {
	// Try changing manager to the zero address.
	s.requireTxFails(s.vault.ChangeManager(s.signer, zeroAddress()))
}

// TestChangeManagerProtected makes sure changeManager is protected.
func (s *VaultSuite) TestChangeManagerProtected() {
	manager := s.account[1]
	// Try to change the Manager as someone other than owner.
	s.requireTxFails(s.vault.ChangeManager(signer(manager), manager.address()))

	// Change the Manager address as owner.
	s.requireTxWithStrictEvents(s.vault.ChangeManager(s.signer, manager.address()))(
		abi.VaultManagerTransferred{
			PreviousManager: s.owner.address(), NewManager: manager.address(),
		},
	)

	// Confirm the Manager address is changed.
	managerAddress, err := s.vault.Manager(nil)
	s.Require().NoError(err)
	s.Equal(manager.address(), managerAddress)

	// Make sure it's still the case only the owner can change the manager.
	receiver := s.account[2]
	s.requireTxFails(s.vault.ChangeManager(signer(manager), receiver.address()))
	s.requireTxFails(s.vault.ChangeManager(signer(s.account[2]), receiver.address()))
	s.requireTxWithStrictEvents(s.vault.ChangeManager(s.signer, receiver.address()))(
		abi.VaultManagerTransferred{
			PreviousManager: manager.address(), NewManager: receiver.address(),
		},
	)
}

// TestWithdrawTo unit tests the withdrawTo function.
func (s *VaultSuite) TestWithdrawTo() {
	receiver := s.account[2]

	for i, erc20 := range s.erc20s {
		balance, err := erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)

		val := bigInt(uint32(i + 1))
		expected := balance.Sub(balance, val)

		// Make transfer.
		s.requireTxWithStrictEvents(
			s.vault.WithdrawTo(s.signer, s.erc20Addresses[i], val, receiver.address()),
		)(
			abi.BasicERC20Transfer{
				From: s.vaultAddress, To: receiver.address(), Value: val,
			},
			abi.VaultWithdrawal{
				Token: s.erc20Addresses[i], Amount: val, To: receiver.address(),
			},
		)

		// Check that resultant balance is as expected.
		balance, err = erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)
		s.Equal(expected, balance)
	}
}

// TestWithdrawToVoidWithdrawal makes sure we perform the void withdrawal.
func (s *VaultSuite) TestWithdrawToVoidWithdrawal() {
	receiver := s.account[2]

	// Do this for a few ERC20s
	balance, err := s.erc20s[0].BalanceOf(nil, s.vaultAddress)
	s.Require().NoError(err)

	val := bigInt(0)
	expected := balance

	// Make transfer.
	s.requireTxWithStrictEvents(
		s.vault.WithdrawTo(s.signer, s.erc20Addresses[0], val, receiver.address()),
	)(
		abi.BasicERC20Transfer{
			From: s.vaultAddress, To: receiver.address(), Value: val,
		},
		abi.VaultWithdrawal{
			Token: s.erc20Addresses[0], Amount: val, To: receiver.address(),
		},
	)

	// Check that resultant balance is as expected.
	balance, err = s.erc20s[0].BalanceOf(nil, s.vaultAddress)
	s.Require().NoError(err)
	s.Equal(expected, balance)
}

// TestWithdrawToProtected makes sure withdrawTo is protected.
func (s *VaultSuite) TestWithdrawToProtected() {
	manager := s.account[1]
	receiver := s.account[2]
	val := bigInt(1)

	// Set the manager.
	s.requireTxWithStrictEvents(s.vault.ChangeManager(s.signer, manager.address()))(
		abi.VaultManagerTransferred{
			PreviousManager: s.owner.address(), NewManager: manager.address(),
		},
	)

	// Confirm manager can transfer.
	s.requireTxWithStrictEvents(
		s.vault.WithdrawTo(signer(manager), s.erc20Addresses[0], val, receiver.address()),
	)(
		abi.BasicERC20Transfer{
			From: s.vaultAddress, To: receiver.address(), Value: val,
		},
		abi.VaultWithdrawal{
			Token: s.erc20Addresses[0], Amount: val, To: receiver.address(),
		},
	)

	// Confirm owner cannot transfer
	s.requireTxFails(
		s.vault.WithdrawTo(s.signer, s.erc20Addresses[0], val, receiver.address()),
	)

	// Confirm random cannot transfer
	s.requireTxFails(
		s.vault.WithdrawTo(signer(receiver), s.erc20Addresses[0], val, receiver.address()),
	)
}

// TestUpgrade tests that we can upgrade to a new Vault and successfully pass ownership
// back to the original Manager, while maintaining Vault collateral.
func (s *VaultSuite) TestUpgrade() {
	newKey := s.account[3]

	// Set up a Basket.
	weights := []*big.Int{shiftLeft(1, 36), shiftLeft(2, 36), shiftLeft(3, 36)}
	basketAddress, tx, _, err := abi.DeployBasket(
		s.signer, s.node, zeroAddress(), s.erc20Addresses, weights,
	)
	s.requireTxWithStrictEvents(tx, err)
	s.NotEqual(zeroAddress(), basketAddress)

	// Manager.
	managerAddress, tx, manager, err := abi.DeployManager(
		s.signer, s.node,
		s.vaultAddress, s.account[2].address(), s.account[2].address(),
		basketAddress, s.account[2].address(), bigInt(0),
	)

	s.logParsers[managerAddress] = manager
	s.requireTx(tx, err)(abi.ManagerOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.owner.address(),
	})

	// Deploy the new vault.
	newVaultAddress, tx, newVault, err := abi.DeployVaultV2(signer(newKey), s.node)
	s.logParsers[newVaultAddress] = newVault
	s.requireTx(tx, err)(
		abi.VaultV2OwnershipTransferred{PreviousOwner: zeroAddress(), NewOwner: newKey.address()},
	)
	s.requireTxWithStrictEvents(newVault.ChangeManager(signer(newKey), managerAddress))(
		abi.VaultV2ManagerTransferred{PreviousManager: newKey.address(), NewManager: managerAddress},
	)

	// Switch over.
	s.requireTxWithStrictEvents(s.vault.NominateNewOwner(s.signer, newVaultAddress))(
		abi.VaultNewOwnerNominated{PreviousOwner: s.owner.address(), Nominee: newVaultAddress},
	)
	s.requireTxWithStrictEvents(manager.NominateNewOwner(s.signer, newVaultAddress))(
		abi.ManagerNewOwnerNominated{PreviousOwner: s.owner.address(), Nominee: newVaultAddress},
	)
	s.requireTx(newVault.CompleteHandoff(signer(newKey), s.vaultAddress, managerAddress))(
		abi.ManagerVaultChanged{OldVaultAddr: s.vaultAddress, NewVaultAddr: newVaultAddress},
		abi.VaultOwnershipTransferred{PreviousOwner: s.owner.address(), NewOwner: newVaultAddress},
		abi.ManagerOwnershipTransferred{PreviousOwner: s.owner.address(), NewOwner: newVaultAddress},
		abi.ManagerNewOwnerNominated{PreviousOwner: newVaultAddress, Nominee: newKey.address()},
		abi.VaultNewOwnerNominated{PreviousOwner: newVaultAddress, Nominee: newKey.address()},
	)

	// Grab ownership back for the Manager.
	s.requireTxWithStrictEvents(manager.AcceptOwnership(signer(newKey)))(
		abi.ManagerOwnershipTransferred{
			PreviousOwner: newVaultAddress,
			NewOwner:      newKey.address(),
		},
	)

	newManagerOwner, err := manager.Owner(nil)
	s.Require().NoError(err)
	s.Equal(newKey.address(), newManagerOwner)

	// Grab ownership back for the old Vault.
	s.requireTxWithStrictEvents(s.vault.AcceptOwnership(signer(newKey)))(
		abi.VaultOwnershipTransferred{
			PreviousOwner: newVaultAddress,
			NewOwner:      newKey.address(),
		},
	)

	newVaultOwner, err := s.vault.Owner(nil)
	s.Require().NoError(err)
	s.Equal(newKey.address(), newVaultOwner)

	// Assert balances in new vault are same as what was passed into original vault in `BeforeTest`.
	for _, erc20 := range s.erc20s {
		bal, err := erc20.BalanceOf(nil, newVaultAddress)
		s.Require().NoError(err)
		s.Equal(bigInt(1000), bal)
	}

}
