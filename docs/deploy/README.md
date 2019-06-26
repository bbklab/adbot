
## 部署维护文档
  - [部署](#部署)
    + [配置要求](#配置要求)
    + [准备](#准备)
    + [安装](#安装)
    + [启动](#启动)
  - [维护](#维护)
    + [Mongod](#mongod)
    + [Adbot-Master](#adbot-master)
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
systemctl enable mongod

mkdir -p /etc/adbot
cat > /etc/adbot/master.env << EOF
LISTEN_ADDR=0.0.0.0:80
TLS_CERT_FILE=
TLS_KEY_FILE=
DB_TYPE=mongodb
MGO_URL=mongodb://127.0.0.1:27017/adbot
EOF
```
> 可选: 安装 [docker headless vnc container](/docs/deploy/container-vncd.md#install)

#### 启动
```bash
systemctl start mongod
systemctl start adbot-master

adbot user create  --name admin --password 8bb306f380521aba3
adbot login -u admin -p 8bb306f380521aba3

adbot info
adbot settings show
```
> 更多命令行见 [adbot CLI](/docs/cli/README.md)

### 维护

#### Mongod
  - 端口监听:       127.0.0.1:27017
  - 配置文件:       /etc/mongod.conf
  - 系统服务:       mongod.service (enabled)
  - 数据存储:       /var/lib/mongo

#### Master
  - 端口监听:       0.0.0.0:80
  - 二进制:         /usr/bin/adbot
  - 配置文件:       /etc/adbot/master.env
  - 系统服务:       adbot-master.service (enabled)
  - 数据存储:       mongodb storage

### 备份
  - /var/lib/mongo
