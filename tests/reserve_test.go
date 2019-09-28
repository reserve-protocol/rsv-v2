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

func TestReserve(t *testing.T) {
	suite.Run(t, new(ReserveSuite))
}

type ReserveSuite struct {
	TestSuite
}

var (
	// Compile-time check that ReserveSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &ReserveSuite{}
	_ suite.SetupAllSuite    = &ReserveSuite{}
	_ suite.TearDownAllSuite = &ReserveSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *ReserveSuite) SetupSuite() {
	s.setup()
}

// TearDownSuite runs once, after all of the tests in the suite.
func (s *ReserveSuite) TearDownSuite() {
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
func (s *ReserveSuite) BeforeTest(suiteName, testName string) {
	// Re-deploy Reserve and store a handle to the Go binding and the contract address.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		reserveAddress: reserve,
	}

	s.requireTxWithEvents(tx, err)(abi.ReserveOwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.account[0].address(),
	})
	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = s.reserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	deployerAddress := s.account[0].address()

	// Make the deployment account a minter, pauser, and freezer.
	s.requireTxWithEvents(s.reserve.ChangeMinter(s.signer, deployerAddress))(
		abi.ReserveMinterChanged{NewMinter: deployerAddress},
	)
	s.requireTxWithEvents(s.reserve.ChangePauser(s.signer, deployerAddress))(
		abi.ReservePauserChanged{NewPauser: deployerAddress},
	)
}

func (s *ReserveSuite) TestDeploy() {}

func (s *ReserveSuite) TestBalanceOf() {
	s.assertRSVBalance(zeroAddress(), bigInt(0))
}

func (s *ReserveSuite) TestName() {
	name, err := s.reserve.Name(nil)
	s.NoError(err)
	s.Equal("Reserve", name)
}

func (s *ReserveSuite) TestSymbol() {
	symbol, err := s.reserve.Symbol(nil)
	s.NoError(err)
	s.Equal("RSV", symbol)
}

func (s *ReserveSuite) TestDecimals() {
	decimals, err := s.reserve.Decimals(nil)
	s.NoError(err)
	s.Equal(uint8(18), decimals)
}

func (s *ReserveSuite) TestAllowsMinting() {
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to recipient.
	s.requireTxWithEvents(s.reserve.Mint(s.signer, recipient, amount))(
		mintingTransfer(recipient, amount),
	)

	// Check that balances are as expected.
	s.assertRSVBalance(s.account[0].address(), bigInt(0))
	s.assertRSVBalance(recipient, amount)
	s.assertRSVTotalSupply(amount)
}

func (s *ReserveSuite) TestTransfer() {
	sender := s.account[1]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to sender.
	s.requireTxWithEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)

	// Transfer from sender to recipient.
	s.requireTxWithEvents(s.reserve.Transfer(signer(sender), recipient, amount))(
		abi.ReserveTransfer{
			From:  sender.address(),
			To:    recipient,
			Value: amount,
		},
	)

	// Check that balances are as expected.
	s.assertRSVBalance(sender.address(), bigInt(0))
	s.assertRSVBalance(recipient, amount)
	s.assertRSVBalance(s.account[0].address(), bigInt(0))
	s.assertRSVTotalSupply(amount)
}

func (s *ReserveSuite) TestTransferExceedsFunds() {
	sender := s.account[1]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)
	smallAmount := bigInt(10) // must be smaller than amount

	// Mint smallAmount to sender.
	s.requireTxWithEvents(s.reserve.Mint(s.signer, sender.address(), smallAmount))(
		mintingTransfer(sender.address(), smallAmount),
	)

	// Transfer from sender to recipient should fail.
	s.requireTxFails(s.reserve.Transfer(signer(sender), recipient, amount))

	// Balances should be as we expect.
	s.assertRSVBalance(sender.address(), smallAmount)
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.account[0].address(), bigInt(0))
	s.assertRSVTotalSupply(smallAmount)
}

