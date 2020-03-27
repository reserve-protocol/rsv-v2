// +build all

package tests

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
)

func TestRelayer(t *testing.T) {
	suite.Run(t, new(RelayerSuite))
}

type RelayerSuite struct {
	TestSuite

	relayer        *abi.Relayer
	relayerAddress common.Address
}

var (
	// Compile-time check that RelayerSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &RelayerSuite{}
	_ suite.SetupAllSuite    = &RelayerSuite{}
	_ suite.TearDownAllSuite = &RelayerSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *RelayerSuite) SetupSuite() {
	s.setup()
}

// BeforeTest runs before each test in the suite.
func (s *RelayerSuite) BeforeTest(suiteName, testName string) {
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

	s.assertRSVTotalSupply(bigInt(0))

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

	relayerAddress, tx, relayer, err := abi.DeployRelayer(s.signer, s.node, s.reserveAddress)

	s.requireTx(tx, err)()

	s.relayer = relayer
	s.relayerAddress = relayerAddress

	s.logParsers[s.relayerAddress] = s.relayer

	// Make sure Reserve address set correctly.
	deployedRSVAddress, err := s.relayer.TrustedRSV(nil)
	s.Require().NoError(err)
	s.Equal(deployedRSVAddress, s.reserveAddress)

	// Set Reserve's trusted relayer address correctly.
	s.requireTxWithStrictEvents(s.reserve.ChangeRelayer(s.signer, s.relayerAddress))(
		abi.ReserveTrustedRelayerChanged{NewTrustedRelayer: s.relayerAddress},
	)

	// Apparently `ecrecover` is only available on private blockchains after
	// sending wei to its address, which is address `1`.
	// See here: https://solidity.readthedocs.io/en/v0.6.4/units-and-global-variables.html
	nonce, err := s.node.PendingNonceAt(context.Background(), s.account[0].address())
	s.Require().NoError(err)

	tx, err = types.SignTx(
		types.NewTransaction(nonce, common.BytesToAddress([]byte{1}), bigInt(1), 210000, bigInt(1), nil),
		types.HomesteadSigner{},
		s.account[0].key,
	)
	s.node.SendTransaction(context.Background(), tx)
	s.requireTx(tx, err)
}

func (s *RelayerSuite) TestDeploy() {}

// TestTransfer checks that someone with RSV can send RSV to a recipient through a relayer.
func (s *RelayerSuite) TestTransfer() {
	relayer := s.account[4]
	sender := s.account[1]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to sender.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)

	// Check that balances are as expected.
	s.assertRSVBalance(sender.address(), amount)
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))

	nonce, err := s.relayer.Nonce(nil, sender.address())
	s.Require().NoError(err)

	hash := s.transferHash(sender.address(), recipient, amount, bigInt(0), nonce)
	sig, err := crypto.Sign(hash, sender.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay transfer tx
	s.requireTxWithStrictEvents(s.relayer.ForwardTransfer(signer(relayer), sig, sender.address(), recipient, amount, bigInt(0)))(
		abi.RelayerTransferForwarded{
			Sig:    sig,
			From:   sender.address(),
			To:     recipient,
			Amount: amount,
			Fee:    bigInt(0),
		},
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

}

// TestTransferWithFee checks that someone with RSV can send RSV to a recipient through a relayer.
func (s *RelayerSuite) TestTransferWithFee() {
	relayer := s.account[4]
	sender := s.account[1]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)
	fee := bigInt(1)
	total := bigInt(101)

	// Mint initial amount sender.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)
	s.assertRSVBalance(sender.address(), amount)

	// Try to transfer total amount and fail.
	nonce, err := s.relayer.Nonce(nil, sender.address())
	s.Require().NoError(err)
	hash := s.transferHash(sender.address(), recipient, amount, fee, nonce)
	sig, err := crypto.Sign(hash, sender.key)
	s.Require().NoError(err)
	sig = addToLastByte(sig)
	s.requireTxFails(s.relayer.ForwardTransfer(signer(relayer), sig, sender.address(), recipient, amount, fee))

	// Give sender enough to pay fee.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), fee))(
		mintingTransfer(sender.address(), fee),
	)
	s.assertRSVBalance(sender.address(), total)

	// Now transaction should complete
	s.requireTxWithStrictEvents(s.relayer.ForwardTransfer(signer(relayer), sig, sender.address(), recipient, amount, fee))(
		abi.RelayerTransferForwarded{
			Sig:    sig,
			From:   sender.address(),
			To:     recipient,
			Amount: amount,
			Fee:    fee,
		},
		abi.RelayerFeeTaken{
			From:  sender.address(),
			To:    relayer.address(),
			Value: fee,
		},
		abi.ReserveTransfer{
			From:  sender.address(),
			To:    recipient,
			Value: amount,
		},
		abi.ReserveTransfer{
			From:  sender.address(),
			To:    relayer.address(),
			Value: fee,
		},
	)

	// Check that balances are as expected.
	s.assertRSVBalance(sender.address(), bigInt(0))
	s.assertRSVBalance(recipient, amount)
	s.assertRSVBalance(relayer.address(), fee)
}

