pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/SafeERC20.sol";

/**
 * The Basket contract defines what an arbitrary "front-token" is backed by.
 *
 * Most importantly, the `backing` quantities correspond to quantities
 * for a single front-token, NOT for a single atto-front-token. 
*/
contract Basket {
    using SafeMath for uint256;

    uint8 public frontTokenDecimals;
    address[] public tokens;
    uint256[] public backing; // how much of each token is a single front token worth
    uint256 public size;
    mapping(address => uint256) public backingMap;


    constructor(
        address[] memory _tokens, 
        uint256[] memory _backing,
        uint8 _frontTokenDecimals
    ) 
        public 
    {
        require(_tokens.length == _backing.length, "invalid basket");
        require(_tokens.length > 0 && _tokens.length <= 100, "basket bad length");

        tokens = _tokens;
        backing = _backing;
        frontTokenDecimals = _frontTokenDecimals;
    }

    function getTokens() external view returns(address[] memory) {
        return tokens;
    }

    /// Calculates the quantities of tokens required to back `_frontTokenSupply`. 
    function quantitiesRequired(uint256 _frontTokenSupply) external view returns(uint256[] memory) {
        uint256[] memory tokenQuantities = new uint256[](size);

        for (uint i = 0; i < size; i++) {
            tokenQuantities[i] = _frontTokenSupply.mul(backing[i]).div(frontTokenDecimals);
        }

        return tokenQuantities;
    }

    /// Calculates what quantities of tokens are needed to reach `_other` at `_frontTokenSupply`.
    function newQuantitiesRequired(uint256 _frontTokenSupply, Basket _other) external view returns(uint256[] memory) {
        uint256[] memory required = new uint256[](size);

        // Calculate required in terms of backing quantities, that is, per single front token. 
        for (uint i = 0; i < size; i++) {
            if (_other.backingMap(tokens[i]) > backingMap[tokens[i]]) {
                required[i] = _other.backingMap(tokens[i]).sub(backingMap[tokens[i]]);
            }
        }

        // Multiply by `_frontTokenSupply` to get total quantities.
        for (uint i = 0; i < size; i++) {
            required[i] = 
                _frontTokenSupply.mul(required[i]).div(frontTokenDecimals);
        }

        return required;
    }
}
