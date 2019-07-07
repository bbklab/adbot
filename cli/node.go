package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	units "github.com/docker/go-units"
	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/color"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/pkg/template"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

var (
	nodeTableHeader   = "NODE ID\t" + "VERSION\t" + color.Yellow("STATUS") + "\tHOSTNAME\tREMOTE\tGEOGRAPHY\tUPTIME\tLOADAVG\tCPU\tMEMORY\tJOIN AT\tACTIVE AT\t\n"
	nodeTableVersion  = "{{if .Version}}{{.Version}}{{else}}-{{end}}"
	nodeTableStatus   = "{{if eq .Status \"online\"}}{{green .Status}}{{else if eq .Status \"offline\"}}{{red .Status}}{{else if eq .Status \"flagging\"}}{{magenta .Status}}{{else if eq .Status \"deploying\"}}{{cyan .Status}}{{else if eq .Status \"deleting\"}}{{magenta .Status}}{{end}}"
	nodeTableHostname = "{{if .SysInfo}}{{.SysInfo.Hostname}}{{else}}-{{end}}"
	nodeTableLoadavg  = "{{if .SysInfo}}{{.SysInfo.LoadAvgs.One}}{{else}}-{{end}}"
	nodeTableCPU      = "{{if .SysInfo}}{{.SysInfo.CPU.Used}}%({{.SysInfo.CPU.Processor}}){{else}}-{{end}}"
	nodeTableMemory   = "{{if .SysInfo}}{{size .SysInfo.Memory.Used}}/{{size .SysInfo.Memory.Total}}{{else}}-{{end}}"
	nodeTableUptime   = "{{if .SysInfo}}{{dsecformat .SysInfo.UptimeInt}}{{else}}-{{end}}"
	nodeGeoInfo       = "{{if .GeoInfo}}{{.GeoInfo.Country}}/{{.GeoInfo.City}}{{else}}-{{end}}"
	nodeTableLine     = "{{.ID}}\t" + nodeTableVersion + "\t" + nodeTableStatus + "\t" + nodeTableHostname + "\t{{hostof .RemoteAddr}}\t" + nodeGeoInfo + "\t" + nodeTableUptime + "\t" + nodeTableLoadavg + "\t" + nodeTableCPU + "\t" + nodeTableMemory + "\t{{tformat .JoinAt}}\t{{tformat .LastActiveAt}}\t\n"
)

var (
	nodeLabelFilterFlag = cli.StringFlag{
		Name:  "filter,f",
		Usage: "node label filters: key1=val1 key2=val2 ...",
	}

	verboseStatsFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose,v",
			Usage: "display node stats details",
		},
	}

	listNodeFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "only display numeric IDs",
		},
		nodeLabelFilterFlag,
		cli.StringFlag{
			Name:  "online,o",
			Usage: "online status filter, true: list all online nodes, otherwise all offlines",
		},
	}

	execNodeFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "cmd",
			Usage: "command to be executed on the node",
		},
	}

	watchNodeFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "all,a",
			Usage: "watch events on all of nodes",
		},
	}

	removeNodeLabelFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "all,a",
			Usage: "remove all of node labels",
		},
	}
)

// NodeCommand is exported
func NodeCommand() cli.Command {
	return cli.Command{
		Name:  "node",
		Usage: "node management",
		Subcommands: []cli.Command{
			nodeListCommand(),     // ls
			nodeInspectCommand(),  // inspect
			nodeStatsCommand(),    // stats
			nodeTerminalCommand(), // terminal
			nodeExecCommand(),     // exec
			nodeWatchCommand(),    // watch
			nodeLabelCommand(),    // label
			nodeCloseCommand(),    // close
		},
	}
}

func nodeListCommand() cli.Command {
	return cli.Command{
		Name:   "ls",
		Usage:  "list all of nodes",
		Flags:  listNodeFlags,
		Action: listNodes,
	}
}

func nodeInspectCommand() cli.Command {
	return cli.Command{
		Name:      "inspect",
		Usage:     "inspect details of a node",
		ArgsUsage: "NODE",
		Action:    inspectNode,
	}
}

func nodeStatsCommand() cli.Command {
	return cli.Command{
		Name:      "stats",
		Usage:     "watch live stream of node stats",
		ArgsUsage: "NODE",
		Flags:     verboseStatsFlags,
		Action:    statsNode,
	}
}

func nodeTerminalCommand() cli.Command {
	return cli.Command{
		Name:      "terminal",
		Usage:     "open a terminal on one give node",
		ArgsUsage: "NODE",
		Action:    terminalNode,
	}
}

