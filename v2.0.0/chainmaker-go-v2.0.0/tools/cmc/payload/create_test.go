/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package payload

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateConfigUpdatePayload(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cmc")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	createOutput = filepath.Join(tmpDir, "config_collect.pb")

	chainId = "chain1"
	contractName = "contract"
	method = "init"
	kvPairs = "tx_scheduler_timeout:15;tx_scheduler_validate_timeout:20"
	sequence = 8

	err = createConfigUpdatePayload()
	assert.NoError(t, err)
}

func TestContractMgmtPayload(t *testing.T) {
	//tmpDir, err := ioutil.TempDir("", "cmc")
	//assert.NoError(t, err)
	//defer os.RemoveAll(tmpDir)
	//
	//createOutput = filepath.Join(tmpDir, "contract_collect.pb")
	//
	//chainId = "chain1"
	//contractName = "contract"
	//version = "1.0.0"
	//runtime = "WASMER_RUST"
	//method = "init"
	//kvPairs = "tx_scheduler_timeout:15;tx_scheduler_validate_timeout:20"
	//byteCodePath = "../../../test/wasm/fact.wasm"
	//
	//err = createContractMgmtPayload()
	//assert.NoError(t, err)
}
