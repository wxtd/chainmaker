#! /bin/bash

# print functions in the chaincode

. utils.sh

CHAINCODE_FILE=/home/ubuntu/ms/chainmaker/v2.0.0/chainmaker-go-v2.0.0/contract/contract_tinygo/main.go

function print_chaincode_function() {
    cat $CHAINCODE_FILE | sed -n '/^func /p' > $CHAINCODE_RESULT_ADDRESS/result.txt

    cat $CHAINCODE_RESULT_ADDRESS/result.txt
}

if [ ! -d "$CHAINCODE_RESULT_ADDRESS" ]; then
    mkdir -p $CHAINCODE_RESULT_ADDRESS
fi

print_chaincode_function