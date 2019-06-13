package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

var (
	updateSettingsFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Usage: "set log level, eg debug|info|warn|error|fatal",
		},
		cli.StringFlag{
			Name:  "enable-httpmux-debug",
			Usage: "enable httpmux debug or not",
		},
		cli.StringFlag{
			Name:  "unmask-sensitive",
			Usage: "unconver the masked sensitive fields (******) for all of api responses",
		},
		cli.StringFlag{
			Name:  "telegram-bot-token",
			Usage: "telegram bot token",
		},
	}

	removeGlobalAttrFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "all,a",
			Usage: "remove all of global setting attrs",
		},
	}
)

// SettingsCommand is exported
func SettingsCommand() cli.Command {
	return cli.Command{
		Name:  "settings",
		Usage: "settings management",
		Subcommands: []cli.Command{
			showCommand(),   // show
			updateCommand(), // update
			resetCommand(),  // reset
			attrCommand(),   // attr
		},
	}
}

func showCommand() cli.Command {
	return cli.Command{
		Name:   "show",
		Usage:  "show current settings",
		Action: showSettings,
	}
}

func updateCommand() cli.Command {
	return cli.Command{
		Name:   "update",
		Usage:  "update current settings",
		Flags:  updateSettingsFlags,
		Action: updateSettings,
	}
}

func resetCommand() cli.Command {
	return cli.Command{
		Name:   "reset",
		Usage:  "reset to initial default settings",
		Action: resetSettings,
	}
}

func attrCommand() cli.Command {
	return cli.Command{
		Name:  "attr",
		Usage: "manage global setting attrs",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "set global attrs",
				ArgsUsage: "name=value [name=value ...]",
				Action:    setGlobalAttrs,
			},
			{
				Name:   "get",
				Usage:  "get global attrs",
				Action: getGlobalAttrs,
			},
			{
				Name:      "rm",
				Usage:     "remove global attrs",
				ArgsUsage: "key1 [key2 ...]",
				Flags:     removeGlobalAttrFlags,
				Action:    removeGlobalAttrs,
			},
		},
	}
}

func showSettings(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	settings, err := client.GetSettings()
	if err != nil {
		return err
	}

	return utils.PrettyJSON(nil, settings)
}

func updateSettings(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	if c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var req = new(types.UpdateSettingsReq)

	if v := c.String("log-level"); v != "" {
		req.LogLevel = ptype.String(v)
	}
	if v := c.String("enable-httpmux-debug"); v != "" {
		vv, _ := strconv.ParseBool(v)
		req.EnableHTTPMuxDebug = ptype.Bool(vv)
	}
	if v := c.String("unmask-sensitive"); v != "" {
		vv, _ := strconv.ParseBool(v)
		req.UnmarkSensitive = ptype.Bool(vv)
	}
	if v := c.String("telegram-bot-token"); v != "" {
		req.TGBotToken = ptype.String(v)
	}

	if _, err := client.UpdateSettings(req); err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func resetSettings(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	if err := client.ResetSettings(); err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func setGlobalAttrs(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		attrs = c.Args()
	)

	if len(attrs) == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	attrsReq, err := label.Parse(strings.Join(attrs, " "))
	if err != nil {
		return err
	}

	nowattrs, err := client.SetGlobalAttrs(attrsReq)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s\n", nowattrs.String())
	return nil
}

func getGlobalAttrs(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	settings, err := client.GetSettings()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s\n", settings.GlobalAttrs.String())
	return nil
}

func removeGlobalAttrs(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		keys = c.Args()
		all  = c.Bool("all") // if true `keys` won't take effective
	)

	if !all {
		if len(keys) == 0 {
			return cli.ShowSubcommandHelp(c)
		}
	}

	nowattrs, err := client.RemoveGlobalAttrs(all, keys)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s\n", nowattrs.String())
	return nil
}
