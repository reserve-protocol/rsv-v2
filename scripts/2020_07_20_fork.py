from brownie import *
import time

zero_address = "0x0000000000000000000000000000000000000000"
basket_weights = [333334e18, 333333e30, 333333e30]
min_owner_wei = 300000000000000000  # 0.3 ETH
min_daily_wei = 30000000000000000  # 0.03 ETH

owner = accounts[0]
daily = accounts[1]
temp = accounts[2]

owner_signer = {"from": owner}
daily_signer = {"from": daily}
temp_signer = {"from": temp}


def initial_rsv_deployment(ctx):
    # Make USDC, TUSD, and PAX exist
    ctx.usdc = BasicERC20.deploy(owner_signer)
    ctx.tusd = BasicERC20.deploy(owner_signer)
    ctx.pax = BasicERC20.deploy(owner_signer)

    # Now begin our contracts
    ctx.vault = Vault.deploy(owner_signer)
    ctx.rsv = DeployedReserve.deploy(owner_signer)

    # This contract has a namespace conflict when instantiated with `brownie`.
    # I've talked to them about it. They have a solution that will go in `brownie-v2`.
    # In the meantime, we can't use `brownie` to interface with this contract.
    # eternal_storage = ReserveEternalStorage.at(rsv.getEternalStorageAddress())
    # eternal_storage.acceptOwnership(owner_signer)

    ctx.proposal_factory = ProposalFactory.deploy(owner_signer)
    ctx.basket = Basket.deploy(
        zero_address,
        [ctx.usdc.address, ctx.tusd.address, ctx.pax.address],
        basket_weights,
        owner_signer,
    )
    ctx.manager = Manager.deploy(
        ctx.vault.address,
        ctx.rsv.address,
        ctx.proposal_factory.address,
        ctx.basket.address,
        daily.address,
        0,
        owner_signer,
    )

    ctx.vault.changeManager(ctx.manager.address, owner_signer)
    ctx.rsv.changeMinter(ctx.manager.address, owner_signer)
    ctx.rsv.changePauser(daily.address, owner_signer)
    ctx.rsv.changeFeeRecipient(daily.address, owner_signer)
    ctx.rsv.unpause(daily_signer)
    ctx.manager.setEmergency(False, daily_signer)

    ctx.usdc.approve(ctx.manager.address, 100000200, owner_signer)  # $100.0002 USDC
    ctx.tusd.approve(
        ctx.manager.address, 99999900000000000000, owner_signer
    )  # $99.9999 TUSD
    ctx.pax.approve(
        ctx.manager.address, 99999900000000000000, owner_signer
    )  # $99.9999 PAX

    ctx.manager.issue(3e20, owner_signer)  # $300 RSV
    assert ctx.rsv.totalSupply() == 3e20


def first_half_of_fork(ctx):
    # Deploy New RSV
    ctx.rsv_2 = Reserve.deploy(temp_signer)
    ctx.rsv_2.nominateNewOwner(owner.address, temp_signer)

    # Create the new manager and update others to point to it
    ctx.manager_2 = Manager.deploy(
        ctx.vault.address,
        ctx.rsv_2.address,
        ctx.proposal_factory.address,
        ctx.basket.address,
        daily.address,
        0,
        temp_signer,
    )
    ctx.manager_2.nominateNewOwner(owner.address, temp_signer)

    # Deploy relayer
    ctx.relayer = Relayer.deploy(ctx.rsv_2.address, temp_signer)
    ctx.relayer.nominateNewOwner(owner.address, temp_signer)

    # Set the new RSV to point to all the right things
    ctx.rsv_2.changeRelayer(ctx.relayer.address, temp_signer)
    ctx.rsv_2.changeMinter(ctx.manager_2.address, temp_signer)
    ctx.rsv_2.changePauser(daily.address, temp_signer)
    ctx.rsv_2.changeFeeRecipient(daily.address, temp_signer)


