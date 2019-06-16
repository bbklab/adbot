
## Order API

### List
`GET /api/adb_orders`  -  list all adb orders
  
Query Parameters:
  - **search**             - optional: search provided `order id` or `out order id`
  - **order_id**           - optional: list provided `order id` orders
  - **out_order_id**       - optional: list provided `out order id` orders
  - **status**             - optional: list provided `status` orders
  - **cbstatus**           - optional: list provided `callback_status` orders
  - **device_id**          - optional: list provided `device_id` orders
  - **start_at**           - optional: list created_at time > `start_at` orders, time format RFC3339, eg: 2019-04-24T17:45:24+08:00
  - **end_at**             - optional: list created_at time < `end_at` orders, time format RFC3339,  eg: 2019-04-24T17:45:56+08:00
  - **offset**             - optional: paging parameter, default 0
  - **limit**              - optional: paging parameter, default 20

Example Request:
```liquid
GET /api/adb_orders HTTP/1.1

OR

GET /api/adb_orders?device_id=546052d21f384 HTTP/1.1
```

Example Response:
```json
response contains Header: `Total-Records`

[
  {
    "id": "201961711914-cc4b76",   // 订单ID
    "status": "pending",           // 支付状态 pending, paid, timeout
    "node_id": "4a264c130cde9319", 
    "device_id": "546052d21f384",  // 收款设备ID
    "out_order_id": "000003",      // 外部商户的订单ID
    "qrtype": "alipay",            // alipay: 支付宝支付 wxpay: 微信支付
    "fee": 1,                      // 收款金额，单位RMB分
    "attach": "",
    "notify_url": "http://requestbin.net/r/1a228471",   // 订单回调地址
    "response": {
      "code": 1,
      "message": "",
      "qrtext": "alipays://platformapi/startapp?appId=20000123\u0026actionType=scan\u0026biz_data={\"s\":\"money\",\"u\":\"2088032017360044\",\"a\":\"0.01\",\"m\":\"201961711914-cc4b76\"}",
      "qrimage": "",
      "time": "2019-06-17T01:19:14.759+08:00"
    },
    "callback": null,
    "callback_status": "none",       // none(未启动), ongoing(进行中), succeed(成功), error(错误), aborted(重启中断)
    "callback_history": [            // 回调发送历史
      
    ],
    "created_at": "2019-06-17T01:19:14.73+08:00",
    "paid_at": "0001-01-01T00:00:00Z",
    "fee_yuan": 0.01
  },
  {
    "id": "201961711523-734c07",
    "status": "paid",
    "node_id": "4a264c130cde9319",
    "device_id": "546052d21f384",
    "out_order_id": "000002",
    "qrtype": "alipay",
    "fee": 1,
    "attach": "",
    "notify_url": "http://requestbin.net/r/1a228471",
    "response": {
      "code": 1,
      "message": "",
      "qrtext": "alipays://platformapi/startapp?appId=20000123\u0026actionType=scan\u0026biz_data={\"s\":\"money\",\"u\":\"2088032017360044\",\"a\":\"0.01\",\"m\":\"201961711523-734c07\"}",
      "qrimage": "",
      "time": "2019-06-17T01:15:23.909+08:00"
    },
    "callback": {
      "code": 1,
      "out_order_id": "000002",
      "fee": 1,
      "attach": "",
      "time": "2019-06-17T01:17:30.064+08:00"
    },
    "callback_status": "succeed",
    "callback_history": [
      "2019-06-17T01:17:31+08:00: succeed"
    ],
    "created_at": "2019-06-17T01:15:23.88+08:00",
    "paid_at": "2019-06-17T01:17:30.063+08:00",
    "fee_yuan": 0.01
  }
]
```

### Get
`GET /api/adb_orders/{order_id}`  -  query one given adb order
  
Example Request:
```liquid
GET /api/adb_orders/201961711914-cc4b76 HTTP/1.1
```

Example Response:  
```json
similar to one of Listed element
```

### ReCallback
`PUT /api/adb_orders/{order_id}/recallback`  -  resend callback of one adb order
