/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package statesqldb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"chainmaker.org/chainmaker/pb-go/v2/syscontract"

	configPb "chainmaker.org/chainmaker/pb-go/v2/config"

	"chainmaker.org/chainmaker-go/utils"

	"chainmaker.org/chainmaker-go/localconf"
	"chainmaker.org/chainmaker-go/store/dbprovider/rawsqlprovider"
	"chainmaker.org/chainmaker-go/store/serialization"
	"chainmaker.org/chainmaker-go/store/types"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

// StateSqlDB provider a implementation of `statedb.StateDB`
// This implementation provides a mysql based data model
type StateSqlDB struct {
	db          protocol.SqlDBHandle
	contractDbs map[string]protocol.SqlDBHandle
	dbConfig    *localconf.SqlDbConfig
	logger      protocol.Logger
	chainId     string
	sync.Mutex
	dbName string
}

//如果数据库不存在，则创建数据库，然后切换到这个数据库，创建表
//如果数据库存在，则切换数据库，检查表是否存在，不存在则创建表。
func (db *StateSqlDB) initContractDb(contractName string) error {
	dbName := getContractDbName(db.dbConfig, db.chainId, contractName)
	db.logger.Debugf("try to create state db %s", dbName)
	_, err := db.db.CreateDatabaseIfNotExist(dbName)
	if err != nil {
		db.logger.Panic("init state sql db fail")
	}
	db.logger.Debugf("try to create state db table: state_infos for contract[%s]", contractName)
	dbHandle := db.getContractDbHandle(contractName)
	err = dbHandle.CreateTableIfNotExist(&StateInfo{})
	if err != nil {
		db.logger.Panic("init state info sql db table fail:" + err.Error())
	}
	err = dbHandle.CreateTableIfNotExist(&StateRecordSql{})
	if err != nil {
		db.logger.Panic("init state record sql sql db table fail:" + err.Error())
	}
	return nil
}
func (db *StateSqlDB) initSystemStateDb(dbName string) error {
	db.logger.Debugf("try to create state db %s", dbName)
	_, err := db.db.CreateDatabaseIfNotExist(dbName)
	if err != nil {
		panic("init state sql db fail")
	}
	db.logger.Debug("try to create system state db table: state_infos")
	err = db.db.CreateTableIfNotExist(&StateInfo{})
	if err != nil {
		panic("init state sql db table fail:" + err.Error())
	}
	db.logger.Debug("try to create system state db table: save_points")
	err = db.db.CreateTableIfNotExist(&types.SavePoint{})
	if err != nil {
		panic("init state sql db table fail:" + err.Error())
	}
	_, err = db.db.Save(&types.SavePoint{BlockHeight: 0})
	return err
}

// NewStateSqlDB construct a new `StateDB` for given chainId
func NewStateSqlDB(chainId string, dbConfig *localconf.SqlDbConfig, logger protocol.Logger) (*StateSqlDB, error) {
	dbName := getDbName(dbConfig, chainId)
	db := rawsqlprovider.NewSqlDBHandle(dbName, dbConfig, logger)
	return newStateSqlDB(dbName, chainId, db, dbConfig, logger)
}

func newStateSqlDB(dbName, chainId string, db protocol.SqlDBHandle, dbConfig *localconf.SqlDbConfig,
	logger protocol.Logger) (*StateSqlDB, error) {
	stateDB := &StateSqlDB{
		db:          db,
		dbConfig:    dbConfig,
		logger:      logger,
		chainId:     chainId,
		dbName:      dbName,
		contractDbs: make(map[string]protocol.SqlDBHandle),
	}

	return stateDB, nil
}
func (s *StateSqlDB) InitGenesis(genesisBlock *serialization.BlockWithSerializedInfo) error {
	s.Lock()
	defer s.Unlock()
	err := s.initSystemStateDb(s.dbName)
	if err != nil {
		return err
	}
	return s.commitBlock(genesisBlock)
}
func getDbName(dbConfig *localconf.SqlDbConfig, chainId string) string {
	return dbConfig.DbPrefix + "statedb_" + chainId
}

func GetContractDbName(chainId, contractName string) string {
	return getContractDbName(localconf.ChainMakerConfig.StorageConfig.StateDbConfig.SqlDbConfig, chainId, contractName)
}

