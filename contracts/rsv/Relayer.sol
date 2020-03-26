pragma solidity 0.5.7;

import "./IRSV.sol";
import "../ownership/Ownable.sol";
import "../zeppelin/utils/ECDSA.sol";

contract Relayer is Ownable {

    IRSV public trustedRSV;
    mapping(address => uint) public nonce;

    event RSVChanged(address indexed oldRSVAddr, address indexed newRSVAddr);
    event TransferForwarded(bytes sig, address indexed from, address indexed to, uint256 indexed amount, uint256 fee);
    event TransferFromForwarded(bytes sig, address indexed spender, address indexed holder, address indexed to, uint256 amount, uint256 fee);
    event ApproveForwarded(bytes sig, address indexed holder, address indexed spender, uint256 indexed amount, uint256 fee);
    event FeeTaken(address indexed from, address indexed to, uint256 indexed value);

    constructor() public {}

    /// Set the Reserve contract address.
    function setRSV(address newTrustedRSV) external onlyOwner {
        emit RSVChanged(address(trustedRSV), newTrustedRSV);
        trustedRSV = IRSV(newTrustedRSV);
    }

    // note that `fee` is not deducted from `amount`.
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
            "transfer",
            from,
            to,
            amount,
            fee,
            nonce[from]
        ));
        nonce[from]++;

        bytes32 ethMessageHash = ECDSA.toEthSignedMessageHash(hash);
        address recoveredSigner = ECDSA.recover(ethMessageHash, sig);
        require(recoveredSigner == from, "invalid signature");

        if (fee > 0) {
            require(trustedRSV.relayTransfer(from, msg.sender, fee), "fee transfer failed");
            emit FeeTaken(from, msg.sender, fee);
        }
        require(trustedRSV.relayTransfer(from, to, amount));
        emit TransferForwarded(sig, from, to, amount, fee);
    }

    // note that `fee` is not deducted from `amount`, and comes from the `spender` rather
    // than `holder`.
    function forwardTransferFrom(
        bytes calldata sig, 
        address spender,
        address holder,
        address to,
        uint256 amount,
        uint256 fee
    )
        external
    {
        bytes32 hash = keccak256(abi.encodePacked(
            address(trustedRSV),
            "transferFrom",
            spender,
            holder,
            to,
            amount,
            fee,
            nonce[spender]
        ));
        nonce[spender]++;

        bytes32 ethMessageHash = ECDSA.toEthSignedMessageHash(hash);
        address recoveredSigner = ECDSA.recover(ethMessageHash, sig);
        require(recoveredSigner == spender, "invalid signature");

        if (fee > 0) {
            require(trustedRSV.relayTransfer(spender, msg.sender, fee), "fee transfer failed");
            emit FeeTaken(spender, msg.sender, fee);
        }
        require(trustedRSV.relayTransferFrom(spender, holder, to, amount));
        emit TransferFromForwarded(sig, spender, holder, to, amount, fee);
    }

    // note that `fee` is not deducted from `amount`.
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
            "approve",
            holder,
            spender,
            amount,
            fee,
            nonce[holder]
        ));
        nonce[holder]++;

        bytes32 ethMessageHash = ECDSA.toEthSignedMessageHash(hash);
        address recoveredSigner = ECDSA.recover(ethMessageHash, sig);
        require(recoveredSigner == holder, "invalid signature");

        if (fee > 0) {
            require(trustedRSV.relayTransfer(holder, msg.sender, fee), "fee transfer failed");
            emit FeeTaken(holder, msg.sender, fee);
        }
        require(trustedRSV.relayApprove(holder, spender, amount));
        emit ApproveForwarded(sig, holder, spender, amount, fee);
    }
}
