// +build regular

package tests

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
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

// BeforeTest runs before each test in the suite.
func (s *ReserveSuite) BeforeTest(suiteName, testName string) {
	// Re-deploy Reserve and store a handle to the Go binding and the contract address.
	reserveAddress, tx, reserve, err := abi.DeployReserve(s.signer, s.node)

	s.logParsers = map[common.Address]logParser{
		reserveAddress: reserve,
	}

	s.requireTx(tx, err)(
		abi.ReserveOwnershipTransferred{PreviousOwner: zeroAddress(), NewOwner: s.owner.address()},
	)

	// Confirm it begins paused.
	paused, err := reserve.Paused(nil)
	s.Require().NoError(err)
	s.Equal(true, paused)

	// Unpause.
	s.requireTxWithStrictEvents(reserve.Unpause(s.signer))(
		abi.ReserveUnpaused{Account: s.owner.address()},
	)

	s.reserve = reserve
	s.reserveAddress = reserveAddress

	// Get the Go binding and contract address for the new ReserveEternalStorage contract.
	s.eternalStorageAddress, err = s.reserve.GetEternalStorageAddress(nil)
	s.Require().NoError(err)
	s.eternalStorage, err = abi.NewReserveEternalStorage(s.eternalStorageAddress, s.node)
	s.Require().NoError(err)

	s.logParsers[s.eternalStorageAddress] = s.eternalStorage

	// Accept ownership.
	s.requireTxWithStrictEvents(s.eternalStorage.AcceptOwnership(s.signer))(
		abi.ReserveEternalStorageOwnershipTransferred{
			PreviousOwner: s.reserveAddress, NewOwner: s.owner.address(),
		},
	)

	deployerAddress := s.owner.address()

	// Make the deployment account a minter, pauser, and freezer.
	s.requireTxWithStrictEvents(s.reserve.ChangeMinter(s.signer, deployerAddress))(
		abi.ReserveMinterChanged{NewMinter: deployerAddress},
	)
	s.requireTxWithStrictEvents(s.reserve.ChangePauser(s.signer, deployerAddress))(
		abi.ReservePauserChanged{NewPauser: deployerAddress},
	)
	s.requireTxWithStrictEvents(s.reserve.ChangeFeeRecipient(s.signer, deployerAddress))(
		abi.ReserveFeeRecipientChanged{NewFeeRecipient: deployerAddress},
	)
}

func (s *ReserveSuite) TestDeploy() {}

func (s *ReserveSuite) TestConstructor() {
	// `pauser`
	pauser, err := s.reserve.Pauser(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), pauser)

	// `feeRecipient`
	feeRecipient, err := s.reserve.FeeRecipient(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), feeRecipient)

	// `maxSupply`
	maxSupply, err := s.reserve.MaxSupply(nil)
	s.Require().NoError(err)
	s.Equal(maxUint256().String(), maxSupply.String())

	// `paused` is tested by BeforeTest

	// `trustedTxFee`
	trustedTxFee, err := s.reserve.TrustedTxFee(nil)
	s.Require().NoError(err)
	s.Equal(zeroAddress(), trustedTxFee)

	// `trustedData` cannot be read because it is internal
}

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
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, recipient, amount))(
		mintingTransfer(recipient, amount),
	)

	// Check that balances are as expected.
	s.assertRSVBalance(s.owner.address(), bigInt(0))
	s.assertRSVBalance(recipient, amount)
	s.assertRSVTotalSupply(amount)
}

func (s *ReserveSuite) TestChangeMinter() {
	minter, err := s.reserve.Minter(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), minter)

	// Change as owner.
	s.requireTxWithStrictEvents(s.reserve.ChangeMinter(s.signer, s.account[2].address()))(
		abi.ReserveMinterChanged{NewMinter: s.account[2].address()},
	)

	minter, err = s.reserve.Minter(nil)
	s.Require().NoError(err)
	s.Equal(s.account[2].address(), minter)

	// Change as minter.
	s.requireTxWithStrictEvents(s.reserve.ChangeMinter(signer(s.account[2]), s.account[3].address()))(
		abi.ReserveMinterChanged{NewMinter: s.account[3].address()},
	)

	minter, err = s.reserve.Minter(nil)
	s.Require().NoError(err)
	s.Equal(s.account[3].address(), minter)
}

