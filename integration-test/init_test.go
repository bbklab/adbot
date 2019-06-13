package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/client"
	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/types"
)

func init() {
	s, err := newApiSuite()
	if err != nil {
		log.Fatal(err)
	}

	check.Suite(s) // register the test suit
}

// make fit with go test mechanism
func Test(t *testing.T) {
	startAt := time.Now()
	fmt.Println("Test Starting ...")

	// run all test cases
	// check.TestingT(t)

	// run regexp matched test cases
	result := check.RunAll(&check.RunConf{
		Filter: os.Getenv("TESTON"),
	})
	if !result.Passed() {
		costPrintln("All Test Finished", startAt)
		log.Fatal(result.String())
	}

	costPrintln("All Test Finished", startAt)
	fmt.Println(result.String())
}

type ApiSuite struct {
	client client.Client
}

func newApiSuite() (*ApiSuite, error) {
	infHost := os.Getenv("API_HOST")

	if infHost == "" {
		return nil, errors.New("env API_HOST required")
	}

	endPoints := strings.Split(infHost, ",")

	client, err := client.New(endPoints)
	if err != nil {
		return nil, err
	}

	if err := ensureDefaultUser(client); err != nil {
		return nil, err
	}

	token, err := client.Login(&types.ReqLogin{userName, types.Password(userPass)})
	if err != nil {
		return nil, err
	}
	client.SetHeader("Admin-Access-Token", token)

	// no need
	// launch local-cluster already setup the default full-module license
	/*
		if err := loadDefaultLicense(client); err != nil {
			return nil, err
		}
	*/

	err = waitNodeGetOnline(client, time.Second*10)
	if err != nil {
		return nil, err
	}

	return &ApiSuite{
		client: client,
	}, nil
}

func ensureDefaultUser(client client.Client) error {
	hasUser, err := client.AnyUsers()
	if err != nil {
		return err
	}

	if hasUser {
		return nil
	}

	_, err = client.CreateUser(defaultUser)
	return err
}

func waitNodeGetOnline(client client.Client, maxWait time.Duration) error {
	timeout := time.After(maxWait)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errors.New("wait nodes get online timeout")

		case <-ticker.C:
			nodes, err := client.ListNodes(nil, ptype.Bool(true), "") // list online nodes
			if err != nil {
				return err
			}
			var idx int
			for _, node := range nodes {
				nodes[idx] = node
				idx++
			}
			if len(nodes[:idx]) > 0 {
				return nil
			}
		}
	}
}

func costPrintln(msg string, startAt time.Time) {
	fmt.Printf("%s  -  [%s]\n", msg, time.Since(startAt))
}
