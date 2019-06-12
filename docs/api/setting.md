
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
  "tg_bot_token": "",
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
