package scheduler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/balancer"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/pkg/qrcode"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

var (
	errSignError = errors.New("signature error")
)

var (
	// failure retry interval for sending our callback
	sendFailureRetry = []time.Duration{
		time.Second * 10,  // 10s
		time.Second * 30,  // 30s
		time.Second * 120, // 2m
		time.Second * 300, // 5m
		time.Second * 900, // 15m
	}
)

// VerifySignature verify the given request
func VerifySignature(key string, data interface{}) error {
	switch req := data.(type) {
	case nil:
		return errors.New("nil data for signature verify")

	case *types.NewAdbOrderReq:
		var expect = sign(key, req.StringToSign())
		if req.Sign != expect {
			return errSignError
		}

	default:
		return errors.New("unexpected data type for signature verify")
	}
	return nil
}

func sign(key, data string) string {
	return strings.ToUpper(utils.Md5sum([]byte(strings.TrimSpace(data + "&key=" + key))))
}

// SubscribeAdbOrderAndSendCallback subscribe wait given adb order's callback until timeout and
// send our callback to adb order's NotifyURL according by given failure retry interval
//
// note: this may take a long time
func SubscribeAdbOrderAndSendCallback(orderID string) {
	RegisterGoroutine("adb_order_callback", orderID)
	defer DeRegisterGoroutine("adb_order_callback", orderID)

	// wait adb order callback event
	err := SubscribeAdbOrderCallbackEvent(orderID, types.AdbOrderTimeout)
	if err != nil {
		MemoAdbOrderStatus(orderID, types.AdbOrderStatusTimeout)
		return
	}

	// now got callback! send our callback
	MemoAdbOrderCallbackStatus(orderID, types.AdbOrderCallbackStatusOngoing)
	err = SendAdbOrderCallback(orderID, sendFailureRetry)
	if err != nil {
		MemoAdbOrderCallbackStatus(orderID, types.AdbOrderCallbackStatusError) // see callback history
		return
	}
	MemoAdbOrderCallbackStatus(orderID, types.AdbOrderCallbackStatusSucceed)
}

// BootupReCallbackAbortedAdbOrders is called while bootup to
// re-sending callback for aborted adb order callback
//
// note: this may take a long time
func BootupReCallbackAbortedAdbOrders(orders []*types.AdbOrder) {
	for _, order := range orders {
		log.Printf("boot up recallback aborted adb order %s", order.ID)

		MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusOngoing)
		err := SendAdbOrderCallback(order.ID, sendFailureRetry)
		if err != nil {
			MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusError) // see callback history
			continue
		}
		MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusSucceed)
	}
}

// SendAdbOrderCallback send given adb order's callback with given retry interval
//
// note: this may take a long time
func SendAdbOrderCallback(orderID string, retryInterval []time.Duration) error {
	// status must has already been memo update the db adb order
	order, err := store.DB().GetAdbOrder(orderID)
	if err != nil {
		return err
	}

	// skip if notify url not provided
	notifyURL := order.NotifyURL
	if notifyURL == "" {
		log.Printf("adb order %s notify url not provided, skip sending callback", order.ID)
		return nil
	}

	// construct our callback
	callback, err := genCallbackOfAdbOrder(order)
	if err != nil {
		return err
	}
	MemoAdbOrderCallback(orderID, callback)

	// send once
	err = sendCallbackOnce(notifyURL, callback)
	if err == nil { // succeed, return
		AppendAdbOrderCallbackHistory(orderID, "")
		return nil
	}
	AppendAdbOrderCallbackHistory(orderID, err.Error())

	// failed, retry
	for _, timeout := range retryInterval {
		time.Sleep(timeout)
		err = sendCallbackOnce(notifyURL, callback)
		if err == nil {
			AppendAdbOrderCallbackHistory(orderID, "")
			return nil
		}
		AppendAdbOrderCallbackHistory(orderID, err.Error())
	}

	return err // final failed, no more retry
}

// construct our callback
func genCallbackOfAdbOrder(order *types.AdbOrder) (*types.NewAdbOrderCallback, error) {
	if order.Status != types.AdbOrderStatusPaid {
		return nil, errors.New("can't generate callback for `un-paid` order")
	}

	// note: in fact the same signature algoritmo with the request
	// so we don't need to re-caculate the callback signature
	// key := settings.GlobalAttrs.Get(types.GlobalAttrPaygateSecretKey)
	// sign := sign(key, fmt.Sprintf("out_order_id=%s&fee=%d", order.OutOrderID, order.Fee))
	sign := order.Sign

	return &types.NewAdbOrderCallback{
		Code:       1,
		OutOrderID: order.OutOrderID,
		Fee:        order.Fee,
		Attach:     order.Attach,
		Sign:       sign,
		Time:       time.Now(),
	}, nil
}

