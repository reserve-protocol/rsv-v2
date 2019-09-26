pragma solidity ^0.5.8;

/**
 * The Basket contract defines the backing weights for a what a  is backed by.
 *
 * Most importantly, the `backing` quantities correspond to quantities
 * for a single front-token, NOT for a single atto-front-token. 
*/

contract Basket {
    address[] public tokens;
    mapping(address => uint256) public weights;
    mapping(address => bool) public hasToken;
    // INVARIANT: {addr | addr in tokens} == {addr | hasToken[addr] == true}

    // SECURITY PROPERTY: The value of prev is always a Basket, and cannot be set by any user.
    constructor(Basket prev, address[] memory _tokens, uint256[] memory _weights) public {
        require(_tokens.length == _weights.length, "Basket: unequal array lengths");
        require(_tokens.length > 0 && _tokens.length <= 100, "Basket: bad length");
        
        for (uint i = 0; i < _tokens.length; i++) {
            weights[_tokens[i]] = _weights[i];
            hasToken[_tokens[i]] = true;
        }
        tokens = _tokens;

        // If a previous basket is specified, copy over its contents where they were not already set.
        if (prev != Basket(0)) {
            for (uint i = 0; i < prev.size(); i++) {
                address tok = prev.tokens(i);
                if (!hasToken[tok]) {
                    weights[tok] = prev.weights(tok);
                    hasToken[tok] = true;
                    tokens.push(tok);
                }
            }
        }
    }

    function getTokens() external view returns(address[] memory) {
        return tokens;
    }

    function size() external view returns(uint) {
        return tokens.length;
    }
}

/*
    /// Calculates the quantities of tokens required to back `_frontTokenSupply`. 
    function quantitiesRequired(uint256 _frontTokenSupply) external view returns(uint256[] memory) {
        uint256[] memory tokenQuantities = new uint256[](tokens.length);

        for (uint i = 0; i < tokens.length; i++) {
            tokenQuantities[i] = 
                _frontTokenSupply.mul(backingMap[tokens[i]]).div(frontTokenDecimals);
        }

        return tokenQuantities;
    }

    /// Calculates what quantities of tokens are needed to reach `_other` at `_frontTokenSupply`.
    function newQuantitiesRequired(uint256 _frontTokenSupply, Basket _other)
        external view returns(uint256[] memory) {
        uint256[] memory required = new uint256[](tokens.length);

        // Calculate required in terms of backing quantities, that is, per single front token. 
        for (uint i = 0; i < tokens.length; i++) {
            if (_other.backingMap(tokens[i]) > backingMap[tokens[i]]) {
                required[i] = _other.backingMap(tokens[i]).sub(backingMap[tokens[i]]);
            }
        }

        // Multiply by `_frontTokenSupply` to get total quantities.
        for (uint i = 0; i < tokens.length; i++) {
            required[i] = 
                _frontTokenSupply.mul(required[i]).div(frontTokenDecimals);
        }

        return required;
    }
*/
