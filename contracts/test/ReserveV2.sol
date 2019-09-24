pragma solidity ^0.5.8;

import "../rsv/Reserve.sol";
import "../rsv/ReserveEternalStorage.sol";

/**
 * @dev A version of the Reserve Token for testing upgrades.
 */
contract ReserveV2 is Reserve {
    constructor() public {
        paused = true;
    }

    function completeHandoff(address previousImplementation) external onlyOwner {
        Reserve previous = Reserve(previousImplementation);
        data = ReserveEternalStorage(previous.getEternalStorageAddress());
        previous.acceptOwnership();

        // Take control of Eternal Storage.
        previous.transferEternalStorage(address(this));
        previous.changePauser(address(this));

        // Old contract off, new contract on.
        previous.pause();
        paused = false;
        emit Unpaused(pauser);

        // Burn the bridge behind us.
        previous.changeMinter(address(0));
        previous.changePauser(address(0));
        previous.changeFreezer(address(0));
        previous.renounceOwnership();
    }
}
