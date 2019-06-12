package client

import (
	"fmt"
	"io"
	"io/ioutil"
)

// PProfData implement Client interface
func (c *AdbotClient) PProfData(name string, seconds int) (io.ReadCloser, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/pprof/%s?seconds=%d", name, seconds), nil, 0, "", "")
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	return resp.Body, nil
}
