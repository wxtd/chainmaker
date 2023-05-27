#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
function ut_cover() {
  cd ${cm}/$1
  go test -coverprofile cover.out ./...
  total=$(go tool cover -func=cover.out | tail -1)
  echo ${total}
  rm cover.out
  coverage=$(echo ${total} | grep -P '\d+\.\d+(?=\%)' -o) #如果macOS 不支持grep -P选项，可以通过brew install grep更新grep
      # 如果测试覆盖率低于N，认为ut执行失败
  (( $(awk 'BEGIN {print ("'${coverage}'" >= "'$2'")}') )) || (echo "$1 单测覆盖率低于$2%"; exit 1)
}

cm=$(pwd)/..

set -e

ut_cover "module/accesscontrol" 47
ut_cover "module/blockchain" 2.2
ut_cover "module/conf/chainconf" 26
ut_cover "module/conf/localconf" 11
ut_cover "module/consensus" 11.4
ut_cover "module/core" 2.1
ut_cover "module/logger" 46
ut_cover "module/net" 29
ut_cover "module/rpcserver" 0
ut_cover "module/snapshot" 25
ut_cover "module/store" 38
ut_cover "module/sync" 61
ut_cover "module/txpool" 47
ut_cover "module/utils" 27
ut_cover "module/vm" 30
#ut_cover "tools/cmc" 6
