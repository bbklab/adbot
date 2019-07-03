package client

import (
	"fmt"
	"io/ioutil"

	lictypes "github.com/bbklab/adbot/types/lic"
)

// UpdateLicense implement Client interface
func (c *AdbotClient) UpdateLicense(data []byte) error {
	resp, err := c.sendRequest("PUT", "/api/license", data, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%d - %s", code, string(bs))
	}

	return nil

}

// ProductLicenseInfo implement Client interface
func (c *AdbotClient) ProductLicenseInfo() (*lictypes.License, error) {
	resp, err := c.sendRequest("GET", "/api/license", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d - %s", code, string(bs))
	}

	var ret *lictypes.License
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// RemoveLicense implement Client interface
func (c *AdbotClient) RemoveLicense() error {
	resp, err := c.sendRequest("DELETE", "/api/license", nil, 0, "", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 204 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%d - %s", code, string(bs))
	}

	return nil
}
