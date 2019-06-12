package utils

import (
	"bytes"
	"io/ioutil"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func isUTF8(b []byte) bool {
	return utf8.Valid(b)
}

// GBK2UTF8 is exported
func GBK2UTF8(b []byte) []byte {
	if isUTF8(b) {
		return b
	}
	reader := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return b
	}
	return bs
}

// UTF82GBK is exported
func UTF82GBK(b []byte) []byte {
	if !isUTF8(b) {
		return b
	}
	reader := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewEncoder())
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return b
	}
	return bs
}