func sendCallbackOnce(url string, cb *types.NewAdbOrderCallback) error {
	var (
		cbbs, _ = json.Marshal(cb)
		ctype   = "application/json; charset=UTF-8"
	)

	resp, err := utils.InsecureHTTPClient().Post(url, ctype, bytes.NewBuffer(cbbs))
	if err != nil {
		return err
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		return fmt.Errorf("%d - %s", code, utils.Truncate(string(bs), 20))
	}

	return nil
}

// SmartPickupAdbDevice pick up an avaliable device from all of adb devices
func SmartPickupAdbDevice(req *types.NewAdbOrderReq) (*types.AdbDevice, error) {
	// list sutiable adb devices
	query := bson.M{
		"status":     types.AdbDeviceStatusOnline, // only online devices
		"weight":     bson.M{"$gt": 0},            // only weight > 0 devices
		"over_quota": false,                       // only not over-quoted devices
	}
	switch req.QRType {
	case types.QRCodeTypeAlipay:
		query["alipay"] = bson.M{"$ne": nil}
	case types.QRCodeTypeWxpay:
		query["wxpay"] = bson.M{"$ne": nil}
	}
	dvcs, _ := store.DB().ListAdbDevices(nil, query)
	if len(dvcs) == 0 {
		return nil, errors.New("no available adb devices")
	}

	// pick up one from device list by weight balancer
	var (
		wb    = balancer.NewWeight()
		items = make([]balancer.Item, len(dvcs))
	)
	for idx, dvc := range dvcs {
		items[idx] = dvc
	}
	next := wb.Next(items)
	if next == nil {
		return nil, errors.New("can't select a adb device by weight balancer")
	}

	return next.(*types.AdbDevice), nil
}

// GenAdbpayQrCode generate qrcode for given adb device
func GenAdbpayQrCode(dvcid, typ string, fee int, comment string) ([]byte, string, error) {
	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		return nil, "", err
	}

	var (
		feeYuan = fmt.Sprintf("%0.2f", float64(fee)/float64(100))
	)

	switch typ {
	case types.QRCodeTypeAlipay:
		alipay := dvc.Alipay
		if alipay == nil {
			return nil, "", errors.New("device hasn't bind any alipay account yet")
		}

		qrtext := genAlipaySchemeURL(alipay.UserID, feeYuan, comment)
		qrpng, err := qrcode.Encode(qrtext) // png data
		if err != nil {
			return nil, "", err
		}
		return qrpng, qrtext, nil

	case types.QRCodeTypeWxpay:
		wxpay := dvc.Wxpay
		if wxpay == nil {
			return nil, "", errors.New("device hasn't bind any wxpay account yet")
		}
	}

	return nil, "", errors.New("unsupported qrcode type")
}

func genAlipaySchemeURL(alipayUserID, feeYuan, comment string) string {
	appID := "20000123" // TODO use our own appID
	bizdata := `{"s":"money","u":"` + alipayUserID + `","a":"` + feeYuan + `","m":"` + comment + `"}`
	return fmt.Sprintf("alipays://platformapi/startapp?appId=%s&actionType=scan&biz_data=%s", appID, bizdata)
}

// RunAdbDeviceLimitCheckerLoop periodically check all db adb device bill & amount of today and
// memo update the adb device .OverQuota
func RunAdbDeviceLimitCheckerLoop() {
	var (
		loopName = fmt.Sprintf("adb device quota limit checker loop")
	)

	RegisterGoroutine("adb_devices_limit_checker", "system")
	defer DeRegisterGoroutine("adb_devices_limit_checker", "system")

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s, this should never happen", loopName)

	// periodically timer notifier
	ticker := time.NewTicker(time.Minute * 30) // status collect ticker: 30m
	defer ticker.Stop()

	for range ticker.C {
		dvcs, _ := store.DB().ListAdbDevices(nil, nil)
		for _, dvc := range dvcs {
			if checkOneAdbDeviceLimit(dvc) {
				log.Warnf("adb device %s reached perday limit[maxbill=%d,maxamount=%d], mark device as over-quoted", dvc.ID, dvc.MaxBill, dvc.MaxAmount)
				MemoAdbDeviceOverQuota(dvc.ID, true)
			}
		}
	}
}

