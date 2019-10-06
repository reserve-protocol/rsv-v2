pragma solidity ^0.5.7;

import "../Basket.sol";
import "../Manager.sol";
import "../Vault.sol";
import "../rsv/IRSV.sol";

contract VaultV2 is Vault {

    function completeHandoff(address previousVaultAddress, Manager manager) external onlyOwner {
        Vault previousVault = Vault(previousVaultAddress);

        previousVault.acceptOwnership();

        previousVault.changeManager(address(this));

        // Transfer tokens from old vault to new vault.
        Basket trustedBasket = manager.trustedBasket();

        for (uint256 i = 0; i < trustedBasket.size(); i++) {
            address tokenAddr = trustedBasket.tokens(i);
            IERC20 token = IERC20(tokenAddr);

            token.safeTransferFrom(
                previousVaultAddress,
                address(this),
                token.balanceOf(address(previousVaultAddress))
            );
        }

        // Point manager at the new vault.
        manager.setVault(address(this));

    }
}
