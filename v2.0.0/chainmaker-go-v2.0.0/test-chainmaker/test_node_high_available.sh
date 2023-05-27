#!/bin/bash

# Test node's High Availablity

. utils.sh

# docker stop cm-node1

cmd=./test_query_or_invoke.sh

test_query_op() {
    
    process_test_org

    $cmd $MODE 1 $ORG1 $@

    stop_docker cm-node${ORG_NUM}
    
    $cmd $MODE 2 $ORG1 $@
    # $cmd $MODE 2 $ORG2 $@

    start_docker cm-node${ORG_NUM}
}

test_invoke_op() {
    # set -x
    process_test_org

    stop_docker cm-node${ORG_NUM}
    
    $cmd $MODE $ORG1 $@
    # $cmd $MODE $ORG2 $@

    start_docker cm-node${ORG_NUM}
    # set +x
}

# get the orgs need to test
process_test_org() {
    if [ $ORG_NUM -eq 1 ]; then
        ORG1=2
        ORG2=3
    elif [ $ORG_NUM -eq 2 ]; then
        ORG1=1
        ORG2=3
    elif [ $ORG_NUM -eq 3 ]; then
        ORG1=1
        ORG2=2
    else 
        errorln "ORG Unknown"
    fi
}

printHelp() {
    println "HELP"
    println "   Please input params like this :"
    println "       ./test_node_high_available.sh query 1 chain_002 find_by_file_hash file_hash ab3456df5799b87c77e7f88"
    println "   or ./test_node_high_available.sh invoke 1 chain_002 save file_name name010 file_hash db3456df5799b87c77e7f88 time 6543234"
}



if [[ $# -lt 4 ]] ; then
    errorln "Params insufficient!"
    printHelp
    exit 0
else
    MODE=$1
    shift
fi

ORG_NUM=$1
shift

if [ "$MODE" == "query" ]; then
    test_query_op $@
elif [ "$MODE" == "invoke" ]; then
    test_invoke_op $@
else
    errorln "Mode illegal!"
    printHelp
    exit 0
fi