/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockdb

import (
	"chainmaker.org/chainmaker-go/store/serialization"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	storePb "chainmaker.org/chainmaker/pb-go/v2/store"
)

// BlockDB provides handle to block and tx instances
type BlockDB interface {
	InitGenesis(genesisBlock *serialization.BlockWithSerializedInfo) error

	//GetBlockHeaderByHash(blockHash []byte) (*commonPb.BlockHeader, error)
	//GetBlockHeaderByHeight(blockHash []byte) (*commonPb.BlockHeader, error)

	// CommitBlock commits the block and the corresponding rwsets in an atomic operation
	CommitBlock(blockInfo *serialization.BlockWithSerializedInfo) error

	// BlockExists returns true if the block hash exist, or returns false if none exists.
	BlockExists(blockHash []byte) (bool, error)

	// GetBlockByHash returns a block given it's hash, or returns nil if none exists.
	GetBlockByHash(blockHash []byte) (*commonPb.Block, error)

	// GetHeightByHash returns a block height given it's hash, or returns nil if none exists.
	GetHeightByHash(blockHash []byte) (uint64, error)

	// GetBlockHeaderByHeight returns a block header by given it's height, or returns nil if none exists.
	GetBlockHeaderByHeight(height uint64) (*commonPb.BlockHeader, error)

	// GetBlock returns a block given it's block height, or returns nil if none exists.
	GetBlock(height uint64) (*commonPb.Block, error)

	// GetTx retrieves a transaction by txid, or returns nil if none exists.
	GetTx(txId string) (*commonPb.Transaction, error)
	GetTxWithBlockInfo(txId string) (*commonPb.TransactionInfo, error)

	// GetTxHeight retrieves a transaction height by txid, or returns nil if none exists.
	GetTxHeight(txId string) (uint64, error)

	// TxExists returns true if the tx exist, or returns false if none exists.
	TxExists(txId string) (bool, error)

	// TxArchived returns true if the tx archived, or returns false.
	TxArchived(txId string) (bool, error)

	// GetTxConfirmedTime retrieves time of the tx confirmed in the blockChain
	GetTxConfirmedTime(txId string) (int64, error)

	// GetLastBlock returns the last block.
	GetLastBlock() (*commonPb.Block, error)

	// GetFilteredBlock returns a filtered block given it's block height, or return nil if none exists.
	GetFilteredBlock(height uint64) (*storePb.SerializedBlock, error)

	// GetLastSavepoint reurns the last block height
	GetLastSavepoint() (uint64, error)

	// GetLastConfigBlock returns the last config block.
	GetLastConfigBlock() (*commonPb.Block, error)

	// GetBlockByTx returns a block which contains a tx.如果查询不到，则返回nil,nil
	GetBlockByTx(txId string) (*commonPb.Block, error)

	// GetArchivedPivot get archived pivot
	GetArchivedPivot() (uint64, error)

	// ShrinkBlocks archive old blocks in an atomic operation
	ShrinkBlocks(startHeight uint64, endHeight uint64) (map[uint64][]string, error)

	// RestoreBlocks restore blocks from outside block data in an atomic operation
	RestoreBlocks(blockInfos []*serialization.BlockWithSerializedInfo) error

	// Close is used to close database
	Close()
}