func checkOneAdbDeviceLimit(dvc *types.AdbDevice) bool {
	var (
		maxamount = dvc.MaxAmount
		maxbill   = dvc.MaxBill
	)
	if maxamount == 0 && maxbill == 0 { // unlimited device
		return false
	}

	num, fee := countAdbDeviceTodayPaidOrders(dvc.ID)
	if maxamount > 0 && fee > maxamount {
		return true
	}
	if maxbill > 0 && num > maxbill {
		return true
	}
	return false
}

func countAdbDeviceTodayPaidOrders(dvcID string) (int, int) {
	var (
		start, end = utils.Today()
		query      = bson.M{
			"device_id": dvcID,
			"status":    types.AdbOrderStatusPaid,
			"$and":      []bson.M{{"created_at": bson.M{"$gt": start}}, {"created_at": bson.M{"$lt": end}}},
		}
	)
	return store.DB().CountAdbOrders(query)
}

// ResetAllAdbDevicesOverQuotaFlag mark all of adb devices as `not over quoted`
func ResetAllAdbDevicesOverQuotaFlag() {
	log.Println("reset all adb devices over quota flag ...")
	dvcs, _ := store.DB().ListAdbDevices(nil, nil)
	for _, dvc := range dvcs {
		MemoAdbDeviceOverQuota(dvc.ID, false)
	}
}

// runAdbNodeReFreshLoop launch db adb node device refresher until node is removed
// note: launched by node join call back only on the first join in the runtime
//
// - periodically collect node adb devices status and refresh db adb device status
// until the db node is removed
func runAdbNodeReFreshLoop(node *mole.ClusterAgent) {
	var (
		id       = node.ID()
		loopName = fmt.Sprintf("adb node %s db refresher loop", id)
	)

	RegisterGoroutine("adbnode_refresher", id)
	defer DeRegisterGoroutine("adbnode_refresher", id)

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s, node maybe removed", loopName)

	// node event notifier
	nodeEvSub := node.Events()
	defer node.EvictEventSubscriber(nodeEvSub)

	// periodically timer notifier
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	// async refresher notifier
	nodeRefresh := make(chan struct{}, 1)
	AddAdbNodeRefreshNotifier(id, nodeRefresh) // register adb node async refresh notifier
	defer DelAdbNodeRefreshNotifier(id)        // avoid leaks while follings steps met errors

	// trigger adb node refresh once immediately before actually launch the loop
	go RefreshAdbNodeAsync(id)

	var err error

	for {
		select {

		case <-ticker.C: // update periodically
			time.Sleep(time.Second * time.Duration(rand.Int()%5+1)) // randomize the refresh loop to smooth potential burst load
			err = adbNodeRefreshOnce(id)

		case <-nodeRefresh: // async node refresh triggered
			err = adbNodeRefreshOnce(id)

		case ev := <-nodeEvSub: // node events that we should refresh adb node devices
			switch ev.(*mole.NodeEvent).Type {
			case mole.NodeEvJoin, mole.NodeEvRejoin, mole.NodeEvFlagging, mole.NodeEvShutdown, mole.NodeEvDie:
				go RefreshAdbNodeAsync(id)
			}
		}

		if err != nil {
			log.Errorf("%s got error: [%v], retry ...", loopName, err)
		}

		if err == errAdbNodeNotFound {
			return // db node maybe removed
		}
	}
}

var errAdbNodeNotFound = errors.New("db node not found")

func adbNodeRefreshOnce(nodeID string) error {
	// ensure we're the leader, otherwise retry
	if !isLeader() {
		return nil
	}

	// ensure db node exists, otherwise tell the loop to exit
	_, err := store.DB().GetNode(nodeID)
	if err != nil {
		if store.DB().ErrNotFound(err) {
			return errAdbNodeNotFound // use specified type error to identify node not exists
		}
		return err
	}

	// list node adb devices
	dbdvcs, err := store.DB().ListAdbDevices(nil, bson.M{"node_id": nodeID})
	if err != nil {
		return err
	}

	// query node adb devices sysinfo
	dvcsInfo, err := DoNodeQueryAdbDevices(nodeID)
	if err != nil {
		MemoNodeAllAdbDeviceStatus(nodeID, types.AdbDeviceStatusOffline, err.Error()) // mark node all devices as offline
		return err
	}

	// memo update collected adb devices ...
	for id, info := range dvcsInfo {
		if _, err := store.DB().GetAdbDevice(id); store.DB().ErrNotFound(err) {
			store.DB().AddAdbDevice(&types.AdbDevice{
				ID:        id,
				NodeID:    nodeID,
				SysInfo:   info,
				Status:    types.AdbDeviceStatusOffline,
				Errmsg:    "",
				MaxAmount: 0,
				MaxBill:   0,
				OverQuota: false,
				Weight:    0,
				Alipay:    nil,
				Wxpay:     nil,
			})
		} else {
			MemoAdbDeviceStatus(id, types.AdbDeviceStatusOnline, "")
			MemoAdbDeviceSysinfo(id, nodeID, info)
		}
	}

	// caculate the missing(uncollected) adb devices and memo update them
	var missingDvcs []string
	for _, dvc := range dbdvcs {
		if _, ok := dvcsInfo[dvc.ID]; !ok {
			missingDvcs = append(missingDvcs, dvc.ID)
		}
	}
	for _, dvcid := range missingDvcs {
		MemoAdbDeviceStatus(dvcid, types.AdbDeviceStatusOffline, "adb device info not collected on this node")
	}

	return nil
}

