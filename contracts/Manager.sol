pragma solidity ^0.5.8;

import "./zeppelin/token/ERC20/SafeERC20.sol";
import "./zeppelin/token/ERC20/IERC20.sol";
import "./zeppelin/math/SafeMath.sol";
import "./Ownable.sol";
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
    uint256 public constant delay = 24 hours;
    
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

    // Whitelist events
    event Whitelisted(address indexed user);
    event DeWhitelisted(address indexed user);

    // Pause events
    event Paused(address indexed account);
    event Unpaused(address indexed account);

    // Changes
    event OperatorChanged(address indexed account);
    event RSVChanged(address indexed account);
    event VaultChanged(address indexed account);
    event SeigniorageChanged(uint256 oldVal, uint256 newVal);


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
        require(!paused, "contract is paused");
        _;
    }

    /// Modifies a function to run only when the caller is on the whitelist, if it is enabled.
    modifier onlyWhitelist() {
        if (useWhitelist) require(whitelist[_msgSender()], "not on whitelist");
        _;
    }

    /// Modifies a function to run only when the caller is the operator account. 
    modifier onlyOperator() {
        require(_msgSender() == operator, "operator only");
        _;
    }


    // ============================= Externals ================================

    /// Issue a quantity of RSV to the caller and deposit collateral tokens in the Vault.
    function issue(uint256 _rsvQuantity) external notPaused onlyWhitelist {
        _issue(_rsvQuantity);
    }

    /// Issues the maximum amount of RSV to the caller based on their allowances.
    function issueMax() external notPaused onlyWhitelist {
        uint256 max = _calculateMaxIssuable(_msgSender());
        _issue(max);
    }

    /// Redeem a quantity of RSV for collateral tokens. 
    function redeem(uint256 _rsvQuantity) external notPaused onlyWhitelist {
        _redeem(_rsvQuantity);
    }

    /// Redeem `allowance` of RSV from the caller's account. 
    function redeemMax() external notPaused onlyWhitelist {
        uint256 max = rsv.allowance(_msgSender(), address(this));
        _redeem(max);
    }

    // /**
    //  * Proposes an adjustment to the quantities of tokens in the Vault. Importantly, this type of
    //  * proposal does not change token addresses. Therefore, if you want to introduce a new token,
    //  * first use the other proposal type. 
    //  */ 
    // function proposeQuantitiesAdjustment( 
    //     uint256[] calldata _amountsIn,
    //     uint256[] calldata _amountsOut
    // ) 
    //     external returns(uint256)
    // {
    //     require(_amountsIn.length == _amountsOut.length, "quantities mismatched");

    //     proposals[proposalsLength] = new Proposal(
    //         proposalsLength,
    //         _msgSender(),
    //         basket.getTokens(),
    //         _amountsIn,
    //         _amountsOut,
    //         Basket(0)
    //     );

    //     return ++proposalsLength;
    // }

    
    /**
     * Proposes a new basket defined by a list of tokens and their backing quantities. 
     * Importantly, this type of proposal means the balances that will be required from the 
     * proposer at the time of execution are variable. If the supply of RSV changes significantly,
     * then much more tokens could be required to execute the proposal. 
     * 
     */ 
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

        return ++proposalsLength;
    }

    /// Accepts a proposal for a new basket, beginning the required delay.
    function acceptProposal(uint256 _proposalID) external onlyOperator {
        require(proposalsLength > _proposalID, "proposals length < id");
        proposals[_proposalID].accept(now + delay);
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
    }

    /// Executes a proposal by exchanging collateral tokens with the proposer.
    function executeProposal(uint256 _proposalID) external {
        require(proposalsLength > _proposalID, "proposals length < id");
        Proposal proposal = proposals[_proposalID];
        proposal.prepare(rsv.totalSupply(), address(vault), basket);
        address[] memory tokens = proposal.getTokens();
        uint256[] memory quantitiesIn = proposal.getQuantitiesIn();

        // Proposer -> Vault
        IERC20 token;
        for (uint i = 0; i < tokens.length; i++) {
            token = IERC20(tokens[i]);
            require(
                token.allowance(proposal.proposer(), address(this)) >= quantitiesIn[i], 
                "allowances insufficient"
            );
            token.safeTransferFrom(proposal.proposer(), address(vault), quantitiesIn[i]);
        }

        // Vault -> Proposer
        vault.batchWithdrawTo(tokens, proposal.getQuantitiesOut(), proposal.proposer());

        _assertFullyCollateralized();
        proposal.complete();
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

    /// Add user to whitelist.
    function whitelistAccount(address _user) external onlyOwner {
        whitelist[_user] = true;
        emit Whitelisted(_user);
    }

    /// Remove user from whitelist.
    function deWhitelistAccount(address _user) external onlyOwner {
        whitelist[_user] = false;
        emit DeWhitelisted(_user);
    }

    /// Set whether or not to apply the whitelist to Issuance and Redemption. 
    function setUseWhitelist(bool _useWhitelist) external onlyOwner {
        useWhitelist = _useWhitelist;
    }

    /// Set the operator
    function setOperator(address _operator) external onlyOwner {
        operator = _operator;
        emit OperatorChanged(operator);
    }

    /// Set the RSV contract address. 
    function setRSV(address _rsv) external onlyOwner {
        rsv = IRSV(_rsv);
        emit RSVChanged(_rsv);
    }

    // Set the Vault contract address. 
    function setVault(address _vault) external onlyOwner {
        vault = IVault(_vault);
        emit VaultChanged(_vault);
    }

    /// Set the seigniorage, in BPS. 
    function setSegniorage(uint256 _seigniorage) external onlyOwner {
        seigniorage = _seigniorage;
        emit SeigniorageChanged(seigniorage, _seigniorage);
    }

    function clearProposals() external onlyOwner {
        proposalsLength = 0;
        emit ProposalsCleared();
    }

    /// Get the tokens in the basket. 
    function basketTokens() external view returns (address[] memory) {
        return basket.getTokens();
    }

    /// Get requirements required for the proposal to be accepted, in terms of proposal tokens. 
    function requirementsForProposal(uint256 _proposalID) external view returns(address[] memory, uint256[] memory) {
        return (proposals[_proposalID].getTokens(), 
            basket.newQuantitiesRequired(rsv.totalSupply(), proposals[_proposalID].basket()));
    }

    /// Get quantities required to issue a quantity of RSV, in terms of basket tokens.  
    function toIssue(uint256 _rsvQuantity) external view returns (uint256[] memory) {
        return _quantitiesRequiredToIssue(_rsvQuantity);
    }


    // ============================= Internals ================================

    /// Internal function for all issuances to go through.
    function _issue(uint256 _rsvQuantity) internal {
        require(_rsvQuantity > 0, "cannot issue zero RSV");
        uint256[] memory quantities = _quantitiesRequiredToIssue(_rsvQuantity);

        // Intake collateral tokens.
        IERC20 token;
        for (uint i = 0; i < basket.size(); i++) {
            token = IERC20(basket.tokens(i));
            require(token.allowance(_msgSender(), address(this)) >= quantities[i], "please set allowance");
            require(token.balanceOf(_msgSender()) >= quantities[i], "insufficient balance");
            token.safeTransferFrom(_msgSender(), address(vault), quantities[i]);
        }

        // Compensate with RSV.
        rsv.mint(_msgSender(), _rsvQuantity);

        _assertFullyCollateralized();
        emit Issuance(_msgSender(), _rsvQuantity);
    }

    /// Internal function for all redemptions to go through.
    function _redeem(uint256 _rsvQuantity) internal {
        require(_rsvQuantity > 0, "cannot redeem 0 RSV");

        // Burn RSV tokens.
        rsv.burnFrom(_msgSender(), _rsvQuantity);

        // Compensate with collateral tokens.
        vault.batchWithdrawTo(
            basket.getTokens(), 
            basket.quantitiesRequired(_rsvQuantity), 
            _msgSender()
        );

        _assertFullyCollateralized();
        emit Redemption(_msgSender(), _rsvQuantity);
    }

    /// Calculates the quantities of tokens required to issue `_rsvQuantity`. 
    function _quantitiesRequiredToIssue(uint256 _rsvQuantity) internal view returns(uint256[] memory) {
        uint256[] memory quantities = basket.quantitiesRequired(_rsvQuantity);
        uint256 seigniorageMultiplier = uint256(seigniorage.add(BPS_FACTOR));

        for (uint i = 0; i < basket.size(); i++) {
            quantities[i] = quantities[i].mul(seigniorageMultiplier).div(BPS_FACTOR);
        }
    }

    /// Calculates the maximum we could issue to an address based on their allowances.
    function _calculateMaxIssuable(address funder) internal view returns(uint256) {
        uint256 rsvDecimalsFactor = uint256(10) ** rsvDecimals;
        uint256 allowance;
        uint256 balance;
        uint256 available;
        uint256 issuable;
        uint256 minIssuable;

        for (uint i = 0; i < basket.size(); i ++) {
            allowance = IERC20(basket.tokens(i)).allowance(funder, address(this));
            balance = IERC20(basket.tokens(i)).balanceOf(funder);
            available = allowance;
            if (balance < available) available = balance;

            issuable = rsvDecimalsFactor.mul(available).div(basket.backing(i));
            if (issuable < minIssuable) minIssuable = issuable;
        }
        return minIssuable;
    }

    /// Ensure that the Vault is fully collateralized. 
    function _assertFullyCollateralized() internal view {
        uint256[] memory expected = basket.quantitiesRequired(rsv.totalSupply());
        for (uint i = 0; i < basket.size(); i++) {
            assert(IERC20(basket.tokens(i)).balanceOf(address(vault)) >= expected[i]);
        }
    }
}
