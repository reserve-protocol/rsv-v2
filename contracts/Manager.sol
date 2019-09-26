pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/math/SafeMath.sol";
import "./ownership/Ownable.sol";
import "./Basket.sol";
import "./Proposal.sol";


interface IRSV {
    // Standard ERC20 functions
    function transfer(address, uint256) external returns(bool);
    function approve(address, uint256) external returns(bool);
    function transferFrom(address, address, uint256) external returns(bool);
    function totalSupply() external view returns(uint256);
    function balanceOf(address) external view returns(uint256);
    function allowance(address, address) external view returns(uint256);
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed holder, address indexed spender, uint256 value);

    // RSV-specific functions
    function mint(address, uint256) external;
    function burnFrom(address, uint256) external;
    function decimals() external view returns(uint8);
}

interface IVault {
    function changeManger(address) external;
    function batchWithdrawTo(address[] calldata, uint256[] calldata, address) external;
}

/**
 * The Manager contract is the point of contact between the Reserve ecosystem
 * and the surrounding world. It manages the Issuance and Redemption of RSV,
 * a decentralized stablecoin backed by a basket of tokens. 
 *
 * The Manager also implements a Proposal system to handle administration of
 * changes to the backing of RSV. Anyone can propose a change to the backing.
 * Once the `owner` approves the proposal, then after a pre-determined delay
 * the proposal is eligible for execution by anyone. However, the funds to 
 * execute the proposal must come from the proposer.
 *
 * There are two different ways to propose changes to the backing of RSV. 
 * See: 
 * - proposeQuantitiesAdjustment()
 * - proposeNewBasket()
 *
 * In both cases, tokens are exchanged with the Vault and a new RSV backing is 
 * set. You can think of the first type of proposal as being useful when you
 * don't want to change the list of tokens that back RSV, but do want to change
 * the quantities. The second type of proposal is more useful when you want to
 * change the tokens in the basket. The downside of this proposal type is that
 * it's difficult to know what capital will be required come execution of the
 * proposal.  
 */