func (s *ReserveSuite) TestChangePauser() {
	pauser, err := s.reserve.Pauser(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), pauser)

	// Change as owner.
	s.requireTxWithStrictEvents(s.reserve.ChangePauser(s.signer, s.account[2].address()))(
		abi.ReservePauserChanged{NewPauser: s.account[2].address()},
	)

	pauser, err = s.reserve.Pauser(nil)
	s.Require().NoError(err)
	s.Equal(s.account[2].address(), pauser)

	// Change as pauser.
	s.requireTxWithStrictEvents(s.reserve.ChangePauser(signer(s.account[2]), s.account[3].address()))(
		abi.ReservePauserChanged{NewPauser: s.account[3].address()},
	)

	pauser, err = s.reserve.Pauser(nil)
	s.Require().NoError(err)
	s.Equal(s.account[3].address(), pauser)
}

func (s *ReserveSuite) TestChangeFeeRecipient() {
	feeRecipient, err := s.reserve.FeeRecipient(nil)
	s.Require().NoError(err)
	s.Equal(s.owner.address(), feeRecipient)

	// Change as owner.
	s.requireTxWithStrictEvents(s.reserve.ChangeFeeRecipient(s.signer, s.account[2].address()))(
		abi.ReserveFeeRecipientChanged{NewFeeRecipient: s.account[2].address()},
	)

	feeRecipient, err = s.reserve.FeeRecipient(nil)
	s.Require().NoError(err)
	s.Equal(s.account[2].address(), feeRecipient)

	// Change as feeRecipient.
	s.requireTxWithStrictEvents(s.reserve.ChangeFeeRecipient(signer(s.account[2]), s.account[3].address()))(
		abi.ReserveFeeRecipientChanged{NewFeeRecipient: s.account[3].address()},
	)

	feeRecipient, err = s.reserve.FeeRecipient(nil)
	s.Require().NoError(err)
	s.Equal(s.account[3].address(), feeRecipient)
}

func (s *ReserveSuite) TestChangeTxFeeHelper() {
	txFee, err := s.reserve.TrustedTxFee(nil)
	s.Require().NoError(err)
	s.Equal(zeroAddress(), txFee)

	// Change as owner.
	s.requireTxWithStrictEvents(s.reserve.ChangeTxFeeHelper(s.signer, s.account[2].address()))(
		abi.ReserveTxFeeHelperChanged{NewTxFeeHelper: s.account[2].address()},
	)

	txFee, err = s.reserve.TrustedTxFee(nil)
	s.Require().NoError(err)
	s.Equal(s.account[2].address(), txFee)
}

func (s *ReserveSuite) TestChangeMaxSupply() {
	maxSupply, err := s.reserve.MaxSupply(nil)
	s.Require().NoError(err)
	s.Equal(maxUint256(), maxSupply)

	amount := bigInt(10)
	// Change as owner.
	s.requireTxWithStrictEvents(s.reserve.ChangeMaxSupply(s.signer, amount))(
		abi.ReserveMaxSupplyChanged{NewMaxSupply: amount},
	)

	maxSupply, err = s.reserve.MaxSupply(nil)
	s.Require().NoError(err)
	s.Equal(amount, maxSupply)
}

func (s *ReserveSuite) TestTransfer() {
	sender := s.account[1]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to sender.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)

	// Transfer from sender to recipient.
	s.requireTxWithStrictEvents(s.reserve.Transfer(signer(sender), recipient, amount))(
		abi.ReserveTransfer{
			From:  sender.address(),
			To:    recipient,
			Value: amount,
		},
	)

	// Check that balances are as expected.
	s.assertRSVBalance(sender.address(), bigInt(0))
	s.assertRSVBalance(recipient, amount)
	s.assertRSVBalance(s.owner.address(), bigInt(0))
	s.assertRSVTotalSupply(amount)
}

