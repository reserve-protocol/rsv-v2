pragma solidity 0.5.7;

import "../zeppelin/GSN/Context.sol";
import "../zeppelin/math/SafeMath.sol";
/**
 * @dev Contract module which provides a basic access control mechanism, where there is an account
 * (owner) that can be granted exclusive access to specific functions.
 *
 * This module is used through inheritance by using the modifier `onlyOwner`.
 *
 * To change ownership, use a 2-part nominate-accept pattern.
 *
 * To renounce ownership, use a renounce-wait-renounce pattern.
 *
 * This contract is loosely based off of https://git.io/JenNF but additionally requires new owners
 * to accept ownership before the transition occurs.
 */
contract Ownable is Context {
    using SafeMath for uint256;

    address private _owner;
    address public _nominatedOwner;
    uint256 private _abdicationTime;

    event NewOwnerNominated(address indexed previousOwner, address indexed newOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event AbdicationTimeSet(uint256 indexed time);

    /// Initialize this contract; set the deployer as the initial owner.
    constructor () internal {
        address msgSender = _msgSender();
        _owner = msgSender;
        emit OwnershipTransferred(address(0), msgSender);
    }

    /// Return the address of the current owner.
    function owner() public view returns (address) {
        return _owner;
    }

    /// Return the time when the owner may abdicate. Zero means that no time has been set, so
    /// the owner cannot abdicate in less than a week.
    function abdicationTime() public view returns (uint256) {
        return _abdicationTime;
    }

    /// Only allow the owner to call the modified function.
    modifier onlyOwner() {
        _onlyOwner();
        _;
    }

    function _onlyOwner() internal view {
        require(_msgSender() == _owner, "caller is not owner");
    }


    /// Nominate a new owner `newOwner`. Needs a follow-up `acceptOwnership` signed by `newOwner`.
    function nominateNewOwner(address newOwner) external onlyOwner {
        require(newOwner != address(0), "new owner is 0 address");
        emit NewOwnerNominated(_owner, newOwner);
        _nominatedOwner = newOwner;
    }

    /// Accept ownership of the contract.
    /// Can be called only by the nominated owner, sometime following `nominateNewOwner`.
    function acceptOwnership() external {
        require(_nominatedOwner == _msgSender(), "unauthorized");
        require(_nominatedOwner != address(0), "cannot accept for 0 address");
        emit OwnershipTransferred(_owner, _nominatedOwner);
        _owner = _nominatedOwner;
    }


    /// Announce intention to abdicate, on or after `time`.
    /// Call setAbdicationTime(0) to cancel a previous abdication announcement.
    function setAbdicationTime(uint256 time) external onlyOwner {
        require(time == 0 || time >= now.add(1 weeks), "nonzero time < 1 week from now");
        emit AbdicationTimeSet(time);
        _abdicationTime = time;
    }

    /// Abdicate ownership; set `_owner` to the 0 address.
    /// Only do this to deliberately lock in the current permissions.
    /// Only the owner can do this, and only after _abdicationTime has been set for a week.
    function abdicateOwnership() external onlyOwner {
        require(_abdicationTime != 0, "must announce abdication first");
        require(now >= _abdicationTime, "cannot abdicate yet");
        emit OwnershipTransferred(_owner, address(0));
        _owner = address(0);
    }
}
