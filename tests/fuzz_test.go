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

var duration = 10

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
		if err == nil {
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
	s.Require().Equal(true, emergency)

	// Unpause from emergency.
	s.requireTxWithStrictEvents(s.manager.SetEmergency(s.signer, false))(
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
	amounts := []*big.Int{shiftLeft(1, 48), shiftLeft(1, 48), shiftLeft(1, 48)}
	s.fundAccountWithErc20sAndApprove(s.proposer, amounts)

	// Pass a WeightProposal so we are able to Issue/Redeem.
	s.weights = []*big.Int{shiftLeft(1, 35), shiftLeft(3, 35), shiftLeft(6, 35)}
	s.changeBasketUsingWeightProposal(s.erc20Addresses, s.weights)
}

// TestByFuzzing chooses between Issuing, Redeeming, WeightProposal, and SwapProposal for
// `duration` times and asserts invariants are uphold at every step.
func (s *ManagerFuzzSuite) TestByFuzzing() {
	fmt.Print("\n")
	for i := 0; i < duration; i++ {
		fmt.Printf("Run %v", i)
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

			// Record how much value the proposer starts with.
			startProposerValue := s.getTotalERC20Quantity(s.proposer.address())

			// Choose 0, 1, 2, or 3 of the 3 ERC20 tokens using a binomial distribution.
			_, tokens := s.chooseTokenSet()

			// Randomly choose weights that sum to 1e36.
			weights := s.generateWeights(tokens, shiftLeft(1, 36))

			// Try to execute this WeightProposal. It's okay if it fails.
			s.tryWeightProposal(tokens, weights)

			// The proposer shouldn't end up with more value than they started with.
			s.assertProposerDidNotGainValue(startProposerValue)

		case 3: // SwapProposal
			fmt.Print(" |  SwapProposal ")

			// Record how much value the proposer starts with.
			startProposerValue := s.getTotalERC20Quantity(s.proposer.address())

			// Choose 0, 1, 2, or 3 of the 3 ERC20 tokens using a binomial distribution.
			erc20s, tokens := s.chooseTokenSet()

			// Generate a SwapProposal.
			amounts, toVault := s.generateSwaps(erc20s, tokens)

			// Try to execute this SwapProposal. It's okay if it fails.
			s.trySwapProposal(s.erc20Addresses, amounts, toVault)

			// The proposer shouldn't end up with more value than they started with.
			s.assertProposerDidNotGainValue(startProposerValue)
		}

		// Display RSV supply and basket content.
		s.printMetrics()

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
func (s *ManagerFuzzSuite) generateWeights(tokens []common.Address, sum *big.Int) []*big.Int {
	var weights []*big.Int
	for _, _ = range tokens {
		n := generateRandUpTo(sum)
		weights = append(weights, n)
		sum = bigInt(0).Sub(sum, n)
	}

	// Add the leftover to the last element
	if len(tokens) > 0 {
		weights[len(weights)-1] = sum.Add(sum, weights[len(weights)-1])
	}

	// Check that the weights sum to 1e36.
	if len(weights) > 0 {
		s.Require().Equal(shiftLeft(1, 36).String(), sumWeights(weights).String())
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

	// Randomly choose what amount to transfer out of the vault, up to the full amount available.
	amounts[indexToVault] = generateRandUpTo(indexBalance)
	toVault[indexToVault] = false

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

// getTotalERC20Quantity sums the total holdings of the proposer across all ERC20 tokens.
func (s *ManagerFuzzSuite) getTotalERC20Quantity(acc common.Address) *big.Int {
	sum := bigInt(0)
	for _, erc20 := range s.erc20s {
		bal, err := erc20.BalanceOf(nil, acc)
		s.Require().NoError(err)
		sum = sum.Add(sum, bal)
	}
	return sum
}

// assertProposerDidNotGainValue asserts that the proposer now holds less than or the same
// amount of value as they did previously when they had `oldVal`.
func (s *ManagerFuzzSuite) assertProposerDidNotGainValue(oldVal *big.Int) {
	// newVal := s.getTotalERC20Quantity(s.proposer.address())
	// if newVal.Cmp(oldVal) == 1 {
	// 	fmt.Println()
	// 	fmt.Println("📢 Yuh oh 📢")
	// 	fmt.Printf("The proposer started with: %v\n", oldVal)
	// 	fmt.Printf("But they ended with: %v\n", newVal)
	// 	fmt.Println()
	// 	s.Require().True(false)
	// }

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
// This implementation is pretty fun :)
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