func (s *ReserveSuite) TestTransferExceedsFunds() {
	sender := s.account[1]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)
	smallAmount := bigInt(10) // must be smaller than amount

	// Mint smallAmount to sender.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), smallAmount))(
		mintingTransfer(sender.address(), smallAmount),
	)

	// Transfer from sender to recipient should fail.
	s.requireTxFails(s.reserve.Transfer(signer(sender), recipient, amount))

	// Balances should be as we expect.
	s.assertRSVBalance(sender.address(), smallAmount)
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))
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
		s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, recipient, smallAmount))(
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

	// Allowance should start zero.
	s.assertRSVAllowance(owner.address(), spender.address(), bigInt(0))

	// Owner approves spender.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(owner), spender.address(), amount))(
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

	// Allowance should start zero.
	s.assertRSVAllowance(owner.address(), spender.address(), bigInt(0))

	// Owner approves spender through increaseAllowance.
	s.requireTxWithStrictEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), amount))(
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
	s.requireTxWithStrictEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), initialAmount))(
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
	s.requireTxWithStrictEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), initialAmount))(
		abi.ReserveApproval{Owner: owner.address(), Spender: spender.address(), Value: initialAmount},
	)

	// Allowance should be as we expect.
	s.assertRSVAllowance(owner.address(), spender.address(), initialAmount)

	// Owner decreases allowance.
	s.requireTxWithStrictEvents(s.reserve.DecreaseAllowance(signer(owner), spender.address(), decrease))(
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
	s.requireTxWithStrictEvents(s.reserve.IncreaseAllowance(signer(owner), spender.address(), initialAmount))(
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
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, banker.address(), amount))(
		mintingTransfer(banker.address(), amount),
	)
	s.assertRSVBalance(banker.address(), amount)

	// Approve spender to spend bankers funds.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(banker), spender.address(), approveAmount))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: approveAmount},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// Pause.
	s.requireTxWithStrictEvents(s.reserve.Pause(s.signer))(
		abi.ReservePaused{Account: s.owner.address()},
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
	s.requireTxWithStrictEvents(s.reserve.Unpause(s.signer))(
		abi.ReserveUnpaused{Account: s.owner.address()},
	)

	// Transfers are allowed while unpaused.
	s.requireTxWithStrictEvents(s.reserve.Transfer(signer(banker), recipient.address(), amount))(
		abi.ReserveTransfer{From: banker.address(), To: recipient.address(), Value: amount},
	)
	s.assertRSVBalance(recipient.address(), amount)

	// Approving is allowed while unpaused.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(banker), spender.address(), bigInt(2)))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: bigInt(2)},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), bigInt(2))

	// DecreaseAllowance is allowed while unpaused.
	s.requireTxWithStrictEvents(s.reserve.DecreaseAllowance(signer(banker), spender.address(), approveAmount))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: bigInt(1)},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), approveAmount)

	// IncreaseAllowance is allowed while unpaused.
	s.requireTxWithStrictEvents(s.reserve.IncreaseAllowance(signer(banker), spender.address(), approveAmount))(
		abi.ReserveApproval{Owner: banker.address(), Spender: spender.address(), Value: bigInt(2)},
	)
	s.assertRSVAllowance(banker.address(), spender.address(), bigInt(2))
}

func (s *ReserveSuite) TestMintingBurningChain() {
	deployerAddress := s.owner.address()
	// Mint to recipient.
	recipient := s.account[1]
	amount := bigInt(100)

	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)

	// Approve signer for burning.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(recipient), deployerAddress, amount))(
		abi.ReserveApproval{Owner: recipient.address(), Spender: deployerAddress, Value: amount},
	)

	// Burn from recipient.
	s.requireTxWithStrictEvents(s.reserve.BurnFrom(s.signer, recipient.address(), amount))(
		abi.ReserveTransfer{From: recipient.address(), To: zeroAddress(), Value: amount},
		abi.ReserveApproval{Owner: recipient.address(), Spender: deployerAddress, Value: bigInt(0)},
	)

	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestMintingTransferBurningChain() {
	deployerAddress := s.owner.address()
	recipient := s.account[1]
	amount := bigInt(100)

	// Mint to recipient.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)

	// Transfer to target.
	target := s.account[2]
	s.requireTxWithStrictEvents(s.reserve.Transfer(signer(recipient), target.address(), amount))(
		abi.ReserveTransfer{From: recipient.address(), To: target.address(), Value: amount},
	)

	s.assertRSVBalance(target.address(), amount)
	s.assertRSVBalance(recipient.address(), bigInt(0))

	// Approve signer for burning.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(target), s.owner.address(), amount))(
		abi.ReserveApproval{Owner: target.address(), Spender: s.owner.address(), Value: amount},
	)

	// Burn from target.
	s.requireTxWithStrictEvents(s.reserve.BurnFrom(s.signer, target.address(), amount))(
		abi.ReserveTransfer{From: target.address(), To: zeroAddress(), Value: amount},
		abi.ReserveApproval{Owner: target.address(), Spender: deployerAddress, Value: bigInt(0)},
	)

	s.assertRSVBalance(target.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(bigInt(0))
}

