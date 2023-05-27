/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"chainmaker.org/chainmaker-go/accesscontrol"
	"chainmaker.org/chainmaker-go/chainconf"
	"chainmaker.org/chainmaker-go/consensus"
	"chainmaker.org/chainmaker-go/consensus/dpos"
	"chainmaker.org/chainmaker-go/core"
	"chainmaker.org/chainmaker-go/core/cache"
	providerConf "chainmaker.org/chainmaker-go/core/provider/conf"
	"chainmaker.org/chainmaker-go/localconf"
	"chainmaker.org/chainmaker-go/logger"
	"chainmaker.org/chainmaker-go/net"
	"chainmaker.org/chainmaker-go/snapshot"
	"chainmaker.org/chainmaker-go/store"
	"chainmaker.org/chainmaker-go/subscriber"
	blockSync "chainmaker.org/chainmaker-go/sync"
	"chainmaker.org/chainmaker-go/txpool"
	"chainmaker.org/chainmaker-go/utils"
	"chainmaker.org/chainmaker-go/vm"
	consensusPb "chainmaker.org/chainmaker/pb-go/v2/consensus"
	storePb "chainmaker.org/chainmaker/pb-go/v2/store"
	"chainmaker.org/chainmaker/protocol/v2"
)

// Init all the modules.
func (bc *Blockchain) Init() (err error) {
	baseModules := []map[string]func() error{
		// init Subscriber
		{moduleNameSubscriber: bc.initSubscriber},
		// init store module
		{moduleNameStore: bc.initStore},
		// init ledger module
		{moduleNameLedger: bc.initCache},
		// init chain config , must latter than store module
		{moduleNameChainConf: bc.initChainConf},
	}

	if err := bc.initBaseModules(baseModules); err != nil {
		return err
	}

	var extModules []map[string]func() error

	if bc.getConsensusType() == consensusPb.ConsensusType_SOLO {
		// solo
		extModules = []map[string]func() error{
			// init access control
			{moduleNameAccessControl: bc.initAC},
			// init vm instances and module
			{moduleNameVM: bc.initVM},

			// init transaction pool
			{moduleNameTxPool: bc.initTxPool},
			// init core engine
			{moduleNameCore: bc.initCore},
			// init consensus module
			{moduleNameConsensus: bc.initConsensus},
		}
	} else {
		// not solo
		extModules = []map[string]func() error{
			// init access control
			{moduleNameAccessControl: bc.initAC},
			// init net service
			{moduleNameNetService: bc.initNetService},
			// init vm instances and module
			{moduleNameVM: bc.initVM},
			// init dpos service
			{moduleNameDpos: bc.initDpos},

			// init transaction pool
			{moduleNameTxPool: bc.initTxPool},
			// init core engine
			{moduleNameCore: bc.initCore},
			// init consensus module
			{moduleNameConsensus: bc.initConsensus},
			// init sync service module
			{moduleNameSync: bc.initSync},
		}
	}

	bc.log.Debug("start to init blockchain ...")

	if err := bc.initExtModules(extModules); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) initBaseModules(baseModules []map[string]func() error) (err error) {
	moduleNum := len(baseModules)
	for idx, baseModule := range baseModules {
		for name, initFunc := range baseModule {
			if err := initFunc(); err != nil {
				bc.log.Errorf("init module[%s] failed, %s", name, err)
				return err
			}
			bc.log.Infof("BASE INIT STEP (%d/%d) => init base[%s] success :)", idx+1, moduleNum, name)
		}
	}
	return
}

func (bc *Blockchain) initExtModules(extModules []map[string]func() error) (err error) {
	moduleNum := len(extModules)
	for idx, initModule := range extModules {
		for name, initFunc := range initModule {
			if err := initFunc(); err != nil {
				bc.log.Errorf("init module[%s] failed, %s", name, err)
				return err
			}
			bc.log.Infof("MODULE INIT STEP (%d/%d) => init module[%s] success :)", idx+1, moduleNum, name)
		}
	}
	return
}

func (bc *Blockchain) initDpos() (err error) {
	_, ok := bc.initModules[moduleNameDpos]
	if ok {
		bc.log.Infof("dpos service module existed, ignore.")
		return
	}
	bc.dpos = dpos.NewDPoSImpl(bc.chainConf, bc.store)
	bc.initModules[moduleNameNetService] = struct{}{}
	return
}

