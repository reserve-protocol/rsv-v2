pragma solidity 0.5.7;

import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/math/SafeMath.sol";
import "./rsv/IRSV.sol";
import "./ownership/Ownable.sol";
import "./Basket.sol";
import "./Proposal.sol";


interface IVault {
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

    Basket public trustedBasket;
    IVault public trustedVault;
    IRSV public trustedRSV;
    IProposalFactory public trustedProposalFactory;

    // Proposals
    mapping(uint256 => IProposal) public trustedProposals;
    uint256 public proposalsLength;
    uint256 public delay = 24 hours;

    // Controls
    bool public issuancePaused;
    bool public emergency;


    // The spread between issuance and redemption in basis points (BPS).
    uint256 public seigniorage;              // 0.1% spread -> 10 BPS. unit: BPS
    uint256 constant BPS_FACTOR = 10000;     // This is what 100% looks like in BPS. unit: BPS
    uint256 constant WEIGHT_SCALE = 10**18; // unit: aqToken/qToken

    event ProposalsCleared();

    // RSV traded events
    event Issuance(address indexed user, uint256 indexed amount);
    event Redemption(address indexed user, uint256 indexed amount);

    // Pause events
    event IssuancePausedChanged(bool indexed oldVal, bool indexed newVal);
    event EmergencyChanged(bool indexed oldVal, bool indexed newVal);
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