// TestTransferFailsFromScammer checks that other accounts cannot
func (s *RelayerSuite) TestTransferFailsFromScammer() {
	relayer := s.account[4]
	sender := s.account[1]
	scammer := s.account[2]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to sender.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, sender.address(), amount))(
		mintingTransfer(sender.address(), amount),
	)

	// Check that balances are as expected.
	s.assertRSVBalance(sender.address(), amount)
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))

	nonce, err := s.relayer.Nonce(nil, sender.address())
	s.Require().NoError(err)

	hash := s.transferHash(sender.address(), recipient, amount, bigInt(0), nonce)
	sig, err := crypto.Sign(hash, scammer.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay transfer tx
	s.requireTxFails(s.relayer.ForwardTransfer(signer(relayer), sig, sender.address(), recipient, amount, bigInt(0)))

	// Check that balances haven't changed.
	s.assertRSVBalance(sender.address(), amount)
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))
}

func (s *RelayerSuite) TestApproveAndTransferFrom() {
	relayer := s.account[4]
	holder := s.account[1]
	spender := s.account[2]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to holder.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, holder.address(), amount))(
		mintingTransfer(holder.address(), amount),
	)

	// Check that balances and allowances are as expected.
	s.assertRSVBalance(holder.address(), amount)
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))
	s.assertRSVAllowance(holder.address(), spender.address(), bigInt(0))

	nonce, err := s.relayer.Nonce(nil, holder.address())
	s.Require().NoError(err)

	hash := s.approveHash(holder.address(), spender.address(), amount, bigInt(0), nonce)
	sig, err := crypto.Sign(hash, holder.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay approve tx.
	s.requireTxWithStrictEvents(s.relayer.ForwardApprove(signer(relayer), sig, holder.address(), spender.address(), amount, bigInt(0)))(
		abi.RelayerApproveForwarded{
			Sig:     sig,
			Holder:  holder.address(),
			Spender: spender.address(),
			Amount:  amount,
			Fee:     bigInt(0),
		},
		abi.ReserveApproval{
			Owner:   holder.address(),
			Spender: spender.address(),
			Value:   amount,
		},
	)

	// Check that the spender has allowance
	s.assertRSVAllowance(holder.address(), spender.address(), amount)

	// ==== SECOND RELAY === //

	nonce, err = s.relayer.Nonce(nil, spender.address())
	s.Require().NoError(err)

	hash = s.transferFromHash(holder.address(), spender.address(), recipient, amount, bigInt(0), nonce)
	sig, err = crypto.Sign(hash, spender.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay approve tx.
	s.requireTxWithStrictEvents(s.relayer.ForwardTransferFrom(signer(relayer), sig, holder.address(), spender.address(), recipient, amount, bigInt(0)))(
		abi.RelayerTransferFromForwarded{
			Sig:     sig,
			Holder:  holder.address(),
			Spender: spender.address(),
			To:      recipient,
			Amount:  amount,
			Fee:     bigInt(0),
		},
		abi.ReserveTransfer{
			From:  holder.address(),
			To:    recipient,
			Value: amount,
		},
		abi.ReserveApproval{
			Owner:   holder.address(),
			Spender: spender.address(),
			Value:   bigInt(0),
		},
	)

	// Check that the spender has no allowance and that balances have changed.
	s.assertRSVAllowance(holder.address(), spender.address(), bigInt(0))
	s.assertRSVBalance(holder.address(), bigInt(0))
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVBalance(recipient, amount)
	s.assertRSVBalance(s.owner.address(), bigInt(0))
}

// TestApproveWithFee checks that someone can approve RSV while paying a fee to a relayer.
func (s *RelayerSuite) TestApproveWithFee() {
	relayer := s.account[4]
	holder := s.account[1]
	spender := s.account[2]
	amount := bigInt(100)
	fee := bigInt(1)

	// Mint fee to holder.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, holder.address(), fee))(
		mintingTransfer(holder.address(), fee),
	)
	s.assertRSVBalance(holder.address(), fee)

	// Approve spender and pay fee.
	nonce, err := s.relayer.Nonce(nil, holder.address())
	s.Require().NoError(err)
	hash := s.approveHash(holder.address(), spender.address(), amount, fee, nonce)
	sig, err := crypto.Sign(hash, holder.key)
	s.Require().NoError(err)
	sig = addToLastByte(sig)
	s.requireTxWithStrictEvents(s.relayer.ForwardApprove(signer(relayer), sig, holder.address(), spender.address(), amount, fee))(
		abi.RelayerApproveForwarded{
			Sig:     sig,
			Holder:  holder.address(),
			Spender: spender.address(),
			Amount:  amount,
			Fee:     fee,
		},
		abi.RelayerFeeTaken{
			From:  holder.address(),
			To:    relayer.address(),
			Value: fee,
		},
		abi.ReserveApproval{
			Owner:   holder.address(),
			Spender: spender.address(),
			Value:   amount,
		},
		abi.ReserveTransfer{
			From:  holder.address(),
			To:    relayer.address(),
			Value: fee,
		},
	)

	// Check that balances and allowances are as expected.
	s.assertRSVBalance(holder.address(), bigInt(0))
	s.assertRSVBalance(relayer.address(), fee)
	s.assertRSVAllowance(holder.address(), spender.address(), amount)
}