// RunAdbEventWatcherLoop watch all adb events and handle them
func RunAdbEventWatcherLoop() {
	var (
		loopName = fmt.Sprintf("adb events watcher loop")
	)

	RegisterGoroutine("adb_events_watcher", "system")
	defer DeRegisterGoroutine("adb_events_watcher", "system")

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s", loopName)

	// obtain adb device subscriber
	sub := SubscribeAdbDeviceEvents()
	defer EvictAdbDeviceEvents(sub)

	// write adb device event to the client with sse format
	for ev := range sub {
		var (
			adbev = ev.(*adbot.AdbEvent)
			dvcid = adbev.Serial
			typ   = adbev.Type
		)
		switch typ {
		case adbot.AdbEventDeviceDie:
			MemoAdbDeviceStatus(dvcid, types.AdbDeviceStatusOffline, "adb device offline event")
		case adbot.AdbEventDeviceAlive:
			MemoAdbDeviceStatus(dvcid, types.AdbDeviceStatusOnline, "")
		case adbot.AdbEventAlipayOrder:
			go checkDevicePendingOrders(dvcid)
		}
	}
}

// sequential check device pending orders
var dvcpendingmux sync.Mutex

// note: only check pending orders within 5 minutes
func checkDevicePendingOrders(dvcid string) {
	dvcpendingmux.Lock()
	defer dvcpendingmux.Unlock()

	RegisterGoroutine("check_adb_device_pending_orders", dvcid)
	defer DeRegisterGoroutine("check_adb_device_pending_orders", dvcid)

	// query device pending orders
	query := bson.M{"device_id": dvcid, "status": types.AdbOrderStatusPending}
	orders, err := store.DB().ListAdbOrders(nil, query)
	if err != nil {
		log.Errorf("query pending adb orders for device %s error: %v", dvcid, err)
		return
	}

	for _, order := range orders {
		nid, dvcid, orderid := order.NodeID, order.DeviceID, order.ID
		_, err := DoNodeCheckAdbOrder(nid, dvcid, orderid)
		if err != nil {
			log.Warnf("query node %s adb device %s order %s error: %v", nid, dvcid, orderid, err)
			continue
		}
		// now we got the order on node adb device, then
		//  - memo update the node status as paid
		//  - publish adb event -> triger sending order callback
		MemoAdbOrderStatus(order.ID, types.AdbOrderStatusPaid)
		PublishAdbOrderCallbackEvent(order.ID)
	}
}

//
// AdbNode ReFresh Notifier Manager
//

// DelAdbNodeRefreshNotifier is exported
func DelAdbNodeRefreshNotifier(id string) {
	sched.arefreshMgr.Lock()
	if ch, ok := sched.arefreshMgr.m[id]; ok {
		delete(sched.arefreshMgr.m, id)
		close(ch)
	}
	sched.arefreshMgr.Unlock()
}

// AddAdbNodeRefreshNotifier is exported
func AddAdbNodeRefreshNotifier(id string, ch chan struct{}) {
	sched.arefreshMgr.Lock()
	sched.arefreshMgr.m[id] = ch
	sched.arefreshMgr.Unlock()
}

// RefreshAdbNodeAsync is exported
func RefreshAdbNodeAsync(id string) {
	sched.arefreshMgr.Lock()
	defer sched.arefreshMgr.Unlock()

	ch, ok := sched.arefreshMgr.m[id]
	if !ok {
		return
	}

	// send avoid block
	select {
	case ch <- struct{}{}:
	default:
	}
}
