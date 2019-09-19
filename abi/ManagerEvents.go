// This file is auto-generated. Do not edit.

package abi

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *ManagerFilterer) ParseLog(log *types.Log) (fmt.Stringer, error) {
	var event fmt.Stringer
	var eventName string
	switch log.Topics[0].Hex() {
	case "0xea592933e1a0c057f8b0807435fa3d61d8c2f5afb5f860d39708e7a36268c0d3": // DeWhitelisted
		event = new(ManagerDeWhitelisted)
		eventName = "DeWhitelisted"
	case "0x9cb9c14f7bc76e3a89b796b091850526236115352a198b1e472f00e91376bbcb": // Issuance
		event = new(ManagerIssuance)
		eventName = "Issuance"
	case "0xb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af": // NewOwnerNominated
		event = new(ManagerNewOwnerNominated)
		eventName = "NewOwnerNominated"
	case "0x4721129e0e676ed6a92909bb24e853ccdd63ad72280cc2e974e38e480e0e6e54": // OperatorChanged
		event = new(ManagerOperatorChanged)
		eventName = "OperatorChanged"
	case "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0": // OwnershipTransferred
		event = new(ManagerOwnershipTransferred)
		eventName = "OwnershipTransferred"
	case "0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258": // Paused
		event = new(ManagerPaused)
		eventName = "Paused"
	case "0xe89b7d17759242abd80309cee2479e2cd462f8a25d06a0f90fba1de050de6623": // ProposalsCleared
		event = new(ManagerProposalsCleared)
		eventName = "ProposalsCleared"
	case "0x439ce0c367c25c0aaf98743becfd020f9403ba4575882d608ed696f1de63ba63": // RSVChanged
		event = new(ManagerRSVChanged)
		eventName = "RSVChanged"
	case "0xe6c82503aaaa3db78b70f183901ae8668918f895b3982b2c851cf2ffe0c6c639": // Redemption
		event = new(ManagerRedemption)
		eventName = "Redemption"
	case "0x739fc76c925698caa5b5b65517fbbf8148d051b676bfac3769c10dc9f146a751": // SeigniorageChanged
		event = new(ManagerSeigniorageChanged)
		eventName = "SeigniorageChanged"
	case "0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa": // Unpaused
		event = new(ManagerUnpaused)
		eventName = "Unpaused"
	case "0xa49691f0dd6477ccef49962612a236d252e3a31c3be8b61fa6abeff3e74a7572": // VaultChanged
		event = new(ManagerVaultChanged)
		eventName = "VaultChanged"
	case "0xaab7954e9d246b167ef88aeddad35209ca2489d95a8aeb59e288d9b19fae5a54": // Whitelisted
		event = new(ManagerWhitelisted)
		eventName = "Whitelisted"
	default:
		return nil, fmt.Errorf("no such event hash for Manager: %v", log.Topics[0])
	}

	err := c.contract.UnpackLog(event, eventName, *log)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (e ManagerDeWhitelisted) String() string {
	return fmt.Sprintf("Manager.DeWhitelisted(%v)", e.User.Hex())
}

func (e ManagerIssuance) String() string {
	return fmt.Sprintf("Manager.Issuance(%v, %v)", e.User.Hex(), e.Amount)
}

func (e ManagerNewOwnerNominated) String() string {
	return fmt.Sprintf("Manager.NewOwnerNominated(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}

func (e ManagerOperatorChanged) String() string {
	return fmt.Sprintf("Manager.OperatorChanged(%v)", e.Account.Hex())
}

func (e ManagerOwnershipTransferred) String() string {
	return fmt.Sprintf("Manager.OwnershipTransferred(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}

func (e ManagerPaused) String() string {
	return fmt.Sprintf("Manager.Paused(%v)", e.Account.Hex())
}

func (e ManagerProposalsCleared) String() string {
	return fmt.Sprintf("Manager.ProposalsCleared()")
}

func (e ManagerRSVChanged) String() string {
	return fmt.Sprintf("Manager.RSVChanged(%v)", e.Account.Hex())
}

func (e ManagerRedemption) String() string {
	return fmt.Sprintf("Manager.Redemption(%v, %v)", e.User.Hex(), e.Amount)
}

func (e ManagerSeigniorageChanged) String() string {
	return fmt.Sprintf("Manager.SeigniorageChanged(%v, %v)", e.OldVal, e.NewVal)
}

func (e ManagerUnpaused) String() string {
	return fmt.Sprintf("Manager.Unpaused(%v)", e.Account.Hex())
}

func (e ManagerVaultChanged) String() string {
	return fmt.Sprintf("Manager.VaultChanged(%v)", e.Account.Hex())
}

func (e ManagerWhitelisted) String() string {
	return fmt.Sprintf("Manager.Whitelisted(%v)", e.User.Hex())
}
