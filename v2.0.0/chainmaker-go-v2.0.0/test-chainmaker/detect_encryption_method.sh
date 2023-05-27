#! /bin/bash

# Detect encryption method

. utils.sh


NODE_KEY_ADDRESS=$TESTDATA_ADDRESS/crypto-config

function test_node_crypt() {
    getdir $NODE_KEY_ADDRESS
}

function getdir(){
    for element in `ls $1`
    do  
        dir_or_file=$1"/"$element
        if [ -d $dir_or_file ]; then 
            getdir $dir_or_file
        else
            if [[ "$dir_or_file" =~ ".crt" || "$dir_or_file" =~ ".pem" ]]; then
                # println $dir_or_file
                result=`openssl x509 -text -in $dir_or_file | grep Signature\ Algorithm: | awk '{print $3}' | uniq`
                println $result
            fi
        fi  
    done
}

test_node_crypt