package utils

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
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

// GetHTTPRealIP get the real source ip from http request
func GetHTTPRealIP(req *http.Request) string {
	var (
		xff = req.Header.Get("X-Forwarded-For")
		xri = req.Header.Get("X-Real-IP")
		pci = req.Header.Get("Proxy-Client-Ip")
	)

	switch {

	case xff != "":
		for _, addr := range strings.Split(xff, ",") {
			addr = strings.TrimSpace(addr)
			if ip := net.ParseIP(addr); ip != nil {
				return addr
			}
		}

	case xri != "":
		if ip := net.ParseIP(xri); ip != nil {
			return xri
		}

	case pci != "":
		if ip := net.ParseIP(pci); ip != nil {
			return pci
		}
	}

	return ""
}
