pragma solidity ^0.5.8;

import "./zeppelin/GSN/Context.sol";
/**
 * @dev Contract module which provides a basic access control mechanism, where
 * there is an account (owner) that can be granted exclusive access to
 * specific functions.
 *
 * This module is used through inheritance by using the modifier `onlyOwner`.
 * 
 * To change ownership, use a 2-part nominate-accept pattern.
 * 
 * This contract is loosely based off of (https://github.com/OpenZeppelin/openzeppelin-contracts/blob/6f8e672f3fcb93289fb559ecbef72b8fd1cd56e1/contracts/ownership/Ownable.sol) but additionally
 * requires new owners to accept ownership before the transition occurs. 
 */
contract Ownable is Context {
    address public _owner;
    address public _nominatedOwner;

    event NewOwnerNominated(address indexed previousOwner, address indexed newOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    /**
     * @dev Initializes the contract setting the deployer as the initial owner.
     */
    constructor () internal {
        address msgSender = _msgSender();
        _owner = msgSender;
        emit OwnershipTransferred(address(0), msgSender);
    }

    /**
     * @dev Throws if called by any account other than the owner.
     */
    modifier onlyOwner() {
        require(_msgSender() == _owner, "Ownable: caller is not the owner");
        _;
    }

    /**
     * @dev Nominates a new owner `newOwner`.
     * Requires a follow-up `acceptOwnership`. 
     * Can only be called by the current owner.
     */
    function nominateNewOwner(address newOwner) external onlyOwner {
        require(newOwner != address(0), "Ownable: new owner is the zero address");
        emit NewOwnerNominated(_owner, newOwner);
        _nominatedOwner = newOwner;
    }

    /**
     * @dev Accepts ownership of the contract.
     */
    function acceptOwnership() external {
        require(_nominatedOwner == _msgSender(), "Ownable: new owner is the zero address");
        emit OwnershipTransferred(_owner, _nominatedOwner);
        _owner = _nominatedOwner;
    }

    /** Set `_owner` to the 0 address.
     * Only do this to deliberately lock in the current permissions.
     */
    function renounceOwnership() external onlyOwner {
        emit OwnershipTransferred(_owner, address(0));
        _owner = address(0);
    }
}
