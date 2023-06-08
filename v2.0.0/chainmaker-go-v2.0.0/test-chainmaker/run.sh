#! /bin/bash

. utils.sh

query_data="1 chain_002 find_by_file_hash file_hash ab3456df5799b87c77e7f88"
invoke_data="1 chain_002 save file_name name008 file_hash bb3456df5799b87c77e7f88 time 6543234"

function test() {
    local result=$RESULT_ADDRESS/result.txt
    # local flag=">>$result 2>>$LOG"
    # if [[ ! -f "$result" ]]; then
    #     touch $result
    # fi
    > $result
    record_result $(TZ=UTC-8 date +%Y-%m-%d\ %H:%M:%S)
    record_separator

    # set -x

    # record_result "1232 3 232 232 322 3"

    # 测试是否使用微服务架构
    record_result "./test_docker_architecture.sh # Testing if using docker architecture"
    ./test_docker_architecture.sh >>$result 2>>$LOG

    # 基础query & invoke
    # 关于query
    record_result "./test_query_or_invoke.sh query 1 $query_data"
    ./test_query_or_invoke.sh query 1 $query_data >>$result 2>>$LOG
    record_result "./test_query_or_invoke.sh query 2 $query_data # Testing query command"
    ./test_query_or_invoke.sh query 2 $query_data >>$result 2>>$LOG
    # or ./example_query.sh 1 && ./example_query.sh 2

    # 关于invoke
    record_result "./test_query_or_invoke.sh invoke $invoke_data # Testing invoke command"
    ./test_query_or_invoke.sh invoke $invoke_data >>$result 2>>$LOG
    # or ./example_invoke.sh

    # 测试所有加密算法 若bug可替换其中的路径为密钥存放的绝对路径
    record_result "./detect_encryption_method.sh # Testing all encryption algorithms"
    ./detect_encryption_method.sh >>$result 2>>$LOG

    # 测试可维护性 使用混沌工程工具chaosblade制造故障
    record_result "./test_blade.sh # Testing maintainability Creating faults with chaosblade, a chaos engineering tool"
    ./test_blade.sh >>$result 2>>$LOG

    # # 打印链码所有函数（弃用）
    # # ./print_chaincode_function.sh >> $result

    # 打印区块数据 是否加密
    record_result "./test_data_crypt.sh # Detecting whether the block data content is desensitized and encrypted"
    ./test_data_crypt.sh >>$result 2>>$LOG

    # 测试交易幂等性、持久性
    record_result "./test_data_duration.sh # Testing transaction idempotency, persistence"
    ./test_data_duration.sh >>$result 2>>$LOG

    # 测试node高可用性
    record_result "./test_node_high_available.sh invoke 0 $invoke_data # Testing node high availability"
    ./test_node_high_available.sh invoke $invoke_data >>$result 2>>$LOG
    # or in query mod
    # ./test_node_high_available.sh query $query_data >>$result 2>>$LOG

    # 测试node数据一致性
    record_result "./test_node_data_consistency.sh # Detecting node data consistency, support failover or not"
    ./test_node_data_consistency.sh >>$result 2>>$LOG

    # 验证共识节点(Raft)
    record_result "./test_node_raft.sh # Validating Consensus Algorithms"
    ./test_node_raft.sh >>$result 2>>$LOG

    # set +x

    record_separator
    
    cat $result
}

function record_result() {
    infoln $@ >> $result
    # echo $1 >> $result
}

function record_separator() {
    println "-------------------------------------------------" >> $result
}

test

