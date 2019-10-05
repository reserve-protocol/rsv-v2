pragma solidity ^0.5.7;

import "../Vault.sol";

contract VaultV2 is Vault {

    function completeHandoff(address previousVaultAddress) external onlyOwner {
        Vault previousVault = Vault(previousVaultAddress);

        previousVault.acceptOwnership();

        previousVault.changeManager(address(this));

        previousVault.renounceOwnership("I hereby renounce ownership of this contract forever.");
    }
}
