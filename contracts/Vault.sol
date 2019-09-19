pragma solidity ^0.5.8;

import "./zeppelin/contracts/token/ERC20/SafeERC20.sol";
import "./zeppelin/contracts/math/SafeMath.sol";
import "./Ownable.sol";

/**
* The Vault contract has an owner who is able to set the manager. The manager is
* able to perform withdrawals. 
*/
contract Vault is Ownable {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    // Auth role
    address public manager;

    event ManagerTransferred(
        address indexed previousManager,
        address indexed newManager
    );

    event TokenWithdraw(address indexed token, uint256 indexed amount);

    constructor() public {
        // Initialize manager as msg.sender
        manager = msg.sender;
    }

    modifier onlyManager() {
        require(msg.sender == manager, "must be manager");
        _;
    }

    /**
     * @dev Allows the current owner to transfer control of the contract to a newManager.
     * @param newManager The address to transfer manager control to.
     */
    function transferManager(address newManager) external onlyOwner {
        _transferManager(newManager);
    }

    /**
     * @dev Transfers the manager control to newManager.
     * @param newManager The address to transfer manager control to.
     */
    function _transferManager(address newManager) internal {
        require(newManager != address(0));
        manager = newManager;
        emit ManagerTransferred(manager, newManager);
    }

    function batchWithdrawTo(address[] calldata tokens, uint256[] calldata amounts, address to) external onlyManager {
        for (uint i = 0; i < tokens.length; i++) {
            if (amounts[i] > 0) {
                IERC20(tokens[i]).safeTransfer(to, amounts[i]);
                emit TokenWithdraw(tokens[i], amounts[i]);
            }
        }        
    }
}
