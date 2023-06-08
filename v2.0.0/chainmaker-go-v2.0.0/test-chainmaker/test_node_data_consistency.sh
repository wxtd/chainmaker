#! /bin/bash

# test fabric node data consistency

. utils.sh

NODE_RESULT_ADDRESS=$RESULT_ADDRESS/node
LOG_ADDRESS=/home/ubuntu/ms/chainmaker/v2.0.0/chainmaker-go-v2.0.0/scripts/docker/tmp
# node_list=(cm-node1 cm-node2 cm-node3 cm-node4)
log_list=(log1 log2 log3 log4)
# log_list=(log1)

function test_node_data_consistency() {
    get_log

    local j=0
    # 2023-05-20 14:05:43.948	[INFO]	[Core] @chain1	common/block_helper.go:769	commit block [16](count:1,hash:4c94bbdebfcb1416293afb918de0255d36d46ff4499de0ac8504fe2157e1579c), time used(check:0,db:1435,ss:0,conf:0,pool:1,pubConEvent:0,other:0,total:1436)
    for i in ${log_list[*]}
    do 
        cat $NODE_RESULT_ADDRESS/node_test_$i.txt | grep block \
            | sed -n '/commit\sblock\s\[[0-9]\+\]/ p' \
            | awk '{print $7,$8,$9}' ORS="\n" \
            | grep -o '\[.*\]' \
            > $NODE_RESULT_ADDRESS/temp${j}.txt
            # | awk '{for(i=9; i<=14; i++) {print $i}}' ORS="\n"
        let j++
    done
    # cat $NODE_RESULT_ADDRESS/node_test_node1.example.com.txt | grep block \
    #     | sed -n '/Writing\sblock\s\[[0-9]\+\]\s(Raft\sindex:\s[0-9]\+)/ p' \
    #     | awk '{print $9,$10,$11,$12,$13,$14}' ORS="\n" > b_tmp.txt
    local cnt=${#node_list[@]}
    comm -12 $NODE_RESULT_ADDRESS/temp0.txt $NODE_RESULT_ADDRESS/temp1.txt > $NODE_RESULT_ADDRESS/common.txt

    # cat $NODE_RESULT_ADDRESS/temp0.txt
    # cat $NODE_RESULT_ADDRESS/common.txt
    local idx=2
    while [[ idx -lt $cnt ]]
    do
        # echo $idx
        comm -12 $NODE_RESULT_ADDRESS/common.txt $NODE_RESULT_ADDRESS/temp$idx.txt > $NODE_RESULT_ADDRESS/temp.txt
        cat $NODE_RESULT_ADDRESS/temp.txt > $NODE_RESULT_ADDRESS/common.txt
        # cp $NODE_RESULT_ADDRESS/temp.txt $NODE_RESULT_ADDRESS/common.txt
        # sleep 3
        # rm $NODE_RESULT_ADDRESS/common.txt
        # mv $NODE_RESULT_ADDRESS/temp.txt $NODE_RESULT_ADDRESS/common.txt
        # cat $NODE_RESULT_ADDRESS/temp.txt
        let idx++
    done
    # cat $NODE_RESULT_ADDRESS/temp.txt
    if [[ -s $NODE_RESULT_ADDRESS/common.txt ]]; then
        successln "Common part:"
        cat $NODE_RESULT_ADDRESS/common.txt
    else 
        errorln "No common parts!"
    fi
    rm $NODE_RESULT_ADDRESS/common.txt
    rm $NODE_RESULT_ADDRESS/temp*.txt
}

# for range get log
function get_log() {
    for i in ${log_list[*]}
    do
        # docker logs -f $i --tail 10 > result/node_test_$i.txt 2>&1 &
        # sleep 1 && kill -SIGINT $?
        sleep 1
        get_last_log_from_node $i
    done
    # kill processes about docker logs
    # echo ${#node_list[@]}
    # sleep 3 && kill -9 `ps -ef | grep docker\ logs | awk '{print $2}' | head -${#node_list[@]}`
}

# get log from node
function get_last_log_from_node() {
    local node_name=$1
    # get ${node_name}'s last ${row} log 
    local row=1000
    tail -n $row $LOG_ADDRESS/$node_name/system.log > $NODE_RESULT_ADDRESS/node_test_$node_name.txt 2>&1 &
    # docker logs -f $node_name --tail 10 > $NODE_RESULT_ADDRESS/node_test_$node_name.txt 2>&1 &
}

if [ ! -d "$NODE_RESULT_ADDRESS" ]; then
    mkdir -p $NODE_RESULT_ADDRESS
fi

test_node_data_consistency