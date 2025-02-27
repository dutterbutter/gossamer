// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package babe

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/scale"
	"github.com/ChainSafe/gossamer/lib/transaction"
)

// BuildBlock builds a block for the slot with the given parent.
// TODO: separate block builder logic into separate module. The only reason this is exported is so other packages
// can build blocks for testing, but it would be preferred to have the builder functionality separated.
func (b *Service) BuildBlock(parent *types.Header, slot Slot) (*types.Block, error) {
	return b.buildBlock(parent, slot)
}

// construct a block for this slot with the given parent
func (b *Service) buildBlock(parent *types.Header, slot Slot) (*types.Block, error) {
	logger.Trace("build block", "parent", parent, "slot", slot)

	// create pre-digest
	preDigest, err := b.buildBlockPreDigest(slot)
	if err != nil {
		return nil, err
	}

	logger.Trace("built pre-digest")

	// create new block header
	number := big.NewInt(0).Add(parent.Number, big.NewInt(1))
	header, err := types.NewHeader(parent.Hash(), common.Hash{}, common.Hash{}, number, types.NewEmptyDigest())
	if err != nil {
		return nil, err
	}

	// initialise block header
	err = b.rt.InitializeBlock(header)
	if err != nil {
		return nil, err
	}

	logger.Trace("initialised block")

	// add block inherents
	inherents, err := b.buildBlockInherents(slot)
	if err != nil {
		return nil, fmt.Errorf("cannot build inherents: %s", err)
	}

	logger.Trace("built block inherents", "encoded inherents", inherents)

	// add block extrinsics
	included := b.buildBlockExtrinsics(slot)

	logger.Trace("built block extrinsics")

	// finalise block
	header, err = b.rt.FinalizeBlock()
	if err != nil {
		b.addToQueue(included)
		return nil, fmt.Errorf("cannot finalise block: %s", err)
	}

	logger.Trace("finalised block")

	header.ParentHash = parent.Hash()
	header.Number.Add(parent.Number, big.NewInt(1))

	// add BABE header to digest
	header.Digest = append(header.Digest, preDigest)

	// create seal and add to digest
	seal, err := b.buildBlockSeal(header)
	if err != nil {
		return nil, err
	}

	header.Digest = append(header.Digest, seal)

	logger.Trace("built block seal")

	body, err := extrinsicsToBody(inherents, included)
	if err != nil {
		return nil, err
	}

	block := &types.Block{
		Header: header,
		Body:   body,
	}

	return block, nil
}

// buildBlockSeal creates the seal for the block header.
// the seal consists of the ConsensusEngineID and a signature of the encoded block header.
func (b *Service) buildBlockSeal(header *types.Header) (*types.SealDigest, error) {
	encHeader, err := header.Encode()
	if err != nil {
		return nil, err
	}

	hash, err := common.Blake2bHash(encHeader)
	if err != nil {
		return nil, err
	}

	sig, err := b.keypair.Sign(hash[:])
	if err != nil {
		return nil, err
	}

	return &types.SealDigest{
		ConsensusEngineID: types.BabeEngineID,
		Data:              sig,
	}, nil
}

// buildBlockPreDigest creates the pre-digest for the slot.
// the pre-digest consists of the ConsensusEngineID and the encoded BABE header for the slot.
func (b *Service) buildBlockPreDigest(slot Slot) (*types.PreRuntimeDigest, error) {
	babeHeader, err := b.buildBlockBABEPrimaryPreDigest(slot)
	if err != nil {
		return nil, err
	}

	encBABEPrimaryPreDigest := babeHeader.Encode()

	return &types.PreRuntimeDigest{
		ConsensusEngineID: types.BabeEngineID,
		Data:              encBABEPrimaryPreDigest,
	}, nil
}

// buildBlockBABEPrimaryPreDigest creates the BABE header for the slot.
// the BABE header includes the proof of authorship right for this slot.
func (b *Service) buildBlockBABEPrimaryPreDigest(slot Slot) (*types.BabePrimaryPreDigest, error) {
	if b.slotToProof[slot.number] == nil {
		return nil, ErrNotAuthorized
	}

	outAndProof := b.slotToProof[slot.number]
	return types.NewBabePrimaryPreDigest(
		b.epochData.authorityIndex,
		slot.number,
		outAndProof.output,
		outAndProof.proof,
	), nil
}

