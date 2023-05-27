/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package native

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
)

func IsNative(contractName string, txType commonPb.TxType) bool {
	return IsNativeContract(contractName) && IsNativeTxType(txType)
}

// IsNativeContract return is native contract name
func IsNativeContract(contractName string) bool {
	_, ok := syscontract.SystemContract_value[contractName]
	return ok
}

//TODO: Devin: Remove it
// IsNativeTxType return is native contract supported transaction type
func IsNativeTxType(txType commonPb.TxType) bool {
	switch txType {
	case commonPb.TxType_QUERY_CONTRACT,
		commonPb.TxType_INVOKE_CONTRACT:
		//commonPb.TxType_INVOKE_CONTRACT:
		return true
	default:
		return false
	}
}
