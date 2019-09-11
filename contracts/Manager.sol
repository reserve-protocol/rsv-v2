pragma solidity ^0.4.24;

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
    using SafeERC20 for IERC20;

    Vault public vault;
    IRSV public rsv;

    bool public paused;

    // This needs to stay small, but that should be fine.
    mapping(address => bool) public whitelist;

    // Supported tokens and their weights in the basket.
    address[] public collateralTokens;
    mapping(address => uint256) public weights;
    uint256 public sumWeights;


    // This allows the vault to net profit
    uint256 public seigniorage;    // In BPS, e.g seigniorage should be set to 10 to achieve a 0.1% spread
    
    // seigniorage
    event SegniorageChanged(uint256 oldVal, uint256 newVal);

    // Pause events
    event Paused(address indexed account);
    event Unpaused(address indexed account);

    // Vault management events
    event TokenAdded(address indexed tokenAddr, uint256 indexed weight);
    event TokenRemoved(address indexed tokenAddr);

    // Whitelist events
    event Whitelist(address indexed user);
    event DeWhitelisted(address indexed user);

    // Issuance/Redemption events
    event Issuance(address indexed user, uint256 indexed amount, uint256[] breakdown);
    event Redemption(address indexed user, uint256 indexed amount, uint256[] breakdown);


    // Begins paused
    constructor(address vaultAddr, address rsvAddress, uint256 _seigniorage) public {
        vault = Vault(vaultAddr);
        rsv = IRSV(rsvAddress);
        whitelist[msg.sender] = true;
        seigniorage = _seigniorage;
        paused = true;
    }


    // === seigniorage ===

    function setSegniorage(uint256 _seigniorage) external onlyOwner {
        emit SegniorageChanged(seigniorage, _seigniorage);
        seigniorage = _seigniorage;
    }


    // === Pausing ===

    /// Modifies a function to run only when the contract is not paused.
    modifier notPaused() {
        require(!paused, "contract is paused");
        _;
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


    // === Whitelisting ===

    modifier onlyWhitelist() {
        require(whitelist[msg.sender], "unauthorized: not on whitelist");
        _;
    }

    // Add user to whitelist.
    function whitelist(address user) external onlyOwner {
        whitelist[user] = true;
        emit Whitelist(user);
    }

    // Remove user from whitelist.
    function deWhitelist(address user) external onlyOwner {
        whitelist[user] = false;
        emit DeWhitelisted(user);
    }


    // === Vault Management ===

    // Add collateral token to the vault.
    function addCollateralToken(address token, uint256 weight) external onlyOwner {
        uint sum = 0;
        for (uint i = 0; i < collateralTokens.length; i++) {
            require(collateralTokens[i] != token, "collateral token already in vault");
            sum += weights[collateralTokens[i]];
        }

        collateralTokens.push(token);
        weights[token] = weight;
        sum += weight;
        sumWeights += weight;

        emit TokenAdded(token, weight);
    }

    // Remove collateral token from the vault. 
    function removeCollateralToken(address token) external onlyOwner {
        for (uint i = 0; i < collateralTokens.length; i++) {
            if (token == collateralTokens[i]) {
                collateralTokens[i] = collateralTokens[collateralTokens.length - 1];
                delete collateralTokens[collateralTokens.length - 1];
                collateralTokens.length--;
                break;
            }

            require(false, "collateral token missing from vault");
        }

        sumWeights -= weights[token];
        delete weights[token];

        emit TokenRemoved(token);
    }


    // === Issuance ===

    // Issue mints RSV for tokens in a way that moves us closer toward
    // the target collateral backing ratio given by weights.
    function issue(uint256 amount) external notPaused onlyWhitelist {
        // Do checks
        uint256[] memory toBuy = collateralAmountsToBuy(amount);
        uint256 sum = 0;
        for (uint i = 0; i < collateralTokens.length; i++) {
            require(IERC20(collateralTokens[i]).allowance(msg.sender, address(this)) >= toBuy[i], "please set allowance");
            require(IERC20(collateralTokens[i]).balanceOf(msg.sender) >= toBuy[i], "insufficient balance");
            sum += toBuy[i];
        }

        require(sum > amount, "there should be seigniorage");

        // Intake collateral
        for (uint j = 0; j < collateralTokens.length; j++) {
            IERC20(collateralTokens[j]).safeTransferFrom(msg.sender, address(vault), toBuy[j]);
        }

        // Hand out RSV
        rsv.mint(msg.sender, amount);

        emit Issuance(msg.sender, amount, toBuy);
    }

    // Solidity is bad at returning dynamic arrays, so this design may not work. 
    function collateralAmountsToBuy(uint256 amount) public view returns(uint256[]) {
        // Calculate deficit collateral at new increased supply
        uint256 deficitCollateral = 0;
        uint256[] memory toBuy = new uint256[](collateralTokens.length);
        for (uint i = 0; i < collateralTokens.length; i++) {
            uint target = weights[collateralTokens[i]] * (rsv.totalSupply() + amount) / sumWeights; 
            if (target - IERC20(collateralTokens[i]).balanceOf(address(vault)) > 0) {
                toBuy[i] = target - IERC20(collateralTokens[i]).balanceOf(address(vault));
                deficitCollateral += toBuy[i];
            }
        }

        // Normalize to amount of RSV being sold
        for (uint j = 0; j < collateralTokens.length; j++) {
            toBuy[j] = toBuy[j] * amount / deficitCollateral;
            toBuy[j] = toBuy[j] * (10000 + seigniorage) / 10000; // seigniorage
        }

        return toBuy;
    }


    // === Redemption ===

    function redeem(uint256 amount) external notPaused onlyWhitelist {
        // Do checks
        require(rsv.allowance(msg.sender, address(this)) >= amount, "please set allowance");
        require(rsv.balanceOf(msg.sender) >= amount, "insufficient rsv to redeem");

        uint256[] memory toSell = collateralAmountsToSell(amount);
        uint256 sum = 0;
        for (uint i = 0; i < collateralTokens.length; i++) {
            sum += toSell[i];
        }

        require(sum <= amount, "we shouldn't sell more than the redemption amount");

        // Intake RSV
        rsv.burnFrom(msg.sender, amount);

        // Hand out collateral
        vault.batchWithdrawTo(collateralTokens, toSell, msg.sender);

        emit Redemption(msg.sender, amount, toSell);
    }

    // Solidity is bad at returning dynamic arrays, so this design may not work. 
    function collateralAmountsToSell(uint256 amount) public view returns(uint256[]) {
        // Calculate excess collateral at new reduced supply
        uint256 excessCollateral = 0;
        uint256[] memory toSell = new uint256[](collateralTokens.length);
        for (uint i = 0; i < collateralTokens.length; i++) {
            uint target = weights[collateralTokens[i]] * (rsv.totalSupply() - amount) / sumWeights; 
            if (IERC20(collateralTokens[i]).balanceOf(address(vault)) - target > 0) {
                toSell[i] = target - IERC20(collateralTokens[i]).balanceOf(address(vault));
                excessCollateral += toSell[i];
            }
        }

        // Normalize to amount of RSV being bought
        for (uint j = 0; j < collateralTokens.length; j++) {
            toSell[j] = toSell[j] * amount / excessCollateral;
        }

        return toSell;
    }


    // Rebalance function signatures
    // function proposeRebalance(address[] tokenAddresses, uint256[] tokenWeights) public notPaused returns(uint256)
    // function acceptRebalance(uint256 nonce) external onlyOwner
    // Oh frick this can be frontrun can't it
}