func (s *RelayerSuite) TestApproveFailsFromScammer() {
	relayer := s.account[4]
	holder := s.account[1]
	spender := s.account[2]
	scammer := s.account[2] // the scammer is the spender!
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to holder.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, holder.address(), amount))(
		mintingTransfer(holder.address(), amount),
	)

	// Check that balances and allowances are as expected.
	s.assertRSVBalance(holder.address(), amount)
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))
	s.assertRSVAllowance(holder.address(), spender.address(), bigInt(0))

	nonce, err := s.relayer.Nonce(nil, holder.address())
	s.Require().NoError(err)

	hash := s.approveHash(holder.address(), spender.address(), amount, bigInt(0), nonce)
	sig, err := crypto.Sign(hash, scammer.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay approve tx.
	s.requireTxFails(s.relayer.ForwardApprove(signer(relayer), sig, holder.address(), spender.address(), amount, bigInt(0)))

	// Check that the spender did not get allowance.
	s.assertRSVAllowance(holder.address(), spender.address(), bigInt(0))
}

// TestTransferFromWithFee checks that a spender pays a fee to the relayer when spending from holder.
func (s *RelayerSuite) TestTransferFromWithFee() {
	relayer := s.account[4]
	holder := s.account[1]
	spender := s.account[2]
	recipient := s.account[3]
	amount := bigInt(100)
	fee := bigInt(1)

	// Mint amount to holder.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, holder.address(), amount))(
		mintingTransfer(holder.address(), amount),
	)
	s.assertRSVBalance(holder.address(), amount)

	// Mint fee to spender.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, spender.address(), fee))(
		mintingTransfer(spender.address(), fee),
	)
	s.assertRSVBalance(spender.address(), fee)

	// Approve spender.
	nonce, err := s.relayer.Nonce(nil, holder.address())
	s.Require().NoError(err)
	hash := s.approveHash(holder.address(), spender.address(), amount, bigInt(0), nonce)
	sig, err := crypto.Sign(hash, holder.key)
	s.Require().NoError(err)
	sig = addToLastByte(sig)
	s.requireTx(s.relayer.ForwardApprove(signer(relayer), sig, holder.address(), spender.address(), amount, bigInt(0)))

	// Perform transferFrom and pay fee.
	nonce, err = s.relayer.Nonce(nil, spender.address())
	s.Require().NoError(err)
	hash = s.transferFromHash(holder.address(), spender.address(), recipient.address(), amount, fee, nonce)
	sig, err = crypto.Sign(hash, spender.key)
	s.Require().NoError(err)
	sig = addToLastByte(sig)
	s.requireTxWithStrictEvents(s.relayer.ForwardTransferFrom(signer(relayer), sig, holder.address(), spender.address(), recipient.address(), amount, fee))(
		abi.RelayerTransferFromForwarded{
			Sig:     sig,
			Holder:  holder.address(),
			Spender: spender.address(),
			To:      recipient.address(),
			Amount:  amount,
			Fee:     fee,
		},
		abi.RelayerFeeTaken{
			From:  spender.address(),
			To:    relayer.address(),
			Value: fee,
		},
		abi.ReserveTransfer{
			From:  holder.address(),
			To:    recipient.address(),
			Value: amount,
		},
		abi.ReserveTransfer{
			From:  spender.address(),
			To:    relayer.address(),
			Value: fee,
		},
		abi.ReserveApproval{
			Owner:   holder.address(),
			Spender: spender.address(),
			Value:   bigInt(0),
		},
	)

	// Check that balances and allowances are as expected.
	s.assertRSVBalance(holder.address(), bigInt(0))
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVBalance(recipient.address(), amount)
	s.assertRSVBalance(relayer.address(), fee)
	s.assertRSVAllowance(holder.address(), spender.address(), bigInt(0))
}