// As long as Minting cannot overflow a uint256, then `transferFrom` cannot overflow.
func (s *ReserveSuite) TestMintWouldOverflow() {
	interestingRecipients := []common.Address{
		common.BigToAddress(bigInt(1)),
		common.BigToAddress(bigInt(255)),
		common.BigToAddress(bigInt(256)),
		common.BigToAddress(bigInt(256)),
		common.BigToAddress(maxUint160()),
		common.BigToAddress(minInt160AsUint160()),
	}
	for _, recipient := range interestingRecipients {
		smallAmount := bigInt(10) // must be smaller than amount
		overflowCausingAmount := maxUint256()
		overflowCausingAmount = overflowCausingAmount.Sub(overflowCausingAmount, bigInt(8))

		// Mint smallAmount to recipient.
		s.requireTxWithEvents(s.reserve.Mint(s.signer, recipient, smallAmount))(
			mintingTransfer(recipient, smallAmount),
		)

		// Mint a quantity large enough to cause overflow in totalSupply i.e.
		// `10 + (uint256::MAX - 8) > uint256::MAX`
		s.requireTxFails(s.reserve.Mint(s.signer, recipient, overflowCausingAmount))
	}
}

func (s *ReserveSuite) TestApprove() {
	owner := s.account[1]
	spender := s.account[2]
	amount := bigInt(53)

	// Owner approves spender.
	s.requireTxWithEvents(s.reserve.Approve(signer(owner), spender.address(), amount))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: amount},
	)

	// Approval should be reflected in allowance.
	s.assertRSVAllowance(owner.address(), spender.address(), amount)

	// Shouldn't be symmetric.
	s.assertRSVAllowance(spender.address(), owner.address(), bigInt(0))

	// Balances shouldn't change.
	s.assertRSVBalance(owner.address(), bigInt(0))
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestIncreaseAllowance() {
	owner := s.account[1]
	spender := s.account[2]
	amount := bigInt(2000)

	// Owner approves spender through increaseAllowance.
	s.requireTxWithEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), amount))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: amount},
	)

	// Approval should be reflected in allowance.
	s.assertRSVAllowance(owner.address(), spender.address(), amount)

	// Shouldn't be symmetric.
	s.assertRSVAllowance(spender.address(), owner.address(), bigInt(0))

	// Balances shouldn't change.
	s.assertRSVBalance(owner.address(), bigInt(0))
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestIncreaseAllowanceWouldOverflow() {
	owner := s.account[1]
	spender := s.account[2]
	initialAmount := bigInt(10)

	// Owner approves spender for initial amount.
	s.requireTxWithEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), initialAmount))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: initialAmount},
	)

	// Owner should not be able to increase approval high enough to overflow a uint256.
	s.requireTxFails(s.reserve.IncreaseAllowance(signer(owner), spender.address(), maxUint256()))
}

func (s *ReserveSuite) TestDecreaseAllowance() {
	owner := s.account[1]
	spender := s.account[2]
	initialAmount := bigInt(10)
	decrease := bigInt(6)
	final := bigInt(4)

	// Owner approves spender for initial amount.
	s.requireTxWithEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), initialAmount))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: initialAmount},
	)

	// Owner decreases allowance.
	s.requireTxWithEvents(s.reserve.DecreaseAllowance(signer(owner), spender.address(), decrease))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: final},
	)

	// Allowance should be as we expect.
	s.assertRSVAllowance(owner.address(), spender.address(), final)

	// Balances shouldn't change.
	s.assertRSVBalance(owner.address(), bigInt(0))
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestDecreaseAllowanceUnderflow() {
	owner := s.account[1]
	spender := s.account[2]
	initialAmount := bigInt(10)
	decrease := bigInt(11)

	// Owner approves spender for initial amount.
	s.requireTxWithEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), initialAmount))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: initialAmount},
	)

	// Owner decreases allowance fails because of underflow.
	s.requireTxFails(s.reserve.DecreaseAllowance(signer(owner), spender.address(), decrease))

	// Allowance should be as we expect.
	s.assertRSVAllowance(owner.address(), spender.address(), initialAmount)

	// Balances shouldn't change.
	s.assertRSVBalance(owner.address(), bigInt(0))
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))

	// Allowances shouldn't change
	s.assertRSVAllowance(owner.address(), spender.address(), initialAmount)
}

