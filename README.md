# Test-chainmaker

> Chainmaker 搭建 four-node-chainmaker网络 并使用test-chainmaker下的脚本文件进行技术风险检测



## 初始化Chainmaker 网络（v2.0.0）

```shell
cd chainmaker/v2.0.0/chainmaker-go-v2.0.0

cd scripts/docker

# 启动网络
./four-nodes_up.sh

# 查看网络是否生成完毕 4node
docker ps -a

# 关闭网络
# ./four-nodes_down.sh
```

查询、更新

```shell
# 复制证书信息，便于后续操作 可以手动更新sdk_configxx.yml文件中的路径
cp -r config ../../tools/cmc/testdata

cd ../../tools/cmc

# 注册
./cmc client contract user create \
--contract-name=chain_002 \
--runtime-type=GASM \
--byte-code-path=/home/ubuntu/ms/chainmaker/v2.0.0/chainmaker-go-v2.0.0/contract/contract_tinygo/chainmaker-contract-go.wasm \
--version=1.0 \
--sdk-conf-path=./testdata/sdk_config.yml \
--admin-key-file-paths=./testdata/config/four-nodes/wx-org1.chainmaker.org/user/admin1/admin1.tls.key,./testdata/config/four-nodes/wx-org2.chainmaker.org/user/admin1/admin1.tls.key,./testdata/config/four-nodes/wx-org3.chainmaker.org/user/admin1/admin1.tls.key,./testdata/config/four-nodes/wx-org4.chainmaker.org/user/admin1/admin1.tls.key \
--admin-crt-file-paths=./testdata/config/four-nodes/wx-org1.chainmaker.org/user/admin1/admin1.tls.crt,./testdata/config/four-nodes/wx-org2.chainmaker.org/user/admin1/admin1.tls.crt,./testdata/config/four-nodes/wx-org3.chainmaker.org/user/admin1/admin1.tls.crt,./testdata/config/four-nodes/wx-org4.chainmaker.org/user/admin1/admin1.tls.crt \
--sync-result=true \
--params="{}"

# save
./cmc client contract user invoke \
--contract-name=chain_002 \
--method=save \
--sdk-conf-path=./testdata/sdk_config.yml \
--params="{\"file_name\":\"name007\",\"file_hash\":\"ab3456df5799b87c77e7f88\",\"time\":\"6543234\"}" \
--sync-result=true

# find_by_file_hash
./cmc client contract user get \
--contract-name=chain_002 \
--method=find_by_file_hash \
--sdk-conf-path=./testdata/sdk_config.yml \
--params="{\"file_hash\":\"ab3456df5799b87c77e7f88\"}"

# org2
./cmc client contract user get \
--contract-name=chain_002 \
--method=find_by_file_hash \
--sdk-conf-path=./testdata/sdk_config2.yml \
--params="{\"file_hash\":\"ab3456df5799b87c77e7f88\"}"


# 吊销
./cmc client contract user revoke \
--contract-name=chain_002 \
--sdk-conf-path=./testdata/sdk_config.yml \
--admin-key-file-paths=./testdata/config/four-nodes/wx-org1.chainmaker.org/user/admin1/admin1.tls.key,./testdata/config/four-nodes/wx-org2.chainmaker.org/user/admin1/admin1.tls.key,./testdata/config/four-nodes/wx-org3.chainmaker.org/user/admin1/admin1.tls.key \
--admin-crt-file-paths=./testdata/config/four-nodes/wx-org1.chainmaker.org/user/admin1/admin1.tls.crt,./testdata/config/four-nodes/wx-org2.chainmaker.org/user/admin1/admin1.tls.crt,./testdata/config/four-nodes/wx-org3.chainmaker.org/user/admin1/admin1.tls.crt \
--org-id=wx-org1.chainmaker.org \
--sync-result=true
```



## 测试

测试中转结果存放在result文件夹中，测试最终结果打印在命令行

```shell
cd ../../test-chainmaker/

cp -r ../scripts/docker/config ./testdata

# 基础query & invoke
# 关于query
./test_query_or_invoke.sh query 1 1 chain_002 find_by_file_hash file_hash ab3456df5799b87c77e7f88
./test_query_or_invoke.sh query 2 1 chain_002 find_by_file_hash file_hash ab3456df5799b87c77e7f88
# or ./example_query.sh 1 && ./example_query.sh 2

# 关于invoke
./test_query_or_invoke.sh invoke 1 chain_002 save file_name name008 file_hash bb3456df5799b87c77e7f88 time 6543234
# or ./example_invoke.sh

# 测试是否使用微服务架构
./test_docker_architecture.sh

# 测试所有加密算法 若bug可替换其中的路径为密钥存放的绝对路径
./detect_encryption_method.sh

# 测试可维护性 使用混沌工程工具chaosblade制造故障
./test_blade.sh

# 打印链码所有函数（弃用）
./print_chaincode_function.sh

# 打印区块数据 是否加密
./test_data_crypt.sh

# 测试交易幂等性、持久性
./test_data_duration.sh

# 测试node高可用性
./test_node_high_available.sh invoke 1 chain_002 save file_name name010 file_hash db3456df5799b87c77e7f88 time 6543234
# or in query mod
# ./test_node_high_available.sh query 1 chain_002 find_by_file_hash file_hash ab3456df5799b87c77e7f88

# 测试node数据一致性
./test_node_data_consistency.sh

# 验证共识节点(Raft)
./test_node_raft.sh
```

