package cli

import (
	"io"
	"os"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
)

// MetricsCommand is exported
func MetricsCommand() cli.Command {
	return cli.Command{
		Name:   "metrics",
		Usage:  "fetch prometheus metrics",
		Action: dumpMetrics,
	}
}

func dumpMetrics(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	stream, err := client.Metrics()
	if err != nil {
		return err
	}
	defer stream.Close()

	io.Copy(os.Stdout, stream)
	return nil
}
