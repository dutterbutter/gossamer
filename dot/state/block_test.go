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

package state

import (
	"math/big"
	"testing"
	"time"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/trie"

	"github.com/ChainSafe/chaindb"
	"github.com/stretchr/testify/require"
)

var testGenesisHeader = &types.Header{
	Number:    big.NewInt(0),
	StateRoot: trie.EmptyHash,
}

func newTestBlockState(t *testing.T, header *types.Header) *BlockState {
	db := NewInMemoryDB(t)
	if header == nil {
		return &BlockState{
			db: chaindb.NewTable(db, blockPrefix),
		}
	}

	bs, err := NewBlockStateFromGenesis(db, header)
	require.NoError(t, err)
	return bs
}

func TestSetAndGetHeader(t *testing.T) {
	bs := newTestBlockState(t, nil)

	header := &types.Header{
		Number:    big.NewInt(0),
		StateRoot: trie.EmptyHash,
		Digest:    types.Digest{},
	}

	err := bs.SetHeader(header)
	require.NoError(t, err)

	res, err := bs.GetHeader(header.Hash())
	require.NoError(t, err)
	require.Equal(t, header, res)
}

func TestHasHeader(t *testing.T) {
	bs := newTestBlockState(t, nil)

	header := &types.Header{
		Number:    big.NewInt(0),
		StateRoot: trie.EmptyHash,
		Digest:    types.Digest{},
	}

	err := bs.SetHeader(header)
	require.NoError(t, err)

	has, err := bs.HasHeader(header.Hash())
	require.NoError(t, err)
	require.Equal(t, true, has)
}

func TestGetBlockByNumber(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)

	blockHeader := &types.Header{
		ParentHash: testGenesisHeader.Hash(),
		Number:     big.NewInt(1),
		Digest:     types.Digest{},
	}

	block := &types.Block{
		Header: blockHeader,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	// AddBlock also sets mapping [blockNumber : hash] in DB
	err := bs.AddBlock(block)
	require.NoError(t, err)

	retBlock, err := bs.GetBlockByNumber(blockHeader.Number)
	require.NoError(t, err)
	require.Equal(t, block, retBlock, "Could not validate returned retBlock as expected")
}

func TestAddBlock(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)

	// Create header
	header0 := &types.Header{
		Number:     big.NewInt(0),
		Digest:     types.Digest{},
		ParentHash: testGenesisHeader.Hash(),
	}
	// Create blockHash
	blockHash0 := header0.Hash()
	// BlockBody with fake extrinsics
	blockBody0 := types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	block0 := &types.Block{
		Header: header0,
		Body:   &blockBody0,
	}

	// Add the block0 to the DB
	err := bs.AddBlock(block0)
	require.NoError(t, err)

	// Create header & blockData for block 1
	header1 := &types.Header{
		Number:     big.NewInt(1),
		Digest:     types.Digest{},
		ParentHash: blockHash0,
	}
	blockHash1 := header1.Hash()

	// Create Block with fake extrinsics
	blockBody1 := types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	block1 := &types.Block{
		Header: header1,
		Body:   &blockBody1,
	}

	// Add the block1 to the DB
	err = bs.AddBlock(block1)
	require.NoError(t, err)

	// Get the blocks & check if it's the same as the added blocks
	retBlock, err := bs.GetBlockByHash(blockHash0)
	require.NoError(t, err)

	require.Equal(t, block0, retBlock, "Could not validate returned block0 as expected")

	retBlock, err = bs.GetBlockByHash(blockHash1)
	require.NoError(t, err)

	// this will panic if not successful, so catch and fail it so
	func() {
		hash := retBlock.Header.Hash()
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("got panic when processing retBlock.Header.Hash() ", r)
			}
		}()
		require.NotEqual(t, hash, common.Hash{})
	}()

	require.Equal(t, block1, retBlock, "Could not validate returned block1 as expected")

	// Check if latestBlock is set correctly
	require.Equal(t, block1.Header.Hash(), bs.BestBlockHash(), "Latest Header Block Check Fail")
}

