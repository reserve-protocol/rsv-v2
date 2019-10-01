pragma solidity 0.5.7;

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
 * The Manager contract is the point of contact between the Reserve ecosystem and the
 * surrounding world. It manages the Issuance and Redemption of RSV, a decentralized stablecoin
 * backed by a basket of tokens.
 *
 * The Manager also implements a Proposal system to handle administration of changes to the
 * backing of RSV. Anyone can propose a change to the backing.  Once the `owner` approves the
 * proposal, then after a pre-determined delay the proposal is eligible for execution by
 * anyone. However, the funds to execute the proposal must come from the proposer.
 *
 * There are two different ways to propose changes to the backing of RSV:
 * - proposeSwap()
 * - proposeWeights()
 *
 * In both cases, tokens are exchanged with the Vault and a new RSV backing is set. You can
 * think of the first type of proposal as being useful when you don't want to rebalance the
 * Vault by exchanging absolute quantities of tokens; its downside is that you don't know
 * precisely what the resulting basket weights will be. The second type of proposal is more
 * useful when you want to fine-tune the Vault weights and accept the downside that it's
 * difficult to know what capital will be required when the proposal is executed.
 */

/* On "unit" comments:
 *
 * The units in use around weight computations are fiddly, and it's pretty annoying to get them
 * properly into the Solidity type system. So, there are many comments of the form "unit:
 * ...". Where such a comment is describing a field, method, or return parameter, the comment means
 * that the data in that place is to be interpreted to have that type. Many places also have
 * comments with more complicated expressions; that's manually working out the dimensional analysis
 * to ensure that the given expression has correct units.
 *
 * Some dimensions used in this analysis:
 * - 1 RSV: 1 Reserve
 * - 1 qRSV: 1 quantum of Reserve.
 *      (RSV & qRSV are convertible by .mul(10**reserve.decimals() qRSV/RSV))
 * - 1 qToken: 1 quantum of an external Token.
 * - 1 aqToken: 1 atto-quantum of an external Token.
 *      (qToken and aqToken are convertible by .mul(10**18 aqToken/qToken)
 * - 1 BPS: 1 Basis Point. Effectively dimensionless; convertible with .mul(10000 BPS).
 *
 * Note that we _never_ reason in units of Tokens or attoTokens.
 */
