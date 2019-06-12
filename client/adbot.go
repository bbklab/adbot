package client

import (
	"fmt"
	"io/ioutil"

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