func (s *ReserveSuite) TestPausing() {
	banker := s.account[1]
	amount := bigInt(1000)
	approveAmount := bigInt(1)
	recipient := s.account[2]
	spender := s.account[3]

	// Give banker funds. Minting is allowed while unpaused.
	s.requireTxWithEvents(s.reserve.Mint(s.signer, banker.address(), amount))(
		mintingTransfer(banker.address(), amount),
	)
	s.assertRSVBalance(banker.address(), amount)

	// Approve spender to spend bankers funds.
	s.requireTxWithEvents(s.reserve.Approve(signer(banker), spender.address(), approveAmount))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: approveAmount},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// Pause.
	s.requireTxWithEvents(s.reserve.Pause(s.signer))(
		abi.ReservePaused{Account: s.account[0].address()},
	)

	// Minting is not allowed while paused.
	s.requireTxFails(s.reserve.Mint(s.signer, recipient.address(), amount))

	// Transfers from are not allowed while paused.
	s.requireTxFails(s.reserve.TransferFrom(s.signer, banker.address(), recipient.address(), amount))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVBalance(banker.address(), amount)

	// Transfers are not allowed while paused.
	s.requireTxFails(s.reserve.Transfer(signer(banker), recipient.address(), amount))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVBalance(banker.address(), amount)

	// Approving is not allowed while paused.
	s.requireTxFails(s.reserve.Approve(signer(banker), spender.address(), amount))
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// IncreaseAllowance is not allowed while paused.
	s.requireTxFails(s.reserve.IncreaseAllowance(signer(banker), spender.address(), amount))
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// DecreaseAllowance is not allowed while paused.
	s.requireTxFails(s.reserve.DecreaseAllowance(signer(banker), spender.address(), amount))
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// Unpause.
	s.requireTxWithEvents(s.reserve.Unpause(s.signer))(
		abi.ReserveUnpaused{Account: s.account[0].address()},
	)

	// Transfers are allowed while unpaused.
	s.requireTxWithEvents(s.reserve.Transfer(signer(banker), recipient.address(), amount))(
		abi.ReserveTransfer{From: banker.address(), To: recipient.address(), Value: amount},
	)
	s.assertRSVBalance(recipient.address(), amount)

	// Approving is allowed while unpaused.
	s.requireTxWithEvents(s.reserve.Approve(signer(banker), spender.address(), bigInt(2)))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: bigInt(2)},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), bigInt(2))

	// DecreaseAllowance is allowed while unpaused.
	s.requireTxWithEvents(s.reserve.DecreaseAllowance(signer(banker), spender.address(), approveAmount))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: bigInt(1)},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// IncreaseAllowance is allowed while unpaused.
	s.requireTxWithEvents(s.reserve.IncreaseAllowance(signer(banker), spender.address(), approveAmount))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: bigInt(2)},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), bigInt(2))
}

