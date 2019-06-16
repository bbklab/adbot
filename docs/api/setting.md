
## Setting API

### Get
`GET /api/settings`  -  query current global settings
  
Example Request:
```liquid
GET /api/settings HTTP/1.1
```

Example Response:
```json
{
    "log_level": "info",
    "enable_httpmux_debug": false,
    "unmask_sensitive": false,
    "tg_bot_token": "",
    "global_attrs": {
        "com_adbbot_paygate_secret": "04409f6be80c3b10d200905532e93dax"
    },
    "updated_at": "2019-06-17T00:16:45.548+08:00",
    "initial": false
}
```

### Update
`PATCH /api/settings`  -  update current global settings

Example Request:
```liquid
PATCH /api/settings HTTP/1.1

Content-Type: application/json

{
  "log_level": "debug",
  "tg_bot_token": "ED7DCE1814334E7DEB3ED93BC7A36B1F"
}
```
