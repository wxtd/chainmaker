/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package chainconf record all the values of the chain config options.
package chainconf

import (
	"errors"
	"fmt"

	"chainmaker.org/chainmaker/pb-go/v2/syscontract"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/config"

	"chainmaker.org/chainmaker-go/logger"
	"chainmaker.org/chainmaker-go/utils"
	"chainmaker.org/chainmaker/common/v2/json"
	"chainmaker.org/chainmaker/protocol/v2"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/groupcache/lru"

	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

var _ protocol.ChainConf = (*ChainConf)(nil)
var log = logger.GetLogger(logger.MODULE_CHAINCONF)

const (
	AllContract = "ALL_CONTRACT"

	blockEmptyErrorTemplate = "block is empty"
)

var blockEmptyError = errors.New(blockEmptyErrorTemplate)

// ChainConf is the config of a chain.
type ChainConf struct {
	log *logger.CMLogger // logger

	options                       // extends options
	ChainConf *config.ChainConfig // chain config

	wLock      sync.RWMutex                             // lock
	watchers   map[string]protocol.Watcher              // config watchers, all watcher will be invoked when chain config changing.
	vmWatchers map[string]map[string]protocol.VmWatcher // contractName ==> module ==> VmWatcher

	lru       *lru.Cache
	configLru *lru.Cache
}

// NewChainConf create a new ChainConf instance.
func NewChainConf(opts ...Option) (*ChainConf, error) {
	chainConf := &ChainConf{
		watchers:   make(map[string]protocol.Watcher),
		vmWatchers: make(map[string]map[string]protocol.VmWatcher),
		lru:        lru.New(100),
		configLru:  lru.New(10),
	}
	if err := chainConf.Apply(opts...); err != nil {
		log.Errorw("NewChainConf apply is error", "err", err)
		return nil, err
	}
	chainConf.log = logger.GetLoggerByChain(logger.MODULE_CHAINCONF, chainConf.chainId)

	return chainConf, nil
}

// Genesis will create new genesis config block of chain.
func Genesis(genesisFile string) (*config.ChainConfig, error) {
	chainConfig := &config.ChainConfig{Contract: &config.ContractConfig{EnableSqlSupport: false}}
	fileInfo := map[string]interface{}{}
	v := viper.New()
	v.SetConfigFile(genesisFile)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&fileInfo); err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(fileInfo)
	if err != nil {
		return nil, err
	}
	log.Debugf("initial genesis config: %s", string(bytes))
	err = json.Unmarshal(bytes, chainConfig)
	if err != nil {
		return nil, err
	}

	// load the trust root certs than set the bytes as value
	// need verify org and root certs
	for _, root := range chainConfig.TrustRoots {
		for i := 0; i < len(root.Root); i++ {
			filePath := root.Root[i]
			if !filepath.IsAbs(filePath) {
				filePath, err = filepath.Abs(filePath)
				if err != nil {
					return nil, err
				}
			}
			log.Infof("load trust root file path: %s", filePath)
			entry, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("fail to read whiltlist file [%s]: %v", filePath, err)
			}
			root.Root[i] = string(entry)
		}
	}

	// load the trust member certs than set the bytes as value
	// need verify org
	trustMemberInfoMap := make(map[string]bool, len(chainConfig.TrustMembers))
	for _, member := range chainConfig.TrustMembers {
		filePath := member.MemberInfo
		if !filepath.IsAbs(filePath) {
			filePath, err = filepath.Abs(filePath)
			if err != nil {
				return nil, err
			}
		}
		log.Infof("load trust member file path: %s", filePath)
		entry, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("fail to read trust memberInfo file [%s]: %v", filePath, err)
		}
		if _, ok := trustMemberInfoMap[string(entry)]; ok {
			return nil, fmt.Errorf("the trust member info is exist, member info: %s", string(entry))
		}
		member.MemberInfo = string(entry)
		trustMemberInfoMap[string(entry)] = true
	}

	// verify
	_, err = VerifyChainConfig(chainConfig)
	if err != nil {
		return nil, err
	}

	return chainConfig, nil
}

// Init chain config.
func (c *ChainConf) Init() error {
	return c.latestChainConfig()
}

