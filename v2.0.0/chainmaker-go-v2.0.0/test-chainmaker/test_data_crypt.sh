#! /bin/bash

# test if data of block is encrypted

. utils.sh

BLOCK_ADDRESS=$RESULT_ADDRESS/block

function test_data_crypt() {
    ${cmc} query block-by-height $1 \
    --chain-id=chain1 \
    --sdk-conf-path=./testdata/sdk_config.yml > $BLOCK_ADDRESS/block_$1.json

    # mv mychannel_$1.block 
    # configtxlator proto_decode --input $BLOCK_ADDRESS/mychannel_$1.block --type common.Block --output $BLOCK_ADDRESS/trace.json

    local value=`cat $BLOCK_ADDRESS/block_$1.json | sed -n "/\"value\":\s/ p" | awk '{print $2}'`
    echo $value
    # echo ${#value}
}

if [ ! -d "$BLOCK_ADDRESS" ]; then
    mkdir -p $BLOCK_ADDRESS
fi

test_data_crypt 7