pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/math/SafeMath.sol";
import "./ownership/Ownable.sol";

/**
* The Vault contract has an owner who is able to set the manager. The manager is
* able to perform withdrawals. 
*/
contract Vault is Ownable {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    address public manager;

    event ManagerTransferred(
        address indexed previousManager,
        address indexed newManager
    );

    constructor() public {
        // Initialize manager as _msgSender()
        manager = _msgSender();
        emit ManagerTransferred(address(0), manager);
    }

    /// Changes the manager account. 
    function changeManager(address newManager) external onlyOwner {
        require(newManager != address(0));
        emit ManagerTransferred(manager, newManager);
        manager = newManager;
    }

    /// Withdraw `amount` of `token` to address `to`. Only callable by `manager`.
    function withdrawTo(address token, uint256 amount, address to) external {
        require(_msgSender() == manager, "must be manager");
        IERC20(token).safeTransfer(to, amount);
    }
}
