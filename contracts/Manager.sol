pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/math/SafeMath.sol";
import "./rsv/IRSV.sol";
import "./ownership/Ownable.sol";
import "./Basket.sol";
import "./Proposal.sol";


interface IVault {
    function changeManger(address) external;
    function withdrawTo(address, uint256, address) external;
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

    // Proposals
    mapping(uint256 => Proposal) public proposals;
    uint256 public proposalsLength;
    uint256 public delay = 24 hours;
    
    // Pausing
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
    event DelayChanged(uint256 oldVal, uint256 newVal);

    // Proposals
    event WeightsProposed(uint256 indexed id,
                          address indexed proposer,
                          address[] tokens,
                          uint256[] weights);

    event SwapProposed(uint256 indexed id,
                       address indexed proposer,
                       address[] tokens,
                       uint256[] amounts,
                       bool[] toVault);
    
    event ProposalAccepted(uint256 indexed id, address indexed proposer);
    event ProposalCanceled(uint256 indexed id, address indexed proposer, address indexed canceler);
    event ProposalExecuted(uint256 indexed id, 
                           address indexed proposer, 
                           address indexed executor,
                           address oldBasket,
                           address newBasket);

    // ============================ Constructor ===============================

    /// Begins paused.
    constructor(address vaultAddr, address rsvAddress, uint256 seigniorage_) public {
        vault = IVault(vaultAddr);
        rsv = IRSV(rsvAddress);
        seigniorage = seigniorage_;
        paused = true;

        // Start with the empty basket. 
        address[] memory tokens = new address[](0);
        uint256[] memory weights = new uint256[](0);
        basket = new Basket(Basket(0), tokens, weights);
    }

    // ============================= Modifiers ================================

    /// Modifies a function to run only when the contract is not paused.
    modifier notPaused() {
        _notPaused();
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

    function _onlyOperator() internal view {
        require(_msgSender() == operator, "operator only");
    }


    // ============================= Public ==================================

    /// Ensure that the Vault is fully collateralized. 
    function isFullyCollateralized() public view returns(bool) {
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            uint256 fullAmount = _weighted(rsv.totalSupply(), basket.weights(token));

            if (IERC20(token).balanceOf(address(vault)) < fullAmount)
                return false;
        }
        return true;
    }

    /// Get amounts of basket tokens required to issue an amount of RSV.
    /// The returned array will be in the same order as the current basket.tokens.
    function toIssue(uint256 rsvAmount) public view returns (uint256[] memory) {
        uint256[] memory amounts = new uint256[](basket.size());
        uint256 feeRate = uint256(seigniorage.add(BPS_FACTOR));

        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            amounts[i] = _weighted(rsvAmount, basket.weights(token)).mul(feeRate).div(BPS_FACTOR);
        }

