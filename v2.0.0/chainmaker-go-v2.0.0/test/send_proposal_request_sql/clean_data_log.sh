#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
rm -rf ../../data
rm -rf ../../log/*/*
rm -rf ../../main/panic.log
rm -rf ../../../cmdata

for((i=1;i<=4;i++))
do
    mysql -uroot -ppassw0rd -P3307 -h192.168.1.35 -e "show databases like 'org${i}_%'" |grep -v org${i}_% | xargs -I{} mysql -uroot -ppassw0rd -P3307 -h192.168.1.35 -e "drop database {}"
done
mysql -uroot -ppassw0rd -P3307 -h192.168.1.35 -e "show databases;"
ps -fe|grep chainmaker|grep -v grep|grep start