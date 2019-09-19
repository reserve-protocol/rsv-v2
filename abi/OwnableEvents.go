// This file is auto-generated. Do not edit.

package abi

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *OwnableFilterer) ParseLog(log *types.Log) (fmt.Stringer, error) {
	var event fmt.Stringer
	var eventName string
	switch log.Topics[0].Hex() {
	case "0xb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af": // NewOwnerNominated
		event = new(OwnableNewOwnerNominated)
		eventName = "NewOwnerNominated"
	case "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0": // OwnershipTransferred
		event = new(OwnableOwnershipTransferred)
		eventName = "OwnershipTransferred"
	default:
		return nil, fmt.Errorf("no such event hash for Ownable: %v", log.Topics[0])
	}

	err := c.contract.UnpackLog(event, eventName, *log)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (e OwnableNewOwnerNominated) String() string {
	return fmt.Sprintf("Ownable.NewOwnerNominated(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}

func (e OwnableOwnershipTransferred) String() string {
	return fmt.Sprintf("Ownable.OwnershipTransferred(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}
