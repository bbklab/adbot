#!/bin/bash
set -ex

# interactive: 
#  just press RETRUN
openssl genrsa -out ssl-key.pem

openssl req -new -key ssl-key.pem -out server.csr
openssl req -new -key ssl-key.pem -out server.csr -subj /C=CN/ST=BeiJing/L=BeiJing/O=BBKLAB/CN=localhost
openssl x509 -in server.csr -out ssl-cert.pem -req -signkey ssl-key.pem -days 36500

rm -f server.csr
echo "+OK"


#
# merge ssl-key.pem & ssl-cert.pem -> cert.pfx
#
# openssl pkcs12 -export -out cert.pfx -inkey ssl-key.pem -in ssl-cert.pem
