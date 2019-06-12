package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/types"
)

var (
	loginFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "username,u",
			Usage: "Username",
			Value: "",
		},
		cli.StringFlag{
			Name:  "password,p",
			Usage: "Password",
			Value: "",
		},
	}
)

// LoginCommand is exported
func LoginCommand() cli.Command {
	return cli.Command{
		Name:   "login",
		Usage:  "log in to the adbot system",
		Flags:  loginFlags,
		Action: login,
	}
}

// LogoutCommand is exported
func LogoutCommand() cli.Command {
	return cli.Command{
		Name:   "logout",
		Usage:  "log out from the adbot system",
		Action: logout,
	}
}

func login(c *cli.Context) error {
	client, currAdbotHost, err := helpers.NewClientNoAuth()
	if err != nil {
		return err
	}

	// using user input username & password
	var (
		username = c.String("username")
		password = c.String("password")
	)

	if username == "" {
		// obtain username & password from stdin
		os.Stdout.Write([]byte("Username: "))
		bs, err := helpers.StdinputLine()
		if err != nil {
			return err
		}
		username = string(bs)
	}

	if password == "" {
		os.Stdout.Write([]byte("Password: "))
		bs, err := helpers.StdinputLine()
		if err != nil {
			return err
		}
		password = string(bs)
	}

	req := &types.ReqLogin{
		UserName: username,
		Password: types.Password(password),
	}

	token, err := client.Login(req)
	if err != nil {
		return err
	}

	if err := helpers.SetAdbotHostAuth(currAdbotHost.Name, username, password, token); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Login Succeed - %s\n", currAdbotHost.Addr)
	return nil
}

func logout(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	// ignore erros, maybe current user already can't login any more (password changed)
	client.Logout()

	// clean up current adbot host saved auths
	currAdbotHost, _ := helpers.CurrentAdbotHost()
	if err := helpers.SetAdbotHostAuth(currAdbotHost.Name, "", "", ""); err != nil {
		if strings.Contains(err.Error(), "not exists") {
			return errors.New("no need, not login yet")
		}
		return err
	}

	fmt.Fprintf(os.Stdout, "Logout Succeed - %s\n", currAdbotHost.Addr)
	return nil
}
