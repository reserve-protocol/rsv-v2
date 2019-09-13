pragma solidity ^0.5.8;

import "./zeppelin/contracts/ownership/Ownable.sol";
import "./zeppelin/contracts/token/ERC20/SafeERC20.sol";
import "./zeppelin/contracts/token/ERC20/IERC20.sol";
import "./zeppelin/contracts/math/SafeMath.sol";
import "./zeppelin/contracts/utils/ReentrancyGuard.sol";
import "./Basket.sol";
import "./Vault.sol";


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
    function getDecimals() external view returns(uint8);
    function mint(address account, uint256 value) external;
    function burnFrom(address account, uint256 value) external;
}


contract Manager is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;
    using SafeMath for uint256;


    // TYPES

    struct Proposal {
        uint256 id;
        Basket b;
        bool closed;
    }


    // DATA

    Basket public basket;
    Vault public vault;
    IRSV public rsv;
    uint256 public rsvDecimalsFactor;

    // Proposals
    mapping(uint256 => Proposal) public proposals;
    uint256 public proposalsLength;

    // No issuance or redemption allowed while the Manager is paused.
    // You CAN accept a new basket and rebalance the vault while paused. 
    bool public paused;

    // Used to control who can issue and redeem.
    mapping(address => bool) public whitelist;
    bool public useWhitelist;

    // The spread between issuance and redemption in BPS.
    uint256 public seigniorage;                 // 0.1% spread -> 10 BPS
    uint256 constant BPS_FACTOR = 10000;        // This is what 100% looks like in BPS.
    

    // EVENTS


    // Basket events
    event ProposalCreated(uint256 indexed id, address indexed proposer, address[] tokens, uint256[] weights);
    event ProposalAccepted(uint256 indexed id, address indexed proposer);
    event ProposalCanceled(uint256 indexed id, address indexed proposer);
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

    // Seigniorage
    event SegniorageUpdated(uint256 oldVal, uint256 newVal);
    event RSVUpdated(address indexed account);
    event VaultUpdated(address indexed account);


    // === Constructor ===

    /// Begins paused.
    constructor(address vaultAddr, address rsvAddress, uint256 seigniorage_) public {
        vault = Vault(vaultAddr);
        rsv = IRSV(rsvAddress);
        whitelist[msg.sender] = true;
        seigniorage = seigniorage_;
        paused = true;
        useWhitelist = true;
        rsvDecimalsFactor = uint256(10) ** rsv.getDecimals(); // how to do with safemath?
    }

    // === Modifiers ===

    /// Modifies a function to run only when the contract is not paused.
    modifier notPaused() {
        require(!paused, "contract is paused");
        _;
    }

    /// Modifies a function to run only when the caller is on the whitelist, if it is enabled.
    modifier onlyWhitelist() {
        if (useWhitelist) require(whitelist[msg.sender], "unauthorized: not on whitelist");
        _;
    }


    // === Externals ===

    /// Issue RSV to the caller and deposit collateral tokens in the Vault.
    function issue(uint256 _rsvQuantity) external notPaused nonReentrant onlyWhitelist {
        IERC20 token;
        uint256[] memory amounts = _getIssuanceAmounts(_rsvQuantity);

        // Intake collateral tokens.
        for (uint i = 0; i < basket.size(); i++) {
            token = IERC20(basket.tokens(i));
            require(token.allowance(msg.sender, address(this)) >= amounts[i], "please set allowance");
            require(token.balanceOf(msg.sender) >= amounts[i], "insufficient balance");
            token.safeTransferFrom(msg.sender, address(vault), amounts[i]);
        }

        // Compensate with RSV.
        rsv.mint(msg.sender, _rsvQuantity);

        _assertFullyCollateralized();
        emit Issuance(msg.sender, _rsvQuantity);
    }

    /// Withdraw collateral tokens from the vault and redeem RSV. 
    function redeem(uint256 _rsvQuantity) external notPaused nonReentrant onlyWhitelist {
        // Burn RSV tokens.
        rsv.burnFrom(msg.sender, _rsvQuantity);

        // Compensate with collateral tokens.
        vault.batchWithdrawTo(
            basket.getTokens(), 
            _getRedemptionAmounts(_rsvQuantity), 
            msg.sender
        );

        _assertFullyCollateralized();
        emit Redemption(msg.sender, _rsvQuantity);
    }

    /// Proposes a new basket. Returns and emits the proposal id. 
    function proposeBasket(
        address[] calldata _tokens, 
        uint256[] calldata _amounts, 
        uint256 _amountsSum
    ) 
        external nonReentrant returns(uint256)
    {
        proposals[proposalsLength] = Proposal({
            id: proposalsLength,
            b: new Basket(msg.sender, _tokens, _amounts, _amountsSum),
            closed: false
        });

        emit ProposalCreated(proposalsLength, msg.sender, _tokens, _amounts);
        return ++proposalsLength;
    }

    /// Accepts a proposal for a new basket and exchanges collateral tokens with the proposer.
    function acceptProposal(uint256 _proposalID) external nonReentrant onlyOwner {
        Basket b = proposals[_proposalID].b;
        require(proposalsLength > _proposalID, "proposals length is shorter than id");
        require(!proposals[_proposalID].closed, "proposal is closed");
        require(b.size() > 0, "proposal at proposalID does not contain a valid basket");

        _rebalance(b);
        _assertFullyCollateralized();
        
        basket = b;
        emit ProposalAccepted(_proposalID, b.proposer());
    }

    // Cancels a proposal. 
    function cancelProposal(uint256 _proposalID) external nonReentrant {
        Proposal storage proposal = proposals[_proposalID];
        require(!proposal.closed, "proposal is already closed");
        require(msg.sender == proposal.b.proposer());
        proposal.closed = true;
        emit ProposalCanceled(_proposalID, msg.sender);
    }

    /// Get amounts required for the proposal to be accepted, in terms of proposal tokens. 
    function getAmountsNecessaryForProposal(uint256 _proposalID) external view returns(uint256[] memory) {
        return _getRebalanceAmounts(basket, proposals[_proposalID].b);
    }

    /// Get amounts required to issue a quantity of RSV, in terms of basket tokens.  
    function getAmountsNecessaryToIssue(uint256 _rsvQuantity) external view returns (uint256[] memory) {
        return _getIssuanceAmounts(_rsvQuantity);
    }

    /// Get the tokens in the basket. 
    function getBasketTokens() external view returns (address[] memory) {
        return basket.getTokens();
    }

    /// Pause the contract.
    function pause() external onlyOwner {
        paused = true;
        emit Paused(msg.sender);
    }

    /// Unpause the contract.
    function unpause() external onlyOwner {
        require(address(basket) != address(0), "can't unpause without a target basket");
        paused = false;
        emit Unpaused(msg.sender);
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

    /// Set the RSV contract address. 
    function setRSV(address _rsv) external onlyOwner {
        rsv = IRSV(_rsv);
        rsvDecimalsFactor = uint256(10) ** rsv.getDecimals(); // how to do with safemath?
        emit RSVUpdated(_rsv);
    }

    // Set the Vault contract address. 
    function setVault(address _vault) external onlyOwner {
        vault = Vault(_vault);
        emit VaultUpdated(_vault);
    }

    /// Set the seigniorage, in BPS. 
    function setSegniorage(uint256 _seigniorage) external onlyOwner {
        seigniorage = _seigniorage;
        emit SegniorageUpdated(seigniorage, _seigniorage);
    }

    function clearProposals() external onlyOwner {
        proposalsLength = 0;
        emit ProposalsCleared();
    }


    // === Internals ===

    /// Rebalance ERC20s across the funder and Vault. 
    function _rebalance(Basket _newBasket) internal {
        // Transfer deficit amounts from funder to the Vault.
        IERC20 token;
        uint256[] memory deficits = _getRebalanceAmounts(_newBasket, basket);
        for (uint i = 0; i < _newBasket.size(); i++) {
            if (deficits[i] > 0) {
                token = IERC20(_newBasket.tokens(i));
                require(
                    token.allowance(_newBasket.proposer(), address(this)) >= deficits[i], 
                    "allowances insufficient"
                );
                token.safeTransferFrom(_newBasket.proposer(), address(vault), deficits[i]);
            }
        }

        // Transfer excess amounts from the Vault to the funder.
        uint256[] memory excesses = _getRebalanceAmounts(basket, _newBasket);
        vault.batchWithdrawTo(basket.getTokens(), excesses, _newBasket.proposer());
    }

    /// Ensure that the Vault is fully collateralized. 
    function _assertFullyCollateralized() internal view {
        address token;
        uint256 expected;
        for (uint i = 0; i < basket.size(); i++) {
            expected = rsv.totalSupply().mul(basket.amounts(i)).div(BPS_FACTOR);
            assert(IERC20(token).balanceOf(address(vault)) >= expected);
        }
    }

    /// Calculates necessary extra tokens needed to go from _b1 to _b2.
    function _getRebalanceAmounts(Basket _b1, Basket _b2) internal view returns(uint256[] memory) {
        uint256[] memory missingAmountsPerRSV = _b2.excessAmountsRelativeToOtherBasket(_b1);
        for (uint i = 0; i < _b2.size(); i++) {
            missingAmountsPerRSV[i] = 
                rsv.totalSupply().mul(missingAmountsPerRSV[i]).div(rsvDecimalsFactor);
        }

        return missingAmountsPerRSV;
    }

    /// Calculates the amounts a user would need in order to issue a quantity of RSV.
    function _getIssuanceAmounts(uint256 _rsvQuantity) internal view returns(uint256[] memory) {
        // There are 10000 BPS in 100%. 
        uint256 seigniorageMultiplier = uint256(seigniorage.add(BPS_FACTOR));
        uint256[] memory amounts = new uint256[](basket.size());
        for (uint i = 0; i < basket.size(); i++) {
            amounts[i] = _rsvQuantity
                .mul(basket.amounts(i))
                .mul(seigniorageMultiplier)
                .div(rsvDecimalsFactor)
                .div(BPS_FACTOR);
        }

        return amounts;
    }

    /// Calculates the amounts a user would receive for redeeming a quantity of RSV.
    function _getRedemptionAmounts(uint256 _rsvQuantity) internal view returns(uint256[] memory) {
        uint256[] memory amounts = new uint256[](basket.size());
        for (uint i = 0; i < basket.size(); i++) {
            amounts[i] = _rsvQuantity.mul(basket.amounts(i)).div(rsvDecimalsFactor);
        }

        return amounts;
    }
}
