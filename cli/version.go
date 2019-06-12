package cli

import (
	"os"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/version"
)

// VersionCommand is exported
func VersionCommand() cli.Command {
	return cli.Command{
		Name:      "version",
		ShortName: "v",
		Usage:     "print version",
		Action: func(c *cli.Context) {
			version.Version().WriteTo(os.Stdout)
		},
	}
}
