package tests

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
	"github.com/reserve-protocol/rsv-beta/soltools"
)

func TestManagerFuzz(t *testing.T) {
	suite.Run(t, new(ManagerFuzzSuite))
}

type ManagerFuzzSuite struct {
	TestSuite
}

var duration = 1000

var (
	// Compile-time check that ManagerFuzzSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &ManagerFuzzSuite{}
	_ suite.SetupAllSuite    = &ManagerFuzzSuite{}
	_ suite.TearDownAllSuite = &ManagerFuzzSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *ManagerFuzzSuite) SetupSuite() {
	s.setup()
	if os.Getenv("FUZZ_DURATION") != "" {
		durationStr := os.Getenv("FUZZ_DURATION")
		asInt, err := strconv.Atoi(durationStr)
		if err != nil {
			duration = asInt
		}
	}
	fmt.Printf("running with fuzz duration: %v\n", duration)

	rand.Seed(time.Now().UnixNano())
}

// TearDownSuite runs once, after all of the tests in the suite.
func (s *ManagerFuzzSuite) TearDownSuite() {
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
func (s *ManagerFuzzSuite) BeforeTest(suiteName, testName string) {
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

func (s *ManagerFuzzSuite) TestByFuzzing() {
	fmt.Print("\n")
	for i := 0; i < duration; i++ {
		fmt.Printf("Run %v", i)

		// Choose between Issuing, Redeeming, WeightProposal, or SwapProposal
		choice := rand.Int31n(4)
		switch choice {
		case 0: // Issue
			fmt.Print("| Issue")
			attoRSV := generateRandUpTo(shiftLeft(1, 30)) // 18 + 12 = 0 to 1 trillion RSV
			s.displayTxResult(s.manager.Issue(signer(s.proposer), attoRSV))
		case 1: // Redeem
			fmt.Print("| Redeem")
			attoRSV := generateRandUpTo(shiftLeft(1, 30)) // 18 + 12 = 0 to 1 trillion RSV
			s.displayTxResult(s.manager.Redeem(signer(s.proposer), attoRSV))
		case 2: // WeightProposal
			fmt.Print("| WeightProposal")
			tokens := generateTokenSet(s.erc20Addresses)
			weights := generateWeights(tokens, shiftLeft(1, 36))
			s.Equal(shiftLeft(1, 36).String(), sumWeights(weights).String())
			s.changeBasketUsingWeightProposal(tokens, weights)
		case 3: // SwapProposal
			fmt.Print("| SwapProposal")
			amounts, toVault := s.generateSwaps()
			s.Equal(bigInt(0).String(), sumSwaps(amounts, toVault).String())
			s.changeBasketUsingSwapProposal(s.erc20Addresses, amounts, toVault)
		}
		fmt.Print("\n")
		s.assertManagerCollateralized()
	}
}

// ===================================== Helpers ===========================================

func generateTokenSet(tokens []common.Address) []common.Address {
	var generated []common.Address
	for _, token := range tokens {
		if rand.Int31n(2) == 0 {
			generated = append(generated, token)
		}
	}
	if len(generated) == 0 {
		generated = append(generated, tokens[0])
	}
	return generated
}

func generateWeights(tokens []common.Address, sum *big.Int) []*big.Int {
	var breakdown []*big.Int
	for i, _ := range tokens {
		n := sum
		if i != len(tokens)-1 {
			n = generateRandUpTo(sum)
			sum = bigInt(0).Sub(sum, n)
		}
		breakdown = append(breakdown, n)
	}
	return breakdown
}

func sumWeights(weights []*big.Int) *big.Int {
	sum := bigInt(0)
	for _, w := range weights {
		sum = sum.Add(sum, w)
	}
	return sum
}

func (s *ManagerFuzzSuite) generateSwaps() ([]*big.Int, []bool) {
	amounts := make([]*big.Int, len(s.erc20s))
	toVault := make([]bool, len(s.erc20s))

	indexToVault := int(rand.Int31n(int32(len(s.erc20s))))
	indexBalance, err := s.erc20s[indexToVault].BalanceOf(nil, s.vaultAddress)
	s.Require().NoError(err)

	amounts[indexToVault] = generateRandUpTo(indexBalance)
	toVault[indexToVault] = true

	remainingTotal := amounts[indexToVault]
	for i, _ := range s.erc20s {
		if i == indexToVault {
			continue
		}

		amounts[i] = remainingTotal

		isIndexLast := 0
		if indexToVault == len(s.erc20s)-1 {
			isIndexLast = 1
		}
		if i != len(s.erc20s)-1-isIndexLast {
			amounts[i] = generateRandUpTo(remainingTotal)
			remainingTotal = remainingTotal.Sub(remainingTotal, amounts[i])
		}
	}
	return amounts, toVault
}

func sumSwaps(amounts []*big.Int, toVault []bool) *big.Int {
	sum := bigInt(0)
	for i, a := range amounts {
		if toVault[i] {
			sum = sum.Add(sum, a)
		} else {
			sum = sum.Sub(sum, a)
		}
	}
	return sum
}

// generateRandUpTo returns a random *big.Int between 0 and `n`
func generateRandUpTo(n *big.Int) *big.Int {
	bound := shiftLeft(1, 18)
	digitsAboveBound := bigInt64(0).Div(n, bound)

	toAdd := bigInt64(0)
	if digitsAboveBound.Cmp(bigInt64(0)) == 1 { // greater than 1e18
		randDigitsAboveBound := generateRandUpTo(digitsAboveBound)
		toAdd = bigInt64(0).Mul(randDigitsAboveBound, bound)
	}
	return bigInt64(0).Add(toAdd, bigInt64(rand.Int63n(bound.Int64())))
}

func bigInt64(n int64) *big.Int {
	return big.NewInt(n)
}

func (s *ManagerFuzzSuite) displayTxResult(tx *types.Transaction, err error) {
	if err == nil {
		receipt, err := bind.WaitMined(context.Background(), s.node, tx)
		s.Require().NoError(err)
		if receipt.Status == types.ReceiptStatusSuccessful {
			fmt.Printf("| Success")
		}
	}
	fmt.Printf("| Failure")
}

func (s *ManagerFuzzSuite) tryWeightProposal(tokens []common.Address, weights []*big.Int) {
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

	// Try to execute the Proposal, but it's okay if it fails.
	s.displayTxResult(s.manager.ExecuteProposal(signer(s.operator), proposalID))
}

func (s *ManagerFuzzSuite) trySwapProposal(tokens []common.Address, amounts []*big.Int, toVault []bool) {
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

	// Try to execute the Proposal.
	s.displayTxResult(s.manager.ExecuteProposal(signer(s.operator), proposalID))
}
