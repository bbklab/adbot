package cli

import (
	"strings"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/agent"
	"github.com/bbklab/adbot/types"
)

var (
	joinFlags = []cli.Flag{
		cli.StringFlag{
			Name:   "addrs",
			Usage:  "The address of masters to join, eg: 127.0.0.1:88,127.0.0.1:89",
			EnvVar: "JOIN_ADDRS",
		},
	}
)

// JoinCommand is exported
func JoinCommand() cli.Command {
	return cli.Command{
		Name:      "join",
		ShortName: "a",
		Usage:     "start adbot agent and join to adbot master",
		Flags:     joinFlags,
		Action:    runAgentAndJoin,
	}
}

func runAgentAndJoin(c *cli.Context) error {
	cfg, err := newAgentConfig(c)
	if err != nil {
		return err
	}

	agent := agent.New(cfg)
	agent.Run() // fatal -> exit

	return nil
}

func newAgentConfig(c *cli.Context) (*types.AgentConfig, error) {
	var (
		addrArgs = c.String("addrs")
	)

	var addrs = []string{}
	if addrArgs != "" {
		addrs = strings.Split(addrArgs, ",")
	}

	cfg := &types.AgentConfig{
		JoinAddrs: addrs,
	}

	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	return cfg, nil
}
