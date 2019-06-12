package utils

import (
	"crypto/tls"
	"net/http"
	"sync"
)

var (
	once sync.Once
	cli  *http.Client
)

// InsecureHTTPClient return a runtime *http.Client with InsecureSkipVerify == true
func InsecureHTTPClient() *http.Client {
	once.Do(func() {
		cli = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})
	return cli
}
