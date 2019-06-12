package cli

import (
	"os"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
)

// InfoCommand is exported
func InfoCommand() cli.Command {
	return cli.Command{
		Name:   "info",
		Usage:  "print summary informations",
		Action: printInfo,
	}
}

func printInfo(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	info, err := client.Info()
	if err != nil {
		return err
	}

	info.WriteTo(os.Stdout)
	return nil
}
