package cli

import (
	"os"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
)

// WhoIsLeaderCommand is exported
func WhoIsLeaderCommand() cli.Command {
	return cli.Command{
		Name:   "who-is-leader",
		Usage:  "tell me who is the leader currently",
		Action: queryCurrentLeader,
	}
}

func queryCurrentLeader(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	info, _ := client.QueryLeader()
	os.Stdout.WriteString("Current Leader Is: " + info + "\r\n")
	return nil
}
