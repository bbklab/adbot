package client

import (
	"io/ioutil"
	"strings"
)

// Panic implement Client interface
func (c *AdbotClient) Panic() error {
	c.SetHeader("Panic-Secret-Token", "7fe64e8c4d89a7a2d204c2f9df9ef5345d95d9fa")

	resp, err := c.sendRequest("GET", "/api/panic", nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, _ := ioutil.ReadAll(resp.Body)

	if code := resp.StatusCode; code != 500 || !strings.Contains(string(bs), "the panic sucks") {
		return &APIError{code, string(bs)}
	}

	return nil
}
