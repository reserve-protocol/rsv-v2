// +build regular

package tests

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/reserve-protocol/rsv-beta/abi"
)

func TestBasket(t *testing.T) {
	suite.Run(t, new(BasketSuite))
}

type BasketSuite struct {
	TestSuite
	weights []*big.Int
}

var (
	// Compile-time check that BasketSuite implements the interfaces we think it does.
	// If it does not implement these interfaces, then the corresponding setup and teardown
	// functions will not actually run.
	_ suite.BeforeTest       = &BasketSuite{}
	_ suite.SetupAllSuite    = &BasketSuite{}
	_ suite.TearDownAllSuite = &BasketSuite{}
)

// SetupSuite runs once, before all of the tests in the suite.
func (s *BasketSuite) SetupSuite() {
	s.setup()
}

// BeforeTest runs before each test in the suite.
func (s *BasketSuite) BeforeTest(suiteName, testName string) {
	// Deploy collateral ERC20s
	s.erc20s = make([]*abi.BasicERC20, 3)
	s.erc20Addresses = make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		erc20Address, _, erc20, err := abi.DeployBasicERC20(s.signer, s.node)
		s.Require().NoError(err)
		s.erc20s[i] = erc20
		s.erc20Addresses[i] = erc20Address
	}

	s.weights = []*big.Int{shiftLeft(1, 36), shiftLeft(2, 36), shiftLeft(3, 36)}

	// Make a simple basket
	basketAddress, tx, basket, err := abi.DeployBasket(
		s.signer,
		s.node,
		zeroAddress(),
		s.erc20Addresses,
		s.weights,
	)

	s.requireTxWithStrictEvents(tx, err)()
	s.basketAddress = basketAddress
	s.basket = basket
}

// TestState checks to make sure state is set up correctly after construction.
func (s *BasketSuite) TestState() {
	// Check that all variables in state are set correctly.
	for i, address := range s.erc20Addresses {
		foundAddress, err := s.basket.Tokens(nil, bigInt(uint32(i)))
		s.Require().NoError(err)
		s.Equal(address, foundAddress)

		foundWeight, err := s.basket.Weights(nil, address)
		s.Require().NoError(err)
		s.Equal(s.weights[i].String(), foundWeight.String())

		foundHas, err := s.basket.Has(nil, address)
		s.Require().NoError(err)
		s.Equal(true, foundHas)
	}
}

// TestViews checks to make sure the view functions work as expected.
func (s *BasketSuite) TestViews() {
	// `getTokens` function.
	tokens, err := s.basket.GetTokens(nil)
	s.Require().NoError(err)
	s.True(reflect.DeepEqual(s.erc20Addresses, tokens))

	// `size` function.
	size, err := s.basket.Size(nil)
	s.Require().NoError(err)
	s.Equal(bigInt(uint32(len(s.erc20Addresses))).String(), size.String())

	// `has` should return false for tokens not in the basket.
	foundHas, err := s.basket.Has(nil, s.account[3].address())
	s.Require().NoError(err)
	s.Equal(false, foundHas)
}

// TestSuccessiveBasketWithEmptyParams tries deploying a second basket from a different account.
// This basket has no tokens, so should carry over tokens from the first basket.
func (s *BasketSuite) TestSuccessiveBasketWithEmptyParams() {
	deployer := s.account[1]

	var emptyTokens []common.Address
	var emptyWeights []*big.Int
	// Deploy a new basket from a different account, but based off the first basket.
	_, tx, basket, err := abi.DeployBasket(
		signer(deployer),
		s.node,
		s.basketAddress,
		emptyTokens,
		emptyWeights,
	)

	s.requireTxWithStrictEvents(tx, err)()

	// Our two baskets should be identical in every way.
	for i, _ := range s.erc20Addresses {
		// State
		firstToken, err := s.basket.Tokens(nil, bigInt(uint32(i)))
		s.Require().NoError(err)
		secondToken, err := basket.Tokens(nil, bigInt(uint32(i)))
		s.Require().NoError(err)
		s.Equal(firstToken, secondToken)

		firstWeight, err := s.basket.Weights(nil, firstToken)
		s.Require().NoError(err)
		secondWeight, err := basket.Weights(nil, firstToken)
		s.Require().NoError(err)
		s.Equal(firstWeight.String(), secondWeight.String())

		firstHas, err := s.basket.Has(nil, firstToken)
		s.Require().NoError(err)
		secondHas, err := basket.Has(nil, firstToken)
		s.Require().NoError(err)
		s.Equal(firstHas, secondHas)
	}

	// `getTokens()`
	firstTokens, err := s.basket.GetTokens(nil)
	s.Require().NoError(err)
	secondTokens, err := basket.GetTokens(nil)
	s.Require().NoError(err)
	s.True(reflect.DeepEqual(firstTokens, secondTokens))

	// `size()`
	firstSize, err := s.basket.Size(nil)
	s.Require().NoError(err)
	secondSize, err := basket.Size(nil)
	s.Equal(firstSize, secondSize)
}