func TestGetSlotForBlock(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)
	expectedSlot := uint64(77)

	babeHeader := types.NewBabePrimaryPreDigest(0, expectedSlot, [32]byte{}, [64]byte{})
	data := babeHeader.Encode()
	preDigest := types.NewBABEPreRuntimeDigest(data)

	block := &types.Block{
		Header: &types.Header{
			ParentHash: testGenesisHeader.Hash(),
			Number:     big.NewInt(int64(1)),
			Digest:     types.Digest{preDigest},
		},
		Body: &types.Body{},
	}

	err := bs.AddBlock(block)
	require.NoError(t, err)

	res, err := bs.GetSlotForBlock(block.Header.Hash())
	require.NoError(t, err)
	require.Equal(t, expectedSlot, res)
}

func TestIsBlockOnCurrentChain(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)
	currChain, branchChains := AddBlocksToState(t, bs, 3)

	for _, header := range currChain {
		onChain, err := bs.isBlockOnCurrentChain(header)
		require.NoError(t, err)

		if !onChain {
			t.Fatalf("Fail: expected block %s to be on current chain", header.Hash())
		}
	}

	for _, header := range branchChains {
		onChain, err := bs.isBlockOnCurrentChain(header)
		require.NoError(t, err)

		if onChain {
			t.Fatalf("Fail: expected block %s not to be on current chain", header.Hash())
		}
	}
}

func TestAddBlock_BlockNumberToHash(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)
	currChain, branchChains := AddBlocksToState(t, bs, 8)

	bestHash := bs.BestBlockHash()
	bestHeader, err := bs.BestBlockHeader()
	require.NoError(t, err)

	var resBlock *types.Block
	for _, header := range currChain {
		resBlock, err = bs.GetBlockByNumber(header.Number)
		require.NoError(t, err)

		if resBlock.Header.Hash() != header.Hash() {
			t.Fatalf("Fail: got %s expected %s for block %d", resBlock.Header.Hash(), header.Hash(), header.Number)
		}
	}

	for _, header := range branchChains {
		resBlock, err = bs.GetBlockByNumber(header.Number)
		require.NoError(t, err)

		if resBlock.Header.Hash() == header.Hash() {
			t.Fatalf("Fail: should not have gotten block %s for branch block num=%d", header.Hash(), header.Number)
		}
	}

	newBlock := &types.Block{
		Header: &types.Header{
			ParentHash: bestHash,
			Number:     big.NewInt(0).Add(bestHeader.Number, big.NewInt(1)),
		},
		Body: &types.Body{},
	}

	err = bs.AddBlock(newBlock)
	require.NoError(t, err)

	resBlock, err = bs.GetBlockByNumber(newBlock.Header.Number)
	require.NoError(t, err)

	if resBlock.Header.Hash() != newBlock.Header.Hash() {
		t.Fatalf("Fail: got %s expected %s for block %d", resBlock.Header.Hash(), newBlock.Header.Hash(), newBlock.Header.Number)
	}
}

func TestFinalizedHash(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)
	h, err := bs.GetFinalizedHash(0, 0)
	require.NoError(t, err)
	require.Equal(t, testGenesisHeader.Hash(), h)

	testhash := common.Hash{1, 2, 3, 4}
	err = bs.SetFinalizedHash(testhash, 1, 1)
	require.NoError(t, err)

	h, err = bs.GetFinalizedHash(1, 1)
	require.NoError(t, err)
	require.Equal(t, testhash, h)
}

func TestLatestFinalizedRound(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)
	r, err := bs.GetRound()
	require.NoError(t, err)
	require.Equal(t, uint64(0), r)

	err = bs.SetRound(99)
	require.NoError(t, err)
	r, err = bs.GetRound()
	require.NoError(t, err)
	require.Equal(t, uint64(99), r)
}

func TestFinalization_DeleteBlock(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)
	AddBlocksToState(t, bs, 5)

	btBefore := bs.bt.DeepCopy()
	t.Log(btBefore)
	before := bs.bt.GetAllBlocks()
	leaves := bs.Leaves()

	// TODO: why isn't arrival time set?
	// for _, n := range before {
	// 	has, err := bs.HasArrivalTime(n)
	// 	require.NoError(t, err)
	// 	require.True(t, has, n)
	// }

	// pick block to finalise
	fin := leaves[len(leaves)-1]
	err := bs.SetFinalizedHash(fin, 1, 1)
	require.NoError(t, err)

	after := bs.bt.GetAllBlocks()
	t.Log(bs.bt)

	isIn := func(arr []common.Hash, b common.Hash) bool {
		for _, a := range arr {
			if b == a {
				return true
			}
		}
		return false
	}

	// assert that every block except finalised has been deleted
	for _, b := range before {
		if b == fin {
			continue
		}

		if isIn(after, b) {
			continue
		}

		isFinalised, err := btBefore.IsDescendantOf(b, fin)
		require.NoError(t, err)

		has, err := bs.HasHeader(b)
		require.NoError(t, err)
		if isFinalised {
			require.True(t, has)
		} else {
			require.False(t, has)
		}

		has, err = bs.HasBlockBody(b)
		require.NoError(t, err)
		if isFinalised {
			require.True(t, has)
		} else {
			require.False(t, has)
		}

		// has, err = bs.HasArrivalTime(b)
		// require.NoError(t, err)
		// if isFinalised && b != bs.genesisHash {
		// 	require.True(t, has, b)
		// } else {
		// 	require.False(t, has)
		// }
	}
}