func (s *ReserveSuite) TestMintingBurningChain() {
	deployerAddress := s.account[0].address()
	// Mint to recipient.
	recipient := s.account[1]
	amount := bigInt(100)

	s.requireTxWithEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)

	// Approve signer for burning.
	s.requireTxWithEvents(s.reserve.Approve(signer(recipient), deployerAddress, amount))(
		abi.ReserveApproval{Owner: recipient.address(), Spender: deployerAddress, Value: amount},
	)

	// Burn from recipient.
	s.requireTxWithEvents(s.reserve.BurnFrom(s.signer, recipient.address(), amount))(
		abi.ReserveTransfer{From: recipient.address(), To: zeroAddress(), Value: amount},
		abi.ReserveApproval{Owner: recipient.address(), Spender: deployerAddress, Value: bigInt(0)},
	)

	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestMintingTransferBurningChain() {
	deployerAddress := s.account[0].address()
	recipient := s.account[1]
	amount := bigInt(100)

	// Mint to recipient.
	s.requireTxWithEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)

	// Transfer to target.
	target := s.account[2]
	s.requireTxWithEvents(s.reserve.Transfer(signer(recipient), target.address(), amount))(
		abi.ReserveTransfer{From: recipient.address(), To: target.address(), Value: amount},
	)

	s.assertRSVBalance(target.address(), amount)
	s.assertRSVBalance(recipient.address(), bigInt(0))

	// Approve signer for burning.
	s.requireTxWithEvents(s.reserve.Approve(signer(target), s.account[0].address(), amount))(
		abi.ReserveApproval{Owner: target.address(), Spender: s.account[0].address(), Value: amount},
	)

	// Burn from target.
	s.requireTxWithEvents(s.reserve.BurnFrom(s.signer, target.address(), amount))(
		abi.ReserveTransfer{From: target.address(), To: zeroAddress(), Value: amount},
		abi.ReserveApproval{Owner: target.address(), Spender: deployerAddress, Value: bigInt(0)},
	)

	s.assertRSVBalance(target.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestBurnFromWouldUnderflow() {
	deployerAddress := s.account[0].address()
	// Mint to recipient.
	recipient := s.account[1]
	amount := bigInt(100)
	causesUnderflowAmount := bigInt(101)

	s.assertRSVTotalSupply(bigInt(0))
	s.requireTxWithEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)

	// Approve signer for burning.
	s.requireTxWithEvents(s.reserve.Approve(signer(recipient), deployerAddress, amount))(
		abi.ReserveApproval{Owner: recipient.address(), Spender: deployerAddress, Value: amount},
	)

	// Burn from recipient.
	s.requireTxFails(s.reserve.BurnFrom(s.signer, recipient.address(), causesUnderflowAmount))

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)
}

func (s *ReserveSuite) TestTransferFrom() {
	sender := s.account[1]
	middleman := s.account[2]
	recipient := s.account[3]

	amount := bigInt(1)
	s.requireTxWithEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)
	s.assertRSVBalance(sender.address(), amount)
	s.assertRSVBalance(middleman.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(amount)

	// Approve middleman to transfer funds from the sender.
	s.requireTxWithEvents(s.reserve.Approve(signer(sender), middleman.address(), amount))(
		abi.ReserveApproval{Owner: sender.address(), Spender: middleman.address(), Value: amount},
	)
	s.assertRSVAllowance(sender.address(), middleman.address(), amount)

	// transferFrom allows the msg.sender to send an existing approval to an arbitrary destination.
	s.requireTxWithEvents(s.reserve.TransferFrom(signer(middleman), sender.address(), recipient.address(), amount))(
		abi.ReserveTransfer{From: sender.address(), To: recipient.address(), Value: amount},
		abi.ReserveApproval{Owner: sender.address(), Spender: middleman.address(), Value: bigInt(0)},
	)
	s.assertRSVBalance(sender.address(), bigInt(0))
	s.assertRSVBalance(middleman.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), amount)

	// Allowance should have been decreased by the transfer
	s.assertRSVAllowance(sender.address(), middleman.address(), bigInt(0))
	// transfers should not change totalSupply.
	s.assertRSVTotalSupply(amount)
}

