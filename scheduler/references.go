package scheduler

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

// node
//

// MemoNodeErrmsg update node's ErrMsg
func MemoNodeErrmsg(nodeID, errmsg string) error {
	update := bson.M{"$set": bson.M{"error": errmsg}}
	return store.DB().UpdateNode(nodeID, update)
}

// MemoNodeStatus update node's Status/LastActiveAt/SysInfo/Latency
func MemoNodeStatus(nodeID, status string, sysinfo *types.SysInfo, latency *time.Duration) error {
	node, err := store.DB().GetNode(nodeID)
	if err != nil {
		return err
	}

	setUpdator := bson.M{}
	if status != "" {
		setUpdator["status"] = status
	}
	if status == types.NodeStatusOnline {
		setUpdator["last_active_at"] = time.Now()
		setUpdator["error"] = ""
	}
	if sysinfo != nil {
		RefreshTrafficRate(node.SysInfo, sysinfo) // caculate the net traffic rate and update the newly sysinfo
		setUpdator["sysinfo"] = sysinfo
	}
	if latency != nil {
		setUpdator["latency"] = ptype.TimeDurationV(latency)
	}

	return store.DB().UpdateNode(nodeID, bson.M{"$set": setUpdator})
}

// node labels
//

// UpsertNodeLabels upsert label pairs to db node
func UpsertNodeLabels(id string, lbs label.Labels) error {
	setUpdator := bson.M{}
	for key, val := range lbs {
		setUpdator[fmt.Sprintf("labels.%s", key)] = val
	}
	return store.DB().UpdateNode(id, bson.M{"$set": setUpdator})
}

// RemoveNodeLabels remove label pair from db node
func RemoveNodeLabels(id string, all bool, keys []string) error {
	if all {
		update := bson.M{"$set": bson.M{"labels": label.New(nil)}}
		return store.DB().UpdateNode(id, update)
	}

	setUpdator := bson.M{}
	for _, key := range keys {
		setUpdator[fmt.Sprintf("labels.%s", key)] = 1
	}
	return store.DB().UpdateNode(id, bson.M{"$unset": setUpdator})
}

// settings attrs(labels)
//

// UpsertSettingsAttr upsert label pair of db Settings.GlobalAttrs
func UpsertSettingsAttr(attrs label.Labels) error {
	setUpdator := bson.M{
		"initial":    false,
		"updated_at": time.Now(),
	}
	for key, val := range attrs {
		setUpdator[fmt.Sprintf("global_attrs.%s", key)] = val
	}
	return store.DB().UpsertSettings(bson.M{"$set": setUpdator})
}

// RemoveSettingsAttr remove label pair from db Settings.GlobalAttrs
func RemoveSettingsAttr(all bool, keys []string) error {
	if all {
		update := bson.M{"$set": bson.M{"global_attrs": label.New(nil), "initial": false, "updated_at": time.Now()}}
		return store.DB().UpsertSettings(update)
	}

	setUpdator := bson.M{
		"initial":    false,
		"updated_at": time.Now(),
	}
	for _, key := range keys {
		setUpdator[fmt.Sprintf("global_attrs.%s", key)] = 1
	}
	return store.DB().UpsertSettings(bson.M{"$unset": setUpdator})
}

// MemoSettings update db Settings
func MemoSettings(update interface{}) error {
	return store.DB().UpsertSettings(update)
}

// MemoSettingsSet update db Settings
func MemoSettingsSet(req *types.UpdateSettingsReq) error {
	setUpdator := bson.M{}
	if req.LogLevel != nil {
		setUpdator["log_level"] = *req.LogLevel
	}
	if req.EnableHTTPMuxDebug != nil {
		setUpdator["enable_httpmux_debug"] = *req.EnableHTTPMuxDebug
	}
	if req.UnmarkSensitive != nil {
		setUpdator["unmask_sensitive"] = *req.UnmarkSensitive
	}
	if req.TGBotToken != nil {
		setUpdator["tg_bot_token"] = *req.TGBotToken
	}
	return store.DB().UpsertSettings(bson.M{"$set": setUpdator})
}

//
// adb devices
//

