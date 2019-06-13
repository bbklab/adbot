package types

import (
	"io"

	"github.com/bbklab/adbot/pkg/template"
)

var summaryInfoTemplate = ` Version:       {{.Version}}
 Listens:       {{.Listens}}
 Uptime:        {{.Uptime}}
 Store:         {{.StoreTyp}}
 AdbNodes:      online:{{.AdbNodes.Online}} offline:{{.AdbNodes.Offline}}
 AdbDevices:    online:{{.AdbDevices.Online}} offline:{{.AdbDevices.Offline}} overquota:{{.AdbDevices.OverQuota}} withinquota:{{.AdbDevices.WithinQuota}}
`

// SummaryInfo is exported
type SummaryInfo struct {
	Version    string           `json:"version"`
	Listens    []string         `json:"listens"`
	Uptime     string           `json:"uptime"`
	StoreTyp   string           `json:"store_type"`
	AdbNodes   AdbNodeSummary   `json:"adb_nodes"`   // status -> num
	AdbDevices AdbDeviceSummary `json:"adb_devices"` // status -> num
}

// AdbNodeSummary is exported
type AdbNodeSummary struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
}

// AdbDeviceSummary is exported
type AdbDeviceSummary struct {
	Total       int `json:"total"`
	Online      int `json:"online"`
	Offline     int `json:"offline"`
	OverQuota   int `json:"over_quota"`
	WithinQuota int `json:"within_quota"`
}

// WriteTo is exported
func (info *SummaryInfo) WriteTo(w io.Writer) (int64, error) {
	parser, err := template.NewParser(summaryInfoTemplate)
	if err != nil {
		return -1, err
	}
	return -1, parser.Execute(w, info) // just make pass govet
}
