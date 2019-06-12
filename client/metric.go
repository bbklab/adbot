package client

import (
	"io"
	"io/ioutil"
)

// Metrics implement Client interface
func (c *AdbotClient) Metrics() (io.ReadCloser, error) {
	settings, err := c.GetSettings()
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest("GET", "/api/metrics", nil, 0, settings.MetricsAuthUser, settings.MetricsAuthPassword)
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{code, string(bs)}
	}

	return resp.Body, nil
}
