
## AdbDevice API

### List
`GET /api/adb_devices`  -  list current adb devices
  
Query Parameters:
  - **brief_all**    - optional: true|false, only list brief `id` for all of adb devices. note: **if true, all of other parameters will be ignored**
  - **search**       - optional: search provided adbnode id or adb device id
  - **status**       - optional: online,offline
  - **over_quota**   - optional: true,false
  - **offset**       - optional: paging parameter, default 0
  - **limit**        - optional: paging parameter, default 20

Example Request:
```liquid
GET /api/adb_devices HTTP/1.1
```

Example Response:
```json
response contains Header: `Total-Records`

[
  {
    "id": "4c8bc08b",        // 设备ID
    "node_id": "4a264c130cde9319",  // 所在分控节点ID
    "sysinfo": {
      "serial_no": "4c8bc08b",
      "device_name": "bbk MP",      // 自定义的设备名
      "manufacturer": "Xiaomi",     // 制造商
      "product_brand": "Xiaomi",
      "product_model": "HM NOTE 1LTE",  // 手机型号
      "product_name": "dior",
      "product_locale": "",
      "release_version": "4.4.4",       // Android版本
      "sdk_version": "19",              // Android SDK版本
      "build_date_utc": "1504795390",
      "time_zone": "Asia/Shanghai",
      "gsm_operator_alpha": "中国移动",    // 移动运营商
      "gsm_operator_country": "cn",
      "gsm_serial": "",
      "gsm_sim_state": "READY",        // SIM卡状态
      "gsm_nitz_time": "1560430327031",
      "gsm_nitz_time_at": "2019-06-13T20:52:07+08:00",
      "boot_time": "1559527604165",
      "boot_time_at": "2019-06-03T10:06:44+08:00", // 开机时间
      "boot_time_for": "10 days",           // 开机运行时长
      "battery": {                 // 电池信息
        "ac_powered": "false",     // 电源线充电
        "usb_powered": "true",     // USB充电
        "wireless_powered": "false", // 无线充电
        "status": "5",
        "level": 100,         // 当前电量
        "scale": 100          // 最大电量数值
      }
    },
    "desc": "",           // 设备描述
    "status": "online",   // 设备状态: online, offline
    "error": "",          // offline时候的错误信息
    "max_amount": 0,      // 单日最大交易金额，0表示不限，单位CNY
    "max_amount_yuan": 0,
    "max_bill": 0,        // 单日最大交易订单数, 0表示不限
    "over_quota": false,  // 当前设备是否已经超出了单日最大配额(上面任意一个配额)
    "weight": 0,          // 权重, 数字0-100, 数字越大表示使用的概率越大，0表示此设备将不被使用
    "today_paid_rate": 50,  // 今日订单成功率
    "recent_adb_orders": {  
      "today": {            // 该设备今日订单统计
        "paid": 1,           // 已支付订单数
        "paid_bill": 0.19,   // 已支付订单金额
        "pending": 0,        // 待支付
        "pending_bill": 0,
        "timeout": 1,        // 等待超时
        "timeout_bill": 0.19
      },
      "month": {           // 该设备本月订单统计
        "paid": 37,
        "paid_bill": 14.11,
        "pending": 0,
        "pending_bill": 0,
        "timeout": 4,
        "timeout_bill": 2.41
      }
    },
    "alipay": {           // 绑定的支付宝账户
        "user_id": "2088032017360044",   // 支付宝UserID   (必填)
        "username": "13619840773",       // 支付宝账号     (必填)
        "nickname": "sldzz"              // 支付宝账号昵称 (非必填)
    },
    "wxpay": null,
  },
  {
    "id": "546052d21f384",
    "node_id": "4a264c130cde9319",
    "sysinfo": {
      "serial_no": "546052d21f384",
      "device_name": "bbk redmi",
      "manufacturer": "Xiaomi",
      "product_brand": "Xiaomi",
      "product_model": "Redmi 4A",
      "product_name": "rolex",
      "product_locale": "zh-CN",
      "release_version": "6.0.1",
      "sdk_version": "23",
      "build_date_utc": "1490366979",
      "time_zone": "Asia/Shanghai",
      "gsm_operator_alpha": "中国移动",
      "gsm_operator_country": "cn",
      "gsm_serial": "2Z721F215568",
      "gsm_sim_state": "READY,ABSENT",
      "gsm_nitz_time": "1560416656012",
      "gsm_nitz_time_at": "2019-06-13T17:04:16+08:00",
      "boot_time": "1557559953185",
      "boot_time_at": "2019-05-11T15:32:33+08:00",
      "boot_time_for": "4 weeks",
      "battery": {
        "ac_powered": "false",
        "usb_powered": "true",
        "wireless_powered": "false",
        "status": "5",
        "level": 100,
        "scale": 100
      }
    },
    "desc": "",
    "status": "online",
    "error": "",
    "max_amount": 20000,
    "max_bill": 200,
    "over_quota": false,
    "weight": 0,
    "alipay": null,
    "wxpay": null,
    "today_bill": 0,
    "today_amount": 0
  }
]
```

### Get
`GET /api/adb_devices/{device_id}`  -  query one given adb device
  
