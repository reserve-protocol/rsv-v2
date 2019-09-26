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
        tokens = new address[](_tokens.length);

        for (uint i = 0; i < _tokens.length; i++) {
            weights[address(_tokens[i])] = _weights[i];
            hasToken[address(_tokens[i])] = true;
            tokens[i] = address(_tokens[i]);
        }

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
