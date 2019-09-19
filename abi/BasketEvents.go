// This file is auto-generated. Do not edit.

package abi

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *BasketFilterer) ParseLog(log *types.Log) (fmt.Stringer, error) {
	var event fmt.Stringer
	var eventName string
	switch log.Topics[0].Hex() {
	default:
		return nil, fmt.Errorf("no such event hash for Basket: %v", log.Topics[0])
	}

	err := c.contract.UnpackLog(event, eventName, *log)
	if err != nil {
		return nil, err
	}
	return event, nil
}