// MemoAdbDeviceStatus update db AdbDevice's Status/Errmsg
func MemoAdbDeviceStatus(id, status, errmsg string) error {
	setUpdator := bson.M{"status": status}
	setUpdator["error"] = errmsg
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoAdbDeviceSysinfo update db AdbDevice's NodeID/SysInfo
func MemoAdbDeviceSysinfo(id, nid string, sysinfo *adbot.AndroidSysInfo) error {
	setUpdator := bson.M{"node_id": nid}
	setUpdator["sysinfo"] = sysinfo
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoNodeAllAdbDeviceStatus update one node's all of AdbDevice's Status/Errmsg
func MemoNodeAllAdbDeviceStatus(nid, status, errmsg string) error {
	dvcs, _ := store.DB().ListAdbDevices(nil, bson.M{"node_id": nid})
	for _, dvc := range dvcs {
		MemoAdbDeviceStatus(dvc.ID, status, errmsg)
	}
	return nil
}

// MemoAdbDeviceDesc update db AdbDevice's Desc
func MemoAdbDeviceDesc(id, desc string) error {
	setUpdator := bson.M{"desc": desc}
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoAdbDeviceMaxBill update db AdbDevice's MaxBill
func MemoAdbDeviceMaxBill(id string, bill int) error {
	setUpdator := bson.M{"max_bill": bill}
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoAdbDeviceMaxAmount update db AdbDevice's MaxAmount
func MemoAdbDeviceMaxAmount(id string, amount int) error {
	setUpdator := bson.M{"max_amount": amount}
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoAdbDeviceWeight update db AdbDevice's Weight
func MemoAdbDeviceWeight(id string, weight int) error {
	setUpdator := bson.M{"weight": weight}
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoAdbDeviceOverQuota update db AdbDevice's OverQuota
func MemoAdbDeviceOverQuota(id string, flag bool) error {
	setUpdator := bson.M{"over_quota": flag}
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

// MemoAdbDeviceAlipay update db AdbDevice's Alipay
func MemoAdbDeviceAlipay(id string, alipay *types.AlipayAccount) error {
	setUpdator := bson.M{"alipay": alipay}
	return store.DB().UpdateAdbDevice(id, bson.M{"$set": setUpdator})
}

//
// adb orders
//

// MemoAdbOrderStatus update db Adb Order's Status
func MemoAdbOrderStatus(orderID, status string) error {
	setUpdator := bson.M{"status": status}
	if status == types.AdbOrderStatusPaid {
		setUpdator["paid_at"] = time.Now()
	}
	update := bson.M{"$set": setUpdator}
	return store.DB().UpdateAdbOrder(orderID, update)
}

// MemoAdbOrderResponse update db Adb Order's Response
func MemoAdbOrderResponse(orderID string, resp *types.NewAdbOrderResp) error {
	update := bson.M{"$set": bson.M{"response": resp}}
	return store.DB().UpdateAdbOrder(orderID, update)
}

// MemoAdbOrderCallback update db Adb Order's Callback
func MemoAdbOrderCallback(orderID string, cb *types.NewAdbOrderCallback) error {
	update := bson.M{"$set": bson.M{"callback": cb}}
	return store.DB().UpdateAdbOrder(orderID, update)
}

// MemoAdbOrderCallbackStatus update db Adb Order's CallbackStatus
func MemoAdbOrderCallbackStatus(orderID, status string) error {
	update := bson.M{"$set": bson.M{"callback_status": status}}
	return store.DB().UpdateAdbOrder(orderID, update)
}

// AppendAdbOrderCallbackHistory push db Adb Order's CallbackHistory
func AppendAdbOrderCallbackHistory(orderID, errmsg string) error {
	var history = fmt.Sprintf("%s: ", time.Now().Format(time.RFC3339))
	if errmsg == "" {
		history += "succeed"
	} else {
		history += errmsg
	}
	update := bson.M{"$push": bson.M{"callback_history": history}}
	return store.DB().UpdateAdbOrder(orderID, update)
}

// the others
// memo(update) ops on db store
//

// MemoUserDesc update db User.Desc
func MemoUserDesc(userID, desc string) error {
	update := bson.M{"$set": bson.M{"desc": desc, "updated_at": time.Now()}}
	return store.DB().UpdateUser(userID, update)
}

// MemoUserPassword update db User.Password
func MemoUserPassword(userID string, password types.Password) error {
	update := bson.M{"$set": bson.M{"password": password, "updated_at": time.Now()}}
	return store.DB().UpdateUser(userID, update)
}

// MemoUserLastLoginAt update db User.LastLoginAt
func MemoUserLastLoginAt(userID string) error {
	update := bson.M{"$set": bson.M{"last_login_at": time.Now()}}
	return store.DB().UpdateUser(userID, update)
}
