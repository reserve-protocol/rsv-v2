// This file is auto-generated. Do not edit.

package abi

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *VaultFilterer) ParseLog(log *types.Log) (fmt.Stringer, error) {
	var event fmt.Stringer
	var eventName string
	switch log.Topics[0].Hex() {
	case "0x68b4558431231278e57e1a969c13c61605ca644c57a73d231bfcc7e2af80d2e0": // BatchWithdrawal
		event = new(VaultBatchWithdrawal)
		eventName = "BatchWithdrawal"
	case "0x9cb45c728de594dab506a1f1a8554e24c8eeaf983618d5ec5dd7bc6f3c49feee": // ManagerTransferred
		event = new(VaultManagerTransferred)
		eventName = "ManagerTransferred"
	case "0xb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af": // NewOwnerNominated
		event = new(VaultNewOwnerNominated)
		eventName = "NewOwnerNominated"
	case "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0": // OwnershipTransferred
		event = new(VaultOwnershipTransferred)
		eventName = "OwnershipTransferred"
	default:
		return nil, fmt.Errorf("no such event hash for Vault: %v", log.Topics[0])
	}

	err := c.contract.UnpackLog(event, eventName, *log)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (e VaultBatchWithdrawal) String() string {
	return fmt.Sprintf("Vault.BatchWithdrawal(%v, %v)", e.Tokens, e.Quantities)
}

func (e VaultManagerTransferred) String() string {
	return fmt.Sprintf("Vault.ManagerTransferred(%v, %v)", e.PreviousManager.Hex(), e.NewManager.Hex())
}

func (e VaultNewOwnerNominated) String() string {
	return fmt.Sprintf("Vault.NewOwnerNominated(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}

func (e VaultOwnershipTransferred) String() string {
	return fmt.Sprintf("Vault.OwnershipTransferred(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}
