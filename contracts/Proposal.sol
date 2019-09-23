pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./Ownable.sol";
import "./Basket.sol";

/**
 * A Proposal represents a suggestion to change the backing for RSV.
 *
 * The lifecycle of a proposal:
 * 1. Creation
 * 2. Acceptance
 * 3. Completion
 *
 * A time can be set during acceptance to determine when completion is eligible.
 * A proposal can also be closed before it is Completed. 
 *
 * This contract is intended to be used in one of two possible ways. Either:
 * - A target RSV basket is suggested, and quantities to be exchanged are  
 *     deduced at the time of proposal execution.
 * - A specific quantity of tokens to be exchanged is suggested, and the 
 *     resultant RSV basket is determined at the time of proposal execution.
 *
 */
contract Proposal is Ownable {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    uint256 public id;
    uint256 public time;
    address public proposer;
    address[] public tokens;
    uint256[] public quantitiesIn; // token quantities to be added to the Vault
    uint256[] public quantitiesOut; // total quantities to be withdrawn from the Vault
    bool public accepted;
    bool public closed;

    Basket public basket;

    // Events
    event ProposalCreated(uint256 indexed id, address indexed proposer, address[] tokens, uint256[] quantitiesIn, uint256[] quantitiesOut);
    event ProposalAccepted(uint256 indexed id, address indexed proposer);
    event ProposalFinished(uint256 indexed id, address indexed proposer);
    event ProposalClosed(uint256 indexed id, address indexed proposer);

    constructor(
        uint256 _id,
        address _proposer,
        address[] memory _tokens,
        uint256[] memory _quantitiesIn, // total quantities for the entire RSV supply, not per RSV
        uint256[] memory _quantitiesOut, // total quantities for the entire RSV supply, not per RSV
        Basket _basket
    ) 
        public 
    {
        require(_quantitiesIn.length == _quantitiesOut.length, "quantities mismatched");
        (id, proposer, tokens, quantitiesIn, quantitiesOut, basket) = 
            (_id, _proposer, _tokens, _quantitiesIn, _quantitiesOut, _basket);
        emit ProposalCreated(_id, _proposer, _tokens, _quantitiesIn, _quantitiesOut);
    }

    /// Moves a proposal from the Created to Accepted state. 
    function accept(uint256 _time) external onlyOwner {
        require(!accepted, "proposal already accepted");
        time = _time;
        accepted = true;
        emit ProposalAccepted(id, proposer);
    }

    /// Closes a proposal if it has not been completed. 
    function close() external onlyOwner {
        require(!closed, "proposal already closed");
        closed = true;
        emit ProposalClosed(id, proposer);
    }

    /// Moves a proposal from the Accepted to Completed state. 
    function complete(
        uint256 _rsvSupply, 
        address _vaultAddr, 
        Basket _prevBasket
    ) 
        external onlyOwner returns(address[] memory, uint256[] memory, uint256[] memory) 
    {
        require(!closed, "proposal already closed");
        require(accepted, "proposal not accepted");
        require(now > time, "wait to execute");

        if (basket == Basket(0)) {
            uint256[] memory newBacking = new uint256[](_prevBasket.size());

            uint256 newQuantity;
            for (uint i = 0; i < _prevBasket.size(); i++) {
                newQuantity = IERC20(tokens[i]).balanceOf(_vaultAddr) + quantitiesIn[i] - quantitiesOut[i];
                require(newQuantity >= 0, "proposal removes too many tokens");
                newBacking[i] = newQuantity.mul(_prevBasket.frontTokenDecimals()).div(_rsvSupply);
            }

            basket = new Basket(tokens, newBacking, _prevBasket.frontTokenDecimals());
            assert(basket.size() == _prevBasket.size());
        }
        quantitiesIn = _prevBasket.newQuantitiesRequired(_rsvSupply, basket);
        quantitiesOut = basket.newQuantitiesRequired(_rsvSupply, _prevBasket);
        closed = true;
        emit ProposalFinished(id, proposer);
        return (tokens, quantitiesIn, quantitiesOut);
    }
}
