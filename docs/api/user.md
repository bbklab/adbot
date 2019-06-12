
## Admin User API

### Any
`GET /api/users/any`  -  query if has any users

Example Request:
```liquid
GET /api/users/any HTTP/1.1
```

Example Response:
```json
{
  "result": true
}
```

### Add
`POST /api/users`  -  create super admin user
  
Example Request:
```json
POST /api/users HTTP/1.1

Content-Type: application/json

{
  "name": "admin",
  "password": "password",
  "desc": "the only privileged user",
}
```

Example Response:
```json
201 - succeed

{
  "id": "18fad7bdc1d7b448",
  "name": "admin",
  "password": "******",
  "desc": "",
  "created_at": "2017-12-05T13:03:14.328027644-05:00",
  "updated_at": "2017-12-05T13:03:14.328028249-05:00"
}
```

### Login
`POST /api/users/login`  -  admin user auth login

Example Request:
```json
POST /api/users/login HTTP/1.1

Content-Type: application/json

{
  "username": "admin",
  "password": "password",
}
```

Example Response:
```liquid
202:
Admin-Access-Token: MTU1NTA2NDU0MXx2YmVEMW0tU3ptczBIcHNQVzRoNUFHdWRmOTlfX0c3QUJTZkI3d05wWW5xc0JvUVlwOUhac3R5OVhFMEVmTllUWnUtR2hSTT18K9XM5r-Pyrx4lO9DWQitdDDoBBt6ie6duhckVxqY0vc=
```
> 管理员登录成功后返回202和Header **Admin-Access-Token**, 后续请求必须带上此Header

### Profile
`GET /api/users/profile`  -  get current user profile
  
Example Request:
```liquid
GET /api/users/profile HTTP/1.1

Admin-Access-Token: MTU1NTA2NDU0MXx2YmVEMW0tU3ptczBIcHNQVzRoNUFHdWRmOTlfX0c3QUJTZkI3d05wWW5xc0JvUVlwOUhac3R5OVhFMEVmTllUWnUtR2hSTT18K9XM5r-Pyrx4lO9DWQitdDDoBBt6ie6duhckVxqY0vc=
```

Example Response:  
```json
{
  "id": "7a4109ac54699fbe",
  "name": "admin",
  "password": "******",
  "desc": "",
  "created_at": "2019-04-02T11:50:54.105+08:00",
  "updated_at": "2019-04-12T17:01:50.814+08:00"
}
```

### Session
`GET /api/users/sessions`  -  get current user sessions
  
Example Request:
```liquid
GET /api/users/sessions HTTP/1.1

Admin-Access-Token: MTU1NTA2NDU0MXx2YmVEMW0tU3ptczBIcHNQVzRoNUFHdWRmOTlfX0c3QUJTZkI3d05wWW5xc0JvUVlwOUhac3R5OVhFMEVmTllUWnUtR2hSTT18K9XM5r-Pyrx4lO9DWQitdDDoBBt6ie6duhckVxqY0vc=
```

Example Response:  
```json
[
  {
    "id": "bcc4e9fd5b37777b",
    "user_id": "d31b526dc1b34bae",
    "remote": "101.200.55.244", // Session登录来源
    "geoinfo": {
      "ip": "67.216.206.57",
      "continent": "North America",
      "country": "United States",
      "country_iso": "US",
      "city": "Los Angeles,California",
      "timezone": "America/Los_Angeles",
      "orgnization": "IT7 Networks Inc"
    },
    "geoinfo_zh": {
      "ip": "67.216.206.57",
      "continent": "北美洲",
      "country": "美国",
      "country_iso": "US",
      "city": "洛杉矶,加利福尼亚州",
      "timezone": "America/Los_Angeles",
      "orgnization": "IT7 Networks Inc"
    },
    "device": "Macintosh",      // Session登录设备
    "os": "Mac OS version: X 10_6_3",
    "browser": "Safari version: 5.0",
    "last_active_at": "2019-04-11T12:05:08.903+08:00"
    "current": true      // 是否是当前登录的Session
  },
  {
    "id": "2842cc48678978d3",
    "user_id": "d31b526dc1b34bae",
    "remote": "101.200.55.244",
    "geoinfo": {
      "ip": "101.200.55.244",
      "continent": "North America",
      "country": "United States",
      "country_iso": "US",
      "city": "Los Angeles,California",
      "timezone": "America/Los_Angeles",
      "orgnization": "IT7 Networks Inc"
    },
    "geoinfo_zh": {
      "ip": "101.200.55.244",
      "continent": "北美洲",
      "country": "美国",
      "country_iso": "US",
      "city": "洛杉矶,加利福尼亚州",
      "timezone": "America/Los_Angeles",
      "orgnization": "IT7 Networks Inc"
    },
    "device": "",
    "os": "",
    "browser": "",
    "last_active_at": "2019-04-11T14:26:19.146+08:00",
    "current": false
  }
]
```

### Kick Session
`DELETE /api/users/sessions/{session_id}`  -  kick out given user session

### Change-Password
`PATCH /api/users/change_password`  -  change admin user`s password

Example Request:
```json
PATCH /api/users/change_password HTTP/1.1

Content-Type: application/json
Admin-Access-Token: MTU1NTA2NDU0MXx2YmVEMW0tU3ptczBIcHNQVzRoNUFHdWRmOTlfX0c3QUJTZkI3d05wWW5xc0JvUVlwOUhac3R5OVhFMEVmTllUWnUtR2hSTT18K9XM5r-Pyrx4lO9DWQitdDDoBBt6ie6duhckVxqY0vc=

{
  "old": "old-password",
  "new": "new-password",
}
```
> 修改密码成功后，所有session会被后端自动注销，所有API都返回401，管理员需要重新登录

### Logout
`DELETE /api/users/logout`  -  logout admin user

Example Request:
```liquid
DELETE /api/users/logout HTTP/1.1

Admin-Access-Token: MTU1NTA2NDU0MXx2YmVEMW0tU3ptczBIcHNQVzRoNUFHdWRmOTlfX0c3QUJTZkI3d05wWW5xc0JvUVlwOUhac3R5OVhFMEVmTllUWnUtR2hSTT18K9XM5r-Pyrx4lO9DWQitdDDoBBt6ie6duhckVxqY0vc=
```
