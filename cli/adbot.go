package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/color"
	"github.com/bbklab/adbot/pkg/template"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

// nolint
var (
	// adb node
	AdbDeviceHeader = fmt.Sprintf("DEVICE(%s/%s/%s)", color.Green("O"), color.Red("L"), color.Cyan("T"))

	AdbNodeTableHeader = "NODE ID\t" + color.Yellow("STATUS") + "\tHOSTNAME\tREMOTE\tGEOGRAPHY\t" + AdbDeviceHeader + "\t\n"

	AdbNodeTableStatus = "{{if eq .Node.Status \"online\"}}{{green .Node.Status}}{{else if eq .Node.Status \"offline\"}}{{red .Node.Status}}{{else if eq .Node.Status \"flagging\"}}{{magenta .Node.Status}}{{else if eq .Node.Status \"deploying\"}}{{cyan .Node.Status}}{{else if eq .Node.Status \"deleting\"}}{{magenta .Node.Status}}{{end}}" // similar as node.go nodeTableStatus

	AdbNodeTableHostname = "{{if .Node.SysInfo}}{{.Node.SysInfo.Hostname}}{{else}}-{{end}}"                       // similar as node.go nodeTableHostname
	AdbNodeGeoInfo       = "{{if .Node.GeoInfo}}{{.Node.GeoInfo.Country}}/{{.Node.GeoInfo.City}}{{else}}-{{end}}" // similar as node.go nodeGeoInfo
	AdbNodeTableDevice   = "{{if eq .NumDevices 0}}{{green \"-\"}}/{{red \"-\"}}/{{cyan \"-\"}}{{else}}{{green .NumOnline}}/{{red .NumOffline}}/{{cyan .NumDevices}}{{end}}"

	AdbNodeTableLine = "{{.Node.ID}}\t" + AdbNodeTableStatus + "\t" + AdbNodeTableHostname + "\t" + "{{hostof .Node.RemoteAddr}}\t" + AdbNodeGeoInfo + "\t" + AdbNodeTableDevice + "\t\n"

	// adb device
	AdbDeviceTableHeader  = "DEVICE ID\tNODE ID\t" + color.Yellow("STATUS") + "\tWEIGHT\tBILL\tAMOUNT\tMANUFACTURER & MODEL\tANDROID VERSION\tBATTERY\tBOOT AT\t\n"
	AdbDeviceTableStatus  = "{{if eq .Status \"online\"}}{{green .Status}}{{else}}{{red .Status}}{{end}}"
	AdbDeviceTableBattery = "{{$level := multiply .SysInfo.Battery.Level 100}}{{divide $level .SysInfo.Battery.Scale 0}}%" // (100*level/scale)%
	AdbDeviceTableLine    = "{{.ID}}\t{{.NodeID}}\t" + AdbDeviceTableStatus + "\t{{.Weight}}\t{{.TodayBill}}/{{.MaxBill}}\t{{.TodayAmount}}/{{.MaxAmount}}\t{{.SysInfo.Manufacturer}} - {{.SysInfo.ProductModel}}\t{{.SysInfo.ReleaseVersion}} - SDK{{.SysInfo.SDKVersion}}\t" + AdbDeviceTableBattery + "\t{{tformat .SysInfo.BootTimeAt}}\t\n"
)

var (
	listAdbNodeFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "only display numeric IDs",
		},
	}

	setAdbDeviceBillFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "value",
			Usage: "max bill per day value, must between [0-10000], 0 means unlimited",
		},
	}

	setAdbDeviceAmountFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "value",
			Usage: "max amount per day value, must between [0-100000000], 0 means unlimited",
		},
	}

	setAdbDeviceWeightFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "value",
			Usage: "weight value, must between [0-100], the higher value means the higher weight, 0 means disabled",
			Value: -1,
		},
	}

	bindAdbDeviceAlipayFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "userid",
			Usage: "alipay user id",
		},
		cli.StringFlag{
			Name:  "username",
			Usage: "alipay username",
		},
		cli.StringFlag{
			Name:  "nickname",
			Usage: "alipay nickname",
		},
	}

	listAdbDeviceFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "only display numeric device IDs",
		},
	}
)

// AdbNodeCommand is exported
func AdbNodeCommand() cli.Command {
	return cli.Command{
		Name:  "adbnode",
		Usage: "adb node management",
		Subcommands: []cli.Command{
			adbNodeListCommand(),    // ls
			adbNodeInspectCommand(), // inspect
		},
	}
}

func adbNodeListCommand() cli.Command {
	return cli.Command{
		Name:   "ls",
		Usage:  "list all of adb nodes",
		Flags:  listAdbNodeFlags,
		Action: listAdbNodes,
	}
}

func adbNodeInspectCommand() cli.Command {
	return cli.Command{
		Name:   "inspect",
		Usage:  "inspect one adb node",
		Action: inspectAdbNode,
	}
}

// AdbDeviceCommand is exported
func AdbDeviceCommand() cli.Command {
	return cli.Command{
		Name:  "adbdevice",
		Usage: "adb device management",
		Subcommands: []cli.Command{
			adbDeviceListCommand(),         // ls
			adbDeviceInspectCommand(),      // inspect
			adbDeviceSetBillCommand(),      // set-bill
			adbDeviceSetAmountCommand(),    // set-amount
			adbDeviceSetWeightCommand(),    // set-weight
			adbDeviceBindAlipayCommand(),   // bind-alipay
			adbDeviceRevokeAlipayCommand(), // revoke-alipay
			// adbDeviceRemoveCommand(),    // rm
		},
	}
}

