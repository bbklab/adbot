package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// GetTLSConfig generate a *tls.Config with given ca,cert,key file
func GetTLSConfig(fcert, fkey, fca string) (*tls.Config, error) {
	var (
		tlscfg = new(tls.Config)
	)

	// load cert
	cert, err := tls.LoadX509KeyPair(fcert, fkey)
	if err != nil {
		return nil, fmt.Errorf("load cert-key pair error: %v", err)
	}
	tlscfg.Certificates = []tls.Certificate{cert}

	// fca is optional
	if fca == "" {
		return tlscfg, nil
	}

	// load root ca
	ca, err := ioutil.ReadFile(fca)
	if err != nil {
		return nil, fmt.Errorf("load ca file error: %v", err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	tlscfg.RootCAs = pool

	return tlscfg, nil
}