Example Request:
```liquid
GET /api/adb_devices/7504a96528b15e69 HTTP/1.1
```

Example Response:  
```json
similar to one of Listed element
```

### Update
`PATCH /api/adb_devices/{device_id}`  -  update one adb device description
  
Example Request:
```liquid
PATCH /api/adb_devices/7504a96528b15e69 HTTP/1.1

Content-Type: application/json

{
  "desc": "description text ...",
}
```

### Set Weight
`PUT /api/adb_devices/{device_id}/weight?val={value}`  -  set adb device weight

### Set Amount
`PUT /api/adb_devices/{device_id}/amount?val={value}`  -  set adb device max amount perday, by CNY

### Set Bill
`PUT /api/adb_devices/{device_id}/bill?val={value}`  -  set adb device max bill perday

### Bind Alipay
`PUT /api/adb_devices/{device_id}/alipay`  -  bind aliapy account to one adb device

Example Request:
```liquid
PUT /api/adb_devices/7504a96528b15e69 HTTP/1.1

Content-Type: application/json

{
  "user_id": "2088032017360044",
  "username": "13619840773",
  "nickname": "sldzz"
}
```

### Revoke Alipay
`DELETE /api/adb_devices/{device_id}/alipay`  -  revoke aliapy account from one adb device

### Verify
`GET /api/adb_devices/{device_id}/verify`  -  generate qrcode image for manually verify the adb device pay charging

Query Parameters:
  - **fee**       - fee, by RMB cent

### ScreenCap
`GET /api/adb_devices/{device_id}/screencap`  -  get screencap image on adb device

### UINodes
`GET /api/adb_devices/{device_id}/uinodes`  -  get current ui nodes of adb device

Example Response:
```json
[
  {
    "index": "0",
    "text": "我的",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[31,72][103,121]",
    "xy": [
      67,
      96
    ]
  },
  {
    "index": "0",
    "text": "设置",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[617,72][689,122]",
    "xy": [
      653,
      97
    ]
  },
  {
    "index": "0",
    "text": "bbk-ng",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[166,145][287,193]",
    "xy": [
      226,
      169
    ]
  },
  {
    "index": "1",
    "text": "bbklab@qq.com",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[166,205][410,249]",
    "xy": [
      288,
      227
    ]
  },
  {
    "index": "0",
    "text": "支付宝会员",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,301][284,357]",
    "xy": [
      204,
      329
    ]
  },
  {
    "index": "0",
    "text": "堆堆乐复活卡限时兑",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[304,310][609,348]",
    "xy": [
      456,
      329
    ]
  },
  {
    "index": "0",
    "text": "账单",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,409][188,465]",
    "xy": [
      156,
      437
    ]
  },
  {
    "index": "0",
    "text": "总资产",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,501][220,557]",
    "xy": [
      172,
      529
    ]
  },
  {
    "index": "0",
    "text": "余额",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,593][188,649]",
    "xy": [
      156,
      621
    ]
  },
  {
    "index": "0",
    "text": "0.01 元",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[208,602][649,640]",
    "xy": [
      428,
      621
    ]
  },
  {
    "index": "0",
    "text": "余额宝",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,685][220,741]",
    "xy": [
      172,
      713
    ]
  },
  {
    "index": "0",
    "text": "银行卡",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,777][220,833]",
    "xy": [
      172,
      805
    ]
  },
  {
    "index": "0",
    "text": "芝麻信用",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,885][252,941]",
    "xy": [
      188,
      913
    ]
  },
  {
    "index": "0",
    "text": "蚂蚁保险",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,977][252,1033]",
    "xy": [
      188,
      1005
    ]
  },
  {
    "index": "0",
    "text": "网商银行",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[124,1069][252,1125]",
    "xy": [
      188,
      1097
    ]
  },
  {
    "index": "0",
    "text": "首页",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[45,1186][98,1266]",
    "xy": [
      71,
      1226
    ]
  },
  {
    "index": "0",
    "text": "财富",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[189,1186][242,1266]",
    "xy": [
      215,
      1226
    ]
  },
  {
    "index": "0",
    "text": "口碑",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[333,1186][386,1266]",
    "xy": [
      359,
      1226
    ]
  },
  {
    "index": "0",
    "text": "朋友",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[477,1186][530,1266]",
    "xy": [
      503,
      1226
    ]
  },
  {
    "index": "0",
    "text": "我的",
    "resource_id": "",
    "package": "com.eg.android.AlipayGphone",
    "content_desc": "",
    "bounds": "[621,1186][674,1266]",
    "xy": [
      647,
      1226
    ]
  }
]
```

### Click
`PATCH /api/adb_devices/{device_id}/click`  -  click adb device UI Coordinate

Query Parameters:
  - **x**    - must: Coordinate X value
  - **y**    - must: Coordinate Y value

### Goback
`PATCH /api/adb_devices/{device_id}/goback`  -  tap adb device back key

### GotoHome
`PATCH /api/adb_devices/{device_id}/gotohome`  -  tap adb device home key

### Reboot
`PATCH /api/adb_devices/{device_id}/reboot`  -  reboot node adb device


### Remove
> TODO
