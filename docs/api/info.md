
## Info API

`GET /api/info`  -  query summary informations
  
Example Request:
```liquid
GET /api/info HTTP/1.1
```

Example Response:
```json
{
  "version": "1.0.0-beta-7444c96",
  "listens": [
    "/var/run/adbot/adbot.sock",
    "0.0.0.0:8008"
  ],
  "uptime": "4.945601886s",
  "store_type": "mongodb",
  "adb_nodes": {
    "total": 1,
    "online": 0,
    "offline": 1
  },
  "adb_devices": {
    "total": 2,
    "online": 0,
    "offline": 2,
    "over_quota": 0,
    "within_quota": 2
  }
}
```
