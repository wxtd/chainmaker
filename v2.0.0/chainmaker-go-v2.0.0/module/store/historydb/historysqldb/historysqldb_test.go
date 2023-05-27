/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package historysqldb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/syscontract"

	"chainmaker.org/chainmaker-go/localconf"
	"chainmaker.org/chainmaker-go/store/dbprovider/rawsqlprovider"
	"chainmaker.org/chainmaker-go/store/historydb"
	"chainmaker.org/chainmaker-go/store/serialization"
	acPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	storePb "chainmaker.org/chainmaker/pb-go/v2/store"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/test"
	"github.com/stretchr/testify/assert"
)

var log = &test.GoLogger{}

func generateBlockHash(chainId string, height uint64) []byte {
	blockHash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", chainId, height)))
	return blockHash[:]
}

func generateTxId(chainId string, height uint64, index int) string {
	txIdBytes := sha256.Sum256([]byte(fmt.Sprintf("%s-%d-%d", chainId, height, index)))
	return hex.EncodeToString(txIdBytes[:32])
}

func createConfigBlock(chainId string, height uint64) *storePb.BlockWithRWSet {
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:     chainId,
			BlockHeight: height,
			Proposer: &acPb.Member{
				OrgId:      "org1",
				MemberInfo: []byte("User1"),
			},
		},
		Txs: []*commonPb.Transaction{
			{
				Payload: &commonPb.Payload{
					ChainId:      chainId,
					TxType:       commonPb.TxType_INVOKE_CONTRACT,
					ContractName: syscontract.SystemContract_CHAIN_CONFIG.String(),
				},
				Sender: &commonPb.EndorsementEntry{
					Signer: &acPb.Member{
						OrgId:      "org1",
						MemberInfo: []byte("Admin"),
					},
					Signature: []byte("signature1"),
				},
				Result: &commonPb.Result{
					Code: commonPb.TxStatusCode_SUCCESS,
					ContractResult: &commonPb.ContractResult{
						Result: []byte("ok"),
					},
				},
			},
		},
	}

	block.Header.BlockHash = generateBlockHash(chainId, height)
	block.Txs[0].Payload.TxId = generateTxId(chainId, height, 0)
	return &storePb.BlockWithRWSet{
		Block:    block,
		TxRWSets: []*commonPb.TxRWSet{},
	}
}

func createBlockAndRWSets(chainId string, height uint64, txNum int) *storePb.BlockWithRWSet {
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:     chainId,
			BlockHeight: height,
			Proposer: &acPb.Member{
				OrgId:      "org1",
				MemberInfo: []byte("User1"),
				MemberType: 0,
			},
		},
	}

	for i := 0; i < txNum; i++ {

		tx := &commonPb.Transaction{
			Payload: &commonPb.Payload{
				ChainId:      chainId,
				TxId:         generateTxId(chainId, height, i),
				TxType:       commonPb.TxType_INVOKE_CONTRACT,
				ContractName: "contract1",
				Method:       "Function1",
				Parameters:   nil,
			},
			Sender: &commonPb.EndorsementEntry{
				Signer: &acPb.Member{
					OrgId:      "org1",
					MemberInfo: []byte("User" + strconv.Itoa(i)),
				},
				Signature: []byte("signature1"),
			},
			Result: &commonPb.Result{
				Code: commonPb.TxStatusCode_SUCCESS,
				ContractResult: &commonPb.ContractResult{
					Result: []byte("ok"),
				},
			},
		}
		block.Txs = append(block.Txs, tx)
	}

	block.Header.BlockHash = generateBlockHash(chainId, height)
	var txRWSets []*commonPb.TxRWSet
	for i := 0; i < txNum; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		txRWset := &commonPb.TxRWSet{
			TxId: block.Txs[i].Payload.TxId,
			TxWrites: []*commonPb.TxWrite{
				{
					Key:          []byte(key),
					Value:        []byte(value),
					ContractName: "contract1",
				},
			},
		}
		txRWSets = append(txRWSets, txRWset)
	}

	return &storePb.BlockWithRWSet{
		Block:    block,
		TxRWSets: txRWSets,
	}
}

var testChainId = "testchainid_1"
var block0 = createConfigBlock(testChainId, 0)
var block1 = createBlockAndRWSets(testChainId, 1, 10)
var block2 = createBlockAndRWSets(testChainId, 2, 2)

/*var block3, _ = createBlockAndRWSets(testChainId, 3, 2)
var configBlock4 = createConfigBlock(testChainId, 4)
var block5, _ = createBlockAndRWSets(testChainId, 5, 3)*/