// latestChainConfig load latest chainConfig
func (c *ChainConf) latestChainConfig() error {
	// load chain config from store
	bytes, err := c.blockchainStore.ReadObject(syscontract.SystemContract_CHAIN_CONFIG.String(), []byte(syscontract.SystemContract_CHAIN_CONFIG.String()))
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return errors.New("chainConfig is empty")
	}
	var chainConfig config.ChainConfig
	err = proto.Unmarshal(bytes, &chainConfig)
	if err != nil {
		return err
	}

	c.ChainConf = &chainConfig

	// compatible with versions before v1.1.1
	if c.ChainConf.Contract == nil {
		c.ChainConf.Contract = &config.ContractConfig{EnableSqlSupport: false} //by default disable sql support
	}
	return nil
}

// GetChainConfigFromFuture get a future chain config.
func (c *ChainConf) GetChainConfigFromFuture(futureBlockHeight uint64) (*config.ChainConfig, error) {
	c.log.Debugf("GetChainConfig from futureBlockHeiht", "futureBlockHeight", futureBlockHeight)
	if futureBlockHeight > 0 {
		futureBlockHeight--
	}
	return GetChainConfigAt(c.log, c.lru, c.configLru, c.blockchainStore, futureBlockHeight)
}

// GetChainConfigAt get chain config with block height.
func (c *ChainConf) GetChainConfigAt(futureBlockHeight uint64) (*config.ChainConfig, error) {
	return GetChainConfigAt(c.log, c.lru, c.configLru, c.blockchainStore, futureBlockHeight)
}

// GetChainConfigAt get the lasted block info of chain config.
// The blockHeight must exist in store.
// If it is a config block , return the current config info.
func GetChainConfigAt(log protocol.Logger, lru *lru.Cache, configLru *lru.Cache, blockchainStore protocol.BlockchainStore, blockHeight uint64) (*config.ChainConfig, error) {
	var (
		block *common.Block
		err   error
	)
	block = getBlockInCache(lru, configLru, blockHeight)

	if block == nil {
		block, err = getBlockFromStore(blockchainStore, blockHeight)
		if err != nil {
			return nil, err
		}
	}

	if block == nil {
		log.Errorf("block is empty(height: %d)", blockHeight)
		return nil, blockEmptyError
	}
	if lru != nil {
		lru.Add(blockHeight, block)
	}

	if !utils.IsConfBlock(block) {
		block, err = getBlockFromStore(blockchainStore, block.Header.PreConfHeight)
		if err != nil {
			return nil, err
		}
		if block.Txs == nil {
			log.Errorf("block(height: %d) is not config block", block.Header.PreConfHeight)
			return nil, errors.New("block is not config block")
		}
	}
	if configLru != nil {
		configLru.Add(block.Header.BlockHeight, block)
	}

	txConfig := block.Txs[0]
	if txConfig.Result == nil || txConfig.Result.ContractResult == nil || txConfig.Result.ContractResult.Result == nil {
		log.Errorw("tx(id: %s) is not config tx", txConfig.Payload.TxId)
		return nil, errors.New("tx is not config tx")
	}
	result := txConfig.Result.ContractResult.Result
	chainConfig := &config.ChainConfig{}
	err = proto.Unmarshal(result, chainConfig)
	if err != nil {
		return nil, err
	}

	return chainConfig, nil
}

func getBlockInCache(lru *lru.Cache, configLru *lru.Cache, blockHeight uint64) *common.Block {
	var block *common.Block
	if configLru != nil {
		if value, ok := configLru.Get(blockHeight); ok {
			block = value.(*common.Block)
		}
	}
	if block == nil && lru != nil {
		if value, ok := lru.Get(blockHeight); ok {
			block = value.(*common.Block)
		}
	}
	return block
}

func getBlockFromStore(blockchainStore protocol.BlockchainStore, blockHeight uint64) (*common.Block, error) {
	var block *common.Block
	var err error
	block, err = blockchainStore.GetBlock(blockHeight)
	if err != nil {
		log.Errorf("get block(height: %d) from store failed, %s", blockHeight, err)
		return nil, err
	}
	return block, err
}

// ChainConfig return the chain config.
func (c *ChainConf) ChainConfig() *config.ChainConfig {
	return c.ChainConf
}

