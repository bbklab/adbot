package cli

import (
	"io"
	"os"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/utils"
)

var (
	dumpFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "goroutine",
			Usage: "only dump goroutines stack",
		},
		cli.BoolFlag{
			Name:  "general",
			Usage: "only dump general debug infos",
		},
		cli.BoolFlag{
			Name:  "config",
			Usage: "only dump master runtime configs",
		},
		cli.BoolFlag{
			Name:  "application",
			Usage: "only dump application debug infos",
		},
		cli.StringFlag{
			Name:  "output,o",
			Usage: "FILE  Write output to <file> instead of stdout",
		},
	}
)

// DumpCommand is exported
func DumpCommand() cli.Command {
	return cli.Command{
		Name:      "dump",
		ShortName: "d",
		Usage:     "dump debug informations",
		Flags:     dumpFlags,
		Action:    dumpDebug,
	}
}

func dumpDebug(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dumpName    = "general"
		goroutine   = c.Bool("goroutine")
		general     = c.Bool("general")
		config      = c.Bool("config")
		application = c.Bool("application")
		foutput     = c.String("output")
		w           = io.Writer(os.Stdout)
	)

	switch {
	case goroutine:
		dumpName = "goroutine"
	case general:
		dumpName = "general"
	case config:
		dumpName = "config"
	case application:
		dumpName = "application"
	}

	if foutput != "" {
		fd, err := os.OpenFile(foutput, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer fd.Close()
		w = io.Writer(fd)
	}

	bs, info, cfg, app, err := client.DebugDump(dumpName)
	if err != nil {
		return err
	}

	switch dumpName {
	case "goroutine":
		w.Write(bs)
	case "general":
		info.WriteTo(w)
	case "config":
		utils.PrettyJSON(w, cfg)
	case "application":
		utils.PrettyJSON(w, app)
	}

	return nil
}
