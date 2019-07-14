package extensions

import (
	"net"
	"strings"

	"github.com/cloudfoundry/gosigar"

	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
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
	g.info.Traffics = make(map[string]*types.NetTraffic)
	return nil
}

func (g *gatherer) diskstats() error {
	g.info.DisksIO = make(map[string]*types.DiskIOInfo)
	return nil
}

// obtain local docker informations, perform like followings:
//   curl -s --unix-socket /var/run/docker.sock http:/info | jq
func (g *gatherer) docker() error {
	g.info.Docker = types.DockerInfo{}
	return nil
}

func (g *gatherer) bbr() error {
	g.info.BBREnabled = false
	return nil
}

func (g *gatherer) withMaster() error {
	g.info.WithMaster = false
	return nil
}

func (g *gatherer) manufacturer() error {
	g.info.Manufacturer, _ = utils.GetHardwareProductName()
	return nil
}