//getContractDbName calculate contract db name, if name length>64, keep start 50 chars add 10 hash chars and 4 tail
func getContractDbName(dbConfig *localconf.SqlDbConfig, chainId, contractName string) string {
	if _, ok := syscontract.SystemContract_value[contractName]; ok { //如果是系统合约，不为每个合约构建数据库，使用统一个statedb数据库
		return getDbName(dbConfig, chainId)
	}
	dbName := dbConfig.DbPrefix + "statedb_" + chainId + "_" + contractName
	if len(dbName) > 64 { //for mysql only support 64 chars
		h := sha256.New()
		h.Write([]byte(dbName))
		sum := h.Sum(nil)
		dbName = dbName[:50] + hex.EncodeToString(sum)[:10] + contractName[len(contractName)-4:]
	}
	return dbName
}

// CommitBlock commits the state in an atomic operation
func (s *StateSqlDB) CommitBlock(blockWithRWSet *serialization.BlockWithSerializedInfo) error {
	s.Lock()
	defer s.Unlock()
	return s.commitBlock(blockWithRWSet)
}
func (s *StateSqlDB) commitBlock(blockWithRWSet *serialization.BlockWithSerializedInfo) error {
	block := blockWithRWSet.Block
	txRWSets := blockWithRWSet.TxRWSets
	txKey := block.GetTxKey()
	if len(txRWSets) == 0 && len(block.Header.ConsensusArgs) == 0 {
		s.logger.Warnf("block[%d] don't have any read write set data", block.Header.BlockHeight)
		return nil
	}
	//1. get a db transaction
	dbTx, err := s.db.GetDbTransaction(txKey)
	s.logger.Infof("GetDbTransaction db:%v,err:%s", dbTx, err)
	processStateDbSqlOutside := false
	if err == nil { //外部已经开启了事务，不用重复创建事务
		s.logger.Debugf("db transaction[%s] already created outside, don't need process statedb sql again", txKey)
		processStateDbSqlOutside = true
	}
	//没有在外部开启事务，则开启事务，进行数据写入
	if !processStateDbSqlOutside {
		dbTx, err = s.db.BeginDbTransaction(txKey)
		if err != nil {
			return err
		}
	}

	//2. 如果是新建合约，则创建对应的数据库，并执行DDL
	if utils.IsContractMgmtBlock(block) {
		//创建对应合约的数据库
		payload := block.Txs[0].Payload
		runtime := payload.GetParameter(syscontract.InitContract_CONTRACT_RUNTIME_TYPE.String())
		contractId := &commonPb.Contract{
			Name:        string(payload.GetParameter(syscontract.InitContract_CONTRACT_NAME.String())),
			Version:     string(payload.GetParameter(syscontract.InitContract_CONTRACT_VERSION.String())),
			RuntimeType: commonPb.RuntimeType(commonPb.RuntimeType_value[string(runtime)]),
		}
		//if contractId.RuntimeType == commonPb.RuntimeType_EVM {
		//	address, _ := evmutils.MakeAddressFromString(contractId.Name)
		//	contractId.Name = address.String()
		//}
		err = s.updateStateForContractInit(dbTx, block, contractId, txRWSets[0].TxWrites, processStateDbSqlOutside)
		if err != nil {
			err2 := s.db.RollbackDbTransaction(txKey)
			if err2 != nil {
				return err2
			}
			return err
		}
	} else {
		//3. 不是新建合约，是普通的合约调用，则在事务中更新数据
		for _, txRWSet := range txRWSets {
			for _, txWrite := range txRWSet.TxWrites {
				err = s.operateDbByWriteSet(dbTx, block, txWrite, processStateDbSqlOutside)
				if err != nil {
					err2 := s.db.RollbackDbTransaction(txKey)
					if err2 != nil {
						return err2
					}
					return err
				}
			}
		}
		//3.5 处理BlockHeader中ConsensusArgs对应的合约状态数据更新
		if len(block.Header.ConsensusArgs) > 0 {
			err = s.updateConsensusArgs(dbTx, block)
			if err != nil {
				err2 := s.db.RollbackDbTransaction(txKey)
				if err2 != nil {
					return err2
				}
				return err
			}
		}
	}
	//4. 更新SavePoint
	err = s.updateSavePoint(dbTx, block.Header.BlockHeight)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	//5. commit transaction
	err = s.db.CommitDbTransaction(txKey)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	s.logger.Debugf("chain[%s]: commit state block[%d]", block.Header.ChainId, block.Header.BlockHeight)
	return nil
}
func (s *StateSqlDB) operateDbByWriteSet(dbTx protocol.SqlDBTransaction,
	block *commonPb.Block, txWrite *commonPb.TxWrite, processStateDbSqlOutside bool) error {
	contractDbName := getContractDbName(s.dbConfig, s.chainId, txWrite.ContractName)
	if txWrite.ContractName != "" { //切换DB
		err := dbTx.ChangeContextDb(contractDbName)
		if err != nil {
			return err
		}
	}
	if len(txWrite.Key) == 0 { //是sql
		if !processStateDbSqlOutside { // 没有在外面处理过，则在这里进行处理
			sql := string(txWrite.Value)
			if _, err := dbTx.ExecSql(sql); err != nil {
				s.logger.Errorf("execute sql[%s] get error:%s", txWrite.Value, err.Error())
				return err
			}
		}
	} else {
		stateInfo := NewStateInfo(txWrite.ContractName, txWrite.Key, txWrite.Value,
			uint64(block.Header.BlockHeight), block.GetTimestamp())
		if _, err := dbTx.Save(stateInfo); err != nil {
			s.logger.Errorf("save state key[%s] get error:%s", txWrite.Key, err.Error())
			return err
		}
	}
	return nil
}
func (s *StateSqlDB) updateSavePoint(dbTx protocol.SqlDBTransaction, height uint64) error {
	err := dbTx.ChangeContextDb(s.dbName)
	if err != nil {
		return err
	}
	_, err = dbTx.ExecSql("update save_points set block_height=?", height)
	if err != nil {
		s.logger.Errorf("update save point to %d get an error:%s", height, err)
		return err
	}
	return nil
}

