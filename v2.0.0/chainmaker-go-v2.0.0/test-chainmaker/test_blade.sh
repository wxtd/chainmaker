#! /bin/bash

. utils.sh

# BLADE_ADDRESS=/home/ubuntu/ms/ChaosEngineer/chaos1.5/chaosblade-1.5.0/target/chaosblade-1.5.0
BLADE_ADDRESS=./tools/target/chaosblade-1.5.0
BLADE=${BLADE_ADDRESS}/blade
TIMING=30
TIME_FLAG=--timeout ${TIMING}


function test_blade() {
    query1

    ${BLADE} c cpu load --cpu-percent 60 ${TIME_FLAG}
    sleep_for_some_time
    query2
    sleep_for_some_time
    ${BLADE} c disk burn --read --write ${TIME_FLAG}
    sleep_for_some_time
    query2
    sleep_for_some_time
    ${BLADE} c mem load --mode ram --mem-percent 50 ${TIME_FLAG}
    sleep_for_some_time
    query2
    sleep_for_some_time
}

function sleep_for_some_time() {
    sleep 15
}

test_blade