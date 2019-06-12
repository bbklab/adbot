
## Info API

`GET /api/info`  -  query summary informations
  
Example Request:
```liquid
GET /api/info HTTP/1.1
```

Example Response:
```json
{
  "version": "1.0.0-beta-cb49fe6",
  "listens": [
    "/var/run/adbot/adbot.sock",
    "0.0.0.0:8008"
  ],
  "uptime": "2h20m6.477281416s",
  "store_type": "mongodb",
  "adb_nodes": {
    "total": 1,
    "online": 1,
    "offline": 0
  },
  "adb_devices": {
    "total": 2,
    "online": 2,
    "offline": 0,
    "over_quota": 0
  }
}
```
