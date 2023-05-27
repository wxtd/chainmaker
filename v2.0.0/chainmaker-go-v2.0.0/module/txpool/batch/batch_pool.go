/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package batch

import (
	"sync"

	"chainmaker.org/chainmaker/common/v2/sortedmap"
	txpoolPb "chainmaker.org/chainmaker/pb-go/v2/txpool"
)

type nodeBatchPool struct {
	pool *sortedmap.IntKeySortedMap
}

func newNodeBatchPool() *nodeBatchPool {
	return &nodeBatchPool{pool: sortedmap.NewIntKeySortedMap()}
}

func (p *nodeBatchPool) PutIfNotExist(batch *txpoolPb.TxBatch) bool {
	batchId := int(batch.BatchId)
	ok := p.pool.Contains(batchId)
	if ok {
		return false
	}
	p.pool.Put(batchId, batch)
	return true
}

func (p *nodeBatchPool) RemoveIfExist(batch *txpoolPb.TxBatch) bool {
	batchId := int(batch.BatchId)
	_, ok := p.pool.Remove(batchId)
	return ok
}

func (p *nodeBatchPool) currentSize() int {
	return p.pool.Length()
}

func (p *nodeBatchPool) GetBatch(batchId int32) *txpoolPb.TxBatch {
	if val, ok := p.pool.Get(int(batchId)); ok {
		return val.(*txpoolPb.TxBatch)
	}
	return nil
}

type pendingBatchPool struct {
	l    sync.RWMutex
	pool map[int32]*txpoolPb.TxBatch
}

func newPendingBatchPool() *pendingBatchPool {
	return &pendingBatchPool{pool: make(map[int32]*txpoolPb.TxBatch)}
}

func (p *pendingBatchPool) PutIfNotExist(batch *txpoolPb.TxBatch) bool {
	p.l.Lock()
	defer p.l.Unlock()
	batchId := batch.BatchId
	_, ok := p.pool[batchId]
	if !ok {
		p.pool[batchId] = batch
		return true
	}
	return false
}

func (p *pendingBatchPool) RemoveIfExist(batch *txpoolPb.TxBatch) bool {
	p.l.Lock()
	defer p.l.Unlock()
	batchId := batch.BatchId
	_, ok := p.pool[batchId]
	if ok {
		delete(p.pool, batchId)
		return true
	}
	return false
}

func (p *pendingBatchPool) GetBatch(batchId int32) *txpoolPb.TxBatch {
	p.l.RLock()
	defer p.l.RUnlock()
	if val, ok := p.pool[batchId]; ok {
		return val
	}
	return nil
}

func (p *pendingBatchPool) Range(f func(batch *txpoolPb.TxBatch) (isContinue bool)) {
	p.l.RLock()
	defer p.l.RUnlock()
	for _, batch := range p.pool {
		if !f(batch) {
			break
		}
	}
}

func (p *pendingBatchPool) currentSize() int {
	p.l.RLock()
	defer p.l.RUnlock()
	return len(p.pool)
}