func (bc *Blockchain) initNetService() (err error) {
	_, ok := bc.initModules[moduleNameNetService]
	if ok {
		bc.log.Infof("net service module existed, ignore.")
		return
	}
	var netServiceFactory net.NetServiceFactory
	if bc.netService, err = netServiceFactory.NewNetService(bc.net, bc.chainId, bc.ac, bc.chainConf, net.WithMsgBus(bc.msgBus)); err != nil {
		bc.log.Errorf("new net service failed, %s", err)
		return
	}
	bc.initModules[moduleNameNetService] = struct{}{}
	return
}

func (bc *Blockchain) initStore() (err error) {
	_, ok := bc.initModules[moduleNameStore]
	if ok {
		bc.log.Infof("store module existed, ignore.")
		return
	}
	var storeFactory store.Factory
	storeLogger := logger.GetLoggerByChain(logger.MODULE_STORAGE, bc.chainId)
	if bc.store, err = storeFactory.NewStore(bc.chainId, &localconf.ChainMakerConfig.StorageConfig, storeLogger); err != nil {
		bc.log.Errorf("new store failed, %s", err.Error())
		return err
	}
	bc.initModules[moduleNameStore] = struct{}{}
	return
}

func (bc *Blockchain) initChainConf() (err error) {
	_, ok := bc.initModules[moduleNameChainConf]
	if ok {
		bc.log.Infof("chain config module existed, ignore.")
		return
	}
	bc.chainConf, err = chainconf.NewChainConf(
		chainconf.WithChainId(bc.chainId),
		chainconf.WithMsgBus(bc.msgBus),
		chainconf.WithBlockchainStore(bc.store),
	)
	if err != nil {
		bc.log.Errorf("new chain config failed, %s", err.Error())
		return err
	}
	err = bc.chainConf.Init()
	if err != nil {
		bc.log.Errorf("init chain config failed, %s", err)
		return err
	}
	bc.chainNodeList, err = bc.chainConf.GetConsensusNodeIdList()
	if err != nil {
		bc.log.Errorf("load node list of chain config failed, %s", err)
		return err
	}
	bc.initModules[moduleNameChainConf] = struct{}{}

	// register myself as config watcher
	bc.chainConf.AddWatch(bc)
	//if localconf.ChainMakerConfig.StorageConfig.StateDbConfig.IsSqlDB() {
	//	panic("init chain conf fail. sql the future feature")
	//}
	return
}

func (bc *Blockchain) initCache() (err error) {
	_, ok := bc.initModules[moduleNameLedger]
	if ok {
		bc.log.Infof("ledger module existed, ignore.")
		return
	}
	// create genesis block
	// 1) if not exist on chain, create it
	// 2) if exist on chain, load the config in genesis, it will be changed to load the config in config transactions in the future
	bc.lastBlock, err = bc.store.GetLastBlock()
	if err != nil { //可能是全新数据库没有任何数据，而且还没创世，所以可能报错，不返回错误，继续进行创世操作即可
		bc.log.Infof("get last block failed, if it's a genesis block, ignore this error, %s", err.Error())
	}

	if bc.lastBlock != nil {
		bc.log.Infof("get last block [chainId:%s]/[height:%d]/[blockhash:%s] success, no need to create genesis block",
			bc.lastBlock.GetHeader().ChainId, bc.lastBlock.GetHeader().BlockHeight, hex.EncodeToString(bc.lastBlock.GetHeader().BlockHash))
	} else {
		chainConfig, err := chainconf.Genesis(bc.genesis)
		if err != nil {
			bc.log.Errorf("invoke chain config genesis failed, %s", err)
			return err
		}
		genesisBlock, rwSetList, err := utils.CreateGenesis(chainConfig)
		if err != nil {
			return fmt.Errorf("create chain [%s] genesis failed, %s", bc.chainId, err.Error())
		}
		if err = bc.store.InitGenesis(&storePb.BlockWithRWSet{Block: genesisBlock, TxRWSets: rwSetList, ContractEvents: nil}); err != nil {
			return fmt.Errorf("put chain[%s] genesis block failed, %s", bc.chainId, err.Error())
		}

		bc.lastBlock = genesisBlock
	}

	//// load chain config with genesis block info
	//if err := ledger.ChainConfigBlock2CMConf(*cc, genesisBlock); err != nil {
	//	return fmt.Errorf("chainConfigBlock2CMConf failed, %s", err.Error())
	//}

	// cache the lasted config block
	bc.ledgerCache = cache.NewLedgerCache(bc.chainId)
	bc.ledgerCache.SetLastCommittedBlock(bc.lastBlock)
	bc.proposalCache = cache.NewProposalCache(bc.chainConf, bc.ledgerCache)
	bc.log.Debugf("go last block: %+v", bc.lastBlock)
	bc.initModules[moduleNameLedger] = struct{}{}
	return nil
}

