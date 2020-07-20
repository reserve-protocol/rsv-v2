pragma solidity 0.5.7;

import "../rsv/Reserve.sol";
import "../rsv/ReserveEternalStorage.sol";

/**
 * @dev A version of the Reserve Token for testing upgrades.
 */
contract ReserveV2 is Reserve {

    string public constant version = "2";

    constructor() Reserve() public {
        trustedData = ReserveEternalStorage(address(0));
    }


    /// Accept upgrade from previous RSV instance. Can only be called once. 
    function acceptUpgrade(address previousImplementation) external onlyOwner {
        require(address(trustedData) == address(0), "can only be run once");
        Reserve previous = Reserve(previousImplementation);
        trustedData = ReserveEternalStorage(previous.getEternalStorageAddress());

        // Copy values from old contract
        maxSupply = previous.maxSupply();
        totalSupply = previous.totalSupply();
        
        // Unpause.
        paused = false;
        emit Unpaused(pauser);

        previous.acceptOwnership();

        // Take control of Eternal Storage.
        previous.changePauser(address(this));
        previous.pause();
        previous.transferEternalStorage(address(this));

        // Burn the bridge behind us.
        previous.changeMinter(address(0));
        previous.changePauser(address(0));
        previous.renounceOwnership("I hereby renounce ownership of this contract forever.");
    }

}
