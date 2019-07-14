package extensions

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cloudfoundry/gosigar"

	"github.com/bbklab/adbot/pkg/file"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

const (
	diskSectorSize = 512
)

var (
	procd = "/proc"
	sysd  = "/sys"
)

// refers:
//  https://github.com/prometheus/node_exporter/blob/751996761903af4cffe18cf2d980b8ae9a202204/collector/filesystem_linux.go#L29
var (
	ignoredMountPoints        = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"
	ignoredFSTypes            = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"
	ignoredMountPointsPattern = regexp.MustCompile(ignoredMountPoints)
	ignoredFSTypesPattern     = regexp.MustCompile(ignoredFSTypes)
)

func (g *gatherer) ips() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	ips := make(map[string][]string)
	for _, iface := range ifaces {
		if iface.Name == "" {
			continue
		}
		if isVirtualNetDev(iface.Name) { // skip virtual netdev, eg : lo, vethx, docker0
			continue
		}
		iface.Name = strings.NewReplacer([]string{
			".", "-",
			"$", "-",
		}...).Replace(iface.Name)
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		ifaddrs := []string{}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ifaddrs = append(ifaddrs, ipnet.IP.String())
				}
			}
		}
		if len(ifaddrs) > 0 {
			ips[iface.Name] = ifaddrs
		}
	}

	g.info.IPs = ips
	return nil
}

func (g *gatherer) disks() error {
	fslist := sigar.FileSystemList{}
	if err := fslist.Get(); err != nil {
		return err
	}

	disks := make(map[string]*types.DiskInfo)
	for _, fs := range fslist.List {
		if !strings.HasPrefix(fs.DevName, "/dev") {
			continue
		}
		if ignoredMountPointsPattern.MatchString(fs.DirName) { // ignore mount point
			continue
		}
		if ignoredFSTypesPattern.MatchString(fs.SysTypeName) { // ignore fs type
			continue
		}
		usage := sigar.FileSystemUsage{}
		if err := usage.Get(fs.DirName); err != nil {
			return err
		}
		disks[fs.DevName] = &types.DiskInfo{
			DevName: fs.DevName,
			MountAt: fs.DirName,
			Total:   usage.Total,
			Used:    usage.Used,
			Free:    usage.Free,
			Inode:   usage.Files,
			Ifree:   usage.FreeFiles,
		}
	}

	g.info.Disks = disks
	return nil
}

func (g *gatherer) traffics() error {
	info := make(map[string]*types.NetTraffic) // inet name -> inet traffic

	err := processFile(filepath.Join(procd, "net/dev"), func(line string) {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if !strings.ContainsRune(fields[0], ':') {
			return
		}

		devname := fields[0][:len(fields[0])-1]

		//skip virtual netdev, eg : lo, vethx, docker0
		if isVirtualNetDev(devname) {
			return
		}

		//skip inactive netdev
		if fields[1] == "0" && fields[9] == "0" {
			return
		}

		nc := new(types.NetTraffic)
		nc.Name = devname
		nc.RxBytes, _ = strconv.ParseUint(fields[1], 10, 64)
		nc.TxBytes, _ = strconv.ParseUint(fields[9], 10, 64)
		nc.RxPackets, _ = strconv.ParseUint(fields[2], 10, 64)
		nc.TxPackets, _ = strconv.ParseUint(fields[10], 10, 64)
		nc.Mac, _ = getMacAddr(devname)
		nc.Time = time.Now()
		info[devname] = nc
	})

	if err != nil {
		return err
	}

	g.info.Traffics = info
	return nil
}

func (g *gatherer) diskstats() error {
	info := make(map[string]*types.DiskIOInfo)

	err := processFile(filepath.Join(procd, "diskstats"), func(line string) {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 14 {
			return
		}

		devname := fields[2]

		// skip virtual block device, eg : loop0, dm-1
		if isVirtualBlock(devname) {
			return
		}

		// skip inactive block device
		if fields[6] == "0" && fields[10] == "0" {
			return
		}

		reads, err := strconv.ParseUint(fields[6], 10, 64)
		if err != nil {
			return
		}
		writes, err := strconv.ParseUint(fields[10], 10, 64)
		if err != nil {
			return
		}

		info[devname] = &types.DiskIOInfo{
			DevName:    devname,
			ReadBytes:  reads * diskSectorSize,
			WriteBytes: writes * diskSectorSize,
		}
	})

	if err != nil {
		return err
	}

	g.info.DisksIO = info
	return nil
}

// obtain local docker informations, perform like followings:
//   curl -s --unix-socket /var/run/docker.sock http:/info | jq
func (g *gatherer) docker() error {
	var (
		info = types.DockerInfo{}
	)

	dinfo, err := localDockerInfo()
	if err != nil {
		goto END
	}

	info.Version = dinfo.ServerVersion
	info.NumImages = dinfo.Images
	info.NumContainers = dinfo.Containers
	info.NumRunningContainers = dinfo.ContainersRunning
	info.Driver = dinfo.Driver
	info.DriverStatus = make(map[string]string)
	for _, status := range dinfo.DriverStatus {
		if len(status) >= 2 {
			info.DriverStatus[status[0]] = status[1]
		}
	}

END:
	g.info.Docker = info
	return nil
}

func (g *gatherer) bbr() error {
	bs, _ := ioutil.ReadFile("/proc/sys/net/ipv4/tcp_congestion_control")
	g.info.BBREnabled = strings.ToLower(strings.TrimSpace(string(bs))) == "bbr"
	return nil
}

func (g *gatherer) withMaster() error {
	g.info.WithMaster = file.Exists("/var/run/adbot/adbot.sock")
	return nil
}

func (g *gatherer) manufacturer() error {
	g.info.Manufacturer, _ = utils.GetHardwareProductName()
	return nil
}

type dockerInfo struct {
	Containers        int64      `json:"Containers"`
	ContainersRunning int64      `json:"ContainersRunning"`
	Images            int64      `json:"Images"`
	ServerVersion     string     `json:"ServerVersion"`
	Driver            string     `json:"Driver"`
	DriverStatus      [][]string `json:"DriverStatus"`
}

// utils
//
//
func localDockerInfo() (*dockerInfo, error) {
	var (
		dinfo  = new(dockerInfo)
		req, _ = http.NewRequest("GET", "http://what-ever/info", nil)
	)
	req.Close = true
	req.Header.Set("Connection", "close")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := dockerClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return dinfo, json.NewDecoder(resp.Body).Decode(&dinfo)
}

func isVirtualNetDev(devname string) bool {
	err := syscall.Access(filepath.Join(sysd, "devices/virtual/net", devname), syscall.F_OK)
	return err == nil
}

func isVirtualBlock(devname string) bool {
	err := syscall.Access(filepath.Join(sysd, "devices/virtual/block", devname), syscall.F_OK)
	return err == nil
}

func getMacAddr(devname string) (string, error) {
	mac, err := ioutil.ReadFile(filepath.Join(sysd, "class/net", devname, "address"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(mac)), nil
}