func (s *ReserveSuite) TestTransferFromWouldUnderflow() {
	sender := s.account[1]
	middleman := s.account[2]
	recipient := s.account[3]

	approveAmount := bigInt(2)
	s.requireTxWithEvents(s.reserve.Mint(s.signer, sender.address(), approveAmount))(
		mintingTransfer(sender.address(), approveAmount),
	)
	s.assertRSVBalance(sender.address(), approveAmount)
	s.assertRSVBalance(middleman.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(approveAmount)

	// Approve middleman to transfer funds from the sender.
	s.requireTxWithEvents(s.reserve.Approve(signer(sender), middleman.address(), approveAmount))(
		abi.ReserveApproval{Owner: sender.address(), Spender: middleman.address(), Value: approveAmount},
	)
	s.assertRSVAllowance(sender.address(), middleman.address(), approveAmount)

	// now reduce the approveAmount in the sender's account to less than the approval for the middleman
	s.requireTxWithEvents(s.reserve.Transfer(signer(sender), recipient.address(), bigInt(1)))(
		abi.ReserveTransfer{From: sender.address(), To: recipient.address(), Value: bigInt(1)},
	)

	// Attempt to transfer more funds than the sender's current balance, but
	// passing the approval checks. Should fail when subtracting value from
	// sender's current balance.
	s.requireTxFails(s.reserve.TransferFrom(signer(middleman), sender.address(), recipient.address(), approveAmount))

	s.assertRSVBalance(sender.address(), bigInt(1))
	s.assertRSVBalance(middleman.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(1))

	// Allowance should not have been changed
	s.assertRSVAllowance(sender.address(), middleman.address(), approveAmount)
	// should not change totalSupply.
	s.assertRSVTotalSupply(approveAmount)
}

///////////////////////

func (s *ReserveSuite) TestPauseFailsForNonPauser() {
	s.requireTxFails(s.reserve.Pause(signer(s.account[2])))
}

func (s *ReserveSuite) TestUnpauseFailsForNonPauser() {
	deployerAddress := s.account[0].address()
	s.requireTxWithEvents(s.reserve.Pause(s.signer))(
		abi.ReservePaused{Account: deployerAddress},
	)
	s.requireTxFails(s.reserve.Unpause(signer(s.account[1])))
}

func (s *ReserveSuite) TestChangePauserFailsForNonPauser() {
	s.requireTxFails(s.reserve.ChangePauser(signer(s.account[2]), s.account[1].address()))
}

///////////////////////

func (s *ReserveSuite) TestMintFailsForNonMinter() {
	recipient := common.BigToAddress(bigInt(1))
	s.requireTxFails(s.reserve.Mint(signer(s.account[2]), recipient, bigInt(7)))
}

func (s *ReserveSuite) TestChangeMinterFailsForNonMinter() {
	s.requireTxFails(s.reserve.ChangeMinter(signer(s.account[2]), s.account[1].address()))
}

///////////////////////

func (s *ReserveSuite) TestUpgrade() {
	recipient := s.account[1]
	amount := big.NewInt(100)

	// Mint to recipient.
	s.requireTxWithEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	// Deploy new contract.
	newKey := s.account[2]
	newTokenAddress, tx, newToken, err := abi.DeployReserveV2(signer(newKey), s.node)
	s.logParsers[newTokenAddress] = newToken
	s.requireTxWithEvents(tx, err)(abi.ReserveV2OwnershipTransferred{
		PreviousOwner: zeroAddress(), NewOwner: s.account[2].address(),
	})

	// Make the switch.
	s.requireTxWithEvents(s.reserve.NominateNewOwner(s.signer, newTokenAddress))(abi.ReserveNewOwnerNominated{
		PreviousOwner: s.account[0].address(), NewOwner: newTokenAddress,
	})
	s.requireTxWithEvents(newToken.CompleteHandoff(signer(newKey), s.reserveAddress)) /*
		not asserting events because there are a lot and we don't care much about them
	*/

	// Old token should not be functional.
	s.requireTxFails(s.reserve.Mint(s.signer, recipient.address(), big.NewInt(1500)))
	s.requireTxFails(s.reserve.Transfer(signer(recipient), s.account[3].address(), big.NewInt(10)))
	s.requireTxFails(s.reserve.Pause(s.signer))
	s.requireTxFails(s.reserve.Unpause(s.signer))

	// assertion function for new token
	assertRSVBalance := func(address common.Address, amount *big.Int) {
		balance, err := newToken.BalanceOf(nil, address)
		s.NoError(err)
		s.Equal(amount.String(), balance.String()) // assert.Equal can mis-compare big.Ints, so compare strings instead
	}

	// New token should be functional.
	assertRSVBalance(recipient.address(), amount)
	s.requireTxWithEvents(newToken.ChangeMinter(signer(newKey), newKey.address()))(
		abi.ReserveV2MinterChanged{NewMinter: newKey.address()},
	)
	s.requireTxWithEvents(newToken.ChangePauser(signer(newKey), newKey.address()))(
		abi.ReserveV2PauserChanged{NewPauser: newKey.address()},
	)
	s.requireTxWithEvents(newToken.Mint(signer(newKey), recipient.address(), big.NewInt(1500)))(
		abi.ReserveV2Transfer{From: zeroAddress(), To: recipient.address(), Value: bigInt(1500)},
	)
	s.requireTxWithEvents(newToken.Transfer(signer(recipient), s.account[3].address(), big.NewInt(10)))(
		abi.ReserveV2Transfer{From: recipient.address(), To: s.account[3].address(), Value: bigInt(10)},
	)
	s.requireTxWithEvents(newToken.Pause(signer(newKey)))(
		abi.ReserveV2Paused{Account: newKey.address()},
	)
	s.requireTxWithEvents(newToken.Unpause(signer(newKey)))(
		abi.ReserveV2Unpaused{Account: newKey.address()},
	)
	assertRSVBalance(recipient.address(), big.NewInt(100+1500-10))
	assertRSVBalance(s.account[3].address(), big.NewInt(10))
}

// Test that we can use the escape hatch in ReserveEternalStorage.
func (s *ReserveSuite) TestEternalStorageEscapeHatch() {
	assertOwner := func(expected common.Address) {
		owner, err := s.eternalStorage.Owner(nil)
		s.NoError(err)
		s.Equal(expected, owner)
	}

	assertEscapeHatch := func(expected common.Address) {
		escapeHatch, err := s.eternalStorage.EscapeHatch(nil)
		s.NoError(err)
		s.Equal(expected, escapeHatch)
	}

	// Check that owner and escapeHatch are initialized in the way we expect.
	assertOwner(s.reserveAddress)
	assertEscapeHatch(s.account[0].address())

	newEscapeHatch := s.account[3]

	// Change escapeHatch address and check it is what we expect.
	s.requireTxWithEvents(s.eternalStorage.TransferEscapeHatch(s.signer, newEscapeHatch.address()))(
		abi.ReserveEternalStorageEscapeHatchTransferred{
			OldEscapeHatch: s.account[0].address(),
			NewEscapeHatch: newEscapeHatch.address(),
		},
	)

	// Check that escapeHatch changed and owner didn't.
	assertOwner(s.reserveAddress)
	assertEscapeHatch(newEscapeHatch.address())

	newOwner := s.account[4]

	// Change owner as escapeHatch account.
	s.requireTxWithEvents(s.eternalStorage.TransferOwnership(signer(newEscapeHatch), newOwner.address()))(
		abi.ReserveEternalStorageOwnershipTransferred{
			OldOwner: s.reserveAddress,
			NewOwner: newOwner.address(),
		},
	)

	// Check that owner changed and escapeHatch didn't.
	assertOwner(newOwner.address())
	assertEscapeHatch(newEscapeHatch.address())

	// Check that owner cannot change escapeHatch.
	s.requireTxFails(s.eternalStorage.TransferEscapeHatch(signer(newOwner), s.account[5].address()))

	// Check that escapeHatch can make the change the owner could not.
	s.requireTxWithEvents(s.eternalStorage.TransferEscapeHatch(signer(newEscapeHatch), s.account[5].address()))(
		abi.ReserveEternalStorageEscapeHatchTransferred{
			OldEscapeHatch: newEscapeHatch.address(),
			NewEscapeHatch: s.account[5].address(),
		},
	)
}

// Test that setBalance works as expected on ReserveEternalStorage.
// It is not used by the current Reserve contract, but is present as a bit
// of potential future-proofing for upgrades.
func (s *ReserveSuite) TestEternalStorageSetBalance() {
	newOwner := s.account[1]
	amount := bigInt(1300)

	// Check that we can't call setBalance before becoming the owner.
	s.requireTxFails(s.eternalStorage.SetBalance(signer(newOwner), newOwner.address(), amount))

	// Transfer ownership of Eternal Storage to external account.
	s.requireTxWithEvents(s.eternalStorage.TransferOwnership(s.signer, newOwner.address()))(
		abi.ReserveEternalStorageOwnershipTransferred{
			OldOwner: s.reserveAddress,
			NewOwner: newOwner.address(),
		},
	)

	// Check that we can now call setBalance.
	s.requireTxWithEvents(s.eternalStorage.SetBalance(signer(newOwner), newOwner.address(), amount))( /* assert zero events */ )

	// Balance should have changed.
	balance, err := s.eternalStorage.Balance(nil, newOwner.address())
	s.NoError(err)
	s.Equal(amount.String(), balance.String())
}
