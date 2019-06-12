package client

import (
	"io/ioutil"
	"time"
)

// Ping implement Client interface
func (c *AdbotClient) Ping() error {
	resp, err := c.sendRequest("GET", "/api/ping", nil, time.Second*10, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var (
		code  = resp.StatusCode
		bs, _ = ioutil.ReadAll(resp.Body)
	)

	if code != 200 || string(bs) != "OK" {
		return &APIError{code, string(bs)}
	}

	return nil
}

// QueryLeader implement Client interface
func (c *AdbotClient) QueryLeader() (string, bool) {
	resp, err := c.sendRequest("GET", "/api/query_leader", nil, time.Second*10, "", "")
	if err != nil {
		return err.Error(), false
	}
	defer resp.Body.Close()

	bs, _ := ioutil.ReadAll(resp.Body)
	return string(bs), resp.StatusCode == 200
}
