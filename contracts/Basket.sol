pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/SafeERC20.sol";

/**
 * The Basket contract defines what an arbitrary "front-token" is backed by.
 *
 * Most importantly, the `backing` quantities correspond to quantities
 * for a single front-token, NOT for a single atto-front-token. 
*/

/* TODO(elder): instead of parallel lists (which are error-prone), use structs to pass around
 * tokens and their in/out-quantities or their new weights. If Solidity permits. */
contract Basket {
    using SafeMath for uint256;

    uint8 public frontTokenDecimals;
    address[] public tokens;
    mapping(address => uint256) public backingMap; // TODO(elder): rename to "weights"

    constructor(
        address[] memory _tokens, 
        uint256[] memory _backing,
        uint8 _frontTokenDecimals
    ) 
        public 
    {
        require(_tokens.length == _backing.length, "invalid basket");
        require(_tokens.length > 0 && _tokens.length <= 100, "basket bad length");

        for (uint i = 0; i < _tokens.length; i++) {
            backingMap[_tokens[i]] = _backing[i];
        }
        tokens = _tokens;
        frontTokenDecimals = _frontTokenDecimals;
    }

    function getTokens() external view returns(address[] memory) {
        return tokens;
    }

    function getSize() external view returns(uint) { // TODO(elder): just "size()"?
        return tokens.length;
    }

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
}
