#! /bin/bash

. utils.sh

test_data_duration() {
    query1
    sleep 1
    query2
    local interval=30
    sleep ${interval}
    
    query2
}

test_data_duration