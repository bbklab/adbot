package cli

import (
	"io/ioutil"
	"os"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/template"
	lictypes "github.com/bbklab/adbot/types/lic"
)

var (
	updateLicenseFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "input,i",
			Usage: "load from a given license pem file",
			Value: "",
		},
	}
)

// LicenseCommand is exported
func LicenseCommand() cli.Command {
	return cli.Command{
		Name:  "license",
		Usage: "license management",
		Subcommands: []cli.Command{
			licenseUpdateCommand(), // update
			licenseShowCommand(),   // show
			licenseRemoveCommand(), // rm
		},
	}
}

func licenseUpdateCommand() cli.Command {
	return cli.Command{
		Name:   "update",
		Usage:  "update current license",
		Flags:  updateLicenseFlags,
		Action: updateLicense,
	}
}

func licenseShowCommand() cli.Command {
	return cli.Command{
		Name:   "show",
		Usage:  "show current license",
		Action: showLicense,
	}
}

func licenseRemoveCommand() cli.Command {
	return cli.Command{
		Name:   "rm",
		Usage:  "remove current license",
		Action: removeLicense,
	}
}

func updateLicense(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	if c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var (
		finput = c.String("input")
	)
	if _, err := os.Stat(finput); err != nil {
		return err
	}
	licBytes, err := ioutil.ReadFile(finput)
	if err != nil {
		return err
	}

	err = client.UpdateLicense(licBytes)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func showLicense(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	lic, err := client.ProductLicenseInfo()
	if err != nil {
		return err
	}

	parser, err := template.NewParser(lictypes.LicenseTemplate)
	if err != nil {
		return err
	}
	return parser.Execute(os.Stdout, lic)
}

func removeLicense(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	err = client.RemoveLicense()
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}