func (bc *Blockchain) initAC() (err error) {
	_, ok := bc.initModules[moduleNameAccessControl]
	if ok {
		bc.log.Infof("access control module existed, ignore.")
		return
	}
	// initialize access control: policy list and resource-policy mapping
	nodeConfig := localconf.ChainMakerConfig.NodeConfig
	skFile := nodeConfig.PrivKeyFile
	if !filepath.IsAbs(skFile) {
		skFile, err = filepath.Abs(skFile)
		if err != nil {
			return err
		}
	}
	certFile := nodeConfig.CertFile
	if !filepath.IsAbs(certFile) {
		certFile, err = filepath.Abs(certFile)
		if err != nil {
			return err
		}
	}
	acLog := logger.GetLoggerByChain(logger.MODULE_ACCESS, bc.chainId)
	//bc.ac, err = accesscontrol.NewAccessControlWithChainConfig(bc.chainConf, nodeConfig.OrgId, bc.store, acLog)
	//if err != nil {
	//	bc.log.Errorf("get organization information failed, %s", err.Error())
	//	return
	//}
	acFactory := accesscontrol.ACFactory()
	bc.ac, err = acFactory.NewACProvider("CERT", bc.chainConf, nodeConfig.OrgId, bc.store, acLog)
	if err != nil {
		bc.log.Errorf("get organization information failed, %s", err.Error())
		return
	}

	bc.identity, err = accesscontrol.InitCertSigningMember(bc.chainConf.ChainConfig(), nodeConfig.OrgId,
		nodeConfig.PrivKeyFile, nodeConfig.PrivKeyPassword, nodeConfig.CertFile)
	if err != nil {
		bc.log.Errorf("initialize identity failed, %s", err.Error())
		return
	}

	bc.initModules[moduleNameAccessControl] = struct{}{}
	return
}

func (bc *Blockchain) initTxPool() (err error) {
	_, ok := bc.initModules[moduleNameTxPool]
	if ok {
		bc.log.Infof("tx pool module existed, ignore.")
		return
	}
	// init transaction pool
	var (
		txPoolFactory txpool.TxPoolFactory
		txType        = txpool.SINGLE
	)
	if strings.ToUpper(localconf.ChainMakerConfig.TxPoolConfig.PoolType) == string(txpool.BATCH) {
		txType = txpool.BATCH
	}
	txpoolLogger := logger.GetLoggerByChain(logger.MODULE_TXPOOL, bc.chainId)
	bc.txPool, err = txPoolFactory.NewTxPool(txpoolLogger,
		txType,
		txpool.WithNodeId(localconf.ChainMakerConfig.NodeConfig.NodeId),
		txpool.WithMsgBus(bc.msgBus),
		txpool.WithChainId(bc.chainId),
		txpool.WithNetService(bc.netService),
		txpool.WithBlockchainStore(bc.store),
		txpool.WithSigner(bc.identity),
		txpool.WithChainConf(bc.chainConf),
		txpool.WithAccessControl(bc.ac),
	)
	if err != nil {
		bc.log.Errorf("new tx pool failed, %s", err)
		return err
	}
	bc.initModules[moduleNameTxPool] = struct{}{}
	return nil
}

func (bc *Blockchain) initVM() (err error) {
	_, ok := bc.initModules[moduleNameVM]
	if ok {
		bc.log.Infof("vm module existed, ignore.")
		return
	}
	// init VM
	var vmFactory vm.Factory
	if bc.netService == nil {
		bc.vmMgr = vmFactory.NewVmManager(localconf.ChainMakerConfig.StorageConfig.StorePath, bc.ac, &soloChainNodesInfoProvider{}, bc.chainConf)
	} else {
		bc.vmMgr = vmFactory.NewVmManager(localconf.ChainMakerConfig.StorageConfig.StorePath, bc.ac, bc.netService.GetChainNodesInfoProvider(), bc.chainConf)
	}
	bc.initModules[moduleNameVM] = struct{}{}
	return
}