        return amounts;
    }

    /// Get amounts of basket tokens that would be sent upon redeeming an amount of RSV.
    /// The returned array will be in the same order as the current basket.tokens.
    function toRedeem(uint256 rsvAmount) public view returns (uint256[] memory) {
        uint256[] memory amounts = new uint256[](basket.size());

        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            amounts[i] = _weighted(rsvAmount, basket.weights(token));
        }

        return amounts;
    }

    // ============================= External ================================

    /// Issue RSV to the caller and deposit collateral tokens in the Vault.
    function issue(uint256 rsvAmount) external notPaused {
        _issue(rsvAmount);
    }

    /// Redeem RSV for collateral tokens. 
    function redeem(uint256 rsvAmount) external notPaused {
        _redeem(rsvAmount);
    }

    /*
     * Propose an exchange of current Vault tokens for new Vault tokens.
     * 
     * These parameters are phyiscally a set of arrays because Solidity doesn't let you pass
     * around arrays of structs as parameters of transactions. Semantically, read these three
     * lists as a list of triples (token, amount, toVault), where:
     *
     * - token is the address of an ERC-20 token,
     * - amount is the amount of the token that the proposer says they will trade with the vault, and
     * - toVault is the direction of that trade. If toVault is true, the proposer offers to send
     *   `amount` of `token` to the vault. If toVault is false, the proposer expects to receive
     *   `amount` of `token` from the vault.
     * 
     * If this proposal is accepted and executed, this set of absolute transfers will occur,
     * and the Vault's basket weights will be adjusted accordingly. (The expected behavior of
     * proposers is that they will aim to make proposals that move the basket weights towards
     * some target of Reserve's management while maintaining full backing; the expected
     * behavior of Reserve's management is to only accept such proposals.)
     * 
     * Note: This type of proposal does not remove token addresses!
     * If you want to remove token addresses entirely, use proposeWeights.
     * 
     * Returns the new proposal's ID.
     */
    function proposeSwap(
        address[] calldata tokens,
        uint256[] calldata amounts,
        bool[] calldata toVault
    ) 
        external returns(uint256)
    {
        require(tokens.length == amounts.length && amounts.length == toVault.length,
                "proposeSwap: unequal lengths");

        proposals[proposalsLength] = new SwapProposal(_msgSender(), tokens, amounts, toVault);

        emit SwapProposed(proposalsLength, _msgSender(), tokens, amounts, toVault);
        return ++proposalsLength;
    }

    
    /** 
     * Propose a new basket, defined by a list of tokens address, and their basket weights.
     * 
     * Note: With this type of proposal, the allowances of tokens that will be required of the
     * proposer may change between proposition and execution. If the supply of RSV rises or falls,
     * then more or fewer tokens will be required to execute the proposal.
     *
     * Returns the new proposal's ID.
     */

    function proposeWeights(address[] calldata tokens, uint256[] calldata weights)
        external returns(uint256)
    {
        require(tokens.length == weights.length, "proposeWeights: unequal lengths");
        require(tokens.length > 0, "proposeWeights: zero length");

        proposals[proposalsLength] =
            new WeightProposal(_msgSender(), new Basket(Basket(0), tokens, weights));

        emit WeightsProposed(proposalsLength, _msgSender(), tokens, weights);

        return ++proposalsLength;
    }

    /// Accepts a proposal for a new basket, beginning the required delay.
    function acceptProposal(uint256 id) external onlyOperator {
        require(proposalsLength > id, "proposals length < id");
        proposals[id].accept(now + delay);
        emit ProposalAccepted(id, proposals[id].proposer());
    }

    /// Cancels a proposal. This can be done anytime before it is enacted by any of:
    /// 1. Proposer 2. Operator 3. Owner
    function cancelProposal(uint256 id) external {
        require(
            _msgSender() == proposals[id].proposer() ||
            _msgSender() == _owner ||
            _msgSender() == operator, 
            "cannot cancel"
        );
        proposals[id].cancel();
        emit ProposalCanceled(id, proposals[id].proposer(), _msgSender());
    }

    /// Executes a proposal by exchanging collateral tokens with the proposer.
    function executeProposal(uint256 id) external onlyOperator {
        require(proposalsLength > id, "proposals length < id");
        address proposer = proposals[id].proposer();
        Basket oldBasket = basket;

        // Complete proposal and compute new basket
        basket = proposals[id].complete(rsv, address(vault), oldBasket);
        
        // For each token in either basket, perform transfers between proposer and Vault 
        for (uint i = 0; i < oldBasket.size(); i++) {
            address token = oldBasket.tokens(i);
            _executeBasketShift(oldBasket, basket, token, proposer);
        }
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            if (!oldBasket.has(token)) {
                _executeBasketShift(oldBasket, basket, token, proposer);
            }
        }
        
        assert(isFullyCollateralized());
        emit ProposalExecuted(id, proposer, _msgSender(), address(oldBasket), address(basket));
    }
        
    /// Pause the contract.
    function pause() external onlyOwner {
        paused = true;
        emit Paused(_msgSender());
    }

    /// Unpause the contract.
    function unpause() external onlyOwner {
        require(basket.size() > 0, "basket cannot be empty");
        paused = false;
        emit Unpaused(_msgSender());
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


    // ============================= Internal ================================

    /// Handles issuance.
    function _issue(uint256 rsvAmount) internal {
        require(rsvAmount > 0, "cannot issue zero RSV");

        // Accept collateral tokens.
        uint256[] memory amounts = toIssue(rsvAmount);
        for (uint i = 0; i < basket.size(); i++) {
            IERC20(basket.tokens(i)).safeTransferFrom(_msgSender(), address(vault), amounts[i]);
        }

        // Compensate with RSV.
        rsv.mint(_msgSender(), rsvAmount);

        assert(isFullyCollateralized());
        emit Issuance(_msgSender(), rsvAmount);
    }

    /// Handles redemption.
    function _redeem(uint256 rsvAmount) internal {
        require(rsvAmount > 0, "cannot redeem 0 RSV");

        // Burn RSV tokens.
        rsv.burnFrom(_msgSender(), rsvAmount);

        // Compensate with collateral tokens.
        uint256[] memory amounts = toRedeem(rsvAmount);
        for (uint i = 0; i < basket.size(); i++) {
            vault.withdrawTo(basket.tokens(i), amounts[i], _msgSender());
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
        uint256 newWeight = newBasket.weights(token);
        uint256 oldWeight = oldBasket.weights(token);
        if (newWeight > oldWeight) {
            // This token must increase in the vault, so transfer from proposer to vault.
            uint256 transferAmount = _weighted(rsv.totalSupply(), newWeight.sub(oldWeight));
            if (transferAmount > 0) 
                IERC20(token).safeTransferFrom(proposer, address(vault), transferAmount);
        } else if (newWeight < oldWeight) {
            // This token will decrease in the vault, so transfer from vault to proposer.
            uint256 transferAmount = _weighted(rsv.totalSupply(), oldWeight.sub(newWeight));
            if (transferAmount > 0) 
                vault.withdrawTo(token, transferAmount, proposer);
        }
    }

    // From a weighting of RSV (e.g., a basket weight) and an amount of RSV,
    // compute the amount of the weighted token that matches that amount of RSV.
    function _weighted(uint256 amount, uint256 weight)
        internal view returns(uint256) {
        return amount.mul(weight).div(uint256(10)**rsv.decimals());
    }

}
