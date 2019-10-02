pragma solidity 0.5.7;

import "./rsv/IRSV.sol";
import "./rsv/Reserve.sol";
import "./Manager.sol";
import "./Vault.sol";
import "./Basket.sol";
import "./Proposal.sol";

contract Echidna is Manager {

    constructor() public {
        trustedVault = IVault(address(new Vault()));
        trustedRSV = IRSV(address(new Reserve()));
        trustedProposalFactory = IProposalFactory(address(new ProposalFactory()));
        operator = 0x00a329C0648769a73afAC7F9381e08fb43DBEA70;
        seigniorage = uint256(0.0);
        emergency = true; // it's not an emergency, but we want everything to start paused.

        // Start with the empty basket.
        trustedBasket = new Basket(Basket(0), new address[](0), new uint256[](0));
    }

    function echidna_test() public view returns(bool) {
        return true;
    }

    function echidna_collateralized() public view returns(bool) {
        return isFullyCollateralized();
    }

}
