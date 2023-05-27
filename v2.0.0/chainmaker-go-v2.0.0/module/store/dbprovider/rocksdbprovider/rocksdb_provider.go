// +build rocksdb

/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rocksdbprovider

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"chainmaker.org/chainmaker-go/localconf"
	//logImpl "chainmaker.org/chainmaker-go/logger"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/pkg/errors"
	"github.com/tecbot/gorocksdb"
)

const (
	defaultBloomFilterBits          = 10
	defaultWriteBufferNumber        = 16
	defaultMaxBackgroundCompactions = 10
	defaultFixedPrefixTransform     = 5
)

const (
	KiB = 1024
	MiB = KiB * 1024
)

const (
	StoreBlockDBDir   = "store_block"
	StoreStateDBDir   = "store_state"
	StoreHistoryDBDir = "store_history"
)

var DbNameKeySep = []byte{0x00}

// Provider provides handle to db instances
type Provider struct {
	db        *gorocksdb.DB
	dbHandles map[string]*RocksDBHandle
	mutex     sync.Mutex

	logger protocol.Logger
}

//// NewBlockProvider construct a new Rocksdb Provider for block operation with given chainId
//func NewBlockProvider(chainId string) *Provider {
//	return NewProvider(chainId, StoreBlockDBDir)
//}
//
//// NewStateProvider construct a new Rocksdb Provider for state operation with given chainId
//func NewStateProvider(chainId string) *Provider {
//	return NewProvider(chainId, StoreStateDBDir)
//}
//
//// NewHistoryProvider construct a new Rocksdb Provider for history operation with given chainId
//func NewHistoryProvider(chainId string) *Provider {
//	return NewProvider(chainId, StoreHistoryDBDir)
//}

// NewProvider construct a new db Provider for given chainId and dir
func NewProvider(chainId string, dbDir string, logger protocol.Logger) *Provider {
	dbOpts := NewRocksdbConfig()
	writeBufferSize := localconf.ChainMakerConfig.StorageConfig.WriteBufferSize
	if writeBufferSize > 0 {
		dbOpts.writeBufferSize = writeBufferSize * MiB
	}
	bloomFilterBits := localconf.ChainMakerConfig.StorageConfig.BloomFilterBits
	if bloomFilterBits > 0 {
		dbOpts.bloomFilterBits = bloomFilterBits
	}
	dbPath := filepath.Join(localconf.ChainMakerConfig.StorageConfig.StorePath, chainId, dbDir)
	rocksdbOpts := dbOpts.ToOptions()

	err := os.MkdirAll(dbPath, 0755)
	if err != nil {
		panic(fmt.Sprintf("Error create rocksdb path: %s", err))
	}

	db, err := gorocksdb.OpenDb(rocksdbOpts, dbPath)
	if err != nil {
		panic(fmt.Sprintf("Error opening rocksdbdbprovider: %s", err))
	}
	return &Provider{
		db:        db,
		dbHandles: make(map[string]*RocksDBHandle),
		mutex:     sync.Mutex{},

		logger: logger,
	}
}

// GetDBHandle returns a [Rocksdb] DBHandle for given dbname
func (p *Provider) GetDBHandle(dbName string) protocol.DBHandle {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	dbHandle := p.dbHandles[dbName]
	if dbHandle == nil {
		dbHandle = &RocksDBHandle{
			dbName:       dbName,
			db:           p.db,
			readOptions:  gorocksdb.NewDefaultReadOptions(),
			writeOptions: gorocksdb.NewDefaultWriteOptions(),
			logger:       p.logger,
		}
		p.dbHandles[dbName] = dbHandle
	}
	return dbHandle
}

// Close is used to close rocksdb database
func (p *Provider) Close() error {
	p.db.Close()
	return nil
}

// RocksDBConfig config of rocksdb
type RocksDBConfig struct {
	bloomFilterBits          int
	writeBufferSize          int
	maxWriteBufferNumber     int
	maxBackgroundCompactions int
	blockCache               int
}

// NewRocksdbConfig create a new rocksdb config
func NewRocksdbConfig() *RocksDBConfig {
	dbOpts := RocksDBConfig{
		bloomFilterBits:          defaultBloomFilterBits,
		writeBufferSize:          64 * MiB,
		maxWriteBufferNumber:     defaultWriteBufferNumber,
		maxBackgroundCompactions: defaultMaxBackgroundCompactions,
		blockCache:               64 * KiB}
	return &dbOpts
}

// ToOptions convert rocksdb config to options
func (config *RocksDBConfig) ToOptions() *gorocksdb.Options {
	options := gorocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true) // 不存在则创建

	bloomFilter := gorocksdb.NewBloomFilter(config.bloomFilterBits) // 布隆过滤器

	options.SetCreateIfMissing(true)
	options.SetWriteBufferSize(config.writeBufferSize)
	options.SetMaxWriteBufferNumber(config.maxWriteBufferNumber)
	options.SetMaxBackgroundCompactions(config.maxBackgroundCompactions)

	blockBasedTableOptions := gorocksdb.NewDefaultBlockBasedTableOptions()
	blockBasedTableOptions.SetBlockCache(gorocksdb.NewLRUCache(uint64(config.blockCache)))
	blockBasedTableOptions.SetFilterPolicy(bloomFilter)
	blockBasedTableOptions.SetBlockCacheCompressed(gorocksdb.NewLRUCache(uint64(config.blockCache)))
	blockBasedTableOptions.SetCacheIndexAndFilterBlocks(true)
	blockBasedTableOptions.SetIndexType(gorocksdb.KHashSearchIndexType)

	options.SetBlockBasedTableFactory(blockBasedTableOptions)
	options.SetPrefixExtractor(gorocksdb.NewFixedPrefixTransform(defaultFixedPrefixTransform))
	options.SetAllowConcurrentMemtableWrites(false)
	return options
}

