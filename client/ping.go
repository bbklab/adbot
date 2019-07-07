package client

import (
	"errors"
	"fmt"
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

// NodeJoinCheck implement Client interface
func (c *AdbotClient) NodeJoinCheck(id string) error {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/nodes/join_check?node_id=%s", id), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, _ := ioutil.ReadAll(resp.Body)
	switch code := resp.StatusCode; code {
	case 403:
		return errors.New(string(bs))
	case 202:
		return nil
	default:
		return &APIError{code, string(bs)}
	}
}
