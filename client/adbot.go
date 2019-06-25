package client

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/types"
)

//
// adb nodes
//

// ListAdbNodes implement Client interface
func (c *AdbotClient) ListAdbNodes() ([]*types.AdbNode, error) {
	resp, err := c.sendRequest("GET", "/api/adb_nodes", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret []*types.AdbNode
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// InspectAdbNode implement Client interface
func (c *AdbotClient) InspectAdbNode(id string) (*types.AdbNode, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/adb_nodes/%s", id), nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.AdbNode
	err = c.bind(resp.Body, &ret)
	return ret, err
}

//
// adb devices
//

// ListAdbDevices implement Client interface
func (c *AdbotClient) ListAdbDevices() ([]*types.AdbDeviceWrapper, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/adb_devices"), nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret []*types.AdbDeviceWrapper
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// InspectAdbDevice implement Client interface
func (c *AdbotClient) InspectAdbDevice(id string) (*types.AdbDeviceWrapper, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/adb_devices/%s", id), nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.AdbDeviceWrapper
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// ScreenCapAdbDevice implement Client interface
func (c *AdbotClient) ScreenCapAdbDevice(id string) ([]byte, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/adb_devices/%s/screencap", id), nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	return ioutil.ReadAll(resp.Body)
}

// DumpAdbDeviceUINodes implement Client interface
func (c *AdbotClient) DumpAdbDeviceUINodes(id string) ([]*adbot.AndroidUINode, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/adb_devices/%s/uinodes", id), nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret []*adbot.AndroidUINode
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// ClickAdbDevice implement Client interface
func (c *AdbotClient) ClickAdbDevice(id string, x, y int) error {
	resp, err := c.sendRequest("PATCH", fmt.Sprintf("/api/adb_devices/%s/click?x=%d&y=%d", id, x, y), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// GobackAdbDevice implement Client interface
func (c *AdbotClient) GobackAdbDevice(id string) error {
	resp, err := c.sendRequest("PATCH", fmt.Sprintf("/api/adb_devices/%s/goback", id), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// GotoHomeAdbDevice implement Client interface
func (c *AdbotClient) GotoHomeAdbDevice(id string) error {
	resp, err := c.sendRequest("PATCH", fmt.Sprintf("/api/adb_devices/%s/gotohome", id), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// RebootAdbDevice implement Client interface
func (c *AdbotClient) RebootAdbDevice(id string) error {
	resp, err := c.sendRequest("PATCH", fmt.Sprintf("/api/adb_devices/%s/reboot", id), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// RunAdbDeviceCmd implement Client interface
func (c *AdbotClient) RunAdbDeviceCmd(id, cmd string) ([]byte, error) {
	deviceCmd := types.AdbDeviceCmd{Command: cmd}
	resp, err := c.sendRequest("POST", "/api/adb_devices/"+id+"/exec", deviceCmd, 0, "", "")
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{code, string(bs)}
	}

	return ioutil.ReadAll(resp.Body)
}

// SetAdbDeviceBill implement Client interface
func (c *AdbotClient) SetAdbDeviceBill(id string, val int) error {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/api/adb_devices/%s/bill?val=%d", id, val), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// SetAdbDeviceAmount implement Client interface
func (c *AdbotClient) SetAdbDeviceAmount(id string, val int) error {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/api/adb_devices/%s/amount?val=%d", id, val), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// SetAdbDeviceWeight implement Client interface
func (c *AdbotClient) SetAdbDeviceWeight(id string, val int) error {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/api/adb_devices/%s/weight?val=%d", id, val), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// BindAdbDeviceAlipay implement Client interface
func (c *AdbotClient) BindAdbDeviceAlipay(id string, alipay *types.AlipayAccount) error {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/api/adb_devices/%s/alipay", id), alipay, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// RevokeAdbDeviceAlipay implement Client interface
func (c *AdbotClient) RevokeAdbDeviceAlipay(id string) error {
	resp, err := c.sendRequest("DELETE", fmt.Sprintf("/api/adb_devices/%s/alipay", id), nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 && code != 204 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// ReportAdbEvent implement Client interface
func (c *AdbotClient) ReportAdbEvent(ev *adbot.AdbEvent) error {
	resp, err := c.sendRequest("POST", "/api/adb_events", ev, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return &APIError{code, string(bs)}
	}

	return nil
}

// WatchAdbEvents implement Client interface
func (c *AdbotClient) WatchAdbEvents() (io.ReadCloser, error) {
	resp, err := c.sendRequest("GET", "/api/adb_events", nil, 0, "", "")
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