func TestGetHashByNumber(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)

	res, err := bs.GetHashByNumber(big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, bs.genesisHash, res)

	header := &types.Header{
		Number:     big.NewInt(1),
		Digest:     types.Digest{},
		ParentHash: testGenesisHeader.Hash(),
	}

	block := &types.Block{
		Header: header,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err = bs.AddBlock(block)
	require.NoError(t, err)

	res, err = bs.GetHashByNumber(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, header.Hash(), res)
}

func TestAddBlock_WithReOrg(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)

	header1a := &types.Header{
		Number:     big.NewInt(1),
		Digest:     types.Digest{},
		ParentHash: testGenesisHeader.Hash(),
	}

	block1a := &types.Block{
		Header: header1a,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err := bs.AddBlock(block1a)
	require.NoError(t, err)

	block1hash, err := bs.GetHashByNumber(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, header1a.Hash(), block1hash)

	header1b := &types.Header{
		Number:         big.NewInt(1),
		Digest:         types.Digest{},
		ParentHash:     testGenesisHeader.Hash(),
		ExtrinsicsRoot: common.Hash{99},
	}

	block1b := &types.Block{
		Header: header1b,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err = bs.AddBlock(block1b)
	require.NoError(t, err)

	// should still be hash 1a since it arrived first
	block1hash, err = bs.GetHashByNumber(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, header1a.Hash(), block1hash)

	header2b := &types.Header{
		Number:         big.NewInt(2),
		Digest:         types.Digest{},
		ParentHash:     header1b.Hash(),
		ExtrinsicsRoot: common.Hash{99},
	}

	block2b := &types.Block{
		Header: header2b,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err = bs.AddBlock(block2b)
	require.NoError(t, err)

	// should now be hash 1b since it's on the longer chain
	block1hash, err = bs.GetHashByNumber(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, header1b.Hash(), block1hash)

	block2hash, err := bs.GetHashByNumber(big.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, header2b.Hash(), block2hash)

	header2a := &types.Header{
		Number:     big.NewInt(2),
		Digest:     types.Digest{},
		ParentHash: header1a.Hash(),
	}

	block2a := &types.Block{
		Header: header2a,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err = bs.AddBlock(block2a)
	require.NoError(t, err)

	header3a := &types.Header{
		Number:     big.NewInt(3),
		Digest:     types.Digest{},
		ParentHash: header2a.Hash(),
	}

	block3a := &types.Block{
		Header: header3a,
		Body:   &types.Body{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	err = bs.AddBlock(block3a)
	require.NoError(t, err)

	// should now be hash 1a since it's on the longer chain
	block1hash, err = bs.GetHashByNumber(big.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, header1a.Hash(), block1hash)

	// should now be hash 2a since it's on the longer chain
	block2hash, err = bs.GetHashByNumber(big.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, header2a.Hash(), block2hash)

	block3hash, err := bs.GetHashByNumber(big.NewInt(3))
	require.NoError(t, err)
	require.Equal(t, header3a.Hash(), block3hash)
}

func TestAddBlockToBlockTree(t *testing.T) {
	bs := newTestBlockState(t, testGenesisHeader)

	header := &types.Header{
		Number:     big.NewInt(1),
		Digest:     types.Digest{},
		ParentHash: testGenesisHeader.Hash(),
	}

	err := bs.setArrivalTime(header.Hash(), time.Now())
	require.NoError(t, err)

	err = bs.AddBlockToBlockTree(header)
	require.NoError(t, err)
	require.Equal(t, bs.BestBlockHash(), header.Hash())
}
