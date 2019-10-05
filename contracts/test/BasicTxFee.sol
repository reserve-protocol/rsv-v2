pragma solidity 0.5.7;

import "../zeppelin/math/SafeMath.sol";


/**
 * Simple Transaction Fee contract for testing. 
 */
contract BasicTxFee {

    uint256 fee;
    constructor(uint256 _fee) public {
        fee = _fee;
    }

    function calculateFee(address, address, uint256) external returns(uint256) {
        return fee;
    }
}
