#!/bin/bash

# Test Chainmaker Query & Invoke Command

. utils.sh

# Firstly query and then save this result as log1.txt
saveQueryResult() {
    println "Querying~"
    # println $#
    
    processParam $@

    # println $ARGS

    # local org_address=${TESTDATA_ADDRESS}/sdk_config${ORG_NUM}.yml

    # set -x
    ${cmc} client contract user get \
    --contract-name=${CHAINCODE_NAME} \
    --method=${FUNCTION_NAME} \
    --sdk-conf-path=${TESTDATA_ADDRESS}/sdk_config${ORG_NUM}.yml \
    --params="{$ARGS}" \
    >$QUERY_OR_INVOKE_RESULT_ADDRESS/query_result1.txt 2>&1
    # set +x
}


# compare query result with the query result firstly
CompareQueryResult() {
    # println "CompareQueryResult"
    println "Querying~"
    processParam $@

    ${cmc} client contract user get \
    --contract-name=${CHAINCODE_NAME} \
    --method=${FUNCTION_NAME} \
    --sdk-conf-path=${TESTDATA_ADDRESS}/sdk_config${ORG_NUM}.yml \
    --params="{$ARGS}" \
    >$QUERY_OR_INVOKE_RESULT_ADDRESS/query_result2.txt 2>&1

    local result1=`cat $QUERY_OR_INVOKE_RESULT_ADDRESS/query_result1.txt | sed -n "/result:/ p" | awk '{print $5}'`
    # echo $result1
    local result2=`cat $QUERY_OR_INVOKE_RESULT_ADDRESS/query_result2.txt | sed -n "/result:/ p" | awk '{print $5}'`
    # echo $result2
    if [ "$result1" == "$result2" ]; then
        successln "Query result is the same!"
    else
        errorln "Query result is different"
    fi
}

# Invoke cmd result
# Chaincode invoke successful. result: status:200
ivokeChaincode() {
    println "Invoking~"
    processParam $@

    # ADDRESS=$(dirname "$PWD")
    local invoke_result=$QUERY_OR_INVOKE_RESULT_ADDRESS/invoke_result.txt
    
    # set -x
    ${cmc} client contract user invoke \
    --contract-name=${CHAINCODE_NAME} \
    --method=${FUNCTION_NAME} \
    --sdk-conf-path=${TESTDATA_ADDRESS}/sdk_config${ORG_NUM}.yml \
    --params="{$ARGS}" \
    --sync-result=true > $invoke_result 2>&1
    # set +x

    success="[code:0]/[msg:OK]"
    local t=`cat $invoke_result | awk '{print $4}'`
    # echo $t
    if [[ "$t" =~ "$success" ]];then
        successln "Invoke Success!"
    else
        errorln "Invoke Failed"
    fi
}

# Process Parameters
processParam() {
    if [ $# -eq 0 ]; then
        ARGS=$ARGS"\"\""
    else
        ARGS=$ARGS"\"$1\""
        shift
        ARGS=$ARGS":""\"$1\""
        shift
    fi
    while [[ $# -gt 0 ]]
    do
        # println $#
        ARGS=$ARGS",""\"$1\""
        shift
        ARGS=$ARGS":""\"$1\""
        shift
    done
}

# Print Help
printHelp() {
    println "Help:"
    println "   ./test_query_or_invoke.sh [TEST_MOD] [OPTION:TURN] [ORG_NUM] [CHAINCODE_NAME] [FUNCTION_NAME] [ARGS]"
    println "   Params:"
    println "       TEST_MOD: query / invoke"
    println "       TURN: 1 / 2"
    println "       ORG_NUM: 1 / 2 / 3 / 4"
    println "   Please input params like this :"
    println "       ./test_query_or_invoke.sh query 1 1 chain_002 find_by_file_hash file_hash ab3456df5799b87c77e7f88"
    println "    or ./test_query_or_invoke.sh invoke 1 chain_002 save file_name name008 file_hash bb3456df5799b87c77e7f88 time 6543234"
}



if [ ! -d "$QUERY_OR_INVOKE_RESULT_ADDRESS" ]; then
    mkdir -p $QUERY_OR_INVOKE_RESULT_ADDRESS
fi

# if [ ! -f "$QUERY_OR_INVOKE_RESULT_ADDRESS/success_invoke.txt" ]; then
#     println "file not exists!"
#     echo "Chaincode invoke successful. result: status:200" > result/success_invoke.txt
# fi

## Parse mode
if [[ $# -lt 4 ]] ; then
    errorln "Params insufficient!"
    printHelp
    exit 0
else
    MODE=$1
    shift
fi

if [ "$MODE" == "query" ]; then
    TURN=$1
    shift
    ORG_NUM=$1
    CHAINCODE_NAME=$2
    FUNCTION_NAME=$3
    shift 3
    setGlobals $ORG_NUM
    if [ $TURN -eq 1 ]; then
        saveQueryResult $@
    else 
        CompareQueryResult $@
    fi
elif [ "$MODE" == "invoke" ]; then
    ORG_NUM=$1
    CHAINCODE_NAME=$2
    FUNCTION_NAME=$3
    shift 3
    setGlobals $ORG_NUM
    ivokeChaincode $@
else
    errorln "Mode illegal!"
    printHelp
    exit 0
fi
