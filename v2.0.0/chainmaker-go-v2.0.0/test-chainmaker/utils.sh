#!/bin/bash

# Some utils about shells

RESULT_ADDRESS=result
NODE_RESULT_ADDRESS=$RESULT_ADDRESS/node
ORDERER_RESULT_ADDRESS=$RESULT_ADDRESS/orderer
CHAINCODE_RESULT_ADDRESS=$RESULT_ADDRESS/chaincode
QUERY_OR_INVOKE_RESULT_ADDRESS=$RESULT_ADDRESS/node
TESTDATA_ADDRESS=./testdata
cmc=../tools/cmc/cmc
SDK_ADDRESS=${TESTDATA_ADDRESS}/sdk_config.yml
SDK1_ADDRESS=${TESTDATA_ADDRESS}/sdk_config1.yml
SDK2_ADDRESS=${TESTDATA_ADDRESS}/sdk_config2.yml
SDK3_ADDRESS=${TESTDATA_ADDRESS}/sdk_config3.yml
SDK4_ADDRESS=${TESTDATA_ADDRESS}/sdk_config4.yml

setGlobals() {
  local ORG_NUM=$1
  infoln "Using organization ${ORG_NUM}"
  # set -x
  local ADDRESS=$(dirname "$PWD")

  # if [ $ORG_NUM -eq 1 ]; then
  #     export CORE_PEER_ADDRESS=localhost:7051
  # elif [ $ORG_NUM -eq 2 ]; then
  #     export CORE_PEER_ADDRESS=localhost:9051
  # elif [ $ORG_NUM -eq 3 ]; then
  #     export CORE_PEER_ADDRESS=localhost:11051
  # else
  #     errorln "ORG Unknown"
  # fi
  # set +x
}

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_BLUE='\033[0;34m'
C_YELLOW='\033[1;33m'

# println echos string
function println() {
  echo -e "$1"
}

# errorln echos i red color
function errorln() {
  println "${C_RED}${1}${C_RESET}"
}

# successln echos in green color
function successln() {
  println "${C_GREEN}${1}${C_RESET}"
}

# infoln echos in blue color
function infoln() {
  println "${C_BLUE}${1}${C_RESET}"
}

# warnln echos in yellow color
function warnln() {
  println "${C_YELLOW}${1}${C_RESET}"
}

# fatalln echos in red color and exits with fail status
function fatalln() {
  errorln "$1"
  exit 1
}

function query1() {
  ./example_query.sh 1
}

function query2() {
  ./example_query.sh 2
}

stop_docker() { 
  for i in "$@"
  do
    println "Stopping ${i}"
    docker stop $i >/dev/null
    wait
    println "Stop Success"
  done
}

start_docker() { 
  for i in "$@"
  do
    println "Starting ${i}"
    docker start $i >/dev/null
    wait
    sleep 30
    # TODO
    println "Start Success"
  done
}

# function get_block() {
#   # peer channel fetch $1 -c mychannel --orderer orderer0.example.com:7050 \
#   # --tls --cafile /home/ubuntu/ms/fabric-samples/test-network/organizations/ordererOrganizations/example.com/orderers/orderer0.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

#   ${cmc} query block-by-height $1 \
#   --chain-id=chain1 \
#   --sdk-conf-path=./testdata/sdk_config.yml > 

# }


# function getdir(){
#   for element in `ls $1`
#   do  
#     dir_or_file=$1"/"$element
#     if [ -d $dir_or_file ]
#     then 
#       getdir $dir_or_file
#     else
#       echo $dir_or_file
#     fi  
#   done
# }