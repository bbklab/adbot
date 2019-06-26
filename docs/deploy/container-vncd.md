
### Install 
```bash
docker info >/dev/null 2>&1 || (
   yum -y install docker
   systemctl enable docker
   systemctl start docker
)

docker stop vncd || true
docker rm   vncd || true

mkdir -p /data/vncd

docker run -d --name vncd \
        --user 0 \
        --net=host \
        -v /data/vncd:/data \
        -e VNC_PW=my-vnc-9509345613c64f \
        consol/centos-xfce-vnc
```

### Usage
  - novnc html5 client:   `http://host.ip.address:6901`
  - default password:     `my-vnc-9509345613c64f`
  - NOTE, save data:      `/data`