func (s *ReserveSuite) TestBurnFromWouldUnderflow() {
	deployerAddress := s.owner.address()
	// Mint to recipient.
	recipient := s.account[1]
	amount := bigInt(100)
	causesUnderflowAmount := bigInt(101)

	s.assertRSVTotalSupply(bigInt(0))
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVTotalSupply(amount)

	// Approve signer for burning.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(recipient), deployerAddress, amount))(
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
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)
	s.assertRSVBalance(sender.address(), amount)
	s.assertRSVBalance(middleman.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(amount)

	// transferFrom fails before approval.
	s.requireTxFails(s.reserve.TransferFrom(signer(middleman), sender.address(), recipient.address(), amount))

	// Approve middleman to transfer funds from the sender.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(sender), middleman.address(), amount))(
		abi.ReserveApproval{Owner: sender.address(), Spender: middleman.address(), Value: amount},
	)
	s.assertRSVAllowance(sender.address(), middleman.address(), amount)

	// transferFrom allows the msg.sender to send an existing approval to an arbitrary destination.
	s.requireTxWithStrictEvents(s.reserve.TransferFrom(signer(middleman), sender.address(), recipient.address(), amount))(
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
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), approveAmount))(
		mintingTransfer(sender.address(), approveAmount),
	)
	s.assertRSVBalance(sender.address(), approveAmount)
	s.assertRSVBalance(middleman.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), bigInt(0))
	s.assertRSVTotalSupply(approveAmount)

	// Approve middleman to transfer funds from the sender.
	s.requireTxWithStrictEvents(s.reserve.Approve(signer(sender), middleman.address(), approveAmount))(
		abi.ReserveApproval{Owner: sender.address(), Spender: middleman.address(), Value: approveAmount},
	)
	s.assertRSVAllowance(sender.address(), middleman.address(), approveAmount)

	// now reduce the approveAmount in the sender's account to less than the approval for the middleman
	s.requireTxWithStrictEvents(s.reserve.Transfer(signer(sender), recipient.address(), bigInt(1)))(
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
	deployerAddress := s.owner.address()
	s.requireTxWithStrictEvents(s.reserve.Pause(s.signer))(
		abi.ReservePaused{Account: deployerAddress},
	)
	s.requireTxFails(s.reserve.Unpause(signer(s.account[1])))
}

func (s *ReserveSuite) TestChangePauserFailsForNonPauser() {
	s.requireTxFails(s.reserve.ChangePauser(signer(s.account[2]), s.account[1].address()))
}

//////////////////////

func (s *ReserveSuite) TestChangeFeeRecipientFailsForNonFeeRecipient() {
	s.requireTxFails(s.reserve.ChangeFeeRecipient(signer(s.account[2]), s.account[1].address()))
}

func (s *ReserveSuite) TestChangeTxFeeHelperFailsForNonOwner() {
	s.requireTxFails(s.reserve.ChangeTxFeeHelper(signer(s.account[2]), s.account[1].address()))
}

func (s *ReserveSuite) TestChangeMaxSupplyFailsForNonOwner() {
	s.requireTxFails(s.reserve.ChangeMaxSupply(signer(s.account[2]), bigInt(1)))
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

func (s *ReserveSuite) TestTransferEternalStorageFailsWhenUnpaused() {
	s.requireTxFails(s.reserve.TransferEternalStorage(signer(s.account[2]), s.account[1].address()))
}

func (s *ReserveSuite) TestTransferEternalStorageFailsForZeroAddress() {
	s.requireTx(s.reserve.Unpause(s.signer))
	s.requireTxFails(s.reserve.TransferEternalStorage(s.signer, zeroAddress()))
}

///////////////////////

func (s *ReserveSuite) TestUpgrade() {
	recipient := s.account[1]
	amount := big.NewInt(100)

	// Mint to recipient.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, recipient.address(), amount))(
		mintingTransfer(recipient.address(), amount),
	)

	// Deploy new contract.
	newKey := s.account[2]
	newTokenAddress, tx, newToken, err := abi.DeployReserveV2(signer(newKey), s.node)
	s.logParsers[newTokenAddress] = newToken
	s.requireTx(tx, err)(
		abi.ReserveV2OwnershipTransferred{PreviousOwner: zeroAddress(), NewOwner: newKey.address()},
	)

	// Make the switch.
	s.requireTxWithStrictEvents(s.reserve.NominateNewOwner(s.signer, newTokenAddress))(abi.ReserveNewOwnerNominated{
		PreviousOwner: s.owner.address(), Nominee: newTokenAddress,
	})
	s.requireTx(newToken.CompleteHandoff(signer(newKey), s.reserveAddress))(
		abi.ReserveEternalStorageTransferred{NewReserveAddress: newTokenAddress},
	)

	// Old token's owner should be the zero address.
	owner, err := s.reserve.Owner(nil)
	s.Require().NoError(err)
	s.Equal(zeroAddress(), owner)

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
	s.requireTxWithStrictEvents(newToken.ChangeMinter(signer(newKey), newKey.address()))(
		abi.ReserveV2MinterChanged{NewMinter: newKey.address()},
	)
	s.requireTxWithStrictEvents(newToken.ChangePauser(signer(newKey), newKey.address()))(
		abi.ReserveV2PauserChanged{NewPauser: newKey.address()},
	)
	s.requireTxWithStrictEvents(newToken.Mint(signer(newKey), recipient.address(), big.NewInt(1500)))(
		abi.ReserveV2Transfer{From: zeroAddress(), To: recipient.address(), Value: bigInt(1500)},
	)
	s.requireTxWithStrictEvents(newToken.Transfer(signer(recipient), s.account[3].address(), big.NewInt(10)))(
		abi.ReserveV2Transfer{From: recipient.address(), To: s.account[3].address(), Value: bigInt(10)},
	)
	s.requireTxWithStrictEvents(newToken.Pause(signer(newKey)))(
		abi.ReserveV2Paused{Account: newKey.address()},
	)
	s.requireTxWithStrictEvents(newToken.Unpause(signer(newKey)))(
		abi.ReserveV2Unpaused{Account: newKey.address()},
	)
	assertRSVBalance(recipient.address(), big.NewInt(100+1500-10))
	assertRSVBalance(s.account[3].address(), big.NewInt(10))
}

