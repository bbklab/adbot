package session

import (
	sc "github.com/gorilla/securecookie"
)

var (
	hashKey  = []byte("some-thing-that-should-be-very-secret")
	blockKey = []byte("secret-length-is-32-means-AES256")
	h        = sc.New(hashKey, blockKey)
)

// Encode encodes given key, value -> token
func Encode(key, val string) (string, error) {
	return h.Encode(key, val)
}

// Decode decodes give token, key -> value
func Decode(encoded, key string) (string, error) {
	var val = new(string)
	err := h.Decode(key, encoded, val)
	if err != nil {
		return "", err
	}
	return *val, nil
}
