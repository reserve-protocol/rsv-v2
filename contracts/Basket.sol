pragma solidity ^0.5.8;

import "./zeppelin/contracts/token/ERC20/SafeERC20.sol";


contract Basket {
    using SafeMath for uint256;

    address public proposer;
    address[] public tokens;
    uint256[] public amounts;
    uint256 public size;
    mapping(address => uint256) public amountMap;


    constructor(
        address _proposer, 
        address[] memory _tokens, 
        uint256[] memory _amounts, 
        uint256 _amountsSum
    ) 
        public 
    {
        require(_tokens.length == _amounts.length, "invalid basket");
        require(_tokens.length > 0, "basket too small");
        require(_tokens.length <= 1000, "basket too big");

        uint256 sum;
        for (uint i = 0; i < _tokens.length; i++) {
            sum += _amounts[i];
            amountMap[_tokens[i]] = _amounts[i];
        }
        require(sum > 0, "basket cannot be empty");
        require(sum == _amountsSum, "amounts must sum to the expected amount");

        proposer = _proposer;
        tokens = _tokens;
        amounts = _amounts;
    }

    function getTokens() external view returns(address[] memory) {
        return tokens;
    }

    /// Calculates the excess amounts in this basket relative to some other basket.
    function excessAmountsRelativeToOtherBasket(Basket _other) external view returns(uint256[] memory) {
        uint256[] memory excess = new uint256[](size);

        for (uint i = 0; i < size; i++) {
            if (amountMap[tokens[i]] > _other.amountMap(tokens[i])) {
                excess[i] = amountMap[tokens[i]].sub(_other.amountMap(tokens[i]));
            }
        }

        return excess;
    }

}
