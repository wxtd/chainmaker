/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
   SPDX-License-Identifier: Apache-2.0
*/

package scheduler

import (
	"errors"
	"fmt"
	"sort"

	acpb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

// Storage interface for smart contracts
type txSimContextImpl struct {
	txExecSeq        int
	txResult         *commonpb.Result
	txRWSet          *commonpb.TxRWSet
	tx               *commonpb.Transaction
	txReadKeyMap     map[string]*commonpb.TxRead
	txWriteKeyMap    map[string]*commonpb.TxWrite
	txWriteKeySql    []*commonpb.TxWrite
	txWriteKeyDdlSql []*commonpb.TxWrite // record ddl vm run success or failure
	snapshot         protocol.Snapshot
	vmManager        protocol.VmManager
	gasUsed          uint64 // only for callContract
	currentDepth     int
	currentResult    []byte
	hisResult        []*callContractResult
	sqlRowCache      map[int32]protocol.SqlRows
	kvRowCache       map[int32]protocol.StateIterator
	blockVersion     uint32
}

type callContractResult struct {
	contractName string
	method       string
	param        map[string][]byte
	depth        int
	gasUsed      uint64
	result       []byte
}

func (s *txSimContextImpl) Get(contractName string, key []byte) ([]byte, error) {
	// Get from write set
	if value, done := s.getFromWriteSet(contractName, key); done {
		s.putIntoReadSet(contractName, key, value)
		return value, nil
	}

	// Get from read set
	if value, done := s.getFromReadSet(contractName, key); done {
		return value, nil
	}

	// Get from db
	if value, err := s.snapshot.GetKey(s.txExecSeq, contractName, key); err != nil {
		return nil, err
	} else {
		// if get from db success, put into read set
		s.putIntoReadSet(contractName, key, value)
		return value, nil
	}
}
func (s *txSimContextImpl) Put(contractName string, key []byte, value []byte) error {
	s.putIntoWriteSet(contractName, key, value)
	return nil
}

func (s *txSimContextImpl) PutRecord(contractName string, value []byte, sqlType protocol.SqlType) {
	txWrite := &commonpb.TxWrite{
		Key:          nil,
		Value:        value,
		ContractName: contractName,
	}
	s.txWriteKeySql = append(s.txWriteKeySql, txWrite)
	if sqlType == protocol.SqlTypeDdl {
		s.txWriteKeyDdlSql = append(s.txWriteKeyDdlSql, txWrite)
	}
}

func (s *txSimContextImpl) Del(contractName string, key []byte) error {
	s.putIntoWriteSet(contractName, key, nil)
	return nil
}

func (s *txSimContextImpl) Select(contractName string, startKey []byte, limit []byte) (protocol.StateIterator, error) {
	// 将来需要把txRwSet的最新状态填充到Iter中去，覆盖或者替换，才是完整的最新的Iter，否则就只是数据库的状态
	return s.snapshot.GetBlockchainStore().SelectObject(contractName, startKey, limit)
}

func (s *txSimContextImpl) GetCreator(contractName string) *acpb.Member {
	contract, err := s.GetContractByName(contractName)
	if err != nil {
		//TODO log
		return nil
	}
	return contract.Creator
}

func (s *txSimContextImpl) GetSender() *acpb.Member {
	return s.tx.Sender.GetSigner()
}

func (s *txSimContextImpl) putIntoReadSet(contractName string, key []byte, value []byte) {
	s.txReadKeyMap[constructKey(contractName, key)] = &commonpb.TxRead{
		Key:          key,
		Value:        value,
		ContractName: contractName,
		Version:      nil,
	}
}

func (s *txSimContextImpl) putIntoWriteSet(contractName string, key []byte, value []byte) {
	s.txWriteKeyMap[constructKey(contractName, key)] = &commonpb.TxWrite{
		Key:          key,
		Value:        value,
		ContractName: contractName,
	}
}

func (s *txSimContextImpl) getFromReadSet(contractName string, key []byte) ([]byte, bool) {
	if txRead, ok := s.txReadKeyMap[constructKey(contractName, key)]; ok {
		return txRead.Value, true
	}
	return nil, false
}

func (s *txSimContextImpl) getFromWriteSet(contractName string, key []byte) ([]byte, bool) {
	if txWrite, ok := s.txWriteKeyMap[constructKey(contractName, key)]; ok {
		return txWrite.Value, true
	}
	return nil, false
}

// Get the corresponding transaction
func (s *txSimContextImpl) GetTx() *commonpb.Transaction {
	return s.tx
}

// Get blockchain storage
func (s *txSimContextImpl) GetBlockchainStore() protocol.BlockchainStore {
	return s.snapshot.GetBlockchainStore()
}

// GetAccessControl get access control service
func (s *txSimContextImpl) GetAccessControl() (protocol.AccessControlProvider, error) {
	if s.vmManager.GetAccessControl() == nil {
		return nil, errors.New("access control for tx sim context is nil")
	}
	return s.vmManager.GetAccessControl(), nil
}

// Get organization service
func (s *txSimContextImpl) GetChainNodesInfoProvider() (protocol.ChainNodesInfoProvider, error) {
	if s.vmManager.GetChainNodesInfoProvider() == nil {
		return nil, errors.New("chainNodesInfoProvider for tx sim context is nil, may be running in singleton mode")
	}
	return s.vmManager.GetChainNodesInfoProvider(), nil
}

// GetTxRWSet return current transaction read write set
func (s *txSimContextImpl) GetTxRWSet(runVmSuccess bool) *commonpb.TxRWSet {
	if s.txRWSet != nil {
		return s.txRWSet
	}
	s.txRWSet = &commonpb.TxRWSet{
		TxId:     s.tx.Payload.TxId,
		TxReads:  nil,
		TxWrites: nil,
	}

	// read set
	{
		txIds := make([]string, 0, len(s.txReadKeyMap))
		for txId := range s.txReadKeyMap {
			txIds = append(txIds, txId)
		}
		sort.Strings(txIds)
		for _, k := range txIds {
			s.txRWSet.TxReads = append(s.txRWSet.TxReads, s.txReadKeyMap[k])
		}
	}

	// write set
	if runVmSuccess {
		txIds := make([]string, 0, len(s.txWriteKeyMap))
		for txId := range s.txWriteKeyMap {
			txIds = append(txIds, txId)
		}
		sort.Strings(txIds)
		for _, k := range txIds {
			s.txRWSet.TxWrites = append(s.txRWSet.TxWrites, s.txWriteKeyMap[k])
		}
		// sql nil key tx writes
		s.txRWSet.TxWrites = append(s.txRWSet.TxWrites, s.txWriteKeySql...)
	} else {
		// ddl sql tx writes
		s.txRWSet.TxWrites = s.txWriteKeyDdlSql
	}
	return s.txRWSet
}

// Get the height of the corresponding block
func (s *txSimContextImpl) GetBlockHeight() uint64 {
	return s.snapshot.GetBlockHeight()
}

func (s *txSimContextImpl) GetBlockProposer() *acpb.Member {
	return s.snapshot.GetBlockProposer()
}

// Obtain the corresponding transaction execution sequence
func (s *txSimContextImpl) GetTxExecSeq() int {
	return s.txExecSeq
}

// set the corresponding transaction execution sequence
func (s *txSimContextImpl) SetTxExecSeq(txExecSeq int) {
	s.txExecSeq = txExecSeq
}

// Get the tx result
func (s *txSimContextImpl) GetTxResult() *commonpb.Result {
	return s.txResult
}

// Set the tx result
func (s *txSimContextImpl) SetTxResult(txResult *commonpb.Result) {
	s.txResult = txResult
}

// Cross contract call
func (s *txSimContextImpl) CallContract(contract *commonpb.Contract, method string, byteCode []byte,
	parameter map[string][]byte, gasUsed uint64, refTxType commonpb.TxType) (
	*commonpb.ContractResult, commonpb.TxStatusCode) {
	s.gasUsed = gasUsed
	s.currentDepth = s.currentDepth + 1
	if s.currentDepth > protocol.CallContractDepth {
		contractResult := &commonpb.ContractResult{
			Code:    uint32(1),
			Result:  nil,
			Message: fmt.Sprintf("CallContract too depth %d", s.currentDepth),
		}
		return contractResult, commonpb.TxStatusCode_CONTRACT_TOO_DEEP_FAILED
	}
	if s.gasUsed > protocol.GasLimit {
		contractResult := &commonpb.ContractResult{
			Code:    uint32(1),
			Result:  nil,
			Message: fmt.Sprintf("There is not enough gas, gasUsed %d GasLimit %d ", gasUsed, int64(protocol.GasLimit)),
		}
		return contractResult, commonpb.TxStatusCode_CONTRACT_FAIL
	}
	if len(byteCode) == 0 {
		dbByteCode, err := s.GetContractBytecode(contract.Name)
		if err != nil {
			return nil, commonpb.TxStatusCode_CONTRACT_FAIL
		}
		byteCode = dbByteCode
	}
	r, code := s.vmManager.RunContract(contract, method, byteCode, parameter, s, s.gasUsed, refTxType)

	result := callContractResult{
		depth:        s.currentDepth,
		gasUsed:      s.gasUsed,
		result:       r.Result,
		contractName: contract.Name,
		method:       method,
		param:        parameter,
	}
	s.hisResult = append(s.hisResult, &result)
	s.currentResult = r.Result
	s.currentDepth = s.currentDepth - 1
	return r, code
}

// Obtain the execution result of current contract (cross contract)
func (s *txSimContextImpl) GetCurrentResult() []byte {
	return s.currentResult
}

// Get contract call depth
func (s *txSimContextImpl) GetDepth() int {
	return s.currentDepth
}

func constructKey(contractName string, key []byte) string {
	return contractName + string(key)
}

func (s *txSimContextImpl) SetStateSqlHandle(index int32, rows protocol.SqlRows) {
	// 当前交易总是串行执行，故不需要加锁
	s.sqlRowCache[index] = rows
}

func (s *txSimContextImpl) GetStateSqlHandle(index int32) (protocol.SqlRows, bool) {
	data, ok := s.sqlRowCache[index]
	return data, ok
}

func (s *txSimContextImpl) SetStateKvHandle(index int32, rows protocol.StateIterator) {
	// 当前交易总是串行执行，故不需要加锁
	s.kvRowCache[index] = rows
}

func (s *txSimContextImpl) GetStateKvHandle(index int32) (protocol.StateIterator, bool) {
	data, ok := s.kvRowCache[index]
	return data, ok
}
func (s *txSimContextImpl) GetBlockVersion() uint32 {
	return s.blockVersion
}

func (s *txSimContextImpl) GetContractByName(name string) (*commonpb.Contract, error) {
	return s.snapshot.GetBlockchainStore().GetContractByName(name)
}

//GetContractBytecode get contract bytecode
func (s *txSimContextImpl) GetContractBytecode(name string) ([]byte, error) {
	return s.snapshot.GetBlockchainStore().GetContractBytecode(name)
}
