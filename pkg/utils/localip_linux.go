package utils

import (
	"net"
	"path/filepath"
	"syscall"
)

var (
	sysd = "/sys"
)

// ListLocalIPs list local inet device's ipv4 addresses
func ListLocalIPs(skipVirtual bool) []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	ret := make([]string, 0)
	for _, iface := range ifaces {
		name := iface.Name
		if skipVirtual && isVirtualNetDev(name) { // skip virtual network devices
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
				if ip.IP.To4() != nil { // remove mac addres.
					ret = append(ret, ip.IP.String())
				}
			}
		}
	}
	return ret
}

func isVirtualNetDev(devname string) bool {
	err := syscall.Access(filepath.Join(sysd, "devices/virtual/net", devname), syscall.F_OK)
	return err == nil
}