contract Manager is Ownable {
    using SafeERC20 for IERC20;
    using SafeMath for uint256;

    // ROLES

    // Manager is already Ownable, but in addition it also has an `operator`. 
    address operator; 


    // DATA

    Basket public basket;
    IVault public vault;
    IRSV public rsv;
    uint8 public constant rsvDecimals = 18;

    // Proposals
    mapping(uint256 => Proposal) public proposals;
    uint256 public proposalsLength;
    uint256 public delay = 24 hours;
    
    // Issuance and Redemption controls
    mapping(address => bool) public whitelist;
    bool public useWhitelist;
    bool public paused;

    // The spread between issuance and redemption in BPS.
    uint256 public seigniorage;                 // 0.1% spread -> 10 BPS
    uint256 constant BPS_FACTOR = 10000;        // This is what 100% looks like in BPS.
    

    // EVENTS

    event ProposalsCleared();

    // RSV traded events
    event Issuance(address indexed user, uint256 indexed amount);
    event Redemption(address indexed user, uint256 indexed amount);

    // Pause events
    event Paused(address indexed account);
    event Unpaused(address indexed account);

    // Changes
    event OperatorChanged(address indexed oldAccount, address indexed newAccount);
    event SeigniorageChanged(uint256 oldVal, uint256 newVal);
    event WhitelistChanged(address indexed user, bool val);
    event DelayChanged(uint256 oldVal, uint256 newVal);

    // Proposals
    event WeightsProposed(uint256 indexed id,
                          address indexed proposer,
                          address[] tokens,
                          uint256[] backing);

    event SwapProposed(uint256 indexed id,
                       address indexed proposer,
                       address[] tokens,
                       uint256[] amounts,
                       bool[] toVault);
    
    event ProposalAccepted(uint256 indexed id, address indexed proposer);
    event ProposalCanceled(uint256 indexed id, address indexed proposer, address indexed canceler);
    event ProposalExecuted(uint256 indexed id, address indexed proposer, address indexed executor);
    event BasketChanged(address indexed oldBasket, address indexed newBasket);

    // ============================ Constructor ===============================

    /// Begins paused.
    constructor(address vaultAddr, address rsvAddress, uint256 seigniorage_) public {
        vault = IVault(vaultAddr);
        rsv = IRSV(rsvAddress);
        whitelist[_msgSender()] = true;
        seigniorage = seigniorage_;
        paused = true;
        useWhitelist = true;
    }

    // ============================= Modifiers ================================

    /// Modifies a function to run only when the contract is not paused.
    modifier notPaused() {
        _notPaused();
        _;
    }

    /// Modifies a function to run only when the caller is on the whitelist, if it is enabled.
    modifier onlyWhitelist() {
        _onlyWhitelist();
        _;
    }

    /// Modifies a function to run only when the caller is the operator account. 
    modifier onlyOperator() {
        _onlyOperator();
        _;
    }

    // This approach reduces bytecode since solidity inlines all modifiers under the hood. 

    function _notPaused() internal view {
        require(!paused, "contract is paused");
    }

    function _onlyWhitelist() internal view {
        if (useWhitelist) require(whitelist[_msgSender()], "not on whitelist");
    }

    function _onlyOperator() internal view {
        require(_msgSender() == operator, "operator only");
    }


    // ============================= Public ==================================

    /// Ensure that the Vault is fully collateralized. 
    function isFullyCollateralized() public view returns(bool) {
        uint256[] memory expected = basket.quantitiesRequired(rsv.totalSupply());
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            uint256 fullAmount = _amountFromWeight(basket.weights(token));

            if (IERC20(token).balanceOf(address(vault)) < fullAmount)
                return false;
        }
        return true;
    }

    // ============================= Externals ================================

    /// Issue a quantity of RSV to the caller and deposit collateral tokens in the Vault.
    function issue(uint256 _rsvQuantity) external notPaused onlyWhitelist {
        _issue(_rsvQuantity);
    }

    // /// Issues the maximum amount of RSV to the caller based on their allowances.
    // function issueMax() external notPaused onlyWhitelist {
    //     uint256 max = _calculateMaxIssuable(_msgSender());
    //     _issue(max);
    // }

    /// Redeem a quantity of RSV for collateral tokens. 
    function redeem(uint256 _rsvQuantity) external notPaused onlyWhitelist {
        _redeem(_rsvQuantity);
    }

    // /// Redeem `allowance` of RSV from the caller's account. 
    // function redeemMax() external notPaused onlyWhitelist {
    //     uint256 max = rsv.allowance(_msgSender(), address(this));
    //     _redeem(max);
    // }

    /** TODO:rewrite, rename
     * Proposes an adjustment to the quantities of tokens in the Vault. Importantly, this type of
     * proposal does not change token addresses. Therefore, if you want to introduce a new token,
     * first use the other proposal type. 
     */

    // TODO:rename, adjust
    function proposeQuantitiesAdjustment(
        address[] calldata _tokens,
        uint256[] calldata _amountsIn,
        uint256[] calldata _amountsOut
    ) 
        external returns(uint256)
    {
        require(_tokens.length == _amountsIn.length, "token quantities mismatched");
        require(_amountsIn.length == _amountsOut.length, "quantities mismatched");

        proposals[proposalsLength] = new Proposal(
            proposalsLength,
            _msgSender(),
            _tokens,
            _amountsIn,
            _amountsOut,
            Basket(0)
        );

        emit NewQuantityAdjustmentProposalCreated(
            proposalsLength, 
            _msgSender(), 
            _tokens,
            _amountsIn, 
            _amountsOut
        );
        return ++proposalsLength;
    }

    
    /** TODO: rewrite, rename
     * Proposes a new basket defined by a list of tokens and their backing quantities. 
     * Importantly, this type of proposal means the balances that will be required from the 
     * proposer at the time of execution are variable. If the supply of RSV changes significantly,
     * then much more tokens could be required to execute the proposal. 
     * 
     */

    // TODO:rename, adjust
    function proposeNewBasket(
        address[] calldata _tokens,
        uint256[] calldata _backing
    )
        external returns(uint256)
    {
        require(_tokens.length == _backing.length, "mismatched token quantities");
        require(_tokens.length > 0, "no tokens in basket");
        uint256[] memory quantitiesIn;
        uint256[] memory quantitiesOut;

        proposals[proposalsLength] = new Proposal(
            proposalsLength,
            _msgSender(),
            _tokens,
            quantitiesIn,
            quantitiesOut,
            new Basket(_tokens, _backing, rsvDecimals)
        );

        emit NewBasketProposalCreated(proposalsLength, _msgSender(), _tokens, _backing);
        return ++proposalsLength;
    }

    // TODO:rename
    /// Accepts a proposal for a new basket, beginning the required delay.
    function acceptProposal(uint256 _proposalID) external onlyOperator {
        require(proposalsLength > _proposalID, "proposals length < id");
        proposals[_proposalID].accept(now + delay);
        emit ProposalAccepted(_proposalID, proposals[_proposalID].proposer());
    }

    // Cancels a proposal. This can be done anytime before it is enacted by any of:
    // 1. Proposer 2. Operator 3. Owner
    function cancelProposal(uint256 _proposalID) external {
        require(
            _msgSender() == proposals[_proposalID].proposer() ||
            _msgSender() == _owner ||
            _msgSender() == operator, 
            "cannot cancel"
        );
        proposals[_proposalID].close();
        emit ProposalCanceled(_proposalID, proposals[_proposalID].proposer(), _msgSender());
    }

    /// Executes a proposal by exchanging collateral tokens with the proposer.
    function executeProposal(uint256 proposalID) external {
        require(
            _msgSender() == proposals[proposalID].proposer() ||
            _msgSender() == operator,
            "cannot execute"
        );
        require(proposalsLength > proposalID, "proposals length < id");
        address proposer = proposals[proposalID].proposer();
        Basket oldBasket = basket;

        // Complete proposal and compute new basket
        basket = proposals[proposalID].complete(rsv.totalSupply(), address(vault), oldBasket);
        
        // For each token in either basket, perform transfers between proposer and Vault 
        for (uint i = 0; i < oldBasket.size(); i++) {
            address token = oldBasket.tokens[i];
            _executeBasketShift(oldBasket, basket, token, proposer);
        }
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens[i];
            if (!oldBasket.has(token)) {
                _executeBasketShift(oldBasket, basket, token, proposer);
            }
        }
        
        assert(isFullyCollateralized());
        emit BasketChanged(address(oldBasket), address(basket));
        emit ProposalExecuted(proposalID, proposer, _msgSender());
    }
        
    /// Pause the contract.
    function pause() external onlyOwner {
        paused = true;
        emit Paused(_msgSender());
    }

    /// Unpause the contract.
    function unpause() external onlyOwner {
        require(address(basket) != address(0), "basket required to unpause");
        paused = false;
        emit Unpaused(_msgSender());
    }

    /// Add or remove user from whitelist.
    function setWhitelist(address _user, bool _val) external onlyOwner {
        whitelist[_user] = _val;
        emit WhitelistChanged(_user, _val);
    }

    /// Set whether or not to apply the whitelist to Issuance and Redemption. 
    function setUseWhitelist(bool _useWhitelist) external onlyOwner {
        useWhitelist = _useWhitelist;
    }

    /// Set the operator
    function setOperator(address _operator) external onlyOwner {
        emit OperatorChanged(operator, _operator);
        operator = _operator;
    }

    /// Set the seigniorage, in BPS. 
    function setSegniorage(uint256 _seigniorage) external onlyOwner {
        seigniorage = _seigniorage;
        emit SeigniorageChanged(seigniorage, _seigniorage);
    }

    /// Set the Proposal delay in hours.
    function setDelay(uint256 _delay) external onlyOwner {
        emit DelayChanged(delay, _delay);
        delay = _delay;
    }

    /// Clear the list of proposals. 
    function clearProposals() external onlyOwner {
        proposalsLength = 0;
        emit ProposalsCleared();
    }

    /// Get the amounts of all basket tokens required to issue a given amount of RSV.
    function amountsToIssue(uint256 rsvAmount) public view returns (uint256[] memory) {
        uint256[] memory amounts = new uint256[basket.size()];

        uint256 feeRate = uint256(seigniorage.add(BPS_FACTOR));
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket[i];
            amounts[i] = _weight(rsvAmount, basket.weights[token]).mul(feeRate).div(BPS_FACTOR);
        }
        return amounts;
    }

    // ============================= Internals ================================

    /// _issue: Internal function for all issuances to go through.
    function _issue(uint256 rsvAmount) internal {
        require(rsvAmount > 0, "cannot issue zero RSV");

        uint256[] memory amounts = amountsToIssue(rsvAmount);

        // Accept collateral tokens.
        for (uint i = 0; i < basket.size(); i++) {
            IERC20(basket.tokens(i)).safeTransferFrom(_msgSender(), address(vault), amounts[i]);
        }
        // Compensate with RSV.
        rsv.mint(_msgSender(), rsvAmount);

        assert(isFullyCollateralized());
        emit Issuance(_msgSender(), rsvAmount);
    }

    /// _redeem: Internal function for all redemptions to go through.
    function _redeem(uint256 rsvAmount) internal {
        require(rsvAmount > 0, "cannot redeem 0 RSV");

        // Burn RSV tokens.
        rsv.burnFrom(_msgSender(), rsvAmount);

        // Compensate with collateral tokens.
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens[i];
            uint256 amount = _weight(rsvAmount, basket.weight(token));
            vault.withdrawTo(token, amount, _msgSender());
        }
        
        assert(isFullyCollateralized());
        emit Redemption(_msgSender(), rsvAmount);
    }


    /// _executeBasketShift transfers the necessary amount of `token` between vault and `proposer`
    /// to rebalance the vault's balance of token, as it goes from oldBasket to newBasket.
    /// @dev To carry out a proposal, this is executed once per relevant token.
    function _executeBasketShift(
        Basket oldBasket,
        Basket newBasket,
        address token,
        address proposer
    ) internal {
        uint256 newWeight = newBasket.weights[token];
        uint256 oldWeight = oldBasket.weights[token];
        if (newWeight > oldWeight) {
            // This token must increase in the vault, so transfer from proposer to vault.
            uint256 transferAmount = _weight(rsv.totalSupply(), newWeight.sub(oldWeight));
            IERC20(token).safeTransferFrom(proposer, address(vault), transferAmount);
        } else if (newWeight < oldWeight) {
            // This token will decrease in the vault, so transfer from vault to proposer.
            uint256 transferAmount = _weight(rsv.totalSupply(), oldWeight.sub(newWeight));
            vault.withdrawTo(token, transferAmount, proposer);
        }
    }

    // From a weighting of RSV (e.g., a basket weight) and an amount of RSV,
    // compute the amount of the weighted token that matches that amount of RSV.
    function _weight(uint256 rsvAmount, uint256 weight)
        internal view returns(uint256 amount) {
        return supply.mul(weight).div(uint256(10)**rsv.decimals();
    }

}
