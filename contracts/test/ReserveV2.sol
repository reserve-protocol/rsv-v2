pragma solidity 0.5.7;

import "../rsv/Reserve.sol";
import "../rsv/ReserveEternalStorage.sol";

/**
 * @dev A version of the Reserve Token for testing upgrades.
 */
contract ReserveV2 is Reserve {

    function completeHandoff(address previousImplementation) external onlyOwner {
        Reserve previous = Reserve(previousImplementation);
        trustedData = ReserveEternalStorage(previous.getEternalStorageAddress());
        previous.acceptOwnership();

        //Take control of Eternal Storage.
        previous.changePauser(address(this));
        previous.pause();
        previous.transferEternalStorage(address(this));

        // Unpause.
        paused = false;
        emit Unpaused(pauser);

        // Burn the bridge behind us.
        previous.changeMinter(address(0));
        previous.changePauser(address(0));
        previous.renounceOwnership();
    }
}
