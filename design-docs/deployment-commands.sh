#!/bin/echo Don't run this by hand. Run this slowly. Carefully. With a friend, and the checklist.
exit 1


export SOLC_VERSION=0.5.7

# real-run parameters:
export POKE_FROM=hardware
export POKE_NODE=https://mainnet.infura.io/v3/d884cdc2e05b4f0897f6dffd0bdc1821
export POKE_DERIVATION_PATH="m/44'/60'/0'/0/3"   # This is eth0/3. We'll use this for both OWNER and DAILY.
export POKE_GASPRICE=8
#   Or whatever gas price. Keep an eye on ethgasstation, and adjust
#   as needed, either in the env variable, or using the -g flag to poke

# With every poke command:
#    - [ ] record the txn hash and returned output in the keybase log
#    - [ ] post the address in slack
#    - [ ] record the address as a variable for further use in that session
#    - [ ] check the result on etherscan
#    - [ ] check the result from another machine, using the alternate Ethereum access node

# Be fresh!
git fetch
git checkout production
git pull
make json test

# Also check that you have the latest `poke`! You'll have to `go install` it.

cd evm

# devchain test only: export POKE_FROM=@0

##################
# Using key DAILY:
##################

DAILY=$(poke Vault.json address) # (or whatever contract. Vault.json is tiny and good)
echo $DAILY                      # Set the "DAILY" address from here. We'll be using this later!


##################
# Using key OWNER:
##################

# 1. Deploy Vault
poke Vault.json deploy
# The actual address here is just for example. copy-paste the POKE_ADDRESS result off the commandline.
# (The same is true for all the other Something=... lines that follow a `poke deploy` line.)
Vault=0xAeDCFcdD80573c2a312d15d6Bb9d921a01E4FB0f



# 2. Deploy Reserve
poke Reserve.json deploy
Reserve=0x1C5857e110CD8411054660F60B5De6a6958CfAE2



# 3. Take ownership of ReserveEternalStorage
poke --address=${Reserve} Reserve.json getEternalStorageAddress
ReserveES=0xb8bF093045D5f2E4AAc9062bD11A8f599f6565B5

poke --address=${ReserveES} ReserveEternalStorage.json acceptOwnership




# 4. Deploy ProposalFactory
poke ProposalFactory.json deploy
ProposalFactory=0x06FA956887524836C8ff54F2f57695762fF0A037



# 5. Deploy Basket

# addresses from etherscan. Triple-check these!
PAX=0x8e870d67f660d95d5be530380d0ec0bd388289e1
USDC=0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
TUSD=0x0000000000085d4780B73119b644AE5ecd22b376

# Zero address
ZERO=0x0000000000000000000000000000000000000000

poke Basket.json deploy $ZERO "[$PAX, $TUSD, $USDC]" "[.333333e36, .333333e36, 333334e18]"
Basket=0x7CC227729270426da6e9E3f51838CF5C7dbc1588

# Really double-check these values! Oh man!
poke --address=$Basket Basket.json weights $PAX
poke --address=$Basket Basket.json weights $USDC
poke --address=$Basket Basket.json weights $TUSD



# 6. Deploy Manager
echo $Vault
echo $Reserve
echo $ProposalFactory
echo $Basket
echo $DAILY
poke Manager.json deploy $Vault $Reserve $ProposalFactory $Basket $DAILY 0
Manager=0x5BA9d812f5533F7Cf2854963f7A9d212f8f28673

poke Manager.json trustedBasket --address=${Manager}
poke Manager.json trustedRSV --address=${Manager}
poke Manager.json trustedProposalFactory --address=${Manager}
poke Manager.json trustedVault --address=${Manager}

# 7. Update auth roles and contract back-links
poke --address=$Vault Vault.json changeManager $Manager
poke --address=$Reserve Reserve.json changeMinter $Manager
poke --address=$Reserve Reserve.json changePauser $DAILY
poke --address=$Reserve Reserve.json changeFeeRecipient $DAILY

poke --address=$Reserve Reserve.json changeMaxSupply 1e23


############################
# STOP! Take out the OWNER key!
# 7. Check the states of things using poke! (see deployment checklist)
############################

# 8. Unpause Reserve and Manager contracts

######################
# Using the DAILY key
######################

# Double-check that you're on the address you expect!
# These should be the same:
echo $DAILY
poke Basket.json address

poke --address=$Reserve Reserve.json unpause
poke --address=$Manager Manager.json setEmergency false

# Further tasks: See the post-deployment checklist!