contract Manager is Ownable {
    using SafeERC20 for IERC20;
    using SafeMath for uint256;

    // ROLES

    // Manager is already Ownable, but in addition it also has an `operator`.
    address public operator;

    // DATA

    Basket public basket;
    IVault public vault;
    IRSV public rsv;

    // Proposals
    mapping(uint256 => Proposal) public proposals;
    uint256 public proposalsLength;
    uint256 public delay = 24 hours;

    // Pausing
    bool public issuancePaused;
    bool public emergency;

    // The spread between issuance and redemption in basis points (BPS).
    uint256 public seigniorage;              // 0.1% spread -> 10 BPS. unit: BPS
    uint256 constant BPS_FACTOR = 10000;     // This is what 100% looks like in BPS. unit: BPS
    uint256 constant WEIGHT_FACTOR = 10**18; // unit: aqToken/qToken

    event ProposalsCleared();

    // RSV traded events
    event Issuance(address indexed user, uint256 indexed amount);
    event Redemption(address indexed user, uint256 indexed amount);

    // Pause events
    event IssuancePaused(address indexed account);
    event IssuanceUnpaused(address indexed account);
    event PausedForEmergency(address indexed account);
    event UnpausedFromEmergency(address indexed account);

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
        emergency = true; // it's not an emergency, but we want everything to start paused.

        // Start with the empty basket.
        address[] memory tokens = new address[](0);
        uint256[] memory weights = new uint256[](0);
        basket = new Basket(Basket(0), tokens, weights);
    }

    // ============================= Modifiers ================================

    /// Modifies a function to run only when issuance is not paused.
    modifier issuanceNotPaused() {
        require(!issuancePaused, "issuance is paused");
        _;
    }

    /// Modifies a function to run only when there is not some emergency that requires upgrades.
    modifier notEmergency() {
        require(!emergency, "contract is paused");
        _;
    }

    /// Modifies a function to run only when the caller is the operator account.
    modifier onlyOperator() {
        require(_msgSender() == operator, "operator only");
        _;
    }

    // ========================= Public + External ============================

    /// Pause issuance.
    function pauseIssuance() external onlyOwner {
        issuancePaused = true;
        emit IssuancePaused(_msgSender());
    }

    /// Unpause issuance.
    function unpauseIssuance() external onlyOwner {
        require(basket.size() > 0, "basket cannot be empty");
        issuancePaused = false;
        emit IssuanceUnpaused(_msgSender());
    }

    /// Pause contract.
    function pauseForEmergency() external onlyOwner {
        emergency = true;
        emit PausedForEmergency(_msgSender());
    }

    /// Unpause contract.
    function unpauseForEmergency() external onlyOwner {
        require(basket.size() > 0, "basket cannot be empty");
        emergency = false;
        emit UnpausedFromEmergency(_msgSender());
    }

    /// Set the operator.
    function setOperator(address _operator) external onlyOwner {
        emit OperatorChanged(operator, _operator);
        operator = _operator;
    }

    /// Set the seigniorage, in BPS.
    function setSeigniorage(uint256 _seigniorage) external onlyOwner {
        emit SeigniorageChanged(seigniorage, _seigniorage);
        seigniorage = _seigniorage;
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

    /// Ensure that the Vault is fully collateralized.  That this is true should be an
    /// invariant of this contract: it's true before and after every txn.
    function isFullyCollateralized() public view returns(bool) {
        uint256 scaleFactor = WEIGHT_FACTOR.mul(uint256(10) ** rsv.decimals());
        // scaleFactor unit: aqToken/qToken * qRSV/RSV

        for (uint i = 0; i < basket.size(); i++) {

            address token = basket.tokens(i);
            uint256 weight = basket.weights(token); // unit: aqToken/RSV
            uint256 balance = IERC20(token).balanceOf(address(vault)); // unit: qRSV

            // Return false if this token is undercollateralized:
            if (rsv.totalSupply().mul(weight) > balance.mul(scaleFactor)) {
                // checking units: [qRSV] * [aqToken/RSV] == [qToken] * [aqToken/qToken * qRSV/RSV]
                return false;
            }
        }
        return true;
    }

    /// Get amounts of basket tokens required to issue an amount of RSV.
    /// The returned array will be in the same order as the current basket.tokens.
    /// return unit: qToken[]
    function toIssue(uint256 rsvAmount) public view returns (uint256[] memory) {
        // rsvAmount unit: qRSV.
        uint256[] memory amounts = new uint256[](basket.size());

        uint256 feeRate = uint256(seigniorage.add(BPS_FACTOR));
        // feeRate unit: BPS
        uint256 effectiveAmount = rsvAmount.mul(feeRate).div(BPS_FACTOR);
        // effectiveAmount unit: qRSV == qRSV*BPS/BPS

        // On issuance, amounts[i] of token i will enter the vault. To maintain full backing,
        // we have to round _up_ each amounts[i].
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            amounts[i] = _weighted(effectiveAmount, basket.weights(token), RoundingMode.UP);
            // unit: qToken = _weighted(qRSV, aqToken/RSV, _)
        }

        return amounts; // unit: qToken[]
    }

    /// Get amounts of basket tokens that would be sent upon redeeming an amount of RSV.
    /// The returned array will be in the same order as the current basket.tokens.
    /// return unit: qToken[]
    function toRedeem(uint256 rsvAmount) public view returns (uint256[] memory) {
        // rsvAmount unit: qRSV
        uint256[] memory amounts = new uint256[](basket.size());

        // On redemption, amounts[i] of token i will leave the vault. To maintain full backing,
        // we have to round _down_ each amounts[i].
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            amounts[i] = _weighted(rsvAmount, basket.weights(token), RoundingMode.DOWN);
            // unit: qToken = _weighted(qRSV, aqToken/RSV, _)
        }

        return amounts;
    }

    /// Handles issuance.
    /// rsvAmount unit: qRSV
    function issue(uint256 rsvAmount) external issuanceNotPaused notEmergency {
        require(rsvAmount > 0, "cannot issue zero RSV");

        // Accept collateral tokens.
        uint256[] memory amounts = toIssue(rsvAmount); // unit: qToken[]
        for (uint i = 0; i < basket.size(); i++) {
            IERC20(basket.tokens(i)).safeTransferFrom(_msgSender(), address(vault), amounts[i]);
            // unit check for amounts[i]: qToken.
        }

        // Compensate with RSV.
        rsv.mint(_msgSender(), rsvAmount);
        // unit check for rsvAmount: qRSV.

        assert(isFullyCollateralized());
        emit Issuance(_msgSender(), rsvAmount);
    }

    /// Handles redemption.
    /// rsvAmount unit: qRSV
    function redeem(uint256 rsvAmount) external notEmergency {
        require(rsvAmount > 0, "cannot redeem 0 RSV");

        // Burn RSV tokens.
        rsv.burnFrom(_msgSender(), rsvAmount);
        // unit check: rsvAmount is qRSV.

        // Compensate with collateral tokens.
        uint256[] memory amounts = toRedeem(rsvAmount); // unit: qToken[]
        for (uint i = 0; i < basket.size(); i++) {
            vault.withdrawTo(basket.tokens(i), amounts[i], _msgSender());
            // unit check for amounts[i]: qToken.
        }

        assert(isFullyCollateralized());
        emit Redemption(_msgSender(), rsvAmount);
    }

    /**
     * Propose an exchange of current Vault tokens for new Vault tokens.
     *
     * These parameters are phyiscally a set of arrays because Solidity doesn't let you pass
     * around arrays of structs as parameters of transactions. Semantically, read these three
     * lists as a list of triples (token, amount, toVault), where:
     *
     * - token is the address of an ERC-20 token,
     * - amount is the amount of the token that the proposer says they will trade with the vault,
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
        uint256[] calldata amounts, // unit: qToken
        bool[] calldata toVault
    )
        external notEmergency returns(uint256)
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
        external notEmergency returns(uint256)
    {
        require(tokens.length == weights.length, "proposeWeights: unequal lengths");
        require(tokens.length > 0, "proposeWeights: zero length");

        proposals[proposalsLength] =
            new WeightProposal(_msgSender(), new Basket(Basket(0), tokens, weights));

        emit WeightsProposed(proposalsLength, _msgSender(), tokens, weights);

        return ++proposalsLength;
    }

    /// Accepts a proposal for a new basket, beginning the required delay.
    function acceptProposal(uint256 id) external onlyOperator notEmergency {
        require(proposalsLength > id, "proposals length < id");
        proposals[id].accept(now + delay);
        emit ProposalAccepted(id, proposals[id].proposer());
    }

    /// Cancels a proposal. This can be done anytime before it is enacted by any of:
    /// 1. Proposer 2. Operator 3. Owner
    function cancelProposal(uint256 id) external notEmergency {
        require(
            _msgSender() == proposals[id].proposer() ||
            _msgSender() == owner() ||
            _msgSender() == operator,
            "cannot cancel"
        );
        proposals[id].cancel();
        emit ProposalCanceled(id, proposals[id].proposer(), _msgSender());
    }

    /// Executes a proposal by exchanging collateral tokens with the proposer.
    function executeProposal(uint256 id) external onlyOperator notEmergency {
        require(proposalsLength > id, "proposals length < id");
        address proposer = proposals[id].proposer();
        Basket oldBasket = basket;

        // Complete proposal and compute new basket
        basket = proposals[id].complete(rsv, oldBasket);

        // For each token in either basket, perform transfers between proposer and Vault
        for (uint i = 0; i < oldBasket.size(); i++) {
            address token = oldBasket.tokens(i);
            _executeBasketShift(oldBasket.weights(token), basket.weights(token), token, proposer);
        }
        for (uint i = 0; i < basket.size(); i++) {
            address token = basket.tokens(i);
            if (!oldBasket.has(token)) {
                _executeBasketShift(
                    oldBasket.weights(token), 
                    basket.weights(token), 
                    token, 
                    proposer
                );
            }
        }

        assert(isFullyCollateralized());
        emit ProposalExecuted(id, proposer, _msgSender(), address(oldBasket), address(basket));
    }


    // ============================= Internal ================================

    /// _executeBasketShift transfers the necessary amount of `token` between vault and `proposer`
    /// to rebalance the vault's balance of token, as it goes from oldBasket to newBasket.
    /// @dev To carry out a proposal, this is executed once per relevant token.
    function _executeBasketShift(
        uint256 oldWeight, // unit: aqTokens/RSV
        uint256 newWeight, // unit: aqTokens/RSV
        address token,
        address proposer
    ) internal {
        if (newWeight > oldWeight) {
            // This token must increase in the vault, so transfer from proposer to vault.
            // (Transfer into vault: round up)
            uint256 transferAmount =
                _weighted(rsv.totalSupply(), newWeight.sub(oldWeight), RoundingMode.UP);
                // transferAmount unit: qTokens
            if (transferAmount > 0)
                IERC20(token).safeTransferFrom(proposer, address(vault), transferAmount);

        } else if (newWeight < oldWeight) {
            // This token will decrease in the vault, so transfer from vault to proposer.
            // (Transfer out of vault: round down)
            uint256 transferAmount =
                _weighted(rsv.totalSupply(), oldWeight.sub(newWeight), RoundingMode.DOWN);
                // transferAmount unit: qTokens
            if (transferAmount > 0)
                vault.withdrawTo(token, transferAmount, proposer);
        }
    }

    // When you perform a weighting of some amount of RSV, it will involve a division, and
    // precision will be lost. When it rounds, do you want to round UP or DOWN? Be maximally
    // conservative.
    enum RoundingMode {UP, DOWN}

    /// From a weighting of RSV (e.g., a basket weight) and an amount of RSV,
    /// compute the amount of the weighted token that matches that amount of RSV.
    function _weighted(
        uint256 amount, // unit: qRSV
        uint256 weight, // unit: aqToken/RSV
        RoundingMode rnd
        ) internal view returns(uint256) // return unit: qTokens
    {
        // This wouldn't work properly with negative numbers, but we don't need them here.
        require(amount >= 0 && weight >= 0, "weight or amount negative");

        uint256 decimalsDivisor = WEIGHT_FACTOR.mul(uint256(10)**(rsv.decimals()));
        // decimalsDivisor unit: aqTokens/qTokens * qRSV/RSV
        uint256 shiftedWeight = amount.mul(weight);
        // shiftedWeight unit: qRSV/RSV * aqTokens

        // If the weighting is precise, or we're rounding down, then use normal division.
        if (rnd == RoundingMode.DOWN || shiftedWeight.mod(decimalsDivisor) == 0) {
            return shiftedWeight.div(decimalsDivisor);
            // return unit: qTokens == qRSV/RSV * aqTokens * (qTokens/aqTokens * RSV/qRSV)
        }
        return shiftedWeight.div(decimalsDivisor).add(1); // return unit: qTokens
    }
}
