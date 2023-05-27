/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package native_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"chainmaker.org/chainmaker/pb-go/v2/syscontract"

	native "chainmaker.org/chainmaker-go/test/chainconfig_test"
	apiPb "chainmaker.org/chainmaker/pb-go/v2/api"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 查询区块
func TestGetBlockByHeight(t *testing.T) {
	conn, err := native.InitGRPCConnect(isTls)
	if err != nil {
		panic(err)
	}
	client := apiPb.NewRpcNodeClient(conn)

	fmt.Println("============ get block by height============")
	// 构造Payload
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "blockHeight",
			Value: []byte("0"),
		},
		{
			Key:   "withRWSet",
			Value: []byte("false"),
		},
	}

	sk, member := native.GetUserSK(1)
	resp, err := native.QueryRequest(sk, member, &client, &native.InvokeContractMsg{TxType: commonPb.TxType_QUERY_CONTRACT,
		ChainId: CHAIN1, ContractName: syscontract.SystemContract_CHAIN_QUERY.String(), MethodName: syscontract.ChainQueryFunction_GET_BLOCK_BY_HEIGHT.String(), Pairs: pairs})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("WARN: client.call err: deadline")

			}
		}

		fmt.Printf("ERROR: client.call err: %v\n", err)
		return
	}
	fmt.Printf("response: %v\n", resp)
	//result := &commonPb.CertInfos{}
	//err = proto.Unmarshal(resp.ContractResult.Result, result)
	//fmt.Printf("send tx resp: code:%d, msg:%s, CertInfos:%+v\n", resp.Code, resp.Message, result)

	blockInfo := &commonPb.BlockInfo{}
	if err = proto.Unmarshal(resp.ContractResult.Result, blockInfo); err != nil {
		panic(err)
	}
	fmt.Println(blockInfo)

	blockHash := blockInfo.Block.Header.BlockHash
	fmt.Println("blockHash", string(blockHash), hex.EncodeToString(blockHash))
}

// 查询区块
func TestGetBlockByHash(t *testing.T) {
	conn, err := native.InitGRPCConnect(isTls)
	if err != nil {
		panic(err)
	}
	client := apiPb.NewRpcNodeClient(conn)

	fmt.Println("============ get block by height============")
	// 构造Payload
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "blockHash",
			Value: []byte("54d54331b4a341353c19b82ec7ad4a6f15b78c9cc4ba8caa84759d1805f4ad1f"),
		},
		{
			Key:   "withRWSet",
			Value: []byte("false"),
		},
	}

	sk, member := native.GetUserSK(1)
	resp, err := native.QueryRequest(sk, member, &client, &native.InvokeContractMsg{TxType: commonPb.TxType_QUERY_CONTRACT,
		ChainId: CHAIN1, ContractName: syscontract.SystemContract_CHAIN_QUERY.String(), MethodName: syscontract.ChainQueryFunction_GET_BLOCK_BY_HASH.String(), Pairs: pairs})

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("WARN: client.call err: deadline")

			}
		}

		fmt.Printf("ERROR: client.call err: %v\n", err)
		return
	}
	fmt.Printf("response: %v\n", resp)

	blockInfo := &commonPb.BlockInfo{}
	if err = proto.Unmarshal(resp.ContractResult.Result, blockInfo); err != nil {
		panic(err)
	}

	blockHash := blockInfo.Block.Header.BlockHash
	fmt.Println("blockHash", string(blockHash), hex.EncodeToString(blockHash))
}

// 查询交易
func TestGetTxById(t *testing.T) {
	conn, err := native.InitGRPCConnect(isTls)
	if err != nil {
		panic(err)
	}
	client := apiPb.NewRpcNodeClient(conn)

	fmt.Println("============ get tx by txId============")
	// 构造Payload
	pairs := []*commonPb.KeyValuePair{
		{
			Key:   "txId",
			Value: []byte("c6b7033daf96441aab83f33e1abe6706543a410e7158405a90e5cfb02aa50660"),
		},
	}

	sk, member := native.GetUserSK(1)
	resp, err := native.QueryRequest(sk, member, &client, &native.InvokeContractMsg{TxType: commonPb.TxType_QUERY_CONTRACT,
		ChainId: CHAIN1, ContractName: syscontract.SystemContract_CHAIN_QUERY.String(), MethodName: syscontract.ChainQueryFunction_GET_TX_BY_TX_ID.String(), Pairs: pairs})
	if resp.Code != 0 {
		fmt.Println(resp.Message)
		return
	}
	tx := &commonPb.TransactionInfo{}
	if err = proto.Unmarshal(resp.ContractResult.Result, tx); err != nil {
		panic(err)
	}
	//fmt.Println(tx)
	data, _ := json.MarshalIndent(tx, "", "  ")
	fmt.Printf("%s", data)
}
