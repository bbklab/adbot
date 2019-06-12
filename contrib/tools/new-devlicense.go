package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	fullModuleLicenseReq = map[string]interface{}{
		"customer":   "dev.env",
		"reseller":   "bbk",                                             // peer license server must already created it
		"module":     255,                                               // full actived modules
		"max_nodes":  10000,                                             // 10000 nodes
		"expired_at": time.Now().Add(time.Hour * 24 * time.Duration(3)), // 3day expire
	}
)

func main() {
	data, err := generateDevLicense(fullModuleLicenseReq)
	if err != nil {
		log.Fatalln(err)
	}

	os.Remove("naive-license.pem")

	err = ioutil.WriteFile("naive-license.pem", []byte(data), os.FileMode(0400))
	if err != nil {
		log.Fatalln(err)
	}

	os.Stdout.WriteString("+OK\r\n")
}

// get dev test license
func generateDevLicense(licReq interface{}) (string, error) {
	var (
		// just make detection a little harder
		AccessToken = deobfuscate("599dgd28ge7c:89ee:e3e436b99458:3d6851::39488e65c2f47b319:f5e9:ce19b87gc7f8495163")
		LicURL      = deobfuscate("iuuqt;00cclmbc/nf;95550bqj0mjdfotft")
	)

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(licReq); err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", LicURL, buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("License-Access-Token", AccessToken)

	client := &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if code := resp.StatusCode; code != 201 {
		return "", fmt.Errorf("%d - %s", code, string(bs))
	}

	return string(bs), nil
}

func deobfuscate(s string) string {
	var clear string
	for i := 0; i < len(s); i++ {
		clear += string(int(s[i]) - 1)
	}
	return clear
}
