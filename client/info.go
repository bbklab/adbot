package client

import (
	"io/ioutil"

	"github.com/bbklab/adbot/types"
)

// Info implement Client interface
func (c *AdbotClient) Info() (*types.SummaryInfo, error) {
	resp, err := c.sendRequest("GET", "/api/info", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.SummaryInfo
	err = c.bind(resp.Body, &ret)
	return ret, err
}