    /// Begins in `emergency` state.
    constructor(address vaultAddr,
        address rsvAddr,
        address proposalFactoryAddr,
        address operatorAddr,
        uint256 _seigniorage) public {
        require(_seigniorage <= 1000, "max seigniorage 10%");
        trustedVault = IVault(vaultAddr);
        trustedRSV = IRSV(rsvAddr);
        trustedProposalFactory = IProposalFactory(proposalFactoryAddr);
        operator = operatorAddr;
        seigniorage = _seigniorage;
        emergency = true; // it's not an emergency, but we want everything to start paused.

        // Start with the empty basket.
        trustedBasket = new Basket(Basket(0), new address[](0), new uint256[](0));
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

    /// Modifies a function to run and complete only if the vault is collateralized.
    modifier vaultCollateralized() {
        require(isFullyCollateralized(), "undercollateralized");
        _;
        assert(isFullyCollateralized());
    }

    // ========================= Public + External ============================

    /// Set if issuance should be paused. 
    function setIssuancePaused(bool val) external onlyOwner {
        emit IssuancePausedChanged(issuancePaused, val);
        issuancePaused = val;
    }

    /// Set if all contract actions should be paused.
    function setEmergency(bool val) external onlyOwner {
        emit EmergencyChanged(emergency, val);
        emergency = val;
    }

    /// Set the operator.
    function setOperator(address _operator) external onlyOwner {
        emit OperatorChanged(operator, _operator);
        operator = _operator;
    }

    /// Set the seigniorage, in BPS.
    function setSeigniorage(uint256 _seigniorage) external onlyOwner {
        require(_seigniorage <= 1000, "max seigniorage 10%");
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
        uint256 scaleFactor = WEIGHT_SCALE.mul(uint256(10) ** trustedRSV.decimals());
        // scaleFactor unit: aqToken/qToken * qRSV/RSV

        for (uint256 i = 0; i < trustedBasket.size(); i++) {

            address trustedToken = trustedBasket.tokens(i);
            uint256 weight = trustedBasket.weights(trustedToken); // unit: aqToken/RSV
            uint256 balance = IERC20(trustedToken).balanceOf(address(trustedVault)); //unit: qToken

            // Return false if this token is undercollateralized:
            if (trustedRSV.totalSupply().mul(weight) > balance.mul(scaleFactor)) {
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
        uint256[] memory amounts = new uint256[](trustedBasket.size());

        uint256 feeRate = uint256(seigniorage.add(BPS_FACTOR));
        // feeRate unit: BPS
        uint256 effectiveAmount = rsvAmount.mul(feeRate).div(BPS_FACTOR);
        // effectiveAmount unit: qRSV == qRSV*BPS/BPS

        // On issuance, amounts[i] of token i will enter the vault. To maintain full backing,
        // we have to round _up_ each amounts[i].
        for (uint256 i = 0; i < trustedBasket.size(); i++) {
            address trustedToken = trustedBasket.tokens(i);
            amounts[i] = _weighted(
                effectiveAmount, 
                trustedBasket.weights(trustedToken), 
                RoundingMode.UP
            );
            // unit: qToken = _weighted(qRSV, aqToken/RSV, _)
        }

        return amounts; // unit: qToken[]
    }

    /// Get amounts of basket tokens that would be sent upon redeeming an amount of RSV.
    /// The returned array will be in the same order as the current basket.tokens.
    /// return unit: qToken[]
    function toRedeem(uint256 rsvAmount) public view returns (uint256[] memory) {
        // rsvAmount unit: qRSV
        uint256[] memory amounts = new uint256[](trustedBasket.size());

        // On redemption, amounts[i] of token i will leave the vault. To maintain full backing,
        // we have to round _down_ each amounts[i].
        for (uint256 i = 0; i < trustedBasket.size(); i++) {
            address trustedToken = trustedBasket.tokens(i);
            amounts[i] = _weighted(
                rsvAmount, 
                trustedBasket.weights(trustedToken), 
                RoundingMode.DOWN
            );
            // unit: qToken = _weighted(qRSV, aqToken/RSV, _)
        }

        return amounts;
    }

    /// Handles issuance.
    /// rsvAmount unit: qRSV
    function issue(uint256 rsvAmount) external  
        issuanceNotPaused 
        notEmergency 
        vaultCollateralized 
    {
        require(rsvAmount > 0, "cannot issue zero RSV");
        require(trustedBasket.size() > 0, "basket cannot be empty");

        // Accept collateral tokens.
        uint256[] memory amounts = toIssue(rsvAmount); // unit: qToken[]
        for (uint256 i = 0; i < trustedBasket.size(); i++) {
            IERC20(trustedBasket.tokens(i)).safeTransferFrom(
                _msgSender(), 
                address(trustedVault), 
                amounts[i]
            );
            // unit check for amounts[i]: qToken.
        }

        // Compensate with RSV.
        trustedRSV.mint(_msgSender(), rsvAmount);
        // unit check for rsvAmount: qRSV.

        emit Issuance(_msgSender(), rsvAmount);
    }

    /// Handles redemption.
    /// rsvAmount unit: qRSV
    function redeem(uint256 rsvAmount) external notEmergency vaultCollateralized {
        require(rsvAmount > 0, "cannot redeem 0 RSV");
        require(trustedBasket.size() > 0, "basket cannot be empty");

        // Burn RSV tokens.
        trustedRSV.burnFrom(_msgSender(), rsvAmount);
        // unit check: rsvAmount is qRSV.

        // Compensate with collateral tokens.
        uint256[] memory amounts = toRedeem(rsvAmount); // unit: qToken[]
        for (uint256 i = 0; i < trustedBasket.size(); i++) {
            trustedVault.withdrawTo(trustedBasket.tokens(i), amounts[i], _msgSender());
            // unit check for amounts[i]: qToken.
        }

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
     * If and when this proposal is accepted and executed, then:
     *
     * 1. The Manager checks that the proposer has allowed adequate funds, for the proposed
     *    transfers from the proposer to the vault.
     * 2. The proposed set of token transfers occur between the Vault and the proposer.
     * 3. The Vault's basket weights are raised and lowered, based on these token transfers and the
     *    total supply of RSV **at the time when the proposal is executed**.
     *
     * Note that the set of token transfers will almost always be at very slightly lower volumes
     * than requested, due to the rounding error involved in (a) adjusting the weights at execution
     * time and (b) keeping the Vault fully collateralized. The contracts should never attempt to
     * trade at higher volumes than requested.
     *
     * The intended behavior of proposers is that they will make proposals that shift the Vault
     * composition towards some known target of Reserve's management while maintaining full
     * backing; the expected behavior of Reserve's management is to accept only such proposals,
     * excepting during dire emergencies.
     *
     * Note: This type of proposal does not reliably remove token addresses!
     * If you want to remove token addresses entirely, use proposeWeights.
     *
     * Returns the new proposal's ID.
     */
    function proposeSwap(
        address[] calldata tokens,
        uint256[] calldata amounts, // unit: qToken
        bool[] calldata toVault
    )
    external notEmergency vaultCollateralized returns(uint256)
    {
        require(tokens.length == amounts.length && amounts.length == toVault.length,
            "proposeSwap: unequal lengths");
        uint256 proposalID = proposalsLength++;

        trustedProposals[proposalID] = trustedProposalFactory.createSwapProposal(
            _msgSender(), 
            tokens, 
            amounts, 
            toVault
        );
        trustedProposals[proposalID].acceptOwnership();

        emit SwapProposed(proposalID, _msgSender(), tokens, amounts, toVault);
        return proposalID;
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
    external notEmergency vaultCollateralized returns(uint256)
    {
        require(tokens.length == weights.length, "proposeWeights: unequal lengths");
        require(tokens.length > 0, "proposeWeights: zero length");

        uint256 proposalID = proposalsLength++;

        trustedProposals[proposalID] = trustedProposalFactory.createWeightProposal(
            _msgSender(), 
            new Basket(Basket(0), tokens, weights)
        );
        trustedProposals[proposalID].acceptOwnership();

        emit WeightsProposed(proposalID, _msgSender(), tokens, weights);
        return proposalID;
    }

    /// Accepts a proposal for a new basket, beginning the required delay.
    function acceptProposal(uint256 id) external onlyOperator notEmergency vaultCollateralized {
        require(proposalsLength > id, "proposals length > id");
        trustedProposals[id].accept(now.add(delay));
        emit ProposalAccepted(id, trustedProposals[id].proposer());
    }

    /// Cancels a proposal. This can be done anytime before it is enacted by any of:
    /// 1. Proposer 2. Operator 3. Owner
    function cancelProposal(uint256 id) external notEmergency vaultCollateralized {
        require(
            _msgSender() == trustedProposals[id].proposer() ||
            _msgSender() == owner() ||
            _msgSender() == operator,
            "cannot cancel"
        );
        require(proposalsLength > id, "proposals length > id");
        trustedProposals[id].cancel();
        emit ProposalCanceled(id, trustedProposals[id].proposer(), _msgSender());
    }

    /// Executes a proposal by exchanging collateral tokens with the proposer.
    function executeProposal(uint256 id) external onlyOperator notEmergency vaultCollateralized {
        require(proposalsLength > id, "proposals length > id");
        address proposer = trustedProposals[id].proposer();
        Basket trustedOldBasket = trustedBasket;

        // Complete proposal and compute new basket
        trustedBasket = trustedProposals[id].complete(trustedRSV, trustedOldBasket);

        // For each token in either basket, perform transfers between proposer and Vault
        for (uint256 i = 0; i < trustedOldBasket.size(); i++) {
            address trustedToken = trustedOldBasket.tokens(i);
            _executeBasketShift(
                trustedOldBasket.weights(trustedToken),
                trustedBasket.weights(trustedToken),
                trustedToken,
                proposer
            );
        }
        for (uint256 i = 0; i < trustedBasket.size(); i++) {
            address trustedToken = trustedBasket.tokens(i);
            if (!trustedOldBasket.has(trustedToken)) {
                _executeBasketShift(
                    trustedOldBasket.weights(trustedToken),
                    trustedBasket.weights(trustedToken),
                    trustedToken,
                    proposer
                );
            }
        }

        emit ProposalExecuted(
            id,
            proposer,
            _msgSender(),
            address(trustedOldBasket),
            address(trustedBasket)
        );
    }


    // ============================= Internal ================================

    /// _executeBasketShift transfers the necessary amount of `token` between vault and `proposer`
    /// to rebalance the vault's balance of token, as it goes from oldBasket to newBasket.
    /// @dev To carry out a proposal, this is executed once per relevant token.
    function _executeBasketShift(
        uint256 oldWeight, // unit: aqTokens/RSV
        uint256 newWeight, // unit: aqTokens/RSV
        address trustedToken,
        address proposer
    ) internal {
        if (newWeight > oldWeight) {
            // This token must increase in the vault, so transfer from proposer to vault.
            // (Transfer into vault: round up)
            uint256 transferAmount =_weighted(
                trustedRSV.totalSupply(), 
                newWeight.sub(oldWeight), 
                RoundingMode.UP
            );
            // transferAmount unit: qTokens

            if (transferAmount > 0) {
                IERC20(trustedToken).safeTransferFrom(
                    proposer, 
                    address(trustedVault), 
                    transferAmount
                );
            }

        } else if (newWeight < oldWeight) {
            // This token will decrease in the vault, so transfer from vault to proposer.
            // (Transfer out of vault: round down)
            uint256 transferAmount =_weighted(
                trustedRSV.totalSupply(), 
                oldWeight.sub(newWeight), 
                RoundingMode.DOWN
            );
            // transferAmount unit: qTokens
            if (transferAmount > 0) {
                trustedVault.withdrawTo(trustedToken, transferAmount, proposer);
            }
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
        uint256 scaleFactor = WEIGHT_SCALE.mul(uint256(10)**(trustedRSV.decimals()));
        // scaleFactor unit: aqTokens/qTokens * qRSV/RSV
        uint256 shiftedWeight = amount.mul(weight);
        // shiftedWeight unit: qRSV/RSV * aqTokens

        // If the weighting is precise, or we're rounding down, then use normal division.
        if (rnd == RoundingMode.DOWN || shiftedWeight.mod(scaleFactor) == 0) {
            return shiftedWeight.div(scaleFactor);
            // return unit: qTokens == qRSV/RSV * aqTokens * (qTokens/aqTokens * RSV/qRSV)
        }
        return shiftedWeight.div(scaleFactor).add(1); // return unit: qTokens
    }
}

