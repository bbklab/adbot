
## Order API

### List
`GET /api/orders`  -  list all (or merchant) orders
  
Query Parameters:
  - **search**             - optional: search provided `order id` or `merchant id` or `merchant order id`
  - **order_id**           - optional: list provided `order id` orders
  - **merchant_order_id**  - optional: list provided `merchant order id` orders
  - **merchant_id**        - optional: list provided `merchant id or merchant name` orders
  - **status**             - optional: list provided `status` orders
  - **cbstatus**           - optional: list provided `callback_status` orders
  - **payway_id**          - optional: list provided `payway_id` orders
  - **start_at**           - optional: list created_at time > `start_at` orders, time format RFC3339, eg: 2019-04-24T17:45:24+08:00
  - **end_at**             - optional: list created_at time < `end_at` orders, time format RFC3339,  eg: 2019-04-24T17:45:56+08:00
  - **offset**             - optional: paging parameter, default 0
  - **limit**              - optional: paging parameter, default 20

Example Request:
```liquid
GET /api/orders HTTP/1.1

OR

GET /api/orders?merchant_id=c37faba8e084e047 HTTP/1.1
```

Example Response:
```json
response contains Header: `Total-Records`

[
  {
    "id": "b26aa5ddbaa5a5da",
    "status": "paid",   // pending(未支付), paid(已支付), error(错误)
    "payway_id": "7504a96528b15e69",
    "payway_name": "PayJS(微信扫码)",
    "merch_id": "1795bebb16ca5e9c",
    "merch_order_id": "006",
    "qrtype": "wxpay", // alipay(支付宝支付), wxpay(微信支付)
    "fee": 1,
    "title": "收款0.01元",
    "attach": "Attach Datas",
    "notify_url": "http://requestbin.fullcontact.com/1d0vus71",
    "sign": "ED7DCE1814334E7DEB3ED93BC7A36B1F",
    "response": {
      "code": 1,
      "message": "",
      "qrcode": "https://payjs.cn/qrcode/d2VpeGluOi8vd3hwYXkvYml6cGF5dXJsP3ByPTRsZlVwRGs=",
      "time": "2019-04-07T11:00:16.539+08:00"
    },
    "callback": {
      "code": 1,
      "error": "",
      "merch_id": "1795bebb16ca5e9c",
      "merch_order_id": "006",
      "fee": 1,
      "attach": "Attach Datas",
      "time": "2019-04-07T11:01:55.591+08:00"
    },
    "callback_history": [
      "2019-04-07T11:00:47+08:00: 409 - \u003c!DOCTYPE html\u003e\n\u003c!-- ...",
      "2019-04-07T11:00:57+08:00: 409 - \u003c!DOCTYPE html\u003e\n\u003c!-- ...",
      "2019-04-07T11:01:28+08:00: 409 - \u003c!DOCTYPE html\u003e\n\u003c!-- ...",
      "2019-04-07T11:01:32+08:00: callback aborted while restart",
      "2019-04-07T11:01:33+08:00: 409 - \u003c!DOCTYPE html\u003e\n\u003c!-- ...",
      "2019-04-07T11:01:43+08:00: 409 - \u003c!DOCTYPE html\u003e\n\u003c!-- ...",
      "2019-04-07T11:01:55+08:00: callback aborted while restart",
      "2019-04-07T11:01:57+08:00: succeed"
    ],
    "callback_status": "succeed",  // none(未启动), ongoing(进行中), succeed(成功), error(错误), aborted(重启中断)
    "inner_request": {
      "order_id": "b26aa5ddbaa5a5da",
      "fee": 1,
      "title": "收款0.01元",
      "comment": "Attach Datas",
      "time": "2019-04-07T11:00:16.057+08:00"
    },
    "inner_response": {
      "paysys_order_id": "2019040711001600442305617",
      "message": "SUCCESS",
      "qrcode": "https://payjs.cn/qrcode/d2VpeGluOi8vd3hwYXkvYml6cGF5dXJsP3ByPTRsZlVwRGs=",
      "annotations": {
        "code_url": "weixin://wxpay/bizpayurl?pr=4lfUpDk",
        "fee": 1,
        "msg": "",
        "out_trade_no": "b26aa5ddbaa5a5da",
        "return_code": 1,
        "signature": "76CF4150A71B6A97B211D3E0799751CD"
      },
      "time": "2019-04-07T11:00:16.539+08:00"
    },
    "inner_callback": {
      "callback": {
        "attach": "Attach Datas",
        "mchid": "1524564751",
        "openid": "o7LFAwTeHfcJfa0ST77KZRHk6tz4",
        "out_trade_no": "b26aa5ddbaa5a5da",
        "payjs_order_id": "2019040711001600442305617",
        "return_code": "1",
        "sign": "11A71A1761B5B866D4063849CBC945E6",
        "time_end": "2019-04-07 11:00:43",
        "total_fee": "1",
        "transaction_id": "4200000305201904070254331028"
      },
      "time": "2019-04-07T11:00:46.962+08:00"
    },
    "created_at": "2019-04-07T11:00:16.056+08:00",
    "updated_at": "2019-04-07T11:01:57.014+08:00",
    "paid_at": "2019-05-14T17:45:20.332+08:00"
  },
  {
    "id": "44f731082b8b515d",
    "status": "paid",
    "payway_id": "7504a96528b15e69",
    "payway_name": "PayJS(微信扫码)",
    "merch_id": "1795bebb16ca5e9c",
    "merch_order_id": "005",
    "qrtype": "wxpay",
    "fee": 1,
    "title": "收款0.01元",
    "attach": "Attach Datas",
    "notify_url": "http://requestbin.fullcontact.com/1d0vus71",
    "sign": "E55DE1A4835CEA0F1E1488C595A05C21",
    "response": {
      "code": 1,
      "message": "",
      "qrcode": "https://payjs.cn/qrcode/d2VpeGluOi8vd3hwYXkvYml6cGF5dXJsP3ByPVNxSjdNYW0=",
      "time": "2019-04-07T10:52:57.335+08:00"
    },
    "callback": {
      "code": 1,
      "error": "",
      "merch_id": "1795bebb16ca5e9c",
      "merch_order_id": "005",
      "fee": 1,
      "attach": "Attach Datas",
      "time": "2019-04-07T10:55:31.407+08:00"
    },
    "callback_history": [
      "2019-04-07T10:55:33+08:00: succeed"
    ],
    "callback_status": "succeed",
    "inner_request": {
      "order_id": "44f731082b8b515d",
      "fee": 1,
      "title": "收款0.01元",
      "comment": "Attach Datas",
      "time": "2019-04-07T10:52:56.733+08:00"
    },
    "inner_response": {
      "paysys_order_id": "2019040710525600193532688",
      "message": "SUCCESS",
      "qrcode": "https://payjs.cn/qrcode/d2VpeGluOi8vd3hwYXkvYml6cGF5dXJsP3ByPVNxSjdNYW0=",
      "annotations": {
        "code_url": "weixin://wxpay/bizpayurl?pr=SqJ7Mam",
        "fee": 1,
        "msg": "",
        "out_trade_no": "44f731082b8b515d",
        "return_code": 1,
        "signature": "7E3AB9CCAF2E7CA97234923041532AC1"
      },
      "time": "2019-04-07T10:52:57.335+08:00"
    },
    "inner_callback": {
      "callback": {
        "attach": "Attach Datas",
        "mchid": "1524564751",
        "openid": "o7LFAwTeHfcJfa0ST77KZRHk6tz4",
        "out_trade_no": "44f731082b8b515d",
        "payjs_order_id": "2019040710525600193532688",
        "return_code": "1",
        "sign": "42D123121E895022E5F1464E838FB476",
        "time_end": "2019-04-07 10:53:31",
        "total_fee": "1",
        "transaction_id": "4200000298201904078156089506"
      },
      "time": "2019-04-07T10:53:34.269+08:00"
    },
    "created_at": "2019-04-07T10:52:56.73+08:00",
    "updated_at": "2019-04-07T10:55:33.073+08:00",
    "paid_at": "2019-05-14T17:45:20.332+08:00"
  }
]
```

### Get
`GET /api/orders/{order_id}`  -  query one given order
  
Example Request:
```liquid
GET /api/orders/b26aa5ddbaa5a5da HTTP/1.1
```

Example Response:  
```json
similar to one of Listed element
```

### ReCallback
`PUT /api/orders/{order_id}/recallback`  -  resend callback of one order
