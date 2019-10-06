// +build all fuzz

package tests

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
)

func TestManagerFuzz(t *testing.T) {
	suite.Run(t, new(ManagerFuzzSuite))
}

type ManagerFuzzSuite struct {
	TestSuite
	numTokens         int
	decimals          []uint32
	addressToDecimals map[common.Address]uint32
}

var duration = flag.Int("runs", 10, "transactions to randomly generate")
var decimals = flag.String("decimals", "6,18,6", "number of decimals for each token")

// Limitations: Only up to 10 tokens max

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
	rand.Seed(time.Now().UnixNano())
	for _, d := range strings.Split(*decimals, ",") {
		dInt, err := strconv.Atoi(d)
		s.Require().NoError(err)
		s.decimals = append(s.decimals, uint32(dInt))
	}
	s.numTokens = len(s.decimals)
	s.addressToDecimals = make(map[common.Address]uint32)
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

	// Deploy collateral ERC20s.
	s.erc20s = make([]*abi.BasicERC20, s.numTokens)
	s.erc20Addresses = make([]common.Address, s.numTokens)
	for i := 0; i < s.numTokens; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)

		s.addressToDecimals[erc20Address] = s.decimals[i]
		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
		s.logParsers[erc20Address] = erc20
	}

	// Make a simple basket
	weights := s.generateWeights(s.erc20Addresses)
	basketAddress, tx, basket, err := abi.DeployBasket(
		s.signer,
		s.node,
		zeroAddress(),
		s.erc20Addresses,
		weights,
	)

	s.requireTxWithStrictEvents(tx, err)()
	s.basketAddress = basketAddress
	s.basket = basket

	// Manager.
	managerAddress, tx, manager, err := abi.DeployManager(
		s.signer,
		s.node,
		vaultAddress,
		reserveAddress,
		propFactoryAddress,
		basketAddress,
		s.operator.address(),
		bigInt(0),
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
	s.Require().Equal(true, emergency)

	// Unpause from emergency.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(signer(s.operator), false))(
		abi.ManagerEmergencyChanged{OldVal: true, NewVal: false},
	)

	// Confirm we are unpaused from emergency.
	emergency, err = s.manager.Emergency(nil)
	s.Require().NoError(err)
	s.Require().Equal(false, emergency)

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

	// Fund and set allowances.
	var amounts []*big.Int
	for i := 0; i < s.numTokens; i++ {
		amounts = append(amounts, shiftLeft(1, 48))
	}
	s.fundAccountWithErc20sAndApprove(s.proposer, amounts)
}

// TestByFuzzing chooses between Issuing, Redeeming, WeightProposal, and SwapProposal for
// `duration` times and asserts invariants are uphold at every step.
func (s *ManagerFuzzSuite) TestByFuzzing() {
	fmt.Print("\n")
	fmt.Printf("Running fuzzing with %v tokens with decimals: %v\n", s.numTokens, s.decimals)
	for i := 0; i < *duration; i++ {
		fmt.Printf("Run %v", i)

		// Record how much value the proposer starts with.
		erc20Balances := s.getERC20Balances()

		// Choose between Issuing, Redeeming, WeightProposal, or SwapProposal
		choice := rand.Int31n(4)
		switch choice {
		case 0: // Issue
			fmt.Print(" |      Issue    ")

			// Uniformly select a random amount of attoRSV from 0 to 1 trillion RSV.
			attoRSV := generateRandUpTo(shiftLeft(1, 30))

			// Issue that much RSV.
			s.displayTxResult(s.manager.Issue(signer(s.proposer), attoRSV))

		case 1: // Redeem
			fmt.Print(" |     Redeem    ")

			// Get the amount of RSV in circulation.
			rsvSupply, err := s.reserve.TotalSupply(nil)
			s.Require().NoError(err)

			// Uniformly select a random amount of attoRSV from 0 to `rsvSupply`.
			attoRSV := generateRandUpTo(rsvSupply)

			// Approve the Manager to spend that amount of RSV from the proposer's address.
			s.requireTx(s.reserve.Approve(signer(s.proposer), s.managerAddress, attoRSV))

			// Redeem that much RSV.
			s.displayTxResult(s.manager.Redeem(signer(s.proposer), attoRSV))

		case 2: // WeightProposal
			fmt.Print(" | WeightProposal")

			// Choose a number of tokens from the full set using a binomial distribution.
			_, tokens := s.chooseTokenSet()

			weights := s.generateWeights(tokens)

			// Try to execute this WeightProposal. It's okay if it fails.
			s.tryWeightProposal(tokens, weights)

		case 3: // SwapProposal
			fmt.Print(" |  SwapProposal ")

			// Choose a number of tokens from the full set using a binomial distribution.
			erc20s, tokens := s.chooseTokenSet()

			// Generate a SwapProposal.
			amounts, toVault := s.generateSwaps(erc20s, tokens)

			// Try to execute this SwapProposal. It's okay if it fails.
			s.trySwapProposal(tokens, amounts, toVault)
		}

		// Display RSV supply and basket content.
		s.printMetrics()

		//
		if choice >= 2 {
			// The proposer shouldn't end up with more value than they started with.
			s.printRoundingError(erc20Balances)
		}

		// Check our on-chain invariant.
		s.assertManagerCollateralized()

		// This is an off-chain calculation that should be identical to the on-chain one.
		s.assertManagerCollateralizedOffChain()
		fmt.Print("\n")
	}
}