// TestSuccessiveBasketWithAdditionalParams deploys a 2nd basket with a new token, and one token
// that overlaps with the previous basket.
func (s *BasketSuite) TestSuccessiveBasketWithAdditionalTokens() {
	deployer := s.account[1]
	newToken := s.account[2].address()
	newWeight := bigInt(uint32(9))
	recurringToken := s.erc20Addresses[0]
	recurringTokenNewWeight := bigInt(uint32(10))

	moreTokens := []common.Address{newToken, recurringToken}
	moreWeights := []*big.Int{newWeight, recurringTokenNewWeight}

	// Deploy a new basket from a different account, but based off the first basket.
	_, tx, newBasket, err := abi.DeployBasket(
		signer(deployer),
		s.node,
		s.basketAddress,
		moreTokens,
		moreWeights,
	)

	s.requireTxWithStrictEvents(tx, err)()

	// The second newBasket should be bigger by 1.
	firstSize, err := s.basket.Size(nil)
	s.Require().NoError(err)
	secondSize, err := newBasket.Size(nil)
	s.Equal(bigInt(0).Add(firstSize, bigInt(1)), secondSize)

	// The token lists should differ by 1 token address.
	firstTokens, err := s.basket.GetTokens(nil)
	s.Require().NoError(err)
	secondTokens, err := newBasket.GetTokens(nil)
	s.Require().NoError(err)
	expectedTokens := []common.Address{newToken, recurringToken}
	for _, tok := range firstTokens {
		if tok != recurringToken {
			expectedTokens = append(expectedTokens, tok)
		}
	}
	s.True(reflect.DeepEqual(expectedTokens, secondTokens))

	// The new token should have the right weight.
	weight, err := newBasket.Weights(nil, newToken)
	s.Require().NoError(err)
	s.Equal(newWeight.String(), weight.String())

	// The recurring token should have the new weight, not the old one.
	weight, err = newBasket.Weights(nil, recurringToken)
	s.Require().NoError(err)
	s.Equal(recurringTokenNewWeight.String(), weight.String())

	// After that, our two baskets should be identical in every way for the 3 original tokens, except for
	// the recurring token's new value.
	for i, _ := range s.erc20Addresses {
		firstToken, err := s.basket.Tokens(nil, bigInt(uint32(i)))
		s.Require().NoError(err)
		secondToken, err := newBasket.Tokens(nil, bigInt(uint32(i+1)))
		s.Require().NoError(err)
		s.Equal(firstToken, secondToken)

		firstWeight, err := s.basket.Weights(nil, firstToken)
		s.Require().NoError(err)
		secondWeight, err := newBasket.Weights(nil, firstToken)
		s.Require().NoError(err)
		if firstToken == recurringToken {
			s.NotEqual(firstWeight.String(), secondWeight.String())
		} else {
			s.Equal(firstWeight.String(), secondWeight.String())
		}

		firstHas, err := s.basket.Has(nil, firstToken)
		s.Require().NoError(err)
		secondHas, err := newBasket.Has(nil, firstToken)
		s.Require().NoError(err)
		s.Equal(firstHas, secondHas)
	}

	// Finally the new basket should have the new token.
	newHas, err := newBasket.Has(nil, newToken)
	s.Require().NoError(err)
	s.Equal(true, newHas)
}

// TestNegativeCases checks to make sure invalid basket constructions revert.
func (s *BasketSuite) TestNegativeCases() {
	// Case 1: Tokens is longer than Weights.
	deployer := s.account[1]
	tokens := s.erc20Addresses
	weights := []*big.Int{bigInt(0)}
	_, tx, _, err := abi.DeployBasket(
		signer(deployer),
		s.node,
		s.basketAddress,
		tokens,
		weights,
	)
	s.requireTxFails(tx, err)

	// Case 2: Weights is longer than Tokens.
	tokens = []common.Address{s.account[1].address()}
	weights = []*big.Int{bigInt(1), bigInt(2)}
	_, tx, _, err = abi.DeployBasket(
		signer(deployer),
		s.node,
		s.basketAddress,
		tokens,
		weights,
	)
	s.requireTxFails(tx, err)

	// Case 3: Basket is too big after addition of tokens from old address.
	var longTokens []common.Address
	var longWeights []*big.Int
	for i := 0; i < 98; i++ {
		longTokens = append(longTokens, common.BigToAddress(bigInt(uint32(100+i))))
		longWeights = append(longWeights, bigInt(1))
	}
	_, tx, _, err = abi.DeployBasket(
		signer(deployer),
		s.node,
		s.basketAddress,
		longTokens,
		longWeights,
	)
	s.requireTxFails(tx, err)

	// Case 4: Basket is too big, even with the zero address basket as its prev.
	var extraLongTokens []common.Address
	var extraLongWeights []*big.Int
	for i := 0; i < 101; i++ {
		extraLongTokens = append(extraLongTokens, common.BigToAddress(bigInt(uint32(100+i))))
		extraLongWeights = append(extraLongWeights, bigInt(1))
	}
	_, tx, _, err = abi.DeployBasket(
		signer(deployer),
		s.node,
		zeroAddress(),
		extraLongTokens,
		extraLongWeights,
	)
	s.requireTxFails(tx, err)

	// Case 5: PrevBasket is not actually a basket.
	tokens = []common.Address{s.account[2].address()}
	weights = []*big.Int{bigInt(1)}
	_, tx, _, err = abi.DeployBasket(
		signer(deployer),
		s.node,
		s.account[3].address(),
		tokens,
		weights,
	)
	s.requireTxFails(tx, err)

	// Case 6: Duplicate tokens in the basket.
	tokens = []common.Address{s.account[2].address(), s.account[2].address()}
	weights = []*big.Int{bigInt(1), bigInt(2)}
	_, tx, _, err = abi.DeployBasket(
		signer(deployer),
		s.node,
		s.basketAddress,
		tokens,
		weights,
	)
	s.requireTxFails(tx, err)
}
