from brownie import *
import time

zero_address = "0x0000000000000000000000000000000000000000"
basket_weights = [333334e18, 333333e30, 333333e30]
min_owner_wei = 300000000000000000  # 0.3 ETH
min_daily_wei = 30000000000000000  # 0.03 ETH


def main():
    owner = accounts[0]
    daily = accounts[1]
    owner_signer = {"from": owner}
    daily_signer = {"from": daily}

    print(
        f"Owner {owner.address} has {owner.balance() / 1e18} but needs {min_owner_wei / 1e18} ETH"
    )
    print(
        f"Daily {daily.address} has {daily.balance() / 1e18} but needs {min_daily_wei / 1e18} ETH"
    )
    while owner.balance() < min_owner_wei or daily.balance() < min_daily_wei:
        print("Waiting on ETH...")
        time.sleep(5)

    usdc = BasicERC20.deploy(owner_signer)
    tusd = BasicERC20.deploy(owner_signer)
    pax = BasicERC20.deploy(owner_signer)
    basket = Basket.deploy(
        zero_address,
        [usdc.address, tusd.address, pax.address],
        basket_weights,
        owner_signer,
    )
    vault = Vault.deploy(owner_signer)
    rsv = Reserve.deploy(owner_signer)
    relayer = Relayer.deploy(rsv.address, owner_signer)
    rsv.changeRelayer(relayer.address)

    # TODO: make this work by resolving namespace conflict over `balance`
    # eternal_storage = ReserveEternalStorage.at(rsv.getEternalStorageAddress())
    # eternal_storage.acceptOwnership(owner_signer)

    proposal_factory = ProposalFactory.deploy(owner_signer)

    manager = Manager.deploy(
        vault.address,
        rsv.address,
        proposal_factory.address,
        basket.address,
        daily.address,
        0,
        owner_signer,
    )

    vault.changeManager(manager.address, owner_signer)
    rsv.changeMinter(manager.address, owner_signer)
    rsv.changePauser(daily.address, owner_signer)
    rsv.changeFeeRecipient(daily.address, owner_signer)
    rsv.unpause(daily_signer)
    manager.setEmergency(False, daily_signer)
    time.sleep(0.1)
