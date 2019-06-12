package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// GenSSHKeypair generate a new ssh private-public key pair
func GenSSHKeypair() (privBytes, pubBytes []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err == nil {
		err = priv.Validate()
	}
	if err != nil {
		return nil, nil, err
	}

	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}
	privBytes = pem.EncodeToMemory(&privBlk)

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	pubKeyBytes := pub.Marshal()
	pubBytes = make([]byte, base64.StdEncoding.EncodedLen(len(pubKeyBytes)))
	base64.StdEncoding.Encode(pubBytes, pubKeyBytes)
	pubBytes = append([]byte("ssh-rsa "), pubBytes...)
	pubBytes = append(pubBytes, []byte(" robot@bbklab.net")...)
	return
}