func adbDeviceListCommand() cli.Command {
	return cli.Command{
		Name:   "ls",
		Usage:  "list adb devices",
		Flags:  listAdbDeviceFlags,
		Action: listAdbDevices,
	}
}

func adbDeviceInspectCommand() cli.Command {
	return cli.Command{
		Name:      "inspect",
		Usage:     "inspect details of an adb device",
		ArgsUsage: "DEVICE",
		Action:    inspectAdbDevice,
	}
}

func adbDeviceSetBillCommand() cli.Command {
	return cli.Command{
		Name:      "set-bill",
		Usage:     "set abb device max bill perday, must between [0-10000], 0 means unlimited",
		ArgsUsage: "DEVICE",
		Flags:     setAdbDeviceBillFlags,
		Action:    setAdbDeviceBill,
	}
}

func adbDeviceSetAmountCommand() cli.Command {
	return cli.Command{
		Name:      "set-amount",
		Usage:     "set abb device max amount perday, must between [0-100000000], 0 means unlimited",
		ArgsUsage: "DEVICE",
		Flags:     setAdbDeviceAmountFlags,
		Action:    setAdbDeviceAmount,
	}
}

func adbDeviceSetWeightCommand() cli.Command {
	return cli.Command{
		Name:      "set-weight",
		Usage:     "set adb device weight value, must between [0-100], the higher value means the higher weight, 0 means disabled",
		ArgsUsage: "DEVICE",
		Flags:     setAdbDeviceWeightFlags,
		Action:    setAdbDeviceWeight,
	}
}

func adbDeviceBindAlipayCommand() cli.Command {
	return cli.Command{
		Name:      "bind-alipay",
		Usage:     "bind abb device with alipay account",
		ArgsUsage: "DEVICE",
		Flags:     bindAdbDeviceAlipayFlags,
		Action:    bindAdbDeviceAlipay,
	}
}

func adbDeviceRevokeAlipayCommand() cli.Command {
	return cli.Command{
		Name:      "revoke-alipay",
		Usage:     "revoke abb device alipay account",
		ArgsUsage: "DEVICE",
		Action:    revokeAdbDeviceAlipay,
	}
}

// adb node
//

func listAdbNodes(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	nodes, err := client.ListAdbNodes()
	if err != nil {
		return err
	}

	// only print ids
	if c.Bool("quiet") {
		for _, node := range nodes {
			fmt.Fprintln(os.Stdout, node.Node.ID)
		}
		return nil
	}

	var (
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	)

	parser, err := template.NewParser(AdbNodeTableLine)
	if err != nil {
		return err
	}

	fmt.Fprint(w, AdbNodeTableHeader)
	for _, node := range nodes {
		parser.Execute(w, node)
	}
	w.Flush()

	return nil
}

func inspectAdbNode(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		nodeID = c.Args().First()
	)

	if nodeID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	node, err := client.InspectAdbNode(nodeID)
	if err != nil {
		return err
	}

	return utils.PrettyJSON(nil, node)
}

// adb device
//

func listAdbDevices(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	dvcs, err := client.ListAdbDevices()
	if err != nil {
		return err
	}

	// only print ids
	if c.Bool("quiet") {
		for _, dvc := range dvcs {
			fmt.Fprintln(os.Stdout, dvc.ID)
		}
		return nil
	}

	var (
		w         = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
		parser, _ = template.NewParser(AdbDeviceTableLine)
	)

	fmt.Fprint(w, AdbDeviceTableHeader)
	for _, dvc := range dvcs {
		parser.Execute(w, dvc)
	}
	w.Flush()

	return nil
}

func inspectAdbDevice(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dvcID = c.Args().First()
	)

	if dvcID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	dvc, err := client.InspectAdbDevice(dvcID)
	if err != nil {
		return err
	}

	return utils.PrettyJSON(nil, dvc)
}

func setAdbDeviceBill(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dvcID = c.Args().First()
		bill  = c.Int("value")
	)

	if dvcID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	err = client.SetAdbDeviceBill(dvcID, bill)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func setAdbDeviceAmount(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dvcID  = c.Args().First()
		amount = c.Int("value")
	)

	if dvcID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	err = client.SetAdbDeviceAmount(dvcID, amount)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func setAdbDeviceWeight(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dvcID  = c.Args().First()
		weight = c.Int("value")
	)

	if dvcID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	err = client.SetAdbDeviceWeight(dvcID, weight)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func bindAdbDeviceAlipay(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dvcID  = c.Args().First()
		alipay = &types.AlipayAccount{
			UserID:   c.String("userid"),
			Username: c.String("username"),
			Nickname: c.String("nickname"),
		}
	)

	if dvcID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	err = client.BindAdbDeviceAlipay(dvcID, alipay)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}

func revokeAdbDeviceAlipay(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		dvcID = c.Args().First()
	)

	if dvcID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	err = client.RevokeAdbDeviceAlipay(dvcID)
	if err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}