// GetConsensusNodeIdList return the node id list of all consensus node.
func (c *ChainConf) GetConsensusNodeIdList() ([]string, error) {
	chainNodeList := make([]string, 0)
	for _, node := range c.ChainConf.Consensus.Nodes {
		for _, nid := range node.NodeId {
			chainNodeList = append(chainNodeList, nid)
		}
	}
	c.log.Debugf("consensus node id list: %v", chainNodeList)
	return chainNodeList, nil
}

// BlockComplete complete the block. Invoke all config watchers.
func (c *ChainConf) CompleteBlock(block *common.Block) error {
	if block == nil {
		c.log.Error(blockEmptyErrorTemplate)
		return blockEmptyError
	}
	if block.Txs == nil || len(block.Txs) == 0 {
		return nil
	}
	tx := block.Txs[0]

	c.wLock.Lock()
	defer c.wLock.Unlock()

	if utils.IsValidConfigTx(tx) { // tx is chainConfig
		// watch chainConfig
		if err := c.callbackChainConfigWatcher(); err != nil {
			return err
		}
	}

	// watch native contract
	contract, ok := IsNativeTxSucc(tx)
	if ok {
		// is native tx
		// callback the watcher by sync
		payloadData, _ := tx.Payload.Marshal()
		if err := c.callbackContractVmWatcher(contract, payloadData); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChainConf) callbackChainConfigWatcher() error {
	err := c.latestChainConfig()
	if err != nil {
		return err
	}
	// callback the watcher by sync
	for m, w := range c.watchers {
		err = w.Watch(c.ChainConf)
		if err != nil {
			c.log.Errorw("chainConf notify err", "module", m, "err", err)
			return err
		}
	}
	return nil
}

func (c *ChainConf) callbackContractVmWatcher(contract string, requestPayload []byte) error {
	// watch the all contract
	if vmWatchers, ok := c.vmWatchers[AllContract]; ok {
		for m, w := range vmWatchers {
			err := w.Callback(contract, requestPayload)
			if err != nil {
				c.log.Errorf("vm watcher callback failed(contract: %s, module: %s), %s", contract, m, err)
				return err
			}
		}
	}

	// watch some contract
	if vmWatchers, ok := c.vmWatchers[contract]; ok {
		for m, w := range vmWatchers {
			err := w.Callback(contract, requestPayload)
			if err != nil {
				c.log.Errorf("vm watcher callback failed(contract: %s, module: %s), %s", contract, m, err)
				return err
			}
		}
	}
	return nil
}

// AddWatch register a config watcher.
func (c *ChainConf) AddWatch(w protocol.Watcher) {
	c.wLock.Lock()
	defer c.wLock.Unlock()
	c.watchers[w.Module()] = w
}

// AddVmWatch add vm watcher
func (c *ChainConf) AddVmWatch(w protocol.VmWatcher) {
	c.wLock.Lock()
	defer c.wLock.Unlock()
	if w != nil {
		contractNames := w.ContractNames()
		if contractNames == nil {
			// watch all contract
			c.addVmWatcherWithAllContract(w)
		} else {
			c.addVmWatcherWithContracts(w)
		}
	}
}

func (c *ChainConf) addVmWatcherWithAllContract(w protocol.VmWatcher) {
	watchers, ok := c.vmWatchers[AllContract]
	if !ok {
		watchers = make(map[string]protocol.VmWatcher)
	}
	if _, ok := watchers[w.Module()]; ok {
		c.log.Errorf("vm watcher existed(contract: %s, module: %s)", AllContract, w.Module())
		return
	}
	watchers[w.Module()] = w
	c.vmWatchers[AllContract] = watchers
}

func (c *ChainConf) addVmWatcherWithContracts(w protocol.VmWatcher) {
	for _, contractName := range w.ContractNames() {
		watchers, ok := c.vmWatchers[contractName]
		if !ok {
			watchers = make(map[string]protocol.VmWatcher)
		} else if _, ok := watchers[w.Module()]; ok {
			c.log.Errorf("vm watcher existed(contract: %s, module: %s)", contractName, w.Module())
			return
		}
		watchers[w.Module()] = w
		c.vmWatchers[contractName] = watchers
	}
}