func (s *RelayerSuite) TestTransferFromFailsFromScammer() {
	relayer := s.account[4]
	holder := s.account[1]
	spender := s.account[2]
	scammer := s.account[3]
	recipient := common.BigToAddress(bigInt(1))
	amount := bigInt(100)

	// Mint to holder.
	s.requireTxWithStrictEvents(s.reserve.Mint(s.signer, holder.address(), amount))(
		mintingTransfer(holder.address(), amount),
	)

	// Check that balances and allowances are as expected.
	s.assertRSVBalance(holder.address(), amount)
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))
	s.assertRSVAllowance(holder.address(), spender.address(), bigInt(0))

	nonce, err := s.relayer.Nonce(nil, holder.address())
	s.Require().NoError(err)

	hash := s.approveHash(holder.address(), spender.address(), amount, bigInt(0), nonce)
	sig, err := crypto.Sign(hash, holder.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay approve tx.
	s.requireTxWithStrictEvents(s.relayer.ForwardApprove(signer(relayer), sig, holder.address(), spender.address(), amount, bigInt(0)))(
		abi.RelayerApproveForwarded{
			Sig:     sig,
			Holder:  holder.address(),
			Spender: spender.address(),
			Amount:  amount,
			Fee:     bigInt(0),
		},
		abi.ReserveApproval{
			Owner:   holder.address(),
			Spender: spender.address(),
			Value:   amount,
		},
	)

	// Check that the spender has allowance
	s.assertRSVAllowance(holder.address(), spender.address(), amount)

	// ==== SECOND RELAY, this one should fail === //

	nonce, err = s.relayer.Nonce(nil, spender.address())
	s.Require().NoError(err)

	hash = s.transferFromHash(holder.address(), spender.address(), recipient, amount, bigInt(0), nonce)
	sig, err = crypto.Sign(hash, scammer.key)
	s.Require().NoError(err)

	// Add 27 to 65th byte.
	sig = addToLastByte(sig)

	// Relay approve tx.
	s.requireTxFails(s.relayer.ForwardTransferFrom(signer(relayer), sig, holder.address(), spender.address(), recipient, amount, bigInt(0)))

	// Check that the spender still has allowance and that balances haven't changed.
	s.assertRSVAllowance(holder.address(), spender.address(), amount)
	s.assertRSVBalance(holder.address(), amount)
	s.assertRSVBalance(spender.address(), bigInt(0))
	s.assertRSVBalance(recipient, bigInt(0))
	s.assertRSVBalance(s.owner.address(), bigInt(0))

}