// ===================================== Helpers ===========================================

// chooseTokenSet chooses a subset of `tokens` using a binomial distribution.
func (s *ManagerFuzzSuite) chooseTokenSet() ([]*abi.BasicERC20, []common.Address) {
	var addresses []common.Address
	var bindings []*abi.BasicERC20
	for i, token := range s.erc20Addresses {
		if rand.Int31n(2) == 0 {
			addresses = append(addresses, token)
			bindings = append(bindings, s.erc20s[i])
		}
	}
	return bindings, addresses
}

// generateWeights randomly chooses weights that add to `sum`.
// I'm not really sure what distribution this is. It certainly biases toward
// putting the largest weights on the early tokens, and fewest on the last.
func (s *ManagerFuzzSuite) generateWeights(tokens []common.Address) []*big.Int {
	var weights []*big.Int
	sum := shiftLeft(1, 18)
	for _, _ = range tokens {
		n := generateRandUpTo(sum)
		weights = append(weights, n)
		sum = bigInt(0).Sub(sum, n)
	}

	// Add the leftover to the last element
	if len(tokens) > 0 {
		weights[len(weights)-1] = sum.Add(sum, weights[len(weights)-1])
	}

	// Check that the weights sum to 1e18.
	if len(weights) > 0 {
		s.Require().Equal(shiftLeft(1, 18).String(), sumWeights(weights).String())
	}

	// Multiply each weight by 1eDecimals
	for i, _ := range weights {
		weights[i] = bigInt(0).Mul(weights[i], shiftLeft(1, s.addressToDecimals[tokens[i]]))
	}
	return weights
}

// generateSwaps generates a list of `amounts` that in net based on `toVault` add to 0.
// Again, here the random distribution of `amounts` isn't uniform.
func (s *ManagerFuzzSuite) generateSwaps(erc20s []*abi.BasicERC20, tokens []common.Address) ([]*big.Int, []bool) {
	amounts := make([]*big.Int, len(erc20s))
	toVault := make([]bool, len(erc20s))

	// Return null results when the list of tokens is either of size 0 or 1.
	if len(erc20s) == 0 {
		return amounts, toVault
	} else if len(erc20s) == 1 {
		amounts[0] = bigInt(0)
		return amounts, toVault
	}

	// Randomly choose a token that will be the one to transfer out of the vault.
	// All other tokens will be transferred into the vault.
	indexToVault := int(rand.Int31n(int32(len(erc20s))))
	indexBalance, err := erc20s[indexToVault].BalanceOf(nil, s.vaultAddress)
	s.Require().NoError(err)
	toVault[indexToVault] = false

	// Multiply 1e100 by everything and then remove zeroes later.
	amounts[indexToVault] = generateRandUpTo(indexBalance)
	amounts[indexToVault] = bigInt(0).Mul(amounts[indexToVault], shiftLeft(1, 100))

	// Now we have to meet the constraint that the sum of the rest of the tokens must equal this chosen amount.
	remainingTotal := bigInt(0).Add(bigInt(0), amounts[indexToVault])
	for i, _ := range erc20s {
		if i == indexToVault {
			continue
		}

		toVault[i] = true
		amounts[i] = generateRandUpTo(remainingTotal)
		remainingTotal = remainingTotal.Sub(remainingTotal, amounts[i])
	}

	// Add the remaining total to the last token going into the Vault.
	if indexToVault == len(erc20s)-1 {
		amounts[len(erc20s)-2] = remainingTotal.Add(remainingTotal, amounts[len(erc20s)-2])
	} else {
		amounts[len(erc20s)-1] = remainingTotal.Add(remainingTotal, amounts[len(erc20s)-1])
	}

	// In net this should not be an exchange of value.
	s.Require().Equal(bigInt(0).String(), sumSwaps(amounts, toVault).String())

	// Now we have to put everything in terms of their own decimals.
	baseDecimals := s.addressToDecimals[tokens[indexToVault]]
	for i, _ := range amounts {
		decimal := s.addressToDecimals[tokens[i]]
		decimalDiff := baseDecimals - decimal
		amounts[i] = bigInt(0).Div(amounts[i], shiftLeft(1, 100+decimalDiff))
	}

	return amounts, toVault
}

