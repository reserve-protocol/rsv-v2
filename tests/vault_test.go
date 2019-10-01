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

// TestBatchWithdrawToVoidWithdrawal makes sure we perform the void withdrawal.
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
