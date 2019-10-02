pragma solidity 0.5.7;

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
 * A time can be set during acceptance to determine when completion is eligible.  A proposal can
 * also be cancelled before it is completed. If a proposal is cancelled, it can no longer become
 * Completed.
 *
 * This contract is intended to be used in one of two possible ways. Either:
 * - A target RSV basket is proposed, and quantities to be exchanged are deduced at the time of
 *   proposal execution.
 * - A specific quantity of tokens to be exchanged is proposed, and the resultant RSV basket is
 *   determined at the time of proposal execution.
 */

interface IProposal {
    function proposer() external returns(address);
    function accept(uint256 time) external;
    function cancel() external;
    function complete(IRSV rsv, Basket oldBasket) external returns(Basket);
    function nominateNewOwner(address newOwner) external;
    function acceptOwnership() external;
}

interface IProposalFactory {
    function createSwapProposal(address,
        address[] calldata tokens,
        uint256[] calldata amounts,
        bool[] calldata toVault
    ) external returns (IProposal);

    function createWeightProposal(address proposer, Basket basket) external returns (IProposal);
}

contract ProposalFactory is IProposalFactory {
    function createSwapProposal(
        address proposer,
        address[] calldata tokens,
        uint256[] calldata amounts,
        bool[] calldata toVault
    )
        external returns (IProposal)
    {
        IProposal proposal = IProposal(new SwapProposal(proposer, tokens, amounts, toVault));
        proposal.nominateNewOwner(msg.sender);
        return proposal;
    }

    function createWeightProposal(address proposer, Basket basket) external returns (IProposal) {
        IProposal proposal = IProposal(new WeightProposal(proposer, basket));
        proposal.nominateNewOwner(msg.sender);
        return proposal;
    }
}

contract Proposal is IProposal, Ownable {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    uint256 public time;
    address public proposer;

    enum State { Created, Accepted, Cancelled, Completed }
    State public state;
    
    event ProposalCreated(address indexed proposer);
    event ProposalAccepted(address indexed proposer, uint256 indexed time);
    event ProposalCancelled(address indexed proposer);
    event ProposalCompleted(address indexed proposer, address indexed basket);

    constructor(address _proposer) public {
        proposer = _proposer;
        state = State.Created;
        emit ProposalCreated(proposer);
    }

    /// Moves a proposal from the Created to Accepted state.
    function accept(uint256 _time) external onlyOwner {
        require(state == State.Created, "proposal not created");
        time = _time;
        state = State.Accepted;
        emit ProposalAccepted(proposer, _time);
    }

    /// Cancels a proposal if it has not been completed.
    function cancel() external onlyOwner {
        require(state != State.Completed);
        state = State.Cancelled;
        emit ProposalCancelled(proposer);
    }

    /// Moves a proposal from the Accepted to Completed state.
    /// Returns the tokens, quantitiesIn, and quantitiesOut, required to implement the proposal.
    function complete(IRSV rsv, Basket oldBasket)
        external onlyOwner returns(Basket)
    {
        require(state == State.Accepted, "proposal must be accepted");
        require(now > time, "wait to execute");
        state = State.Completed;

        Basket b = _newBasket(rsv, oldBasket);
        emit ProposalCompleted(proposer, address(b));
        return b;
    }

    /// Returns the newly-proposed basket. This varies for different types of proposals,
    /// so it's abstract here.
    function _newBasket(IRSV trustedRSV, Basket oldBasket) internal returns(Basket);
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
    Basket public trustedBasket;

    constructor(address _proposer, Basket _trustedBasket) Proposal(_proposer) public {
        require(_trustedBasket.size() > 0, "proposal cannot be empty");
        trustedBasket = _trustedBasket;
    }

    /// Returns the newly-proposed basket
    function _newBasket(IRSV, Basket) internal returns(Basket) {
        return trustedBasket;
    }
}

/**
 * A SwapProposal represents a suggestion to transfer fixed amounts of tokens into and out of the
 * vault. Whereas a WeightProposal designates how much a _single RSV_ should be backed by,
 * a SwapProposal first designates what quantities of tokens to transfer in total and then
 * solves for the new resultant basket later.
 *
 * When this proposal is completed, it calculates what the weights for the new basket will be
 * and returns it. If RSV supply is 0, this kind of Proposal cannot be used. 
 */

// On "unit" comments, see comment at top of Manager.sol.
contract SwapProposal is Proposal {
    address[] public tokens;
    uint256[] public amounts; // unit: qToken
    bool[] public toVault;

    uint256 constant WEIGHT_SCALE = uint256(10)**18; // unit: aqToken / qToken

    constructor(address _proposer,
                address[] memory _tokens,
                uint256[] memory _amounts, // unit: qToken
                bool[] memory _toVault )
        Proposal(_proposer) public
    {
        require(_tokens.length > 0, "proposal cannot be empty");
        require(_tokens.length == _amounts.length && _amounts.length == _toVault.length,
                "unequal array lengths");
        tokens = _tokens;
        amounts = _amounts;
        toVault = _toVault;
    }

    /// Return the newly-proposed basket, based on the current vault and the old basket.
    function _newBasket(IRSV trustedRSV, Basket trustedOldBasket) internal returns(Basket) {

        uint256[] memory weights = new uint256[](tokens.length);
        // unit: aqToken/RSV

        uint256 scaleFactor = WEIGHT_SCALE.mul(uint256(10)**(trustedRSV.decimals()));
        // unit: aqToken/qToken * qRSV/RSV

        uint256 rsvSupply = trustedRSV.totalSupply();
        // unit: qRSV

        for (uint i = 0; i < tokens.length; i++) {
            uint256 oldWeight = trustedOldBasket.weights(tokens[i]);
            // unit: aqToken/RSV

            if (toVault[i]) {
                // We require that the execution of a SwapProposal takes in no more than the funds
                // offered in its proposal -- that's part of the premise. It turns out that,
                // because we're rounding down _here_ and rounding up in
                // Manager._executeBasketShift(), it's possible for the naive implementation of
                // this mechanism to overspend the proposer's tokens by 1 qToken. We avoid that,
                // here, by making the effective proposal one less. Yeah, it's pretty fiddly.
                
                weights[i] = oldWeight.add( (amounts[i].sub(1)).mul(scaleFactor).div(rsvSupply) );
                //unit: aqToken/RSV == aqToken/RSV == [qToken] * [aqToken/qToken*qRSV/RSV] / [qRSV]
            } else {
                weights[i] = oldWeight.sub( amounts[i].mul(scaleFactor).div(rsvSupply) );
                //unit: aqToken/RSV
            }
        }

        return new Basket(trustedOldBasket, tokens, weights);
        // unit check for weights: aqToken/RSV
    }
}