//如果是创建或者升级合约，那么需要创建对应的数据库和state_infos表，然后执行DDL语句，然后如果是KV数据，保存数据
func (s *StateSqlDB) updateStateForContractInit(dbTx protocol.SqlDBTransaction, block *commonPb.Block,
	contractId *commonPb.Contract, writes []*commonPb.TxWrite, processStateDbSqlOutside bool) error {

	dbName := getContractDbName(s.dbConfig, block.Header.ChainId, contractId.Name)
	s.logger.Debugf("start init new db:%s for contract[%s]", dbName, contractId.Name)
	err := s.initContractDb(contractId.Name) //创建合约的数据库和KV表
	if err != nil {
		return err
	}
	s.logger.DebugDynamic(func() string {
		str := "WriteSet:"
		for i, w := range writes {
			str += fmt.Sprintf("id:%d,contract:%s,key:%s,value len:%d;", i, w.ContractName, w.Key, len(w.Value))
		}
		return str
	})

	for _, txWrite := range writes {
		if len(txWrite.Key) == 0 { //这是SQL语句
			// 已经在VM执行的时候执行了SQL则不处理，只有快速同步的时候，没有经过VM执行，才需要直接把写集的SQL运行
			if !processStateDbSqlOutside {
				writeDbName := getContractDbName(s.dbConfig, block.Header.ChainId, txWrite.ContractName)
				err = dbTx.ChangeContextDb(writeDbName)
				if err != nil {
					s.logger.Errorf("change context db to %s get an error:%s", writeDbName, err)
					return err
				}
				_, err = dbTx.ExecSql(string(txWrite.Value)) //运行用户自定义的建表语句
				if err != nil {
					s.logger.Errorf("execute sql[%s] get an error:%s", string(txWrite.Value), err)
					return err
				}
			}
		} else { //是KV数据，直接存储到StateInfo表
			stateInfo := NewStateInfo(txWrite.ContractName, txWrite.Key, txWrite.Value,
				uint64(block.Header.BlockHeight), block.GetTimestamp())
			writeDbName := getContractDbName(s.dbConfig, block.Header.ChainId, txWrite.ContractName)
			err = dbTx.ChangeContextDb(writeDbName)
			if err != nil {
				return err
			}
			s.logger.Debugf("try save state key[%s] to db[%s]", txWrite.Key, writeDbName)
			if err = saveStateInfo(dbTx, stateInfo); err != nil {
				s.logger.Errorf("save state key[%s] to db[%s] get error:%s", txWrite.Key, writeDbName, err.Error())
				return err
			}
		}
	}
	s.logger.Debugf("chain[%s]: save state block[%d]",
		block.Header.ChainId, block.Header.BlockHeight)
	return nil
}