type soloChainNodesInfoProvider struct{}

func (s *soloChainNodesInfoProvider) GetChainNodesInfo() ([]*protocol.ChainNodeInfo, error) {
	return []*protocol.ChainNodeInfo{}, nil
}

func (bc *Blockchain) initCore() (err error) {
	_, ok := bc.initModules[moduleNameCore]
	if ok {
		bc.log.Infof("core engine module existed, ignore.")
		return
	}
	// create snapshot manager
	var snapshotFactory snapshot.Factory
	if bc.chainConf.ChainConfig().Snapshot != nil && bc.chainConf.ChainConfig().Snapshot.EnableEvidence {
		bc.snapshotManager = snapshotFactory.NewSnapshotEvidenceMgr(bc.store)
	} else {
		bc.snapshotManager = snapshotFactory.NewSnapshotManager(bc.store)
	}
	// init coreEngine module
	coreEngineConfig := &providerConf.CoreEngineConfig{
		ChainId:         bc.chainId,
		TxPool:          bc.txPool,
		SnapshotManager: bc.snapshotManager,
		MsgBus:          bc.msgBus,
		Identity:        bc.identity,
		LedgerCache:     bc.ledgerCache,
		ChainConf:       bc.chainConf,
		AC:              bc.ac,
		BlockchainStore: bc.store,
		Log:             logger.GetLoggerByChain(logger.MODULE_CORE, bc.chainId),
		VmMgr:           bc.vmMgr,
		ProposalCache:   bc.proposalCache,
		Subscriber:      bc.eventSubscriber,
	}

	coreEngineFactory := core.Factory()
	bc.coreEngine, err = coreEngineFactory.NewConsensusEngine(bc.getConsensusType().String(), coreEngineConfig)
	if err != nil {
		bc.log.Errorf("new core engine failed, %s", err.Error())
		return err
	}
	bc.initModules[moduleNameCore] = struct{}{}
	return
}

func (bc *Blockchain) initConsensus() (err error) {
	// init consensus module
	var consensusFactory consensus.Factory
	id := localconf.ChainMakerConfig.NodeConfig.NodeId
	nodes := bc.chainConf.ChainConfig().Consensus.Nodes
	nodeIds := make([]string, len(nodes))
	for i, node := range nodes {
		for _, nid := range node.NodeId {
			nodeIds[i] = nid
		}
	}
	_, ok := bc.initModules[moduleNameConsensus]
	if ok {
		bc.log.Infof("consensus module existed, ignore.")
		return
	}
	dbHandle := bc.store.GetDBHandle(protocol.ConsensusDBName)
	bc.consensus, err = consensusFactory.NewConsensusEngine(
		bc.getConsensusType(),
		bc.chainId,
		id,
		nodeIds,
		bc.identity,
		bc.ac,
		dbHandle,
		bc.ledgerCache,
		bc.proposalCache,
		bc.coreEngine.GetBlockVerifier(),
		bc.coreEngine.GetBlockCommitter(),
		bc.netService,
		bc.msgBus,
		bc.chainConf,
		bc.store,
		bc.coreEngine.GetHotStuffHelper(),
		bc.dpos)
	if err != nil {
		bc.log.Errorf("new consensus engine failed, %s", err)
		return err
	}
	bc.initModules[moduleNameConsensus] = struct{}{}
	return
}

func (bc *Blockchain) initSync() (err error) {
	_, ok := bc.initModules[moduleNameSync]
	if ok {
		bc.log.Infof("sync module existed, ignore.")
		return
	}
	// init sync service module
	bc.syncServer = blockSync.NewBlockChainSyncServer(
		bc.chainId,
		bc.netService,
		bc.msgBus,
		bc.store,
		bc.ledgerCache,
		bc.coreEngine.GetBlockVerifier(),
		bc.coreEngine.GetBlockCommitter(),
	)
	bc.initModules[moduleNameSync] = struct{}{}
	return
}

func (bc *Blockchain) initSubscriber() error {
	_, ok := bc.initModules[moduleNameSubscriber]
	if ok {
		bc.log.Infof("subscriber module existed, ignore.")
		return nil
	}
	bc.eventSubscriber = subscriber.NewSubscriber(bc.msgBus)
	bc.initModules[moduleNameSubscriber] = struct{}{}
	return nil
}

func (bc *Blockchain) isModuleInit(moduleName string) bool {
	_, ok := bc.initModules[moduleName]
	return ok
}
