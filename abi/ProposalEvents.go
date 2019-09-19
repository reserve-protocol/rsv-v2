// This file is auto-generated. Do not edit.

package abi

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *ProposalFilterer) ParseLog(log *types.Log) (fmt.Stringer, error) {
	var event fmt.Stringer
	var eventName string
	switch log.Topics[0].Hex() {
	case "0xb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af": // NewOwnerNominated
		event = new(ProposalNewOwnerNominated)
		eventName = "NewOwnerNominated"
	case "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0": // OwnershipTransferred
		event = new(ProposalOwnershipTransferred)
		eventName = "OwnershipTransferred"
	case "0x488e676b8b729cd92586573f5b7b42787c118396c4a19f570e9c9e422e4dbf18": // ProposalAccepted
		event = new(ProposalProposalAccepted)
		eventName = "ProposalAccepted"
	case "0x9f4919593d9037fda1c872a81da93898e49ce460fadf3cbf6e8ae5b64e80e3cc": // ProposalClosed
		event = new(ProposalProposalClosed)
		eventName = "ProposalClosed"
	case "0xf53540d9e8bf8fe3396baa36e7fd34999c1a6b57bb70364c3d0c4dcbbe0baf55": // ProposalCreated
		event = new(ProposalProposalCreated)
		eventName = "ProposalCreated"
	case "0x867eabc55a8ac1bf0c7f26d0d4902538fd1165506b0fe99946359e9bd4d07fb6": // ProposalFinished
		event = new(ProposalProposalFinished)
		eventName = "ProposalFinished"
	default:
		return nil, fmt.Errorf("no such event hash for Proposal: %v", log.Topics[0])
	}

	err := c.contract.UnpackLog(event, eventName, *log)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (e ProposalNewOwnerNominated) String() string {
	return fmt.Sprintf("Proposal.NewOwnerNominated(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}

func (e ProposalOwnershipTransferred) String() string {
	return fmt.Sprintf("Proposal.OwnershipTransferred(%v, %v)", e.PreviousOwner.Hex(), e.NewOwner.Hex())
}

func (e ProposalProposalAccepted) String() string {
	return fmt.Sprintf("Proposal.ProposalAccepted(%v, %v)", e.Id, e.Proposer.Hex())
}

func (e ProposalProposalClosed) String() string {
	return fmt.Sprintf("Proposal.ProposalClosed(%v, %v)", e.Id, e.Proposer.Hex())
}

func (e ProposalProposalCreated) String() string {
	return fmt.Sprintf("Proposal.ProposalCreated(%v, %v, %v, %v, %v)", e.Id, e.Proposer.Hex(), e.Tokens, e.QuantitiesIn, e.QuantitiesOut)
}

func (e ProposalProposalFinished) String() string {
	return fmt.Sprintf("Proposal.ProposalFinished(%v, %v)", e.Id, e.Proposer.Hex())
}