func saveStateInfo(tx protocol.SqlDBTransaction, stateInfo *StateInfo) error {
	updateSql := `update state_infos set object_value=?,block_height=?
where contract_name=? and object_key=?`
	result, err := tx.ExecSql(updateSql, stateInfo.ObjectValue, stateInfo.BlockHeight, stateInfo.ContractName,
		stateInfo.ObjectKey)
	if result == 0 || err != nil {
		insertSql := `INSERT INTO state_infos (contract_name,object_key,object_value,block_height)
VALUES (?,?,?,?)`
		_, err = tx.ExecSql(insertSql, stateInfo.ContractName, stateInfo.ObjectKey, stateInfo.ObjectValue,
			stateInfo.BlockHeight)
	}
	return err
}

// ReadObject returns the state value for given contract name and key, or returns nil if none exists.
func (s *StateSqlDB) ReadObject(contractName string, key []byte) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	db := s.getContractDbHandle(contractName)
	sql := "select object_value from state_infos where contract_name=? and object_key=?"
	res, err := db.QuerySingle(sql, contractName, key)
	if err != nil {
		s.logger.Errorf("failed to read state, contract:%s, key:%s,error:%s", contractName, key, err)
		return nil, err
	}
	if res.IsEmpty() {
		s.logger.Debugf(" read empty state, contract:%s, key:%s", contractName, key)
		return nil, nil
	}
	var stateValue []byte

	err = res.ScanColumns(&stateValue)
	if err != nil {
		s.logger.Errorf("failed to read state, contract:%s, key:%s", contractName, key)
		return nil, err
	}
	//s.logger.Debugf(" read right state, contract:%s, key:%s valLen:%d", contractName, key, len(stateValue))
	return stateValue, nil
}

// SelectObject returns an iterator that contains all the key-values between given key ranges.
// startKey is included in the results and limit is excluded.
func (s *StateSqlDB) SelectObject(contractName string, startKey []byte, limit []byte) (protocol.StateIterator, error) {
	s.Lock()
	defer s.Unlock()
	db := s.getContractDbHandle(contractName)
	sql := "select * from state_infos where object_key >= ? and object_key  < ?"
	rows, err := db.QueryMulti(sql, startKey, limit)
	if err != nil {
		return nil, err
	}
	return newKVIterator(rows), nil
}

// GetLastSavepoint returns the last block height
func (s *StateSqlDB) GetLastSavepoint() (uint64, error) {
	s.Lock()
	defer s.Unlock()
	sql := "select block_height from save_points"
	row, err := s.db.QuerySingle(sql)
	if err != nil {
		return 0, err
	}
	var height *uint64
	err = row.ScanColumns(&height)
	if err != nil {
		return 0, err
	}
	if height == nil {
		return 0, nil
	}
	return *height, nil
}
func (s *StateSqlDB) getContractDbHandle(contractName string) protocol.SqlDBHandle {
	if handle, ok := s.contractDbs[contractName]; ok {
		s.logger.Debugf("reuse exist db handle for contract[%s],handle:%p", contractName, handle)
		return handle
	}
	if s.dbConfig.SqlDbType == "sqlite" { //sqlite is a file db, don't create multi connection.
		s.contractDbs[contractName] = s.db
		return s.db
	}
	dbName := getContractDbName(s.dbConfig, s.chainId, contractName)
	db := rawsqlprovider.NewSqlDBHandle(dbName, s.dbConfig, s.logger)
	s.contractDbs[contractName] = db
	s.logger.Infof("create new sql db handle[%p] database[%s] for contract[%s]", db, dbName, contractName)
	return db
}

// Close is used to close database, there is no need for gorm to close db
func (s *StateSqlDB) Close() {
	s.Lock()
	defer s.Unlock()
	s.logger.Info("close state sql db")
	s.db.Close()
	for contract, db := range s.contractDbs {
		s.logger.Infof("close state sql db for contract:%s", contract)
		db.Close()
	}
}