// RocksDBHandle Rocksdb database handle
type RocksDBHandle struct {
	dbName       string
	readOptions  *gorocksdb.ReadOptions
	writeOptions *gorocksdb.WriteOptions
	db           *gorocksdb.DB

	logger protocol.Logger
}

// Get get value from rocksdb
func (dbHandle *RocksDBHandle) Get(key []byte) ([]byte, error) {
	value, err := dbHandle.db.GetBytes(dbHandle.readOptions, makeKeyWithDbName(dbHandle.dbName, key))
	if err != nil {
		dbHandle.logger.Errorf("getting rocksdbprovider key [%#v], err:%s", key, err.Error())
		return nil, errors.Wrapf(err, "error getting rocksdbprovider key [%#v]", key)
	}
	return value, nil
}

// Put put key,value to rocksdb
func (dbHandle *RocksDBHandle) Put(key []byte, value []byte) error {
	if value == nil {
		dbHandle.logger.Warn("writing rocksdbprovider key [%#v] with nil value", key)
		return errors.New("error writing rocksdbprovider with nil value")
	}
	err := dbHandle.db.Put(dbHandle.writeOptions, makeKeyWithDbName(dbHandle.dbName, key), value)
	if err != nil {
		dbHandle.logger.Errorf("writing rocksdbprovider key [%#v]", key)
		return errors.Wrapf(err, "error writing rocksdbprovider key [%#v]", key)
	}
	return err
}

// Has check if exist for key in rocksdb
func (dbHandle *RocksDBHandle) Has(key []byte) (bool, error) {
	value, err := dbHandle.db.Get(dbHandle.readOptions, makeKeyWithDbName(dbHandle.dbName, key))
	if value == nil {
		dbHandle.logger.Errorf("can not get rocksdbprovider key [%#v]", key)
		return false, errors.Wrapf(err, "can not get rocksdbprovider key [%#v]", key)
	}
	if err != nil {
		dbHandle.logger.Errorf("getting rocksdbprovider key [%#v], err:%s", key, err.Error())
		return false, errors.Wrapf(err, "error getting rocksdbprovider key [%#v]", key)
	}
	return value.Exists(), nil
}

// Delete delete key from rocksdb
func (dbHandle *RocksDBHandle) Delete(key []byte) error {
	err := dbHandle.db.Delete(dbHandle.writeOptions, makeKeyWithDbName(dbHandle.dbName, key))
	if err != nil {
		dbHandle.logger.Errorf("deleting rocksdbprovider key [%#v]", key)
		return errors.Wrapf(err, "error deleting rocksdbprovider key [%#v]", key)
	}
	return err
}

// WriteBatch writes a batch in an atomic way
func (dbHandle *RocksDBHandle) WriteBatch(batch protocol.StoreBatcher, sync bool) error {
	if batch.Len() == 0 {
		return nil
	}
	writeBatch := gorocksdb.NewWriteBatch()
	for k, v := range batch.KVs() {
		key := makeKeyWithDbName(dbHandle.dbName, []byte(k))
		if v == nil {
			writeBatch.Delete(key)
		} else {
			writeBatch.Put(key, v)
		}
	}

	if err := dbHandle.db.Write(dbHandle.writeOptions, writeBatch); err != nil {
		dbHandle.logger.Errorf("write batch to rocksdbprovider failed")
		return errors.Wrap(err, "error writing batch to rocksdbprovider")
	}
	return nil
}

// CompactRange compacts the underlying DB for the given key range.
func (dbHandle *RocksDBHandle) CompactRange(start, limit []byte) error {
	dbHandle.db.CompactRange(gorocksdb.Range{
		Start: start,
		Limit: limit,
	})
	return nil
}

// NewIteratorWithRange returns an iterator that contains all the key-values between given key ranges
// start is included in the results and limit is excluded.
func (dbHandle *RocksDBHandle) NewIteratorWithRange(start []byte, limit []byte) protocol.Iterator {
	// todo
	panic("not yet implemented for rocksdb")
}

// NewIteratorWithPrefix returns an iterator that contains all the key-values with given prefix
func (dbHandle *RocksDBHandle) NewIteratorWithPrefix(prefix []byte) protocol.Iterator {
	// todo
	panic("not yet implemented for rocksdb")
}
func (dbHandle *RocksDBHandle) Close() error {
	dbHandle.db.Close()
	return nil
}
func makeKeyWithDbName(column string, key []byte) []byte {
	return append(append([]byte(column), DbNameKeySep...), key...)
}
