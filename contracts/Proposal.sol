pragma solidity ^0.5.8;

import "./zeppelin/contracts/token/ERC20/IERC20.sol";
import "./zeppelin/contracts/token/ERC20/SafeERC20.sol";
import "./Basket.sol";
import "./Ownable.sol";

contract Proposal is Ownable {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    uint256 public id;
    uint256 public time;
    address public proposer;
    address[] public tokens;
    uint256[] public quantitiesIn; // total quantities for the entire RSV supply, not per RSV
    uint256[] public quantitiesOut; // total quantities for the entire RSV supply, not per RSV
    bool public accepted;
    bool public closed;

    Basket public basket;

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
        id = _id;
        proposer = _proposer;
        tokens = _tokens;
        quantitiesIn = _quantitiesIn;
        quantitiesOut = _quantitiesOut;
        basket = _basket;
        emit ProposalCreated(_id, _proposer, _tokens, _quantitiesIn, _quantitiesOut);
    }

    function accept(uint256 _time) external onlyOwner {
        require(!accepted, "proposal already accepted");
        time = _time;
        accepted = true;
        emit ProposalAccepted(id, proposer);
    }

    function close() external onlyOwner {
        require(!closed, "proposal already closed");
        closed = true;
        emit ProposalClosed(id, proposer);
    }

    function prepare(uint256 _rsvSupply, address _vaultAddr, Basket _prevBasket) external onlyOwner {
        if (basket == Basket(0)) {
            uint256[] memory newBacking = new uint256[](_prevBasket.size());

            uint256 newQuantity;
            for (uint i = 0; i < _prevBasket.size(); i++) {
                newQuantity = IERC20(tokens[i]).balanceOf(_vaultAddr) + quantitiesIn[i] - quantitiesOut[i];
                require(newQuantity >= 0, "proposal tries to remove too many tokens");
                newBacking[i] = newQuantity.mul(_prevBasket.frontTokenDecimals()).div(_rsvSupply);
            }

            basket = new Basket(tokens, newBacking,_prevBasket.frontTokenDecimals());
            assert(basket.size() == _prevBasket.size());

            // Sanity checking
            uint256[] memory expectedAmountsIn = _prevBasket.newQuantitiesRequired(_rsvSupply, basket);
            uint256[] memory expectedAmountsOut = basket.newQuantitiesRequired(_rsvSupply, _prevBasket);
            for (uint i = 0; i < basket.size(); i++) {
                assert(expectedAmountsIn[i] == quantitiesIn[i]);
                assert(expectedAmountsOut[i] == quantitiesOut[i]);
            }
        } else {
            quantitiesIn = _prevBasket.newQuantitiesRequired(_rsvSupply, basket);
            quantitiesOut = basket.newQuantitiesRequired(_rsvSupply, _prevBasket);
        }
    }

    function finish() external onlyOwner {
        assert(basket != Basket(0));
        require(!closed, "proposal already closed");
        require(accepted, "proposal not accepted");
        require(now > time, "wait to execute");
        closed = true;
        emit ProposalFinished(id, proposer);
    }

    function getTokens() external view returns(address[] memory) {
        return tokens;
    }

    function getQuantitiesIn() external view returns(uint256[] memory) {
        return quantitiesIn;
    }

    function getQuantitiesOut() external view returns(uint256[] memory) {
        return quantitiesOut;
    }

}
