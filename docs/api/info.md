
## Info API

`GET /api/info`  -  query summary informations
  
Example Request:
```liquid
GET /api/info HTTP/1.1
```

Example Response:
```json
{
  "version": "1.0.0-beta-de7b1f9",
  "listens": [
    "/var/run/adbot/adbot.sock",
    "0.0.0.0:8008"
  ],
  "uptime": "18m30.828631195s",
  "store_type": "mongodb",
  "adb_nodes": {  // 分控节点
    "total": 1,
    "online": 1,
    "offline": 0
  },
  "adb_devices": {  // 设备
    "total": 2,
    "online": 1,
    "offline": 1,
    "over_quota": 0,
    "within_quota": 2
  },
  "adb_orders": {  // 订单
    "total": {     // 总计
      "paid": 38,
      "paid_bill": 15.34,
      "pending": 0,
      "pending_bill": 0,
      "timeout": 6,
      "timeout_bill": 3.65
    },
    "today": {   // 今日
      "paid": 1,
      "paid_bill": 0.19,
      "pending": 0,
      "pending_bill": 0,
      "timeout": 1,
      "timeout_bill": 0.19
    },
    "month": {  // 本月
      "paid": 38,
      "paid_bill": 15.34,
      "pending": 0,
      "pending_bill": 0,
      "timeout": 6,
      "timeout_bill": 3.65
    }
  }
}
```
