package types

import (
	"errors"
	"fmt"
	"net"
	"os"

	mgo "gopkg.in/mgo.v2"

	"github.com/bbklab/adbot/pkg/file"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/pkg/validator"
)

// MasterConfig is exported
type MasterConfig struct {
	Listen        string       `json:"listen"`             // must
	TLSCert       string       `json:"tls_cert,omitempty"` // optional, if given, serving additional https protocol
	TLSKey        string       `json:"tls_key,omitempty"`  // optional, if given, serving additional https protocol
	AdvertiseAddr string       `json:"advertise_addr"`     // must
	PublicKeyData string       `json:"public_key_data"`    // must, file or text
	PromAddr      string       `json:"prometheus_addr"`    // must
	UnixSock      string       `json:"unix_sock"`          // optional
	PidFile       string       `json:"pid_file"`           // optional
	Store         *StoreConfig `json:"store"`              // must
}

// StoreConfig is exported
type StoreConfig struct {
	Type          string         `json:"type"`
	MongodbConfig *MongodbConfig `json:"mongodb_config,omitempty"`
}

// MongodbConfig is exported
type MongodbConfig struct {
	MgoURL string `json:"mgo_url"`
}

// RequireServeTLS is exported
func (c *MasterConfig) RequireServeTLS() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

// Valid is exported
func (c *MasterConfig) Valid() error {
	if err := validator.String(c.Listen, 1, 1024, nil); err != nil {
		return fmt.Errorf("listen addr %v", err)
	}

	if c.RequireServeTLS() {
		for _, file := range []string{c.TLSCert, c.TLSKey} {
			if _, err := os.Stat(file); err != nil {
				return err
			}
		}
	}

	if err := validator.String(c.AdvertiseAddr, 1, 1024, nil); err != nil {
		return fmt.Errorf("advertise addr %v", err)
	}

	if _, _, err := net.SplitHostPort(c.AdvertiseAddr); err != nil {
		return fmt.Errorf("advertise addr [%s] should be the format host:port", c.AdvertiseAddr)
	}

	// we're expecting a valid pem public key
	// require this validation check may lead to start up failure under circleci environment,
	// thus we enable this only while we got an exists public key file
	if file.Exists(c.PublicKeyData) {
		if _, err := utils.LoadRSAPublicKey(c.PublicKeyData); err != nil {
			return fmt.Errorf("try load public key [%s] error: %v", c.PublicKeyData, err)
		}
	}

	if err := validator.String(c.PromAddr, 1, 1024, nil); err != nil {
		return fmt.Errorf("prometheus addr %v", err)
	}

	if c.Store == nil {
		return errors.New("db store options required")
	}

	return c.Store.Valid()
}

// Valid is exported
func (c *StoreConfig) Valid() error {
	switch typ := c.Type; typ {
	case "mongodb", "mongo":
		if c.MongodbConfig == nil {
			return errors.New("mongodb store config required")
		}
		if _, err := mgo.ParseURL(c.MongodbConfig.MgoURL); err != nil {
			return fmt.Errorf("parse mongodb url %v", err)
		}

	default:
		return errors.New("unsupported db store type: " + typ)
	}

	return nil
}

// AgentConfig is exported
type AgentConfig struct {
	JoinAddrs []string `json:"join_addrs"`
}

// Valid is exported
func (c *AgentConfig) Valid() error {
	if len(c.JoinAddrs) == 0 {
		return errors.New("at least one join addr required")
	}

	for _, addr := range c.JoinAddrs {
		_, _, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("join addr [%s] should be the format host:port", addr)
		}
	}

	return nil
}