func createBlock(chainId string, height uint64) *commonPb.Block {
	block := &commonPb.Block{
		Header: &commonPb.BlockHeader{
			ChainId:     chainId,
			BlockHeight: height,
			Proposer: &acPb.Member{
				OrgId:      "org1",
				MemberInfo: []byte("User1"),
				MemberType: 0,
			},
		},
		Txs: []*commonPb.Transaction{
			{
				Payload: &commonPb.Payload{
					ChainId: chainId,
				},
				Sender: &commonPb.EndorsementEntry{
					Signer: &acPb.Member{
						OrgId:      "org1",
						MemberInfo: []byte("User1"),
					},
					Signature: []byte("signature1"),
				},
				Result: &commonPb.Result{
					Code: commonPb.TxStatusCode_SUCCESS,
					ContractResult: &commonPb.ContractResult{
						Result: []byte("ok"),
					},
				},
			},
		},
	}

	block.Header.BlockHash = generateBlockHash(chainId, height)
	block.Txs[0].Payload.TxId = generateTxId(chainId, height, 0)
	return block
}

func initProvider() protocol.SqlDBHandle {
	conf := &localconf.SqlDbConfig{}
	conf.Dsn = ":memory:"
	conf.SqlDbType = "sqlite"
	conf.SqlLogMode = "Info"
	p := rawsqlprovider.NewSqlDBHandle("chain1", conf, log)
	return p
}

//初始化DB并同时初始化创世区块
func initSqlDb() *HistorySqlDB {
	db, _ := newHistorySqlDB(testChainId, initProvider(), log)
	_, blockInfo, _ := serialization.SerializeBlock(block0)
	db.InitGenesis(blockInfo)
	return db
}

func TestHistorySqlDB_CommitBlock(t *testing.T) {
	db := initSqlDb()
	block1.TxRWSets[0].TxWrites[0].Value = nil
	_, blockInfo, err := serialization.SerializeBlock(block1)
	assert.Nil(t, err)
	err = db.CommitBlock(blockInfo)
	assert.Nil(t, err)
}

func TestHistorySqlDB_GetLastSavepoint(t *testing.T) {
	db := initSqlDb()
	_, block1, err := serialization.SerializeBlock(block1)
	assert.Nil(t, err)
	err = db.CommitBlock(block1)
	assert.Nil(t, err)
	height, err := db.GetLastSavepoint()
	assert.Nil(t, err)
	assert.Equal(t, uint64(block1.Block.Header.BlockHeight), height)

	_, block2, err := serialization.SerializeBlock(block2)
	assert.Nil(t, err)
	err = db.CommitBlock(block2)
	assert.Nil(t, err)
	height, err = db.GetLastSavepoint()
	assert.Nil(t, err)
	assert.Equal(t, uint64(block2.Block.Header.BlockHeight), height)
}
func TestHistorySqlDB_GetHistoryForKey(t *testing.T) {
	db := initSqlDb()
	block1.TxRWSets[0].TxWrites[0].Value = nil
	_, blockInfo, err := serialization.SerializeBlock(block1)
	assert.Nil(t, err)
	err = db.CommitBlock(blockInfo)
	assert.Nil(t, err)
	result, err := db.GetHistoryForKey("contract1", []byte("key_1"))
	assert.Nil(t, err)

	assert.Equal(t, 1, getCount(result))

}
func getCount(i historydb.HistoryIterator) int {
	count := 0
	for i.Next() {
		count++
	}
	return count
}
func TestHistorySqlDB_GetAccountTxHistory(t *testing.T) {
	db := initSqlDb()
	block1.TxRWSets[0].TxWrites[0].Value = nil
	_, blockInfo, err := serialization.SerializeBlock(block1)
	assert.Nil(t, err)
	err = db.CommitBlock(blockInfo)
	assert.Nil(t, err)
	result, err := db.GetAccountTxHistory([]byte("User1"))
	assert.Nil(t, err)
	assert.Equal(t, 1, getCount(result))
	for result.Next() {
		v, _ := result.Value()
		t.Logf("%#v", v)
	}
}
func TestHistorySqlDB_GetContractTxHistory(t *testing.T) {
	db := initSqlDb()
	block1.TxRWSets[0].TxWrites[0].Value = nil
	_, blockInfo, err := serialization.SerializeBlock(block1)
	err = db.CommitBlock(blockInfo)
	assert.Nil(t, err)
	result, err := db.GetContractTxHistory("contract1")
	assert.Nil(t, err)
	assert.Equal(t, 10, getCount(result))
	for result.Next() {
		v, _ := result.Value()
		t.Logf("%#v", v)
	}
}