def second_half_of_fork(ctx):

    # Transfer all owners to the permanent owner key
    ctx.rsv_2.acceptOwnership(owner_signer)
    ctx.relayer.acceptOwnership(owner_signer)
    ctx.manager_2.acceptOwnership(owner_signer)

    # Set the vault to point to the new manager
    ctx.manager.setEmergency(True, daily_signer)
    ctx.vault.changeManager(ctx.manager_2.address, owner_signer)

    # Finalize the fork
    ctx.rsv.nominateNewOwner(ctx.rsv_2.address, owner_signer)
    ctx.rsv_2.acceptUpgrade(ctx.rsv.address, owner_signer)
    ctx.manager_2.setEmergency(False, daily_signer)


def final_confirmations(ctx):

    # Check redemptions
    assert ctx.rsv_2.totalSupply() == 3e20
    ctx.rsv_2.approve(ctx.manager_2.address, 2e20, owner_signer)
    ctx.manager_2.redeem(2e20, owner_signer)
    assert ctx.rsv_2.totalSupply() == 1e20
    print("successfully redeemed!")

    # Check issuance
    ctx.usdc.approve(ctx.manager_2.address, 100000200, owner_signer)  # $100.0002 USDC
    ctx.tusd.approve(
        ctx.manager_2.address, 99999900000000000000, owner_signer
    )  # $99.9999 TUSD
    ctx.pax.approve(
        ctx.manager_2.address, 99999900000000000000, owner_signer
    )  # $99.9999 PAX
    ctx.manager_2.issue(3e20, owner_signer)
    assert ctx.rsv_2.totalSupply() == 4e20
    print("and issued!")

    # Make sure the old stuff is turned off
    assert ctx.rsv.paused()
    assert ctx.manager.emergency()

    # Make sure the vault points to the new manager
    assert ctx.vault.manager() == ctx.manager_2.address

    # Check the entire new Manager state
    assert ctx.manager_2.operator() == daily.address
    assert ctx.manager_2.trustedBasket() == ctx.basket.address
    assert ctx.manager_2.trustedVault() == ctx.vault.address
    assert ctx.manager_2.trustedRSV() == ctx.rsv_2.address
    assert ctx.manager_2.trustedProposalFactory() == ctx.proposal_factory.address
    assert ctx.manager_2.proposalsLength() == 0
    assert not ctx.manager_2.issuancePaused()
    assert not ctx.manager_2.emergency()
    assert ctx.manager_2.seigniorage() == 0

    # Check the entire new RSV state
    assert ctx.rsv_2.getEternalStorageAddress() == ctx.rsv.getEternalStorageAddress()
    assert ctx.rsv_2.trustedTxFee() == zero_address
    assert ctx.rsv_2.trustedRelayer() == ctx.relayer.address
    assert ctx.rsv_2.maxSupply() == ctx.rsv.maxSupply()
    assert not ctx.rsv_2.paused()
    assert ctx.rsv_2.minter() == ctx.manager_2.address
    assert ctx.rsv_2.pauser() == daily.address
    assert ctx.rsv_2.feeRecipient() == daily.address

    # Make sure the bridge is burnt behind us
    assert ctx.rsv.minter() == zero_address
    assert ctx.rsv.pauser() == zero_address
    assert ctx.rsv.owner() == zero_address
    assert ctx.rsv.paused()


class Context:
    pass


def main():
    print(
        f"Owner {owner.address} has {owner.balance() / 1e18} but needs {min_owner_wei / 1e18} ETH"
    )
    print(
        f"Daily {daily.address} has {daily.balance() / 1e18} but needs {min_daily_wei / 1e18} ETH"
    )
    while owner.balance() < min_owner_wei or daily.balance() < min_daily_wei:
        print("Waiting on ETH...")
        time.sleep(5)

    ctx = Context()

    initial_rsv_deployment(ctx)
    first_half_of_fork(ctx)
    second_half_of_fork(ctx)
    final_confirmations(ctx)
    print("and we're done!")
    time.sleep(0.5)
