#!/bin/bash

secret="04409f6be80c3b10d200905532e93dax"
out_order_id="000107"                            # uniq id
fee="19"                                         # by RMB CENT
notify_url="http://requestbin.net/r/tmtwe1tm"    # callback url
# notify_url="http://requestbin.net/r/1evwwgr1"    # callback url
sign=$(echo -en "out_order_id=${out_order_id}&fee=${fee}&key=${secret}" | md5sum | awk '{print $1}' | tr '[a-z]' '[A-Z]')
body='{"out_order_id": "'${out_order_id}'", "qrtype":"alipay", "fee": '${fee}', "notify_url":"'${notify_url}'", "attach": "anything", "sign": "'${sign}'"}'

curl -s -XPOST -H "Content-Type: application/json" \
	http://127.0.0.1:8008/api/adb_paygate/new \
	-d "$body"  | jq
