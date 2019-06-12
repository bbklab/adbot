package client

import (
	"errors"
	"io/ioutil"
	"net/url"

	"github.com/bbklab/adbot/debug"
	"github.com/bbklab/adbot/types"
)

// DebugDump implement Client interface
// parameter name:
//  - goroutine      -  []byte
//  - general        -  debug.Info
//  - application    -  map[string]interface{}
func (c *AdbotClient) DebugDump(name string) ([]byte, *debug.Info, *types.MasterConfig, map[string]interface{}, error) {
	resp, err := c.sendRequest("GET", "/api/debug/dump?name="+url.QueryEscape(name), nil, 0, "", "")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, nil, nil, nil, &APIError{code, string(bs)}
	}

	switch name {

	case "goroutine":
		bs, err := ioutil.ReadAll(resp.Body)
		return bs, nil, nil, nil, err

	case "general":
		var ret *debug.Info
		err = c.bind(resp.Body, &ret)
		return nil, ret, nil, nil, err

	case "config":
		var ret *types.MasterConfig
		err = c.bind(resp.Body, &ret)
		return nil, nil, ret, nil, err

	case "application":
		var ret map[string]interface{}
		err = c.bind(resp.Body, &ret)
		return nil, nil, nil, ret, err

	default:
		return nil, nil, nil, nil, errors.New("unsupported dump name")
	}
}