// tryWeightProposal tries to execute a WeightProposal and prints the result.
func (s *ManagerFuzzSuite) tryWeightProposal(tokens []common.Address, weights []*big.Int) {
	// Propose the new basket.
	_, err := s.manager.ProposeWeights(signer(s.proposer), tokens, weights)
	if err != nil {
		fmt.Printf(" | ❌")
		return
	}

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
	s.Require().True(reflect.DeepEqual(basketTokens, tokens))

	// Size
	basketSize, err := basket.Size(nil)
	s.Require().NoError(err)
	s.Require().Equal(bigInt(uint32(len(tokens))).String(), basketSize.String())

	// Weights
	for i := 0; i < len(weights); i++ {
		foundBacking, err := basket.Weights(nil, tokens[i])
		s.Require().NoError(err)
		s.Require().Equal(weights[i], foundBacking)
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

// trySwapProposal tries to execute a SwapProposal and prints the result.
func (s *ManagerFuzzSuite) trySwapProposal(tokens []common.Address, amounts []*big.Int, toVault []bool) {
	// Propose the new basket.
	_, err := s.manager.ProposeSwap(signer(s.proposer), tokens, amounts, toVault)
	if err != nil {
		fmt.Printf(" | ❌")
		return
	}

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

// printMetrics prints the current RSV supply in scientific notation.
func (s *ManagerFuzzSuite) printMetrics() {
	rsvSupply, err := s.reserve.TotalSupply(nil)
	s.Require().NoError(err)

	fmt.Printf(" | RSV: %v", toScientificNotation(rsvSupply))

	basketAddress, err := s.manager.TrustedBasket(nil)
	s.Require().NoError(err)

	basket, err := abi.NewBasket(basketAddress, s.node)
	s.Require().NoError(err)

	tokens, err := basket.GetTokens(nil)
	s.Require().NoError(err)

	weights := make([]*big.Int, len(s.erc20s))
	balances := make([]*big.Int, len(s.erc20s))
	for i, token := range s.erc20Addresses {
		tokenErc20, err := abi.NewBasicERC20(token, s.node)
		s.Require().NoError(err)

		tokBal, err := tokenErc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)

		balances[i] = tokBal

		if hasAddress(tokens, token) {
			tokenWeight, err := basket.Weights(nil, token)
			s.Require().NoError(err)
			weights[i] = tokenWeight
		} else {
			weights[i] = bigInt(0)
		}
	}

	weightsSum := sumWeights(weights)
	fmt.Print(" |")
	for i, _ := range s.erc20Addresses {
		fmt.Printf(" %v", toFraction(weights[i], weightsSum))
	}
	// fmt.Print(" |")
	// for i, _ := range s.erc20Addresses {

	// 	fmt.Printf(" %v", bigInt(0).Div(weights[i], shiftLeft(1, s.decimals[i])))
	// }
}

// assertManagerCollateralizedOffChain is the same calculation that happens on-chain.
func (s *ManagerFuzzSuite) assertManagerCollateralizedOffChain() {
	basketAddress, err := s.manager.TrustedBasket(nil)
	s.Require().NoError(err)

	basket, err := abi.NewBasket(basketAddress, s.node)
	s.Require().NoError(err)

	tokens, err := basket.GetTokens(nil)
	s.Require().NoError(err)

	rsvTotalSupply, err := s.reserve.TotalSupply(nil)
	s.Require().NoError(err)
	s.Require().True(len(tokens) > 0)

	for _, token := range tokens {
		tokenErc20, err := abi.NewBasicERC20(token, s.node)
		s.Require().NoError(err)

		tokBal, err := tokenErc20.BalanceOf(nil, s.vaultAddress)
		s.Require().NoError(err)

		tokenWeight, err := basket.Weights(nil, token)
		s.Require().NoError(err)

		leftSide := bigInt(0).Mul(rsvTotalSupply, tokenWeight)
		rightSide := bigInt(0).Mul(tokBal, shiftLeft(1, 36))
		s.Require().True(leftSide.Cmp(rightSide) <= 0)
	}

}

// getERC20Balances returns the list of ERC20 balances the proposer has.
func (s *ManagerFuzzSuite) getERC20Balances() []*big.Int {
	var balances []*big.Int
	for _, erc20 := range s.erc20s {
		bal, err := erc20.BalanceOf(nil, s.proposer.address())
		s.Require().NoError(err)
		balances = append(balances, bal)
	}
	return balances
}

// printRoundingError prints the total winnings for the proposer.
func (s *ManagerFuzzSuite) printRoundingError(oldBalances []*big.Int) {
	newBalances := s.getERC20Balances()

	total := bigInt(0)
	// Multiply every gain by 1^100 so we can successfully divide out decimals.
	for i := 0; i < s.numTokens; i++ {
		diff := bigInt(0).Sub(newBalances[i], oldBalances[i])
		diff = bigInt(0).Mul(diff, shiftLeft(1, 100))
		diff = bigInt(0).Div(diff, shiftLeft(1, s.decimals[i]))
		total = bigInt(0).Add(total, diff)
	}

	// Now we have a total that is 1e100 bigger than it should be.
	// Divide by 1e94 to get things in terms of millionths.
	total = bigInt(0).Div(total, shiftLeft(1, 94))
	// fmt.Println(total.String())

	// fmt.Println(oldBalances)
	// fmt.Println(newBalances)

	if total.Cmp(bigInt(0)) == 1 {
		fmt.Printf(" -- Cumulative qtoken gain by proposer: %v/million", total.String())
	}
}

// displayTxResult prints whether or not the tx succeeded.
func (s *ManagerFuzzSuite) displayTxResult(tx *types.Transaction, err error) {
	if err == nil {
		receipt, err := bind.WaitMined(context.Background(), s.node, tx)
		s.Require().NoError(err)
		if receipt.Status == types.ReceiptStatusSuccessful {
			fmt.Printf(" | ✅")
		}
	} else {
		fmt.Printf(" | ❌")
	}
}

// =========================================== Utility ===========================================

// sumWeights returns the sum of a `big.Int` array.
func sumWeights(weights []*big.Int) *big.Int {
	sum := bigInt(0)
	for _, w := range weights {
		sum = sum.Add(sum, w)
	}
	return sum
}

// sumSwaps sums `amounts` based on whether they are going to or from the vault.
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

// generateRandUpTo returns a random *big.Int between 0 and `n`.
// This implementation is pretty fun :) check it out
func generateRandUpTo(n *big.Int) *big.Int {
	if n.Cmp(bigInt(0)) == 0 { // the void basecase
		return bigInt(0)
	}

	bound := shiftLeft(1, 18)
	above, below := bigInt(0).DivMod(n, bound, bigInt(0))

	// The real basecase
	if above.Cmp(bigInt(0)) == 0 {
		return big.NewInt(rand.Int63n(below.Int64()))
	}

	// The recursive case.
	inheritance := generateRandUpTo(above) // recurse!
	toReturn := big.NewInt(rand.Int63n(bound.Int64()))

	// Add the two sets of random digits
	toReturn = toReturn.Add(toReturn, bigInt(0).Mul(inheritance, bound))
	return toReturn
}

// toScientificNotation returns `n` in scientific notation.
func toScientificNotation(n *big.Int) string {
	nStr := n.String()
	numDigits := len(n.String())

	frontStr := nStr[0]
	backStr := nStr[1:numDigits]
	if numDigits > 3 {
		backStr = nStr[1:3]
	}
	return string(frontStr) + "." + string(backStr) + "e" + strconv.Itoa(numDigits-1)
}

// toFraction returns `n` as a percentage of `total`.
func toFraction(n *big.Int, total *big.Int) string {
	biggerN := bigInt(0).Mul(n, shiftLeft(1, 2))
	frac := bigInt(0).Div(biggerN, total).Int64()

	return fmt.Sprintf("%d%%", frac)
}

// hasAddress returns whether `x` is in `arr`.
func hasAddress(arr []common.Address, x common.Address) bool {
	for _, val := range arr {
		if x == val {
			return true
		}
	}
	return false
}
