package scheduler

import (
	"time"

	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
	"github.com/bbklab/adbot/version"
	"gopkg.in/mgo.v2/bson"
)

// SummaryInfo is exported
func SummaryInfo() (*types.SummaryInfo, error) {
	info := &types.SummaryInfo{
		Version:    version.GetVersion() + "-" + version.GetGitCommit(),
		Uptime:     time.Since(sched.startAt).String(),
		StoreTyp:   store.DB().Type(),
		Listens:    make([]string, 0, 0),
		AdbNodes:   types.AdbNodeSummary{},
		AdbDevices: types.AdbDeviceSummary{},
		AdbOrders:  types.AdbOrderSummary{},
	}

	// adb nodes
	info.AdbNodes.Total = store.DB().CountNodes(nil)
	info.AdbNodes.Online = store.DB().CountNodes(bson.M{"status": types.NodeStatusOnline})
	info.AdbNodes.Offline = store.DB().CountNodes(bson.M{"status": types.NodeStatusOffline})

	// adb devices
	info.AdbDevices.Total = store.DB().CountAdbDevices(nil)
	info.AdbDevices.Online = store.DB().CountAdbDevices(bson.M{"status": types.AdbDeviceStatusOnline})
	info.AdbDevices.Offline = store.DB().CountAdbDevices(bson.M{"status": types.AdbDeviceStatusOffline})
	info.AdbDevices.OverQuota = store.DB().CountAdbDevices(bson.M{"over_quota": true})
	info.AdbDevices.WithinQuota = store.DB().CountAdbDevices(bson.M{"over_quota": false})

	// adb orders
	info.AdbOrders.Total = CountAdbOrdersByStatus(nil)
	todayStartAt, _ := utils.Today()
	info.AdbOrders.Today = CountAdbOrdersByStatus(bson.M{"created_at": bson.M{"$gt": todayStartAt}})
	monthStartAt, _ := utils.CurrMonth()
	info.AdbOrders.Month = CountAdbOrdersByStatus(bson.M{"created_at": bson.M{"$gt": monthStartAt}})

	return info, nil
}

// CountAdbOrdersByStatus count given query matched orders to types.AdbOrderStatistics
func CountAdbOrdersByStatus(query bson.M) types.AdbOrderStatistics {
	if query == nil {
		query = bson.M{}
	}

	var ret types.AdbOrderStatistics

	query["status"] = types.AdbOrderStatusPaid
	num, numFee := store.DB().CountAdbOrders(query)
	ret.Paid = num
	ret.PaidBill = float64(numFee) / float64(100)

	query["status"] = types.AdbOrderStatusPending
	num, numFee = store.DB().CountAdbOrders(query)
	ret.Pending = num
	ret.PendingBill = float64(numFee) / float64(100)

	query["status"] = types.AdbOrderStatusTimeout
	num, numFee = store.DB().CountAdbOrders(query)
	ret.Timeout = num
	ret.TimeoutBill = float64(numFee) / float64(100)

	return ret
}