func nodeExecCommand() cli.Command {
	return cli.Command{
		Name:      "exec",
		Usage:     "exec command in a specified node",
		ArgsUsage: "NODE",
		Flags:     execNodeFlags,
		Action:    execNode,
	}
}

func nodeWatchCommand() cli.Command {
	return cli.Command{
		Name:      "watch",
		Usage:     "watch events of a node",
		ArgsUsage: "NODE",
		Flags:     watchNodeFlags,
		Action:    watchNode,
	}
}

func nodeLabelCommand() cli.Command {
	return cli.Command{
		Name:  "label",
		Usage: "manage node labels",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "set labels on a node",
				ArgsUsage: "NODE name=value [name=value ...]",
				Action:    setNodeLabels,
			},
			{
				Name:      "get",
				Usage:     "get labels of a node",
				ArgsUsage: "NODE",
				Action:    getNodeLabels,
			},
			{
				Name:      "rm",
				Usage:     "remove labels from a node",
				ArgsUsage: "NODE key1 [key2 ...]",
				Flags:     removeNodeLabelFlags,
				Action:    removeNodeLabels,
			},
		},
	}
}

func nodeCloseCommand() cli.Command {
	return cli.Command{
		Name:      "close",
		Usage:     "close a specified node once",
		ArgsUsage: "NODE",
		Action:    closeNode,
	}
}

func listNodes(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		filter = c.String("filter") // label filter
		online = c.String("online") // online filter
		cldsvr = c.String("cldsvr") // cloudsvr filter
		status string
	)

	// label filter
	lbsFilter, err := label.Parse(filter)
	if err != nil {
		return err
	}

	// online filter
	if online != "" {
		if flag, _ := strconv.ParseBool(online); flag {
			status = "online"
		} else {
			status = "offline"
		}
	}

	nodes, err := client.ListNodes(lbsFilter, status, cldsvr)
	if err != nil {
		return err
	}

	// only print ids
	if c.Bool("quiet") {
		for _, node := range nodes {
			fmt.Fprintln(os.Stdout, node.ID)
		}
		return nil
	}

	var (
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	)

	parser, err := template.NewParser(nodeTableLine)
	if err != nil {
		return err
	}

	fmt.Fprint(w, nodeTableHeader)
	for _, node := range nodes {
		parser.Execute(w, node)
	}
	w.Flush()

	return nil
}

func inspectNode(c *cli.Context) error {
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

	node, err := client.InspectNode(nodeID)
	if err != nil {
		return err
	}

	return utils.PrettyJSON(nil, node)
}

func statsNode(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		nodeID  = c.Args().First()
		verbose = c.Bool("verbose")
	)

	if nodeID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	stream, err := client.WatchNodeStats(nodeID)
	if err != nil {
		return err
	}
	helpers.TrapExit(func() { stream.Close() })
	defer stream.Close()

	if verbose {
		io.Copy(os.Stdout, stream)
		return nil
	}

	var (
		dec = json.NewDecoder(stream)
		w   = tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	)
	for {
		info := new(types.SysInfo)
		err := dec.Decode(&info)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		time := time.Unix(info.UnixTime, 0).Format(time.RFC3339)
		hostname := info.Hostname
		load := fmt.Sprintf("%0.2f", info.LoadAvgs.One)
		cpu := fmt.Sprintf("%d", info.CPU.Used) + `%`
		memory := units.HumanSize(float64(info.Memory.Used)) + "/" + units.HumanSize(float64(info.Memory.Total))
		swap := units.HumanSize(float64(info.Swap.Used)) + "/" + units.HumanSize(float64(info.Swap.Total))
		var net string
		for inet, traffic := range info.Traffics {
			in := units.HumanSize(float64(traffic.RxBytes))
			out := units.HumanSize(float64(traffic.TxBytes))
			if net != "" {
				net += ","
			}
			net += fmt.Sprintf("%s[in/out]:%s/%s", inet, in, out)
		}

		fmt.Fprint(w, fmt.Sprintf("%s\t%s\tload:%s\tcpu:%s\tmem:%s\tswap:%s\tnet:%s\t\n",
			time, hostname, load, cpu, memory, swap, net))
		w.Flush()
	}
	return nil
}

func terminalNode(c *cli.Context) error {
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

	// firstly ensure the target node online
	node, err := client.InspectNode(nodeID)
	if err != nil {
		return fmt.Errorf("node %s not exists: %v", nodeID, err)
	}
	if node.Status != types.NodeStatusOnline {
		return fmt.Errorf("node %s status %s: %s", nodeID, node.Status, node.ErrMsg)
	}

	return client.OpenNodeTerminal(nodeID, os.Stdin, os.Stdout)
}

