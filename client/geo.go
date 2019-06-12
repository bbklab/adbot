package client

import (
	"io/ioutil"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

// CurrentGeoMetadata implement Client interface
func (c *AdbotClient) CurrentGeoMetadata() (map[string]maxminddb.Metadata, error) {
	resp, err := c.sendRequest("GET", "/api/geo/metadata", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret map[string]maxminddb.Metadata
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// UpdateGeoData implement Client interface
func (c *AdbotClient) UpdateGeoData() (prev, current map[string]maxminddb.Metadata, cost time.Duration, err error) {
	resp, err := c.sendRequest("PATCH", "/api/geo/update", nil, 0, "", "")
	if err != nil {
		return nil, nil, time.Duration(0), err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, nil, time.Duration(0), &APIError{code, string(bs)}
	}

	var ret struct {
		Previous map[string]maxminddb.Metadata `json:"previous"`
		Current  map[string]maxminddb.Metadata `json:"current"`
		Cost     time.Duration                 `json:"cost"`
	}
	err = c.bind(resp.Body, &ret)
	if err != nil {
		return nil, nil, time.Duration(0), err
	}

	return ret.Previous, ret.Current, ret.Cost, nil
}
