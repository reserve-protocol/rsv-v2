pragma solidity ^0.5.11;

import "./zeppelin/contracts/ownership/Ownable.sol";
import "./zeppelin/contracts/token/ERC20/SafeERC20.sol";
import "./zeppelin/contracts/math/SafeMath.sol";
import "./Vault.sol";


interface IRSV {
    // Standard ERC20 functions
    function transfer(address, uint256) external returns (bool);
    function approve(address, uint256) external returns (bool);
    function transferFrom(address, address, uint256) external returns (bool);
    function totalSupply() external view returns (uint256);
    function balanceOf(address) external view returns (uint256);
    function allowance(address, address) external view returns (uint256);
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed holder, address indexed spender, uint256 value);

    // RSV-specific functions
    function mint(address account, uint256 value) external ;
    function burnFrom(address account, uint256 value) external;
}


contract Manager is Ownable {
    using SafeMath for uint256;

    uint256 constant BPS_MULTIPLIER = 10000;


    // TYPES

    struct Basket {
        address[] tokens;
        mapping(address => uint256) weights; // In BPS. A weight of 30% would be stored as 3000
    }

    struct Proposal {
        uint256 id;
        address proposer;
        Basket basket;
        bool closed;
    }


    // DATA

    Basket public currentBasket;
    Vault public vault;
    IRSV public rsv;

    // Proposals
    mapping(uint256 => Proposal) public proposals;
    uint256 public proposalsLength;

    // No issuance, redemption, or rebalancing allowed while the Manager is paused.
    // You CAN set the RSV and Vault addresses while paused. 
    bool public paused;

    // Used to control who can issue and redeem.
    mapping(address => bool) public whitelist;
    bool public useWhitelist;

    // The spread between issuance and redemption in BPS.
    uint256 public seigniorage;    // 0.1% spread -> 10 BPS
    

    // EVENTS

    // Seigniorage
    event SegniorageUpdated(uint256 oldVal, uint256 newVal);

    // Pause events
    event Paused(address indexed account);
    event Unpaused(address indexed account);

    // Basket events
    event BasketProposed(uint256 indexed id, address indexed proposer, address[] tokens, uint256[] weights, uint256 size);
    event BasketAccepted(uint256 indexed id, address indexed proposer);
    event BasketClosed(uint256 indexed id, address indexed proposer);

    // Whitelist events
    event Whitelisted(address indexed user);
    event DeWhitelisted(address indexed user);

    // RSV traded events
    event Issuance(address indexed user, uint256 indexed amount);
    event Redemption(address indexed user, uint256 indexed amount);


    // === Constructor ===

    // Begins paused
    constructor(address vaultAddr, address rsvAddress, uint256 seigniorage_) public {
        vault = Vault(vaultAddr);
        rsv = IRSV(rsvAddress);
        whitelist[msg.sender] = true;
        seigniorage = seigniorage_;
        paused = true;
        useWhitelist = true;
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
    }


    // === Externals ===

    /// Issue RSV to the caller and collect collateral tokens.
    function issue(uint256 _amount) external notPaused onlyWhitelist {
        // Do checks
        uint256[] memory toBuy = _collateralAmountsToTrade(currentBasket.tokens, _amount, seigniorage);
        uint256 sum = 0;
        for (uint i = 0; i < collateralTokens.length; i++) {
            require(SafeERC20(collateralTokens[i]).allowance(msg.sender, address(this)) >= toBuy[i], "please set allowance");
            require(SafeERC20(collateralTokens[i]).balanceOf(msg.sender) >= toBuy[i], "insufficient balance");
            sum += toBuy[i];
        }

        require(sum >= _amount, "there should be seigniorage");

        // Intake collateral
        for (uint j = 0; j < collateralTokens.length; j++) {
            SafeERC20(collateralTokens[j]).safeTransferFrom(msg.sender, address(vault), toBuy[j]);
        }

        // Hand out RSV
        rsv.mint(msg.sender, _amount);

        emit Issuance(msg.sender, _amount, toBuy);
    }

    /// Burn RSV from the caller's account and compensate them with collateral tokens.
    function redeem(uint256 _amount) external notPaused onlyWhitelist {
        require(rsv.allowance(msg.sender, address(this)) >= _amount, "please set allowance");
        require(rsv.balanceOf(msg.sender) >= _amount, "insufficient rsv to redeem");

        uint256[] memory toSell = _collateralAmountsToTrade(currentBasket.tokens, _amount, 0);
        uint256 sum = 0;
        for (uint i = 0; i < collateralTokens.length; i++) {
            sum += toSell[i];
        }

        require(sum <= _amount, "we shouldn't sell more than the redemption amount");

        // Intake RSV
        rsv.burnFrom(msg.sender, _amount);

        // Hand out collateral
        vault.batchWithdrawTo(collateralTokens, toSell, msg.sender);

        emit Redemption(msg.sender, _amount, toSell);
    }

    /// Proposes a new basket. Returns and emits the proposal id. 
    function proposeBasket(address[] _tokenAddresses, uint256[] _weights, uint256 _size) external returns(uint256)        {
        Basket memory b = _createBasket(_tokenAddresses, _weights, _size); // Runs all necessary checks.

        proposals.push(Proposal({
            id: proposalsLength,
            proposer: msg.sender,
            basket: b,
            closed: false
        }));

        emit BasketProposed(proposalsLength, msg.sender, _tokenAddresses, _weights, _size);
        return ++proposalsLength;
    }

    /// Accepts a new basket if and only if the basket can be achieved by exchanging tokens with the proposer. 
    function acceptBasket(uint256 _proposalID) external onlyOwner {
        Basket memory b = proposals[_proposalID].basket;
        require(b.tokens.length > 0, "proposal at proposalID does not contain a valid basket");

        // Rebalance and set the new basket.
        _rebalance(proposals[_proposalID].proposer, currentBasket, b);
        currentBasket = b;
        
        // Double check everything went as planned.
        _assertFullyCollateralized();

        emit BasketAccepted(_proposalID, proposals[_proposalID].proposer);
    }

    /// Set the seigniorage, in BPS. 
    function setSegniorage(uint256 _seigniorage) external onlyOwner {
        emit SegniorageUpdated(seigniorage, _seigniorage);
        seigniorage = _seigniorage;
    }

    /// Pause the contract.
    function pause() external onlyOwner {
        paused = true;
        emit Paused(msg.sender);
    }

    /// Unpause the contract.
    function unpause() external onlyOwner {
        paused = false;
        emit Unpaused(msg.sender);
    }

    /// Add user to whitelist.
    function whitelist(address _user) external onlyOwner {
        whitelist[_user] = true;
        emit Whitelist(_user);
    }

    /// Remove user from whitelist.
    function deWhitelist(address _user) external onlyOwner {
        whitelist[_user] = false;
        emit DeWhitelisted(_user);
    }


    // === Internals ===

    function _collateralAmountsToTrade(address[] _tokens, uint256 rsvQuantity, uint256 seigniorage) internal pure returns(uint256[]) {
        uint256[] memory amounts = new uint256[](_tokens.length);
        uint256 adjustedQuantity = rsvQuantity.mul(seigniorage.add(BPS_MULTIPLIER)).div(BPS_MULTIPLIER);
        uint256 sum;

        for (uint i = 0; i < tokens.length; i++) {
            amounts[i] = adjustedQuantity.mul(currentBasket.weights[_tokens[i]]).div(BPS_MULTIPLIER);
            sum.add(amounts[i]);
        }

        assert(sum == rsvQuantity);
        return amounts;
    }

    /// Create the basket struct and do validation.
    function _createBasket(address[] _tokens, uint256[] _weights, uint256 _size) internal pure returns(Basket) {
        require(_size > 0, "size must be greater than zero");
        require(_size <= 1000, "size must be less than 1000"); // arbitrary max size
        require(_size == _tokens.length, "number of tokens must be equal to basket size");
        require(_size == _weights.length, "number of weights must be equal to basket size");

        mapping(address => uint256) memory weightsMap;
        uint256 sum = 0;
        for (uint i = 0; i < _size; i++) {
            weightsMap[_tokens[i]] = _weights[i];
            sum.add(_weights[i]);
        }

        require(sum == BPS_MULTIPLIER, "weights must be in BPS and sum to " + string(BPS_MULTIPLIER));
        return memory Basket({
            tokens: _tokens,
            weights: weightsMap
        });
    }

    /// Rebalance ERC20s across the funder and Vault. 
    function _rebalance(address funder, Basket _old, Basket _new) internal {
        // Determine what quantities of tokens need to be transferred where. 
        mapping(address => uint256) toDeposit = 
            _calculateMissingQuantities(_old, _new);
        mapping(address => uint256) toWithdraw = 
            _calculateMissingQuantities(_new, _old);

        // Transfer tokens from the funder to the Vault.
        address token;
        for (uint i = 0; i < _new.tokens.length; i++) {
            token = _new.tokens[i]
            if (toDeposit[token] > 0) {
                require(
                    SafeERC20(token).allowance(funder, address(this)) >= toDeposit[token], 
                    "allowances insufficient"
                );
                SafeERC20(token).safeTransferFrom(
                    funder, 
                    address(vault), 
                    toDeposit[token]
                );
            }
        }

        // Transfer tokens from the Vault to the funder.
        uint256[] amounts;
        for (uint i = 0; i < _old.tokens.length; i++) {
            amounts.push(toWithdraw[_old.tokens[i]]);
        }
        vault.batchWithdrawTo(_old.tokens, amounts, funder);
    }

    /// Calculate what quantities of tokens you would need to add to go from _old to _new.
    function _calculateMissingQuantities(Basket _old, Basket _new) internal pure
            returns (mapping(address => uint256)) {
        address token;
        uint256 diff;
        mapping(address => uint256) memory quantities; // quantities that must be added to _old to get _new

        for (uint i = 0; i < _new.tokens.length; i++) {
            token = _new.tokens[i];
            if (_new.weights[token] > _old.weights[token]) {
                diff = _new.weights[token].sub(_old.weights[token]);
                quantities[token] = rsv.totalSupply.mul(diff).div(BPS_MULTIPLIER);
            }
        }

        return quantities;
    }


    function _assertFullyCollateralized() internal pure {
        address token;
        uint256 expected;
        for (uint i = 0; i < currentBasket.tokens.length; i++) {
            token = currentBasket.tokens[i];
            expected = rsv.totalSupply.mul(currentBasket.weights[token]).div(BPS_MULTIPLIER);
            assert(SafeERC20(token).balanceOf(address(vault)) >= expected);
        }
    }
}