func execNode(c *cli.Context) error {
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

	var (
		cmd = c.String("cmd")
	)

	if cmd == "" {
		return cli.ShowSubcommandHelp(c)
	}

	stream, err := client.RunNodeCmd(nodeID, cmd)
	if err != nil {
		return err
	}
	helpers.TrapExit(func() { stream.Close() })
	defer stream.Close()

	io.Copy(os.Stdout, stream)
	return nil
}

func watchNode(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		nodeIDs = c.Args()
		all     = c.Bool("all")
	)

	// if all, query all current node ids
	if all || (len(nodeIDs) > 0 && strings.ToLower(nodeIDs[0]) == helpers.ParamAll) {
		nodes, err := client.ListNodes(nil, "", "")
		if err != nil {
			return err
		}
		nodeIDs = make([]string, 0, len(nodes)) // reset empty and refill with all node ids
		for _, node := range nodes {
			nodeIDs = append(nodeIDs, node.ID)
		}
		if len(nodeIDs) == 0 { // if use -a but empty nodes, directly return
			return nil
		}
	}

	if len(nodeIDs) == 0 { // must provide at least one node id
		return cli.ShowSubcommandHelp(c)
	}

	// on bunch of nodes by concurrency
	var (
		wg sync.WaitGroup
	)

	wg.Add(len(nodeIDs))
	for _, nodeID := range nodeIDs {

		var (
			colorf = color.SeqColorFunc() // sequence (non-repeated) color node prefix
			prefix = colorf(fmt.Sprintf("[%s]  | ", nodeID))
		)

		go func(nodeID, prefix string) {
			defer wg.Done()

			var (
				logf = func(content string) {
					fmt.Fprintln(os.Stdout, prefix+content)
				}
				errf = func(content string) {
					fmt.Fprintln(os.Stderr, prefix+color.Red(content))
				}
			)

			// request on node's event http endpoint
			stream, err := client.WatchNodeEvents(nodeID)
			if err != nil {
				errf(err.Error())
				return
			}
			defer stream.Close()

			// stream copy each line
			var (
				reader = bufio.NewReader(stream)
			)
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						if len(line) > 0 {
							goto PROCESS
						}
						break
					}
					errf(err.Error())
					return
				}

			PROCESS:
				line = bytes.TrimSuffix(line, []byte("\n")) // trim final '\n'
				if bytes.HasSuffix(line, []byte("\r")) {    // trim possible final '\r'
					line = bytes.TrimSuffix(line, []byte("\r"))
				}

				// only process with sse `data ` prefixed line
				if !bytes.HasPrefix(line, []byte("data: ")) {
					continue
				}
				line = bytes.TrimPrefix(line, []byte("data: "))

				ev := new(mole.NodeEvent)
				if err := json.Unmarshal(line, &ev); err != nil {
					errf(fmt.Sprintf("malformat sse event: %s", string(line)))
					continue
				}

				logf(fmt.Sprintf("%s: %s", ev.Time.Format(time.RFC3339), ev.Type))
			}

		}(nodeID, prefix)

	}
	wg.Wait()

	return nil
}

func setNodeLabels(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		nodeID = c.Args().First()
		labels = c.Args().Tail()
	)

	if nodeID == "" || len(labels) == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	lbs, err := label.Parse(strings.Join(labels, " "))
	if err != nil {
		return err
	}

	nowlbs, err := client.UpsertNodeLabels(nodeID, lbs)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s: %s\n", nodeID, nowlbs.String())
	return nil
}

func getNodeLabels(c *cli.Context) error {
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

	node, err := client.InspectNode(nodeID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s: %s\n", nodeID, node.Labels.String())
	return nil
}

func removeNodeLabels(c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		nodeID = c.Args().First()
		keys   = c.Args().Tail()
		all    = c.Bool("all") // if true `keys` won't take effective
	)

	if nodeID == "" {
		return cli.ShowSubcommandHelp(c)
	}

	if !all {
		if len(keys) == 0 {
			return cli.ShowSubcommandHelp(c)
		}
	}

	nowlbs, err := client.RemoveNodeLabels(nodeID, all, keys)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s: %s\n", nodeID, nowlbs.String())
	return nil
}

func closeNode(c *cli.Context) error {
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

	if err := client.CloseNode(nodeID); err != nil {
		return err
	}

	os.Stdout.Write(append([]byte("OK"), '\r', '\n'))
	return nil
}
