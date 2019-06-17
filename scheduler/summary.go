package scheduler

import (
	"time"

	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
	"github.com/bbklab/adbot/version"
)

// SummaryInfo is exported
func SummaryInfo() (*types.SummaryInfo, error) {
	info := &types.SummaryInfo{
		Version:  version.GetVersion() + "-" + version.GetGitCommit(),
		Uptime:   time.Since(sched.startAt).String(),
		StoreTyp: store.DB().Type(),
		Listens:  make([]string, 0, 0),
	}

	nodes, err := store.DB().ListNodes(nil, nil)
	if err != nil {
		return nil, err
	}
	info.AdbNodes = types.AdbNodeSummary{}
	for _, node := range nodes {
		info.AdbNodes.Total++
		switch node.Status {
		case types.NodeStatusOnline:
			info.AdbNodes.Online++
		case types.NodeStatusOffline:
			info.AdbNodes.Offline++
		}
	}

	devices, err := store.DB().ListAdbDevices(nil, nil)
	if err != nil {
		return nil, err
	}
	info.AdbDevices = types.AdbDeviceSummary{}
	for _, device := range devices {
		info.AdbDevices.Total++
		switch device.Status {
		case types.AdbDeviceStatusOnline:
			info.AdbDevices.Online++
		case types.AdbDeviceStatusOffline:
			info.AdbDevices.Offline++
		}
		if device.OverQuota {
			info.AdbDevices.OverQuota++
		} else {
			info.AdbDevices.WithinQuota++
		}
	}

	return info, nil
}
