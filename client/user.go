package client

import (
	"io/ioutil"

	"github.com/bbklab/adbot/types"
)

// AnyUsers implement Client interface
func (c *AdbotClient) AnyUsers() (bool, error) {
	resp, err := c.sendRequest("GET", "/api/users/any", nil, 0, "", "")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return false, &APIError{code, string(bs)}
	}

	var ret map[string]bool
	err = c.bind(resp.Body, &ret)
	return ret["result"], err
}

// ListUsers implement Client interface
func (c *AdbotClient) ListUsers() ([]*types.User, error) {
	resp, err := c.sendRequest("GET", "/api/users", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret []*types.User
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// UserProfile implement Client interface
func (c *AdbotClient) UserProfile() (*types.User, error) {
	resp, err := c.sendRequest("GET", "/api/users/profile", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret *types.User
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// UserSessions implement Client interface
func (c *AdbotClient) UserSessions() ([]*types.UserSessionWrapper, error) {
	resp, err := c.sendRequest("GET", "/api/users/sessions", nil, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var ret []*types.UserSessionWrapper
	err = c.bind(resp.Body, &ret)
	return ret, err
}

// CreateUser implement Client interface
func (c *AdbotClient) CreateUser(user *types.User) (*types.User, error) {
	resp, err := c.sendRequest("POST", "/api/users", user, 0, "", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 201 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{code, string(bs)}
	}

	var created *types.User
	err = c.bind(resp.Body, &created)
	return created, err
}

// ChangeUserPassword implement Client interface
func (c *AdbotClient) ChangeUserPassword(req *types.ReqChangePassword) error {
	resp, err := c.sendRequest("PATCH", "/api/users/change_password", req, 0, "", "")
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
