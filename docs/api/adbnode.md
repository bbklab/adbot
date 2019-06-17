
## AdbNode API

### List
`GET /api/adb_nodes`  -  list current adb nodes

Query Parameters:
  - **node_id**            - optional: list provided `id` nodes
  - **status**             - optional: list provided `status` orders
  - **remote**             - optional: regexp match remote ip
  - **hostname**           - optional: regexp match hostname
  - **with_master**        - optional: true|false, with master or not
  - **labels**             - optional: labels filter, format: key1=val1,key2=val2,key3=val3,...
  - **offset**             - optional: paging parameter, default 0
  - **limit**              - optional: paging parameter, default 20

Example Request:
```liquid
GET /api/adb_nodes HTTP/1.1
```

Example Response:
```json
response contains Header: `Total-Records`

[
  {
    "node": {
      "id": "4a264c130cde9319",
      "status": "online",       // online, offline, flagging
      "version": "1.0.0-beta",
      "error": "",
      "remote_addr": "115.60.5.118:8652",
      "geoinfo": {              // 地理位置英文
        "ip": "115.60.5.118",
        "continent": "Asia",
        "country": "China",
        "country_iso": "CN",
        "city": "Zhengzhou,Henan",
        "timezone": "Asia/Shanghai",
        "orgnization": "CHINA UNICOM China169 Backbone"
      },
      "geoinfo_zh": {           // 地理位置中文
        "ip": "115.60.5.118",
        "continent": "亚洲",
        "country": "中国",
        "country_iso": "CN",
        "city": "郑州,河南",
        "timezone": "Asia/Shanghai",
        "orgnization": "CHINA UNICOM China169 Backbone"
      },
      "sysinfo": {
        "hostname": "bbklab",     // 主机名
        "os": "Ubuntu 18.04.2 LTS",
        "kernel": "4.15.0-51-generic",
        "uptime": "648520.000000",
        "uptime_int": 648520,
        "unixtime": 1560430255,
        "loadavgs": {
          "one": 0.05,
          "five": 0.11,
          "fifteen": 0.15
        },
        "cpu": {
          "processor": 4,
          "physical": 4,
          "used": 0
        },
        "memory": {
          "total": 3925020672,
          "used": 1887739904,
          "cached": 1775230976
        },
        "swap": {
          "total": 2046291968,
          "used": 4456448,
          "free": 2041835520
        },
        "user": {
          "uid": "0",
          "gid": "0",
          "name": "root",
          "sudo": false
        },
        "ips": {
          "wlp2s0": [
            "192.168.0.111"
          ]
        },
        "disks": {
          "/dev/sda2": {
            "dev_name": "/dev/sda2",
            "mount_at": "/",
            "total": 39185848,
            "used": 33639808,
            "free": 5546040,
            "inode": 2498560,
            "ifree": 2065539
          },
          "/dev/sda3": {
            "dev_name": "/dev/sda3",
            "mount_at": "/data",
            "total": 73961680,
            "used": 28774020,
            "free": 45187660,
            "inode": 4710400,
            "ifree": 4534074
          }
        },
        "disksio": {
          "sda": {
            "dev_name": "sda",
            "read_bytes": 75413504,
            "write_bytes": 24725198848
          },
          "sda1": {
            "dev_name": "sda1",
            "read_bytes": 309248,
            "write_bytes": 679936
          },
          "sda2": {
            "dev_name": "sda2",
            "read_bytes": 72941568,
            "write_bytes": 24723367936
          },
          "sda3": {
            "dev_name": "sda3",
            "read_bytes": 2021376,
            "write_bytes": 696320
          },
          "sr0": {
            "dev_name": "sr0",
            "read_bytes": 38912,
            "write_bytes": 0
          }
        "traffics": {
          "wlp2s0": {
            "name": "wlp2s0",
            "mac": "8c:a9:82:6d:86:58",
            "rx_bytes": 743825478,
            "tx_bytes": 128021423,
            "rx_packets": 1656594,
            "tx_packets": 611420,
            "rx_rate": 79,
            "tx_rate": 179,
            "time": "2019-06-13T20:50:56.903+08:00"
          }
        },
        "docker": {
          "version": "18.09.1",
          "num_images": 29,
          "num_containers": 1,
          "num_running_containers": 0,
          "driver": "aufs",
          "driver_status": {
            "Backing Filesystem": "extfs",
            "Dirperm1 Supported": "true",
            "Dirs": "128",
            "Root Dir": "/data/docker/aufs"
          }
        },
        "bbr_enabled": false,
        "with_master": false,
        "manufacturer": "3249A64 (LENOVO)"
      },
      "ssh_config": null,
      "labels": {},
      "join_at": "2019-06-13T17:19:53.58+08:00",
      "last_active_at": "2019-06-13T20:51:23.59+08:00",
      "latency": 55082964,
      "inst_job": "",
      "remote_ip": "115.60.5.118",                // 外部IP地址
      "hwinfo": "4C - 3.66G - 3249A64 (LENOVO)"   // 硬件信息简要
    },
    "num_devices": 2,       // adb设备总数 
    "num_online": 2,        // adb设备在线数
    "num_offline": 0        // adb设备离线数
  }
]
```

### Get
`GET /api/adb_nodes/{node_id}`  -  query one given adb node
  
Example Request:
```liquid
GET /api/adb_nodes/1795bebb16ca5e9c HTTP/1.1
```

Example Response:  
```json
similar to one of Listed element
```