// buildBlockExtrinsics applies extrinsics to the block. it returns an array of included extrinsics.
// for each extrinsic in queue, add it to the block, until the slot ends or the block is full.
// if any extrinsic fails, it returns an empty array and an error.
func (b *Service) buildBlockExtrinsics(slot Slot) []*transaction.ValidTransaction {
	var included []*transaction.ValidTransaction

	for !hasSlotEnded(slot) {
		txn := b.transactionState.Pop()
		// Transaction queue is empty.
		if txn == nil {
			return included
		}

		// Move to next extrinsic.
		if txn.Extrinsic == nil {
			continue
		}

		extrinsic := txn.Extrinsic
		logger.Trace("build block", "applying extrinsic", extrinsic)

		ret, err := b.rt.ApplyExtrinsic(extrinsic)
		if err != nil {
			logger.Warn("failed to apply extrinsic", "error", err, "extrinsic", extrinsic)
			continue
		}

		err = determineErr(ret)
		if err != nil {
			logger.Warn("failed to apply extrinsic", "error", err, "extrinsic", extrinsic)

			// Failure of the module call dispatching doesn't invalidate the extrinsic.
			// It is included in the block.
			if _, ok := err.(*DispatchOutcomeError); !ok {
				continue
			}
		}

		logger.Debug("build block applied extrinsic", "extrinsic", extrinsic)
		included = append(included, txn)
	}

	return included
}

// buildBlockInherents applies the inherents for a block
func (b *Service) buildBlockInherents(slot Slot) ([][]byte, error) {
	// Setup inherents: add timstap0
	idata := types.NewInherentsData()
	err := idata.SetInt64Inherent(types.Timstap0, uint64(time.Now().Unix()))
	if err != nil {
		return nil, err
	}

	// add babeslot
	err = idata.SetInt64Inherent(types.Babeslot, slot.number)
	if err != nil {
		return nil, err
	}

	// add finalnum
	fin, err := b.blockState.GetFinalizedHeader(0, 0)
	if err != nil {
		return nil, err
	}

	err = idata.SetBigIntInherent(types.Finalnum, fin.Number)
	if err != nil {
		return nil, err
	}

	ienc, err := idata.Encode()
	if err != nil {
		return nil, err
	}

	// Call BlockBuilder_inherent_extrinsics which returns the inherents as extrinsics
	inherentExts, err := b.rt.InherentExtrinsics(ienc)
	if err != nil {
		return nil, err
	}

	// decode inherent extrinsics
	exts, err := scale.Decode(inherentExts, [][]byte{})
	if err != nil {
		return nil, err
	}

	// apply each inherent extrinsic
	for _, ext := range exts.([][]byte) {
		in, err := scale.Encode(ext)
		if err != nil {
			return nil, err
		}

		ret, err := b.rt.ApplyExtrinsic(in)
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(ret, []byte{0, 0}) {
			errTxt := determineErr(ret)
			return nil, fmt.Errorf("error applying inherent: %s", errTxt)
		}
	}

	return exts.([][]byte), nil
}

func (b *Service) addToQueue(txs []*transaction.ValidTransaction) {
	for _, t := range txs {
		hash, err := b.transactionState.Push(t)
		if err != nil {
			logger.Trace("Failed to add transaction to queue", "error", err)
		} else {
			logger.Trace("Added transaction to queue", "hash", hash)
		}
	}
}

func hasSlotEnded(slot Slot) bool {
	slotEnd := slot.start.Add(slot.duration)
	return time.Since(slotEnd) >= 0
}

func extrinsicsToBody(inherents [][]byte, txs []*transaction.ValidTransaction) (*types.Body, error) {
	extrinsics := types.BytesArrayToExtrinsics(inherents)

	for _, tx := range txs {
		decExt, err := scale.Decode(tx.Extrinsic, []byte{})
		if err != nil {
			return nil, err
		}
		extrinsics = append(extrinsics, decExt.([]byte))
	}

	return types.NewBodyFromExtrinsics(extrinsics)
}
