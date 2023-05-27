/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package single

import (
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

type mempoolTxs struct {
	isConfigTxs bool
	txs         []*commonPb.Transaction
	source      protocol.TxSource
}

type valInPendingCache struct {
	inBlockHeight uint64
	tx            *commonPb.Transaction
}
