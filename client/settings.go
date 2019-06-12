package client

import (
	"fmt"
	"io/ioutil"

	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/types"
)

// GetSettings implement Client interface
func (c *AdbotClient) GetSettings() (*types.Settings, error) {
	resp, err := c.sendRequest("GET", "/api/settings", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.Settings
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// UpdateSettings implement Client interface
func (c *AdbotClient) UpdateSettings(req *types.UpdateSettingsReq) (*types.Settings, error) {
	resp, err := c.sendRequest("PATCH", "/api/settings", req, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.Settings
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// SetGlobalAttrs implement Client interface
func (c *AdbotClient) SetGlobalAttrs(attrs label.Labels) (label.Labels, error) {
	resp, err := c.sendRequest("PUT", "/api/settings/attrs", attrs, 0, "", "")
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

// RemoveGlobalAttrs implement Client interface
func (c *AdbotClient) RemoveGlobalAttrs(all bool, keys []string) (label.Labels, error) {
	resp, err := c.sendRequest("DELETE", fmt.Sprintf("/api/settings/attrs?all=%t", all), keys, 0, "", "")
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

// ResetSettings implement Client interface
func (c *AdbotClient) ResetSettings() error {
	resp, err := c.sendRequest("PUT", "/api/settings/reset", nil, 0, "", "")
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

// GenAdvertiseQrCode implement Client interface
func (c *AdbotClient) GenAdvertiseQrCode() ([]byte, error) {
	resp, err := c.sendRequest("GET", "/api/settings/advertise_addr/qrcode", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, _ := ioutil.ReadAll(resp.Body)

	if code := resp.StatusCode; code != 200 {
		return nil, &APIError{code, string(bs)}
	}

	return bs, nil
}
