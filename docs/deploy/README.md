
## 部署维护文档
  - [部署](#部署)
    + [配置要求](#配置要求)
    + [准备](#准备)
    + [安装](#安装)
    + [启动](#启动)
    + [增加分控](#增加分控)
  - [维护](#维护)
    + [Mongod](#mongod)
    + [Adbot-Master](#master)
    + [Adbot-Node](#node)
  - [备份](#备份)

### 部署

#### 配置要求
  - OS     :  CentOS 7 x86_64 (better with X window)
  - CPU    :  4C+
  - Memory :  8G+
  - Disk   :  System: 20G+, Data: 100G+

#### 准备
```bash
setenforce 0
sed -i 's/enforcing/disabled/' /etc/selinux/config
systemctl stop postfix
systemctl disable postfix
systemctl stop firewalld
systemctl disable firewalld
iptables -F
iptables -X
yum makecache fast
yum -y install epel-release
yum -y install bind-utils net-tools file telnet nmap wget jq socat
```

#### 安装
> 到发布页面获取最新版本RPM安装包[Release Page](https://github.com/bbklab/adbot/releases)
```bash
rpm -ivh adbot-geolite2-1.0.0-rhel7.x86_64.rpm
rpm -ivh adbot-dependency-1.0.0-rhel7.x86_64.rpm
rpm -ivh adbot-master-latest-rhel7.x86_64.rpm
rpm -ivh /usr/share/adbot/dependency/mongod.pkg

cat > /etc/adbot/master.env << EOF
LISTEN_ADDR=127.0.0.1:8008
TLS_CERT_FILE=
TLS_KEY_FILE=
DB_TYPE=mongodb
MGO_URL=mongodb://127.0.0.1:27017/adbot
EOF
```
> 可选: 安装 [docker headless vnc container](/docs/deploy/container-vncd.md#install)

#### 启动
```bash
systemctl start  mongod
systemctl enable mongod
systemctl start  adbot-master
systemctl enable adbot-master

adbot user create  --name admin --password 8bb306f380521aba3
adbot login -u admin -p 8bb306f380521aba3

adbot info
adbot settings show
```
> 更多命令行见 [adbot CLI](/docs/cli/README.md)

#### 增加分控
  - 准备一台主机，带USB3.0+接口，安装操作系统CentOS 7 x86_64
  - 主机挂接一台10口**带独立电源**的USBHub，同样要支持USB3.0+
  - 将手机通过USB3.0数据线接入USBHub，并依照[手机设置文档](https://github.com/bbklab/adbot/wiki/Redmi-4A%E6%89%8B%E6%9C%BA%E8%AE%BE%E7%BD%AE%E6%AD%A5%E9%AA%A4)开启手机的USB调试和其他设置
  - 将主控上的文件`/usr/share/adbot/agent.pkg`拷贝到新准备的分控主机
  - 在分控主机上安装分控程序并启动:
```bash
rpm -ivh agent.pkg

cat > /etc/adbot/agent.env << EOF
JOIN_ADDRS=adb.master.fqdn.hostname:8008
EOF

systemctl  start adbot-agent
systemctl  enable adbot-agent
```
  - 待分控程序启动后，可以到主控上执行`adbot node ls`和`adbot adb-node ls`查看新加入的分控节点

### 维护

#### Mongod
  - 端口监听:       127.0.0.1:27017
  - 配置文件:       /etc/mongod.conf
  - 系统服务:       mongod.service (enabled)
  - 数据存储:       /var/lib/mongo
  - 运行日志:       journalctl -u mongod
  - 其他日志:       /var/log/mongodb

#### Master
  - 端口监听:       127.0.0.1:8008
  - 二进制:         /usr/bin/adbot
  - 配置文件:       /etc/adbot/master.env
  - 系统服务:       adbot-master.service (enabled)
  - 数据存储:       mongodb storage
  - 运行日志:       journalctl -u adbot-master
  - API审计日志:    /var/log/adbot-audit

#### Node
> 分控节点的维护比较特殊，通常情况下分控节点并不和主控部署在一起，而可能是在任意地理位置的一台主机  
> 只要分控连接上了主控，并且在线的情况下，可以通过主控的命令行CLI: **adbot node terminal**通过反弹Shell  
> 登录到分控节点进行基本的维护作业  

  - 二进制:         /usr/bin/adbot-agent
  - 配置文件:       /etc/adbot/agent.env
  - 系统服务:       adbot-agent.service (enabled)
  - 运行日志:       journalctl -u adbot-agent

### 备份
  - /var/lib/mongo