func (s *RelayerSuite) TestSetRSVProtected() {
	scammer := s.account[1]

	s.requireTxFails(s.relayer.SetRSV(signer(scammer), scammer.address()))

	trustedRSV, err := s.relayer.TrustedRSV(nil)
	s.Require().NoError(err)

	s.Equal(s.reserveAddress.String(), trustedRSV.String())
}

// ========================================== HELPERS ========================================= //

func (s *RelayerSuite) transferHash(
	from common.Address,
	to common.Address,
	amount *big.Int,
	fee *big.Int,
	nonce *big.Int,
) []byte {
	interimHash := crypto.Keccak256Hash(
		s.reserveAddress.Bytes(),
		[]byte("forwardTransfer"),
		from.Bytes(),
		to.Bytes(),
		common.LeftPadBytes(amount.Bytes(), 32),
		common.LeftPadBytes(fee.Bytes(), 32),
		common.LeftPadBytes(nonce.Bytes(), 32),
	)
	return crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(interimHash))),
		interimHash.Bytes(),
	).Bytes()
}

func (s *RelayerSuite) transferFromHash(
	holder common.Address,
	spender common.Address,
	to common.Address,
	amount *big.Int,
	fee *big.Int,
	nonce *big.Int,
) []byte {
	interimHash := crypto.Keccak256Hash(
		s.reserveAddress.Bytes(),
		[]byte("forwardTransferFrom"),
		holder.Bytes(),
		spender.Bytes(),
		to.Bytes(),
		common.LeftPadBytes(amount.Bytes(), 32),
		common.LeftPadBytes(fee.Bytes(), 32),
		common.LeftPadBytes(nonce.Bytes(), 32),
	)
	return crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(interimHash))),
		interimHash.Bytes(),
	).Bytes()
}

func (s *RelayerSuite) approveHash(
	holder common.Address,
	spender common.Address,
	amount *big.Int,
	fee *big.Int,
	nonce *big.Int,
) []byte {
	interimHash := crypto.Keccak256Hash(
		s.reserveAddress.Bytes(),
		[]byte("forwardApprove"),
		holder.Bytes(),
		spender.Bytes(),
		common.LeftPadBytes(amount.Bytes(), 32),
		common.LeftPadBytes(fee.Bytes(), 32),
		common.LeftPadBytes(nonce.Bytes(), 32),
	)
	return crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(interimHash))),
		interimHash.Bytes(),
	).Bytes()
}

func addToLastByte(sig []byte) []byte {
	v := []byte{27}
	v[0] = v[0] + sig[64]
	sig = sig[:64]
	sig = append(sig, v...)
	return sig
}
