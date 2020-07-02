pragma solidity 0.5.7;

import "./IRSV.sol";
import "../ownership/Ownable.sol";
import "../zeppelin/utils/ECDSA.sol";

/**
 * @title The Reserve Relayer Contract
 * @dev A contract to support metatransactions via ECDSA signature verification.
 *
 */
contract Relayer is Ownable {

    IRSV public trustedRSV;
    mapping(address => uint) public nonce;

    event RSVChanged(address indexed oldRSVAddr, address indexed newRSVAddr);

    event TransferForwarded(
        bytes sig,
        address indexed from,
        address indexed to,
        uint256 indexed amount,
        uint256 fee
    );
    event TransferFromForwarded(
        bytes sig,
        address indexed holder,
        address indexed spender,
        address indexed to,
        uint256 amount,
        uint256 fee
    );
    event ApproveForwarded(
        bytes sig,
        address indexed holder,
        address indexed spender,
        uint256 amount,
        uint256 fee
    );
    event FeeTaken(address indexed from, address indexed to, uint256 indexed value);

    constructor(address rsvAddress) public {
        trustedRSV = IRSV(rsvAddress);
    }

    /// Set the Reserve contract address.
    function setRSV(address newTrustedRSV) external onlyOwner {
        emit RSVChanged(address(trustedRSV), newTrustedRSV);
        trustedRSV = IRSV(newTrustedRSV);
    }

    /// Forward a signed `transfer` call to the RSV contract if `sig` matches the signature.
    /// Note that `amount` is not reduced by `fee`; the fee is taken separately.
    function forwardTransfer(
        bytes calldata sig,
        address from,
        address to,
        uint256 amount,
        uint256 fee
    )
        external
    {
        bytes32 hash = keccak256(abi.encodePacked(
            address(trustedRSV),
            "forwardTransfer",
            from,
            to,
            amount,
            fee,
            nonce[from]
        ));
        nonce[from]++;

        address recoveredSigner = _recoverSignerAddress(hash, sig);
        require(recoveredSigner == from, "invalid signature");

        _takeFee(from, fee);

        require(
            trustedRSV.relayTransfer(from, to, amount), 
            "Reserve.sol relayTransfer failed"
        );
        emit TransferForwarded(sig, from, to, amount, fee);
    }

    /// Forward a signed `approve` call to the RSV contract if `sig` matches the signature.
    /// Note that `amount` is not reduced by `fee`; the fee is taken separately.
    function forwardApprove(
        bytes calldata sig,
        address holder,
        address spender,
        uint256 amount,
        uint256 fee
    )
        external
    {
        bytes32 hash = keccak256(abi.encodePacked(
            address(trustedRSV),
            "forwardApprove",
            holder,
            spender,
            amount,
            fee,
            nonce[holder]
        ));
        nonce[holder]++;

        address recoveredSigner = _recoverSignerAddress(hash, sig);
        require(recoveredSigner == holder, "invalid signature");

        _takeFee(holder, fee);

        require(
            trustedRSV.relayApprove(holder, spender, amount), 
            "Reserve.sol relayApprove failed"
        );
        emit ApproveForwarded(sig, holder, spender, amount, fee);
    }

    /// Forward a signed `transferFrom` call to the RSV contract if `sig` matches the signature.
    /// Note that `fee` is not deducted from `amount`, but separate.
    /// Allowance checking is left up to the Reserve contract to do.
    function forwardTransferFrom(
        bytes calldata sig,
        address holder,
        address spender,
        address to,
        uint256 amount,
        uint256 fee
    )
        external
    {
        bytes32 hash = keccak256(abi.encodePacked(
            address(trustedRSV),
            "forwardTransferFrom",
            holder,
            spender,
            to,
            amount,
            fee,
            nonce[spender]
        ));
        nonce[spender]++;

        address recoveredSigner = _recoverSignerAddress(hash, sig);
        require(recoveredSigner == spender, "invalid signature");

        _takeFee(spender, fee);

        require(
            trustedRSV.relayTransferFrom(holder, spender, to, amount), 
            "Reserve.sol relayTransfer failed"
        );
        emit TransferFromForwarded(sig, holder, spender, to, amount, fee);
    }

    /// Recover the signer's address from the hash and signature.
    function _recoverSignerAddress(bytes32 hash, bytes memory sig)
        internal pure
        returns (address)
    {
        bytes32 ethMessageHash = ECDSA.toEthSignedMessageHash(hash);
        return ECDSA.recover(ethMessageHash, sig);
    }

    /// Transfer a fee from payer to sender.
    function _takeFee(address payer, uint256 fee) internal {
        if (fee > 0) {
            require(trustedRSV.relayTransfer(payer, msg.sender, fee), "fee transfer failed");
            emit FeeTaken(payer, msg.sender, fee);
        }
    }

}
