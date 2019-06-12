package ssh

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config represents the ssh options
type Config struct {
	Addr     string `json:"addr" bson:"addr"` // host:port
	User     string `json:"user" bson:"user"`
	Password string `json:"password" bson:"password"`
	PrivKey  string `json:"privkey" bson:"privkey"`
	Timeout  int    `json:"timeout" bson:"timeout"` // by second
}

// Valid is exported
func (cfg *Config) Valid() error {
	if cfg.Addr == "" {
		return errors.New("ssh host address required")
	}
	if _, _, err := net.SplitHostPort(cfg.Addr); err != nil {
		return fmt.Errorf("ssh host address [%s] should be the format host:port", cfg.Addr)
	}

	if cfg.User == "" {
		return errors.New("ssh user required")
	}

	if cfg.Timeout < 0 {
		return errors.New("ssh timeout can't be negative")
	}

	if cfg.Password == "" && cfg.PrivKey == "" {
		return errors.New("at least one ssh auth mechanism [password | public-key] required")
	}

	if key := cfg.PrivKey; len(key) > 0 {
		if _, err := ssh.ParsePrivateKey([]byte(key)); err != nil {
			return fmt.Errorf("ssh.ParsePrivateKey error: %v", err)
		}
	}

	return nil
}

// Client is a ssh client
type Client struct {
	Config *Config
	config *ssh.ClientConfig // converted from Config
}

// NewClient initialize a ssh Client with given configs
func NewClient(config *Config) (*Client, error) {
	c := &Client{Config: config}
	return c, c.init()
}

func (c *Client) init() error {
	// prepare auth methods
	var auths []ssh.AuthMethod

	if p := c.Config.Password; p != "" {
		auths = append(auths, ssh.Password(p))
	}

	if key := c.Config.PrivKey; len(key) > 0 {
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			return fmt.Errorf("ssh.ParsePrivateKey error: %v", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	if len(auths) == 0 {
		return errors.New("no avaliable ssh auth mechanism")
	}

	var timeout = c.Config.Timeout
	if timeout <= 0 {
		timeout = 10 // default timeout
	}

	// prepare ssh client config
	c.config = &ssh.ClientConfig{
		User:            c.Config.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	return nil
}

// Ping perform ssh ping test to verify the connectivity
func (c *Client) Ping() error {
	client, err := ssh.Dial("tcp", c.Config.Addr, c.config)
	if err != nil {
		return err
	}
	client.Close()
	return nil
}

// Run execute a shell command via ssh and wait the result synchronously
func (c *Client) Run(cmd string, sudo bool) ([]byte, error) {
	client, err := ssh.Dial("tcp", c.Config.Addr, c.config)
	if err != nil {
		return nil, fmt.Errorf("ssh.Dial error: %v", err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh.NewSession error: %v", err)
	}
	defer sess.Close()

	if sudo {
		cmd = "sudo " + cmd
	}

	bs, err := sess.CombinedOutput(cmd)
	if err != nil {
		return bs, fmt.Errorf("ssh.Session.Run error: %v", err)
	}

	return bs, nil
}

// RunWithProgress execute a shell command via ssh and return the result by realtime
func (c *Client) RunWithProgress(cmd, execID string, sudo bool) (io.ReadCloser, error) {
	client, err := ssh.Dial("tcp", c.Config.Addr, c.config)
	if err != nil {
		return nil, fmt.Errorf("ssh.Dial error: %v", err)
	}

	sess, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh.NewSession error: %v", err)
	}

	var (
		pipeReader, pipeWriter = io.Pipe()
	)

	sess.Stdout = pipeWriter
	sess.Stderr = pipeWriter

	go func() {
		defer func() {
			sess.Close()
			client.Close()
		}()

		if sudo {
			cmd = "sudo " + cmd
		}

		var msg string
		if err := sess.Run(cmd); err != nil {
			msg = fmt.Sprintf("\r\n%s:FINISH: %s\r\n", execID, err.Error())
		} else {
			msg = fmt.Sprintf("\r\n%s:FINISH: SUCCESS\r\n", execID)
		}

		pipeWriter.Write([]byte(msg))
		pipeWriter.Close() // so the PipeReader got EOF
	}()

	// when the pipeReader was closed, sess.Run will end with error
	return io.ReadCloser(pipeReader), nil
}
