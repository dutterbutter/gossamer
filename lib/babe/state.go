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
	"math/big"
	"time"

	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/lib/common"
	rtstorage "github.com/ChainSafe/gossamer/lib/runtime/storage"
	"github.com/ChainSafe/gossamer/lib/transaction"
)

// BlockState interface for block state methods
type BlockState interface {
	BestBlockHash() common.Hash
	BestBlockHeader() (*types.Header, error)
	BestBlockNumber() (*big.Int, error)
	BestBlock() (*types.Block, error)
	SubChain(start, end common.Hash) ([]common.Hash, error)
	AddBlock(*types.Block) error
	GetAllBlocksAtDepth(hash common.Hash) []common.Hash
	GetHeader(common.Hash) (*types.Header, error)
	GetBlockByNumber(*big.Int) (*types.Block, error)
	GetBlockByHash(common.Hash) (*types.Block, error)
	GetArrivalTime(common.Hash) (time.Time, error)
	GenesisHash() common.Hash
	GetSlotForBlock(common.Hash) (uint64, error)
	GetFinalizedHeader(uint64, uint64) (*types.Header, error)
	IsDescendantOf(parent, child common.Hash) (bool, error)
}

// StorageState interface for storage state methods
type StorageState interface {
	TrieState(hash *common.Hash) (*rtstorage.TrieState, error)
	StoreTrie(ts *rtstorage.TrieState) error
}

// TransactionState is the interface for transaction queue methods
type TransactionState interface {
	Push(vt *transaction.ValidTransaction) (common.Hash, error)
	Pop() *transaction.ValidTransaction
	Peek() *transaction.ValidTransaction
}

// EpochState is the interface for epoch methods
type EpochState interface {
	SetCurrentEpoch(epoch uint64) error
	GetCurrentEpoch() (uint64, error)
	SetEpochData(uint64, *types.EpochData) error
	GetEpochData(epoch uint64) (*types.EpochData, error)
	HasEpochData(epoch uint64) (bool, error)
	GetConfigData(epoch uint64) (*types.ConfigData, error)
	HasConfigData(epoch uint64) (bool, error)
	GetStartSlotForEpoch(epoch uint64) (uint64, error)
	GetEpochForBlock(header *types.Header) (uint64, error)
	SetFirstSlot(slot uint64) error
	GetLatestEpochData() (*types.EpochData, error)
	SkipVerify(*types.Header) (bool, error)
}