// Test that we can use the owner in ReserveEternalStorage.
func (s *ReserveSuite) TestEternalStorageOwner() {
	assertReserveAddress := func(expected common.Address) {
		reserveAddress, err := s.eternalStorage.ReserveAddress(nil)
		s.NoError(err)
		s.Equal(expected, reserveAddress)
	}

	assertOwner := func(expected common.Address) {
		owner, err := s.eternalStorage.Owner(nil)
		s.NoError(err)
		s.Equal(expected, owner)
	}

	// Check that owner and reserveAddress are initialized in the way we expect.
	assertReserveAddress(s.reserveAddress)
	assertOwner(s.owner.address())

	newOwner := s.account[3]

	// Nominate a new owner.
	s.requireTxWithStrictEvents(s.eternalStorage.NominateNewOwner(s.signer, newOwner.address()))(
		abi.ReserveEternalStorageNewOwnerNominated{
			PreviousOwner: s.owner.address(),
			Nominee:       newOwner.address(),
		},
	)

	// Accept ownership.
	s.requireTxWithStrictEvents(s.eternalStorage.AcceptOwnership(signer(newOwner)))(
		abi.ReserveEternalStorageOwnershipTransferred{
			PreviousOwner: s.owner.address(),
			NewOwner:      newOwner.address(),
		},
	)

	// Check that owner changed and reserveAddress didn't.
	assertReserveAddress(s.reserveAddress)
	assertOwner(newOwner.address())

	newReserveAccount := s.account[4]

	// Change reserveAddress as owner account.
	s.requireTxWithStrictEvents(s.eternalStorage.UpdateReserveAddress(signer(newOwner), newReserveAccount.address()))(
		abi.ReserveEternalStorageReserveAddressTransferred{
			OldReserveAddress: s.reserveAddress,
			NewReserveAddress: newReserveAccount.address(),
		},
	)

	// Check that reserveAddress changed and owner didn't.
	assertReserveAddress(newReserveAccount.address())
	assertOwner(newOwner.address())

	// Check that reserveAddress cannot change owner.
	s.requireTxFails(s.eternalStorage.NominateNewOwner(signer(newReserveAccount), s.account[5].address()))

	// Check that owner can make the change the reserveAddress could not.
	s.requireTxWithStrictEvents(s.eternalStorage.NominateNewOwner(signer(newOwner), s.account[5].address()))(
		abi.ReserveEternalStorageNewOwnerNominated{
			PreviousOwner: newOwner.address(),
			Nominee:       s.account[5].address(),
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

	// Set reserveAddress to newOwner.
	s.requireTxWithStrictEvents(s.eternalStorage.UpdateReserveAddress(s.signer, newOwner.address()))(
		abi.ReserveEternalStorageReserveAddressTransferred{
			OldReserveAddress: s.reserveAddress,
			NewReserveAddress: newOwner.address(),
		},
	)

	// Check that we can now call setBalance.
	s.requireTxWithStrictEvents(s.eternalStorage.SetBalance(signer(newOwner), newOwner.address(), amount))

	// Balance should have changed.
	balance, err := s.eternalStorage.Balance(nil, newOwner.address())
	s.NoError(err)
	s.Equal(amount.String(), balance.String())
}
