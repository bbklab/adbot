package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
)

// Md5sum simlar to /usr/bin/md5sum
func Md5sum(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// FileMd5sum compute the MD5 digest of given file path
func FileMd5sum(path string) (string, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Md5sum(bs), nil
}
