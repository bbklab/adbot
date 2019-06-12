package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/gorilla/websocket"

	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/ws"
	"github.com/bbklab/adbot/types"
)

// ListNodes implement Client interface
func (c *AdbotClient) ListNodes(lbsFilter label.Labels, online *bool, cldsvr string) ([]*types.NodeWrapper, error) {
	var lbsQuery string
	for k, v := range lbsFilter {
		if lbsQuery == "" {
			lbsQuery += fmt.Sprintf("%s=%s", k, v)
		} else {
			lbsQuery += fmt.Sprintf(",%s=%s", k, v)
		}
	}

	var uri = fmt.Sprintf("/api/nodes?labels=%s&cldsvr=%s&online=", url.QueryEscape(lbsQuery), cldsvr)
	if online != nil {
		uri += fmt.Sprintf("%t", *online)
	}

	resp, err := c.sendRequest("GET", uri, nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret []*types.NodeWrapper
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// InspectNode implement Client interface
func (c *AdbotClient) InspectNode(id string) (*types.NodeWrapper, error) {
	resp, err := c.sendRequest("GET", "/api/nodes/"+id, nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.NodeWrapper
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// WatchNodeStats implement Client interface
func (c *AdbotClient) WatchNodeStats(id string) (io.ReadCloser, error) {
	resp, err := c.sendRequest("GET", "/api/nodes/"+id+"/stats", nil, 0, "", "")
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

// RunNodeCmd implement Client interface
func (c *AdbotClient) RunNodeCmd(id, cmd string) (io.ReadCloser, error) {
	nodeCmd := types.NodeCmd{Command: cmd}
	resp, err := c.sendRequest("POST", "/api/nodes/"+id+"/exec", nodeCmd, 0, "", "")
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

// OpenNodeTerminal implement Client interface
func (c *AdbotClient) OpenNodeTerminal(id string, input io.Reader, output io.Writer) error {
	uri := fmt.Sprintf("ws://what-ever/api/nodes/%s/terminal", id)
	wsConn, _, err := c.wsDialer.Dial(uri, nil)
	if err != nil {
		return err
	}
	defer wsConn.Close()

	var (
		wsConnWrapper = ws.NewWrappedWsConn(wsConn, websocket.TextMessage)
	)

	// io.Copy between user input <--> server ws conn
	go func() {
		output.Write([]byte("Welcome to adbot node terminal!\r\n\r\n"))
		io.Copy(output, wsConnWrapper)
		output.Write([]byte("\r\nbye ~\r\n"))
	}()
	io.Copy(wsConnWrapper, input)

	return nil
}

// WatchNodeEvents implement Client interface
func (c *AdbotClient) WatchNodeEvents(id string) (io.ReadCloser, error) {
	resp, err := c.sendRequest("GET", "/api/nodes/"+id+"/events", nil, 0, "", "")
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

// UpsertNodeLabels implement Client interface
func (c *AdbotClient) UpsertNodeLabels(id string, lbs label.Labels) (label.Labels, error) {
	resp, err := c.sendRequest("PUT", "/api/nodes/"+id+"/labels", lbs, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{code, string(bs)}
	}

	var ret label.Labels
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// RemoveNodeLabels implement Client interface
func (c *AdbotClient) RemoveNodeLabels(id string, all bool, keys []string) (label.Labels, error) {
	resp, err := c.sendRequest("DELETE", fmt.Sprintf("/api/nodes/%s/labels?all=%t", id, all), keys, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{code, string(bs)}
	}

	var ret label.Labels
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// CloseNode implement Client interface
func (c *AdbotClient) CloseNode(id string) error {
	resp, err := c.sendRequest("DELETE", "/api/nodes/"+id+"/close", nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 204 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}
