pragma solidity 0.5.7;

import "../ownership/Ownable.sol";
import "../zeppelin/math/SafeMath.sol";

/**
 * @title Eternal Storage for the Reserve Token
 *
 * @dev Eternal Storage facilitates future upgrades.
 *
 * If Reserve chooses to release an upgraded contract for the Reserve in the future, Reserve will
 * have the option of reusing the deployed version of this data contract to simplify migration.
 *
 * The use of this contract does not imply that Reserve will choose to do a future upgrade, nor
 * that any future upgrades will necessarily re-use this storage. It merely provides option value.
 */
contract ReserveEternalStorage is Ownable {

    using SafeMath for uint256;


    // ===== auth =====

    address public reserveAddress;

    event ReserveAddressTransferred(
        address indexed oldReserveAddress,
        address indexed newReserveAddress
    );

    /// On construction, set auth fields.
    constructor() public {
        reserveAddress = _msgSender();
        emit ReserveAddressTransferred(address(0), reserveAddress);
    }

    /// Only run modified function if sent by `reserveAddress`.
    modifier onlyReserveAddress() {
        require(_msgSender() == reserveAddress, "onlyReserveAddress");
        _;
    }

    /// Set `reserveAddress`.
    function updateReserveAddress(address newReserveAddress) external {
        require(_msgSender() == reserveAddress || _msgSender() == owner(), "not authorized");
        emit ReserveAddressTransferred(reserveAddress, newReserveAddress);
        reserveAddress = newReserveAddress;
    }



    // ===== balance =====

    mapping(address => uint256) public balance;

    /// Add `value` to `balance[key]`, unless this causes integer overflow.
    ///
    /// @dev This is a slight divergence from the strict Eternal Storage pattern, but it reduces
    /// the gas for the by-far most common token usage, it's a *very simple* divergence, and
    /// `setBalance` is available anyway.
    function addBalance(address key, uint256 value) external onlyReserveAddress {
        balance[key] = balance[key].add(value);
    }

    /// Subtract `value` from `balance[key]`, unless this causes integer underflow.
    function subBalance(address key, uint256 value) external onlyReserveAddress {
        balance[key] = balance[key].sub(value);
    }

    /// Set `balance[key]` to `value`.
    function setBalance(address key, uint256 value) external onlyReserveAddress {
        balance[key] = value;
    }



    // ===== allowed =====

    mapping(address => mapping(address => uint256)) public allowed;

    /// Set `to`'s allowance of `from`'s tokens to `value`.
    function setAllowed(address from, address to, uint256 value) external onlyReserveAddress {
        allowed[from][to] = value;
    }
}
