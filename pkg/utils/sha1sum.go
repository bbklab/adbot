package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
)

// Sha1sum simlar to /usr/bin/sha1sum
func Sha1sum(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// FileSha1sum compute the SHA1 digest of given file path
func FileSha1sum(path string) (string, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Sha1sum(bs), nil
}
