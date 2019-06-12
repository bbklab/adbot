package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/color"
	"github.com/bbklab/adbot/pkg/template"
	"github.com/bbklab/adbot/types"
)

// nolint
var (
	HostTableHeader  = color.Yellow("CURRENT") + "\tNAME\tHOST\tUSER\t\n"
	hostTableCurrent = "{{if .Current}}{{green .Current}}{{else}}{{cyan .Current}}{{end}}"
	hostTableUser    = "{{if eq .User \"\"}}-{{else}}{{.User}}{{end}}"
	HostTableLine    = hostTableCurrent + "\t{{.Name}}\t{{.Addr}}\t" + hostTableUser + "\t\n"
)

var (
	addHostFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "name",
			Usage: "uniq name of adbot daemon host",
		},
		cli.StringFlag{
			Name:  "addr",
			Usage: "adbot daemon socket address to connect to, eg: unix:///var/run/adbot/adbot.sock  http://ip:port",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "adbot daemon host administrator user name",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "adbot daemon administrator user password",
		},
	}
)

// HostCommand is exported
func HostCommand() cli.Command {
	return cli.Command{
		Name:  "host",
		Usage: "adbot daemon host socket(s) management",
		Subcommands: []cli.Command{
			hostListCommand(),   // ls
			hostAddCommand(),    // add
			hostRemoveCommand(), // rm
			hostSwitchCommand(), // switch
			hostResetCommand(),  // reset
		},
	}
}

func hostListCommand() cli.Command {
	return cli.Command{
		Name:   "ls",
		Usage:  "list all of adbot daemon hosts",
		Action: listHosts,
	}
}

func hostAddCommand() cli.Command {
	return cli.Command{
		Name:   "add",
		Usage:  "add a adbot daemon connection",
		Flags:  addHostFlags,
		Action: addHost,
	}
}

func hostRemoveCommand() cli.Command {
	return cli.Command{
		Name:      "rm",
		Usage:     "remove one or more adbot daemon hosts",
		ArgsUsage: "host [host...]",
		Action:    removeHost,
	}
}

func hostSwitchCommand() cli.Command {
	return cli.Command{
		Name:      "switch",
		Usage:     "switch to given adbot daemon host",
		ArgsUsage: "host",
		Action:    switchHost,
	}
}

func hostResetCommand() cli.Command {
	return cli.Command{
		Name:   "reset",
		Usage:  "reset to default configs",
		Action: resetHost,
	}
}

func listHosts(c *cli.Context) error {
	hosts, err := helpers.ListAdbotHosts()
	if err != nil {
		return err
	}

	var (
		w         = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
		parser, _ = template.NewParser(HostTableLine)
	)

	fmt.Fprint(w, HostTableHeader)
	for _, host := range hosts {
		parser.Execute(w, host)
	}
	w.Flush()

	return nil
}

func addHost(c *cli.Context) error {
	if c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var (
		host = &helpers.AdbotHost{
			Name:     c.String("name"),
			Addr:     c.String("addr"),
			User:     c.String("user"),
			Password: c.String("password"),
		}
	)

	// verify
	if err := host.Valid(); err != nil {
		return err
	}

	// verify addr
	client, err := helpers.NewClientNoAuthByAddr(host.Addr)
	if err != nil {
		return err
	}

	// verify user & password
	token, err := client.Login(&types.ReqLogin{host.User, types.Password(host.Password)})
	if err != nil {
		return err
	}

	// set the token
	host.Token = token

	// save to local configs files
	err = helpers.AddAdbotHost(host)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte(host.Name), '\r', '\n'))
	return nil
}

func removeHost(c *cli.Context) error {
	var (
		hosts = c.Args()
		errN  int
	)

	if !c.Args().Present() {
		return cli.ShowSubcommandHelp(c)
	}

	for _, host := range hosts {
		if err := helpers.RemoveAdbotHost(host); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", host, err)
			errN++
		} else {
			fmt.Fprintf(os.Stdout, "Deleted: %s\n", host)
		}
	}

	if errN > 0 {
		return fmt.Errorf("%d removal error", errN)
	}

	return nil

}
func switchHost(c *cli.Context) error {
	var (
		hosts = c.Args()
	)

	if len(hosts) == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	curr, err := helpers.SwitchAdbotHost(hosts[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Using Adbot Host %s - %s\n", curr.Name, curr.Addr)
	return nil
}

func resetHost(c *cli.Context) error {
	err := helpers.ResetAdbotHost()
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}
