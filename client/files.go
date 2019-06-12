package client

import (
	"io"
	"io/ioutil"
	"net/url"
)

// FetchFile implement Client interface
// parameter name -> keys of scheduler.resFileMap
func (c *AdbotClient) FetchFile(name string) (io.ReadCloser, string, error) {
	resp, err := c.sendRequest("GET", "/api/files?name="+url.QueryEscape(name), nil, 0, "", "")
	if err != nil {
		return nil, "", err
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, "", &APIError{code, string(bs)}
	}

	return resp.Body, resp.Header.Get("Sha1sum"), nil
}
