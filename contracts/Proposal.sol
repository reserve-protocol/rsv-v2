pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./rsv/IRSV.sol";
import "./ownership/Ownable.sol";
import "./Basket.sol";

/**
 * A Proposal represents a suggestion to change the backing for RSV.
 *
 * The lifecycle of a proposal:
 * 1. Creation
 * 2. Acceptance
 * 3. Completion
 *
 * A time can be set during acceptance to determine when completion is
 * eligible.  A proposal can also be cancelled before it is completed. If a
 * proposal is cancelled, it can no longer become Completed.
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

    uint256 public time;
    address public proposer;

    enum State { Created, Accepted, Cancelled, Completed }
    State public state;
    
    constructor(address _proposer) public {
        proposer = _proposer;
        state = State.Created;
    }

    /// Moves a proposal from the Created to Accepted state. 
    function accept(uint256 _time) external onlyOwner {
        require(state == State.Created, "proposal not created");
        time = _time;
        state = State.Accepted;
    }

    /// Cancels a proposal if it has not been completed. 
    function cancel() external onlyOwner {
        require(state != State.Completed);
        state = State.Cancelled;
    }

    /// Moves a proposal from the Accepted to Completed state.
    /// Returns the tokens, quantitiesIn, and quantitiesOut, required to implement the proposal.
    function complete(IRSV rsv, Basket oldBasket) 
        external onlyOwner returns(Basket)
    {
        require(state == State.Accepted, "proposal must be accepted");
        require(now > time, "wait to execute");
        state = State.Completed;

        return _newBasket(rsv, oldBasket);
    }

    /// Returns the newly-proposed basket. This varies for different types of proposals,
    /// so it's abstract here.
    function _newBasket(IRSV rsv, Basket oldBasket) internal returns(Basket);
}

/**
 * A WeightProposal represents a suggestion to change the backing for RSV to a new distribution
 * of tokens. You can think of it as designating what a _single RSV_ should be backed by, but 
 * deferring on the precise quantities of tokens that will be need to be exchanged until a later 
 * point in time.
 *
 * When this proposal is completed, it simply returns the target basket. 
 */
contract WeightProposal is Proposal {
    Basket public basket;

    constructor(address _proposer, Basket _basket) Proposal(_proposer) public {
        basket = _basket;
    }

    /// Returns the newly-proposed basket
    function _newBasket(IRSV, Basket) internal returns(Basket) {
        return basket;
    }
}

/**
 * A SwapProposal represents a suggestion to transfer fixed amounts of tokens into and out of the
 * vault. Whereas a WeightProposal designates how much a _single RSV_ should be backed by, 
 * a SwapProposal first designates what quantities of tokens to transfer in total and then 
 * solves for the new resultant basket later. 
 *
 * When this proposal is completed, it calculates what the weights for the new basket will be
 * and returns it. 
 */
contract SwapProposal is Proposal {
    address public proposer;
    address[] public tokens;
    uint256[] public amounts;
    bool[] public toVault;

    constructor(address _proposer,
                address[] memory _tokens,
                uint256[] memory _amounts,
                bool[] memory _toVault )
        Proposal(_proposer) public
    {
        require(_tokens.length == _amounts.length && _amounts.length == _toVault.length,
                "unequal array lengths");
        tokens = _tokens;
        amounts = _amounts;
        toVault = _toVault;
    }

    /// Return the newly-proposed basket, based on the current vault and the old basket.
    function _newBasket(IRSV rsv, Basket oldBasket) internal returns(Basket) {

        uint256[] memory weights = new uint256[](tokens.length);
        uint256 divisor = uint256(10)**rsv.decimals();
        uint256 rsvSupply = rsv.totalSupply();
        
        for (uint i = 0; i < tokens.length; i++) {
            address token = tokens[i];
            uint256 oldWeight = oldBasket.weights(token);
            
            if (toVault[i]) {
                weights[i] = oldWeight.add( amounts[i].mul(divisor).div(rsvSupply) );
            } else {
                weights[i] = oldWeight.sub( amounts[i].mul(divisor).div(rsvSupply) );
            }
        }

        return new Basket(oldBasket, tokens, weights);
    }
}
