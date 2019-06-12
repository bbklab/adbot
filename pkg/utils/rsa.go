package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"github.com/bbklab/adbot/pkg/file"
)

// LoadRSAPrivateKey load given path file or text to *rsa.PrivateKey
func LoadRSAPrivateKey(fileOrText string) (*rsa.PrivateKey, error) {
	var (
		bs  []byte
		err error
	)

	if file.Exists(fileOrText) {
		bs, err = ioutil.ReadFile(fileOrText)
		if err != nil {
			return nil, err
		}
	} else {
		bs = []byte(fileOrText)
	}

	if len(bs) == 0 {
		return nil, errors.New("empty RSA Private Key")
	}

	block, _ := pem.Decode(bs)
	if block == nil {
		return nil, errors.New("no PEM encoded key found")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	if err = key.Validate(); err != nil {
		return nil, err
	}

	return key, nil
}

// LoadRSAPublicKey load given path file or text to *rsa.PublicKey
func LoadRSAPublicKey(fileOrText string) (*rsa.PublicKey, error) {
	var (
		bs  []byte
		err error
	)

	if file.Exists(fileOrText) {
		bs, err = ioutil.ReadFile(fileOrText)
		if err != nil {
			return nil, err
		}
	} else {
		bs = []byte(fileOrText)
	}

	if len(bs) == 0 {
		return nil, errors.New("empty RSA Public Key")
	}

	block, _ := pem.Decode(bs)
	if block == nil {
		return nil, errors.New("no PEM encoded key found")
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// GenerateRSAKeyPairs generate rsa key (private/public) and marshal as PEM bytes
func GenerateRSAKeyPairs() ([]byte, []byte, error) {
	// note: the generated `key` object already contains the corresponding public key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	var (
		pribs = marshalRSAPrivateKey(key)           // private key format
		pubbs = marshalRSAPublicKey(&key.PublicKey) // public key format
	)

	return pribs, pubbs, nil
}

func marshalRSAPrivateKey(key *rsa.PrivateKey) []byte {
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	return pemdata
}

func marshalRSAPublicKey(key *rsa.PublicKey) []byte {
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(key),
		},
	)

	return pemdata
}
