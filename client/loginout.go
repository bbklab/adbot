package client

import (
	"errors"
	"io/ioutil"

	"github.com/bbklab/adbot/types"
)

// Login implement Client interface
// the client has the cookie jar store, after login succeed,
// reusing the client could pass through all afterwards requests.
func (c *AdbotClient) Login(req *types.ReqLogin) (string, error) {
	resp, err := c.sendRequest("POST", "/api/users/login", req, 0, "", "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 202 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return "", &APIError{code, string(bs)}
	}

	var token = resp.Header.Get("Admin-Access-Token")
	if token == "" {
		return "", errors.New("unexpected Admin-Access-Token from response header")
	}

	return token, nil
}

// Logout implement Client interface
func (c *AdbotClient) Logout() error {
	resp, err := c.sendRequest("DELETE", "/api/users/logout", nil, 0, "", "")
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
