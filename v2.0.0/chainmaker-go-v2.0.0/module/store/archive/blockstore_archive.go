/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package archive

import (
	"errors"
	"sync"

	"chainmaker.org/chainmaker/protocol/v2"

	"chainmaker.org/chainmaker-go/store/blockdb"
	"chainmaker.org/chainmaker-go/store/resultdb"
	"chainmaker.org/chainmaker-go/store/serialization"

	"chainmaker.org/chainmaker-go/localconf"
)

const defaultMinUnArchiveBlockHeight = 10
const defaultUnArchiveBlockHeight = 300000 //about 7 days block produces

var (
	errHeightNotReach        = errors.New("target archive height not reach")
	errLastHeightTooLow      = errors.New("chain last height too low to archive")
	errHeightTooLow          = errors.New("target archive height too low")
	errRestoreHeightNotMatch = errors.New("restore block height not match last archived height")
	//ErrInvalidateRestoreBlocks invalidate restore blocks
	ErrInvalidateRestoreBlocks = errors.New("invalidate restore blocks")
	//ErrConfigBlockArchive config block do not need archive
	ErrConfigBlockArchive = errors.New("config block do not need archive")
	//ErrArchivedTx archived transaction
	ErrArchivedTx = errors.New("archived transaction")
	//ErrArchivedRWSet archived RWSet
	ErrArchivedRWSet = errors.New("archived RWSet")
	//ErrArchivedBlock archived block
	ErrArchivedBlock = errors.New("archived block")
)

// ArchiveMgr provide handle to archive instances
type ArchiveMgr struct {
	sync.RWMutex
	archivedPivot        uint64
	unarchiveBlockHeight uint64
	blockDB              blockdb.BlockDB
	resultDB             resultdb.ResultDB
	storeConfig          *localconf.StorageConfig

	logger protocol.Logger
}

// NewArchiveMgr construct a new `ArchiveMgr` with given chainId
func NewArchiveMgr(chainId string, blockDB blockdb.BlockDB, resultDB resultdb.ResultDB,
	storeConfig *localconf.StorageConfig, logger protocol.Logger) (*ArchiveMgr, error) {
	archiveMgr := &ArchiveMgr{
		blockDB:     blockDB,
		resultDB:    resultDB,
		logger:      logger,
		storeConfig: storeConfig,
	}

	unarchiveBlockHeight := uint64(0)
	cfgUnArchiveBlockHeight := archiveMgr.storeConfig.UnArchiveBlockHeight
	if cfgUnArchiveBlockHeight == 0 {
		unarchiveBlockHeight = defaultUnArchiveBlockHeight
		archiveMgr.logger.Infof(
			"config UnArchiveBlockHeight not set, will set to defaultMinUnArchiveBlockHeight:[%d]", defaultUnArchiveBlockHeight)
	} else if cfgUnArchiveBlockHeight <= defaultMinUnArchiveBlockHeight {
		unarchiveBlockHeight = defaultMinUnArchiveBlockHeight
		archiveMgr.logger.Infof(
			"config UnArchiveBlockHeight is too low:[%d], will set to defaultMinUnArchiveBlockHeight:[%d]",
			cfgUnArchiveBlockHeight, defaultMinUnArchiveBlockHeight)
	} else if cfgUnArchiveBlockHeight > defaultMinUnArchiveBlockHeight {
		unarchiveBlockHeight = cfgUnArchiveBlockHeight
	}

	archiveMgr.unarchiveBlockHeight = unarchiveBlockHeight
	if _, err := archiveMgr.GetArchivedPivot(); err != nil {
		return nil, err
	}

	return archiveMgr, nil
}

// ArchiveBlock cache a block with given block height and update batch
func (mgr *ArchiveMgr) ArchiveBlock(archiveHeight uint64) error {
	mgr.Lock()
	defer mgr.Unlock()

	var (
		lastHeight, archivedPivot uint64
		txIdsMap                  map[uint64][]string
		err                       error
	)

	if lastHeight, err = mgr.blockDB.GetLastSavepoint(); err != nil {
		return err
	}

	if archivedPivot, err = mgr.GetArchivedPivot(); err != nil {
		return err
	}

	//archiveHeight should between archivedPivot and lastHeight - unarchiveBlockHeight
	if lastHeight <= mgr.unarchiveBlockHeight {
		return errLastHeightTooLow
	} else if mgr.archivedPivot >= archiveHeight {
		return errHeightTooLow
	} else if archiveHeight >= lastHeight-mgr.unarchiveBlockHeight {
		return errHeightNotReach
	}

	if txIdsMap, err = mgr.blockDB.ShrinkBlocks(archivedPivot+1, archiveHeight); err != nil {
		return err
	}

	if err = mgr.resultDB.ShrinkBlocks(txIdsMap); err != nil {
		return err
	}

	mgr.logger.Infof("archived block from [%d] to [%d], block size:%d",
		mgr.archivedPivot, archiveHeight, archiveHeight-mgr.archivedPivot)

	return nil
}

// RestoreBlock restore block from outside block data
func (mgr *ArchiveMgr) RestoreBlock(blockInfos []*serialization.BlockWithSerializedInfo) error {
	mgr.Lock()
	defer mgr.Unlock()
	if len(blockInfos) == 0 {
		mgr.logger.Warnf("restore block is empty")
		return nil
	}

	//make sure archivedPivot is latest
	if _, err := mgr.GetArchivedPivot(); err != nil {
		return err
	}

	total := len(blockInfos)
	lastRestoreHeight := uint64(blockInfos[total-1].Block.Header.BlockHeight)
	if lastRestoreHeight != mgr.archivedPivot {
		mgr.logger.Errorf("restore last block height[%d] not match node archived height[%d]",
			blockInfos[total-1].Block.Header.BlockHeight, mgr.archivedPivot)
		return errRestoreHeightNotMatch
	}

	//restore block info should be continuous
	curHeight := uint64(lastRestoreHeight)
	for i := 0; i < total; i++ {
		if blockInfos[total-i-1].Block.Header.BlockHeight != curHeight {
			return ErrInvalidateRestoreBlocks
		}
		curHeight = curHeight - 1
	}

	if err := mgr.blockDB.RestoreBlocks(blockInfos); err != nil {
		return err
	}

	if err := mgr.resultDB.RestoreBlocks(blockInfos); err != nil {
		return err
	}

	mgr.logger.Infof("retore block from [%d] to [%d], block size:%d",
		lastRestoreHeight, blockInfos[0].Block.Header.BlockHeight, total)
	return nil
}

// GetArchivedPivot return restore block pivot
func (mgr *ArchiveMgr) GetArchivedPivot() (uint64, error) {
	archivedPivot, err := mgr.blockDB.GetArchivedPivot()
	if err != nil {
		return 0, err
	}

	mgr.archivedPivot = archivedPivot
	return mgr.archivedPivot, nil
}

// GetMinUnArchiveBlockSize return unarchiveBlockHeight
func (mgr *ArchiveMgr) GetMinUnArchiveBlockSize() uint64 {
	return mgr.unarchiveBlockHeight
}