func (s *StateSqlDB) QuerySingle(contractName, sql string, values ...interface{}) (protocol.SqlRow, error) {
	s.Lock()
	defer s.Unlock()
	db := s.getContractDbHandle(contractName)

	row, err := db.QuerySingle(sql, values...)
	if err != nil {
		s.logger.Errorf("execute sql[%s] in statedb[%s] get an error:%s", sql, contractName, err)
		return nil, err
	}
	if row.IsEmpty() {
		s.logger.Infof("query single return empty row. sql:%s,db name:%s", sql, contractName)
	}
	return row, err
}
func (s *StateSqlDB) QueryMulti(contractName, sql string, values ...interface{}) (protocol.SqlRows, error) {
	s.Lock()
	defer s.Unlock()
	db := s.getContractDbHandle(contractName)

	return db.QueryMulti(sql, values...)

}
func (s *StateSqlDB) ExecDdlSql(contractName, sql, version string) error {
	s.Lock()
	defer s.Unlock()
	dbName := getContractDbName(s.dbConfig, s.chainId, contractName)
	exist, err := s.db.CreateDatabaseIfNotExist(dbName)
	if err != nil {
		return err
	}
	if !exist {
		if errTmp := s.initContractDb(contractName); errTmp != nil {
			return errTmp
		}
	}
	db := s.getContractDbHandle(contractName)
	// query ddl from db
	record := NewStateRecordSql(contractName, sql, protocol.SqlTypeDdl, version, 0)
	query, args := record.GetQueryStatusSql()
	s.logger.Debug("Query sql:", query, args)
	row, _ := s.db.QuerySingle(query, args)
	//查询数据库中是否有DDL记录，如果有对应记录，而且状态是1，那么就跳过重复执行DDL的情况
	if row != nil && !row.IsEmpty() {
		status := 0
		err = row.ScanColumns(&status)
		if err != nil {
			return err
		}
		if status == 1 { //SUCCESS
			s.logger.Infof("DDLRecord[%s] already executed, ignore it", sql)
			return nil
		}
	}
	insertSql, args2 := record.GetInsertSql()
	_, err = s.db.ExecSql(insertSql, args2...)
	if err != nil {
		s.logger.Warnf("DDLRecord[%s] save fail. error: %s", sql, err.Error())
	}
	//查询不到记录，或者查询出来后状态是失败，则执行DDL
	s.logger.Debugf("run DDL sql[%s] in db[%s]", sql, dbName)
	_, err = db.ExecSql(sql)

	record.Status = 1
	updateSql, args3 := record.GetUpdateSql()
	_, err2 := s.db.ExecSql(updateSql, args3...)
	if err2 != nil {
		s.logger.Warnf("DDLRecord[%s] update fail. error: %s", sql, err2.Error())
	}
	return err
}
func (s *StateSqlDB) BeginDbTransaction(txName string) (protocol.SqlDBTransaction, error) {
	s.Lock()
	defer s.Unlock()
	return s.db.BeginDbTransaction(txName)
}
func (s *StateSqlDB) GetDbTransaction(txName string) (protocol.SqlDBTransaction, error) {
	s.Lock()
	defer s.Unlock()
	return s.db.GetDbTransaction(txName)

}
func (s *StateSqlDB) CommitDbTransaction(txName string) error {
	s.Lock()
	defer s.Unlock()
	return s.db.CommitDbTransaction(txName)

}
func (s *StateSqlDB) RollbackDbTransaction(txName string) error {
	s.Lock()
	defer s.Unlock()
	s.logger.Warnw("rollback db transaction:", txName)
	return s.db.RollbackDbTransaction(txName)
}

func (s *StateSqlDB) updateConsensusArgs(dbTx protocol.SqlDBTransaction, block *commonPb.Block) error {
	//try to add consensusArgs
	consensusArgs, err := utils.GetConsensusArgsFromBlock(block)
	if err != nil {
		s.logger.Errorf("parse header.ConsensusArgs get an error:%s", err)
		return err
	}
	if consensusArgs.ConsensusData != nil {
		s.logger.Debugf("add consensusArgs ConsensusData to statedb")
		for _, write := range consensusArgs.ConsensusData.TxWrites {
			err = s.operateDbByWriteSet(dbTx, block, write, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *StateSqlDB) GetChainConfig() (*configPb.ChainConfig, error) {
	val, err := s.ReadObject(syscontract.SystemContract_CHAIN_CONFIG.String(),
		[]byte(syscontract.SystemContract_CHAIN_CONFIG.String()))
	if err != nil {
		return nil, err
	}
	conf := &configPb.ChainConfig{}
	err = conf.Unmarshal(val)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
