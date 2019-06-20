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
 AdbOrders:     paid:{{.AdbOrders.Total.Paid}}/{{.AdbOrders.Total.PaidBill}} pending:{{.AdbOrders.Total.Pending}}/{{.AdbOrders.Total.PendingBill}} timeout:{{.AdbOrders.Total.Timeout}}/{{.AdbOrders.Total.TimeoutBill}}
`

// SummaryInfo is exported
type SummaryInfo struct {
	Version    string           `json:"version"`
	Listens    []string         `json:"listens"`
	Uptime     string           `json:"uptime"`
	StoreTyp   string           `json:"store_type"`
	AdbNodes   AdbNodeSummary   `json:"adb_nodes"`   // status -> num
	AdbDevices AdbDeviceSummary `json:"adb_devices"` // status -> num
	AdbOrders  AdbOrderSummary  `json:"adb_orders"`
}

// WriteTo is exported
func (info *SummaryInfo) WriteTo(w io.Writer) (int64, error) {
	parser, err := template.NewParser(summaryInfoTemplate)
	if err != nil {
		return -1, err
	}
	return -1, parser.Execute(w, info) // just make pass govet
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

// AdbOrderSummary is exported
type AdbOrderSummary struct {
	Total           AdbOrderStatistics `json:"total"` // total
	RecentAdbOrders `json:",inline"`
}

// RecentAdbOrders is exported
type RecentAdbOrders struct {
	Today AdbOrderStatistics `json:"today"` // this day
	Month AdbOrderStatistics `json:"month"` // this month
}

// AdbOrderStatistics is exported
type AdbOrderStatistics struct {
	Paid        int     `json:"paid"`
	PaidBill    float64 `json:"paid_bill"`
	Pending     int     `json:"pending"`
	PendingBill float64 `json:"pending_bill"`
	Timeout     int     `json:"timeout"`
	TimeoutBill float64 `json:"timeout_bill"`
}
