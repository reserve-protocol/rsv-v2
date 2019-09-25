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
	if coverageEnabled {
		s.createSlowCoverageNode()
	} else {
		s.createFastNode()
	}
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
		s.requireTxStrongly(erc20.Transfer(s.signer, vaultAddress, val))(
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
	s.requireTxStrongly(s.vault.ChangeManager(s.signer, manager.address()))(
		abi.VaultManagerTransferred{
			PreviousManager: s.owner.address(), NewManager: manager.address(),
		},
	)

	// Confirm the Manager address is changed.
	managerAddress, err := s.vault.Manager(nil)
	s.Require().NoError(err)
	s.Equal(manager.address(), managerAddress)
}

// TestBatchWithdrawTo unit tests the batchWidthdrawTo function.
func (s *VaultSuite) TestBatchWithdrawTo() {
	receiver := s.account[2]
	var initialAmounts []*big.Int
	var withdrawAmounts []*big.Int
	var expectedAmounts []*big.Int
	var eventsToExpect []fmt.Stringer

	// Determine what amounts to transfer and expect.
	for i, erc20 := range s.erc20s {
		balance, err := erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)

		val := bigInt(uint32(i + 1))
		initialAmounts = append(initialAmounts, balance)
		withdrawAmounts = append(withdrawAmounts, val)
		expectedAmounts = append(expectedAmounts, balance.Sub(balance, val))
		eventsToExpect = append(eventsToExpect, abi.BasicERC20Transfer{
			From: s.vaultAddress, To: receiver.address(), Value: val,
		})
	}

	// Transfer amounts.
	eventsToExpect = append(eventsToExpect, abi.VaultBatchWithdrawal{
		Tokens: s.erc20Addresses, Quantities: withdrawAmounts, To: receiver.address(),
	})
	s.requireTxStrongly(
		s.vault.BatchWithdrawTo(s.signer, s.erc20Addresses, withdrawAmounts, receiver.address()))(
		eventsToExpect...,
	)

	// Test expectations.
	for i, erc20 := range s.erc20s {
		balance, err := erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)
		s.Equal(expectedAmounts[i], balance)
	}
}

// TestBatchWithdrawToVoidWithdrawal makes sure we perform the void withdrawal.
func (s *VaultSuite) TestBatchWithdrawToVoidWithdrawal() {
	receiver := s.account[2]
	var initialAmounts []*big.Int
	var withdrawAmounts []*big.Int
	var expectedAmounts []*big.Int

	// Determine what amounts to transfer and expect.
	for _, erc20 := range s.erc20s {
		balance, err := erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)

		val := bigInt(0)
		initialAmounts = append(initialAmounts, balance)
		withdrawAmounts = append(withdrawAmounts, val)
		expectedAmounts = append(expectedAmounts, balance.Sub(balance, val))
	}

	// Perform the void withdrawal.
	s.requireTxStrongly(
		s.vault.BatchWithdrawTo(s.signer, s.erc20Addresses, withdrawAmounts, receiver.address()))(
		abi.VaultBatchWithdrawal{
			Tokens: s.erc20Addresses, Quantities: withdrawAmounts, To: receiver.address(),
		},
	)

	// Check that balances haven't changed.
	for i, erc20 := range s.erc20s {
		balance, err := erc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)
		s.Equal(expectedAmounts[i], balance)
	}
}

// TestBatchWithdrawToBadInput makes sure we perform the void withdrawal.
func (s *VaultSuite) TestBatchWithdrawToBadInput() {
	receiver := s.account[2]
	var withdrawAmounts []*big.Int

	// Determine what amounts to transfer and expect.
	for i, _ := range s.erc20s {
		val := bigInt(uint32(i + 1))
		withdrawAmounts = append(withdrawAmounts, val)
	}
	// Add one more to make it break
	withdrawAmounts = append(withdrawAmounts, bigInt(0))

	// Make sure the withdrawal fails.
	s.requireTxFails(
		s.vault.BatchWithdrawTo(s.signer, s.erc20Addresses, withdrawAmounts, receiver.address()),
	)
}

// TestFunctionsProtected makes sure the endpoint modifiers work as expected.
func (s *VaultSuite) TestFunctionsProtected() {
	// Change the Manager address.
	manager := s.account[1]
	s.requireTxStrongly(s.vault.ChangeManager(s.signer, manager.address()))(
		abi.VaultManagerTransferred{
			PreviousManager: s.owner.address(), NewManager: manager.address(),
		},
	)

	// Confirm the Manager address is changed.
	managerAddress, err := s.vault.Manager(nil)
	s.Require().NoError(err)
	s.Equal(manager.address(), managerAddress)

	// Make sure only the owner can call `changeManager`.
	receiver := s.account[2]
	s.requireTxFails(s.vault.ChangeManager(signer(manager), receiver.address()))
	s.requireTxFails(s.vault.ChangeManager(signer(s.account[2]), receiver.address()))

	// Make sure only the manager can call `batchWithdrawTo`.
	var withdrawAmounts []*big.Int
	s.requireTxFails(
		s.vault.BatchWithdrawTo(s.signer, s.erc20Addresses, withdrawAmounts, receiver.address()),
	)
	s.requireTxFails(
		s.vault.BatchWithdrawTo(signer(receiver), s.erc20Addresses, withdrawAmounts, receiver.address()),
	)

}
