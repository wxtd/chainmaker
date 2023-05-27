/*
 * Copyright (C) BABEC. All rights reserved.
 * Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package multisign

import (
	"fmt"

	"chainmaker.org/chainmaker-go/vm/native/common"

	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

var (
	KEY_ContractMgmtPayload   = "ContractMgmtPayload"
	KEY_SystemContractPayload = "SystemContractPayload"
)

// MultiSignContract multiSign Contract
type MultiSignContract struct {
	methods map[string]common.ContractFunc
	log     protocol.Logger
}

func NewMultiSignContract(log protocol.Logger) *MultiSignContract {
	return &MultiSignContract{
		log:     log,
		methods: registerMultiSignContractMethods(log),
	}
}

func (c *MultiSignContract) GetMethod(methodName string) common.ContractFunc {
	return c.methods[methodName]
}

func registerMultiSignContractMethods(log protocol.Logger) map[string]common.ContractFunc {
	methodMap := make(map[string]common.ContractFunc, 64)

	return methodMap
}

// MultiSignRuntime  mutlSign runtime
type MultiSignRuntime struct {
	log protocol.Logger
}

// payloadInfo the memory payload info
type payloadInfo struct {
	txType      commonPb.TxType
	payload     interface{}
	payloadType string
}

// parsePayload unmarshal bytes
func parsePayload(txType string, payloadBytes []byte) (*payloadInfo, error) {
	switch txType {
	//case commonPb.TxType_MANAGE_USER_CONTRACT.String():
	//	txType1 := commonPb.TxType(commonPb.TxType_value[txType])
	//	payload := new(commonPb.Payload)
	//	err := proto.Unmarshal(payloadBytes, payload)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return &payloadInfo{
	//		txType:      txType1,
	//		payload:     payload,
	//		payloadType: KEY_ContractMgmtPayload,
	//	}, nil
	//case commonPb.TxType_INVOKE_CONTRACT.String(), commonPb.TxType_INVOKE_CONTRACT.String():
	//	txType1 := commonPb.TxType(commonPb.TxType_value[txType])
	//	payload := new(commonPb.Payload)
	//	err := proto.Unmarshal(payloadBytes, payload)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return &payloadInfo{
	//		txType:      txType1,
	//		payload:     payload,
	//		payloadType: KEY_SystemContractPayload,
	//	}, nil
	default:
		return nil, fmt.Errorf("no support the tx_type, tx_type = %s", txType)
	}
}
