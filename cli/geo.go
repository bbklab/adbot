package cli

import (
	"fmt"
	"os"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/utils"
)

var (
	showGeoMetadataFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose,v",
			Usage: "display more details",
		},
	}
)

// GeoCommand is exported
func GeoCommand() cli.Command {
	return cli.Command{
		Name:  "geo",
		Usage: "geo data management",
		Subcommands: []cli.Command{
			geoShowCommand(),   // show
			geoUpdateCommand(), // update
		},
	}
}

func geoShowCommand() cli.Command {
	return cli.Command{
		Name:   "show",
		Usage:  "show geo metadata infomations",
		Flags:  showGeoMetadataFlags,
		Action: showGeoMetadata,
	}
}

func geoUpdateCommand() cli.Command {
	return cli.Command{
		Name:   "update",
		Usage:  "update geo datas",
		Action: updateGeoData,
	}
}

func showGeoMetadata(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	info, err := client.CurrentGeoMetadata()
	if err != nil {
		return err
	}

	if c.Bool("verbose") {
		return utils.PrettyJSON(nil, info)
	}

	fmt.Fprintf(os.Stdout, " Current:\r\n")
	fmt.Fprintf(os.Stdout, formatGeoMetadata(info["asn"]))
	fmt.Fprintf(os.Stdout, formatGeoMetadata(info["city"]))
	return nil

}

func updateGeoData(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	prev, curr, cost, err := client.UpdateGeoData()
	if err != nil {
		return err
	}

	if prev != nil {
		fmt.Fprintf(os.Stdout, " Previous:\r\n")
		fmt.Fprintf(os.Stdout, formatGeoMetadata(prev["asn"]))
		fmt.Fprintf(os.Stdout, formatGeoMetadata(prev["city"]))
	}

	if curr != nil {
		fmt.Fprintf(os.Stdout, " Updated To:\r\n")
		fmt.Fprintf(os.Stdout, formatGeoMetadata(curr["asn"]))
		fmt.Fprintf(os.Stdout, formatGeoMetadata(curr["city"]))
	}

	fmt.Fprintf(os.Stdout, "\r\n Time Cost:%s\r\n", cost.String())
	return nil
}

func formatGeoMetadata(d maxminddb.Metadata) string {
	var (
		typ     = d.DatabaseType
		buildAt = time.Unix(int64(d.BuildEpoch), 0)
		nCount  = d.NodeCount
	)
	return fmt.Sprintf("   - %s(%s): %d nodes\r\n", typ, buildAt.Format(time.RFC3339), nCount)
}
