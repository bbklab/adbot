package helpers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"github.com/bbklab/adbot/client"
	"github.com/bbklab/adbot/pkg/cmd"
)

// NewClient new an adbot client.Client with authications used in the CLI with current using adbot host
func NewClient() (client.Client, error) {
	client, curr, err := NewClientNoAuth()
	if err != nil {
		return nil, err
	}

	// login to get the token and save
	if curr.Token == "" {
		token, err := client.Login(curr.ReqLogin())
		if err != nil {
			return nil, err
		}
		client.SetHeader("Admin-Access-Token", token)
		SetAdbotHostAuth(curr.Name, curr.User, curr.Password, token)
		return client, nil
	}

	// verify the token
	client.SetHeader("Admin-Access-Token", curr.Token)
	_, err = client.UserProfile()
	if err == nil { // token accessable, return
		return client, nil
	}

	// not-401 error, return the error
	if !strings.HasPrefix(err.Error(), "401") {
		return nil, err
	}

	// 401 error, token expired, relogin to get the token and update
	token, err := client.Login(curr.ReqLogin())
	if err != nil {
		return nil, err
	}
	client.SetHeader("Admin-Access-Token", token)
	SetAdbotHostAuth(curr.Name, curr.User, curr.Password, token)
	return client, nil
}

// NewClientNoAuth new an un-authed adbot client.Client in the CLI with current using adbot host
//
// note: the second return parameter is current using adbot host which
// may be used by later auth login step
func NewClientNoAuth() (client.Client, *AdbotHost, error) {
	adbothost, err := CurrentAdbotHost()
	if err != nil {
		return nil, nil, err
	}

	client, err := NewClientNoAuthByAddr(adbothost.Addr)
	if err != nil {
		return nil, nil, err
	}

	return client, adbothost, nil
}

// NewClientNoAuthByAddr new an un-authed adbot client.Client in the CLI with given adbot host addr
func NewClientNoAuthByAddr(addr string) (client.Client, error) {
	return client.New([]string{addr})
}

// PrettyJSON is exported
func PrettyJSON(w io.Writer, data interface{}) error {
	if w == nil {
		w = io.Writer(os.Stdout)
	}

	bs, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	w.Write(append(bs, '\r', '\n'))
	return nil
}

// ContentFromFileOrCLI is exported
func ContentFromFileOrCLI(source string) ([]byte, error) {
	if strings.HasPrefix(source, "file://") {
		return ioutil.ReadFile(CLIFile(source))
	}
	if _, err := os.Stat(source); err == nil {
		return ioutil.ReadFile(source)
	}
	return []byte(source), nil
}

// CLIFile is exported
func CLIFile(source string) string {
	return strings.TrimPrefix(source, "file://")
}

// TrapExit is exported
func TrapExit(cleanFunc func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for range ch {
			if cleanFunc != nil {
				cleanFunc()
			}
			os.Exit(0)
		}
	}()
}

// GetHomeDir is exported
func GetHomeDir() string {
	dir, err := CurrentHomedir()
	if err != nil {
		return "/root"
	}
	return dir
}

// CurrentHomedir is exported
func CurrentHomedir() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", err // maybe met: `user: Current not implemented on linux/amd64`, require CGO_ENABLED=1 and  go>=1.9
	}

	if current.HomeDir == "" {
		return "", errors.New("current user does NOT have a homedir")
	}

	return current.HomeDir, nil
}

// StdinputLine read one line from os.Stdini without tailing \r or \n
func StdinputLine() ([]byte, error) {
	var (
		reader = bufio.NewReader(os.Stdin)
	)

	line, err := reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF && len(line) > 0 {
			goto PROCESS
		}
		return nil, err
	}

PROCESS:
	line = bytes.TrimSuffix(line, []byte("\n")) // trim final '\n'
	if bytes.HasSuffix(line, []byte("\r")) {    // trim possible final '\r'
		line = bytes.TrimSuffix(line, []byte("\r"))
	}

	return line, nil
}

// Exec exec given command string and ignore any output but check the command returned error
func Exec(command string) error {
	_, se, err := cmd.RunCmd(map[string]string{"TERM": "xterm"}, "/bin/sh", "-c", command)
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(se))
	}
	return nil
}

// ExecRedirect exec given command string and redirect output to given io.Writer until finished
// note: this ignore any command returned error
func ExecRedirect(w io.Writer, command string) error {
	stream, err := cmd.RunCmdProgress(map[string]string{"TERM": "xterm"}, false, nil, "/bin/sh", "-c", command)
	if err != nil {
		return err
	}

	io.Copy(w, stream)
	return nil
}
