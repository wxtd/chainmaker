/*
 * Copyright 2020 The SealEVM Authors
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package storage

import (
	"chainmaker.org/chainmaker-go/evm/evm-go/environment"
	"chainmaker.org/chainmaker/common/v2/evmutils"
)

//Teh External Storage,provding a Storage for touching out of current evm
type IExternalStorage interface {
	GetBalance(address *evmutils.Int) (*evmutils.Int, error)
	GetCode(address *evmutils.Int) ([]byte, error)
	GetCodeSize(address *evmutils.Int) (*evmutils.Int, error)
	GetCodeHash(address *evmutils.Int) (*evmutils.Int, error)
	GetBlockHash(block *evmutils.Int) (*evmutils.Int, error)

	CreateAddress(caller *evmutils.Int, tx environment.Transaction) *evmutils.Int
	CreateFixedAddress(caller *evmutils.Int, salt *evmutils.Int, tx environment.Transaction) *evmutils.Int

	CanTransfer(from *evmutils.Int, to *evmutils.Int, amount *evmutils.Int) bool

	Load(n string, k string) (*evmutils.Int, error)
	Store(address string, key string, val []byte)
}
