package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/pkg/validator"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

//
// adb nodes
//

func (s *Server) listAdbNodes(ctx *httpmux.Context) {
	var (
		nodeID     = ctx.Query["node_id"]
		status     = ctx.Query["status"]
		remote     = ctx.Query["remote"] // remote ip search
		hostname   = ctx.Query["hostname"]
		withMaster = ctx.Query["with_master"]
		labels     = ctx.Query["labels"] // key1=val1,key2=val2,key3=val3...
		query      = bson.M{}
	)

	// build query
	if nodeID != "" {
		query["id"] = nodeID
	}
	if status != "" {
		query["status"] = status
	}
	if remote != "" {
		query["remote_addr"] = bson.M{"$regex": bson.RegEx{Pattern: remote}}
	}
	if hostname != "" {
		query["sysinfo.hostname"] = bson.M{"$regex": bson.RegEx{Pattern: hostname}}
	}
	if withMaster != "" {
		withMasterV, _ := strconv.ParseBool(withMaster)
		query["sysinfo.with_master"] = withMasterV
	}
	if labels != "" {
		pairs := strings.Split(labels, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				query[fmt.Sprintf("labels.%s", kv[0])] = kv[1]
			}
		}
	}

	// filter nodes & sort
	nodes, err := store.DB().ListNodes(getPager(ctx), query)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.AdbNode, len(nodes))
	for idx, node := range nodes {
		wraps[idx] = s.wrapAdbNode(node)
	}

	n := store.DB().CountNodes(query)
	ctx.Res.Header().Set("Total-Records", strconv.Itoa(n))
	ctx.JSON(200, wraps)
}

func (s *Server) getAdbNode(ctx *httpmux.Context) {
	var (
		id = ctx.Path["node_id"]
	)

	node, err := store.DB().GetNode(id)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, s.wrapAdbNode(node))
}

//
// adb devices
//

func (s *Server) listAdbDevices(ctx *httpmux.Context) {
	var (
		search      = ctx.Query["search"] // id or node_id
		status      = ctx.Query["status"]
		overquota   = ctx.Query["over_quota"]
		briefAll, _ = strconv.ParseBool(ctx.Query["brief_all"])
		query       = bson.M{}
	)

	if briefAll {
		s.listAllAdbDevicesBrief(ctx)
		return
	}

	// build query
	if search != "" {
		query["$or"] = []bson.M{{"id": search}, {"node_id": search}}
	}
	if status != "" {
		query["status"] = status
	}
	if overquota != "" {
		query["over_quota"], _ = strconv.ParseBool(overquota)
	}

	dvcs, err := store.DB().ListAdbDevices(getPager(ctx), query)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.AdbDeviceWrapper, len(dvcs))
	for idx, dvc := range dvcs {
		wraps[idx] = s.wrapAdbDevice(dvc)
	}

	n := store.DB().CountAdbDevices(query)
	ctx.Res.Header().Set("Total-Records", strconv.Itoa(n))
	ctx.JSON(200, wraps)
}

func (s *Server) getAdbDevice(ctx *httpmux.Context) {
	var (
		dvcid = ctx.Path["device_id"]
	)

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, s.wrapAdbDevice(dvc))
}

func (s *Server) updateAdbDevice(ctx *httpmux.Context) {
	var (
		dvcid = ctx.Path["device_id"]
	)

	var req struct {
		Desc *string `json:"desc"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.BadRequest(err)
		return
	}

	if req.Desc == nil {
		ctx.Status(204)
		return
	}

	if err := validator.String(*req.Desc, -1, 128, nil); err != nil {
		ctx.BadRequest(err)
		return
	}

	err := scheduler.MemoAdbDeviceDesc(dvcid, *req.Desc)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	current, _ := store.DB().GetAdbDevice(dvcid)
	ctx.JSON(200, current)
}

func (s *Server) setAdbDeviceBill(ctx *httpmux.Context) {
	var (
		dvcid = ctx.Path["device_id"]
		bill  = ctx.Query["val"]
	)

	val, err := strconv.Atoi(bill)
	if err != nil {
		ctx.BadRequest(err)
		return
	}

	err = validator.Int(val, 0, 10000)
	if err != nil {
		ctx.BadRequest(fmt.Errorf("max bill %v", err))
		return
	}

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	err = scheduler.MemoAdbDeviceMaxBill(dvc.ID, val)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

func (s *Server) setAdbDeviceAmount(ctx *httpmux.Context) {
	var (
		dvcid  = ctx.Path["device_id"]
		amount = ctx.Query["val"]
	)

	val, err := strconv.Atoi(amount)
	if err != nil {
		ctx.BadRequest(err)
		return
	}

	err = validator.Int(val, 0, 100000000)
	if err != nil {
		ctx.BadRequest(fmt.Errorf("max amount %v", err))
		return
	}

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	err = scheduler.MemoAdbDeviceMaxAmount(dvc.ID, val)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

func (s *Server) setAdbDeviceWeight(ctx *httpmux.Context) {
	var (
		dvcid  = ctx.Path["device_id"]
		weight = ctx.Query["val"]
	)

	val, err := strconv.Atoi(weight)
	if err != nil {
		ctx.BadRequest(err)
		return
	}

	err = validator.Int(val, 0, 100)
	if err != nil {
		ctx.BadRequest(fmt.Errorf("weight %v", err))
		return
	}

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	err = scheduler.MemoAdbDeviceWeight(dvc.ID, val)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

func (s *Server) bindAdbDeviceAlipay(ctx *httpmux.Context) {
	var (
		dvcid  = ctx.Path["device_id"]
		alipay = new(types.AlipayAccount)
	)

	// obtain new alipay account
	if err := ctx.Bind(alipay); err != nil {
		ctx.BadRequest(err)
		return
	}

	if err := alipay.Valid(); err != nil {
		ctx.BadRequest(err)
		return
	}

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if alipay := dvc.Alipay; alipay != nil {
		ctx.Conflict(fmt.Errorf("device already bind alipay account %s", alipay.Username))
		return
	}

	// TODO
	// ensure the alipay account is valid by query this alipay user ???
	//

	// ensure this alipay account not binded to other device
	devices, _ := store.DB().ListAdbDevices(nil, nil)
	for _, device := range devices {
		if device.Alipay != nil && device.Alipay.UserID == alipay.UserID {
			ctx.Conflict(fmt.Errorf("this alipay account already binded to device %s", device.ID))
			return
		}
	}

	err = scheduler.MemoAdbDeviceAlipay(dvc.ID, alipay)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

func (s *Server) revokeAdbDeviceAlipay(ctx *httpmux.Context) {
	var (
		dvcid = ctx.Path["device_id"]
	)

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if alipay := dvc.Alipay; alipay == nil {
		ctx.Status(204)
		return
	}

	// TODO
	// firstly we should ensure this adb device is disabled
	//

	// TODO
	//
	// must ensure no adbpay orders related to this device unfinished ...
	//

	err = scheduler.MemoAdbDeviceAlipay(dvc.ID, nil)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

func (s *Server) verifyAdbDevice(ctx *httpmux.Context) {
	var (
		dvcid  = ctx.Path["device_id"]
		fee, _ = strconv.Atoi(ctx.Query["fee"])
	)

	if fee <= 0 {
		ctx.BadRequest("bad parameter: fee")
		return
	}

	var (
		feeYuan = fmt.Sprintf("%0.2f", float64(fee)/float64(100))
	)

	dvc, err := store.DB().GetAdbDevice(dvcid)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if alipay := dvc.Alipay; alipay == nil {
		ctx.NotFound("device not binded any alipay account")
		return
	}

	orderID := "FOR VERIFY TEST"
	qrpng, _, err := scheduler.GenAdbpayQrCode(dvc.ID, types.QRCodeTypeAlipay, fee, orderID)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	// return the qrcode image
	ctx.Res.Header().Set("Content-Type", "image/png")
	ctx.Res.Header().Set("OrderID", orderID)
	ctx.Res.Header().Set("Fee", strconv.Itoa(fee))
	ctx.Res.Header().Set("FeeYuan", feeYuan)
	ctx.Res.WriteHeader(200)
	ctx.Res.Write(qrpng)
}

//
// adb orders
//

func (s *Server) listAdbOrders(ctx *httpmux.Context) {
	var (
		search     = ctx.Query["search"] // search id or out_order_id
		orderID    = ctx.Query["order_id"]
		outOrderID = ctx.Query["out_order_id"]
		status     = ctx.Query["status"]
		cbstatus   = ctx.Query["cbstatus"]
		device     = ctx.Query["device_id"]
		startAt    = ctx.Query["start_at"]
		endAt      = ctx.Query["end_at"]
		query      = bson.M{}
	)

	// build query
	if search != "" {
		query["$or"] = []bson.M{{"id": search}, {"out_order_id": search}}
	}
	if orderID != "" {
		query["id"] = orderID
	}
	if outOrderID != "" {
		query["out_order_id"] = outOrderID
	}
	if status != "" {
		query["status"] = status
	}
	if cbstatus != "" {
		query["callback_status"] = cbstatus
	}
	if device != "" {
		query["device_id"] = device
	}

	var (
		startTime time.Time
		endTime   time.Time
		err       error
	)
	if startAt != "" {
		startTime, err = time.Parse(time.RFC3339, startAt)
		if err != nil {
			ctx.BadRequest(err)
			return
		}
	}
	if endAt != "" {
		endTime, err = time.Parse(time.RFC3339, endAt)
		if err != nil {
			ctx.BadRequest(err)
			return
		}
	}
	if !startTime.IsZero() && !endTime.IsZero() {
		query["$and"] = []bson.M{
			{
				"created_at": bson.M{"$gt": startTime},
			},
			{
				"created_at": bson.M{"$lt": endTime},
			},
		}
	} else if !startTime.IsZero() {
		query["created_at"] = bson.M{"$gt": startTime}
	} else if !endTime.IsZero() {
		query["created_at"] = bson.M{"$lt": endTime}
	}

	// db query
	orders, err := store.DB().ListAdbOrders(getPager(ctx), query)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.AdbOrderWrapper, len(orders))
	for idx, order := range orders {
		wrap := s.wrapAdbOrder(order)
		wraps[idx] = wrap
	}

	n, fee := store.DB().CountAdbOrders(query)
	ctx.Res.Header().Set("Total-Records", strconv.Itoa(n))
	ctx.Res.Header().Set("Total-Fee-Yuan", fmt.Sprintf("%0.2f", float64(fee)/float64(100)))
	ctx.JSON(200, wraps)
}

func (s *Server) getAdbOrder(ctx *httpmux.Context) {
	var (
		id = ctx.Path["order_id"]
	)

	order, err := store.DB().GetAdbOrder(id)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, s.wrapAdbOrder(order))
}

func (s *Server) reCallbackAdbOrder(ctx *httpmux.Context) {
	var (
		id = ctx.Path["order_id"]
	)

	order, err := store.DB().GetAdbOrder(id)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if order.CallbackStatus == types.AdbOrderCallbackStatusSucceed {
		ctx.Status(204)
		return
	}

	if order.CallbackStatus == types.AdbOrderCallbackStatusOngoing {
		ctx.Conflict("another order callback ongoing")
		return
	}

	scheduler.MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusOngoing)
	err = scheduler.SendAdbOrderCallback(order.ID, nil) // send once, no retry
	if err != nil {
		ctx.AutoError(err)
		scheduler.MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusError) // see callback history
		return
	}

	scheduler.MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusSucceed)
	ctx.Status(200)
}

//
// public api docs
//

func (s *Server) getAdbPublicAPIDocs(ctx *httpmux.Context) {
	file, sha1sum, err := scheduler.GetResFile("public-api")
	if err != nil {
		ctx.AutoError(err)
		return
	}

	fd, err := os.Open(file)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer fd.Close()

	ctx.Res.Header().Set("Content-Type", "application/octet-stream")
	ctx.Res.Header().Set("Content-Disposition", "attachment; filename="+path.Base(file))
	if sha1sum != "" {
		ctx.Res.Header().Set("Sha1sum", sha1sum)
	}
	io.Copy(ctx.Res, fd)
}

//
// adb paygate
//  - public apis (visit by out side pay system)
//  - protected by https (TODO)
//  - protected by ip range check (TODO)
//  - protected by fixed access token: ADB-PAYGATE-SECRET
//
func (s *Server) newAdbOrder(ctx *httpmux.Context) {
	var (
		req     = new(types.NewAdbOrderReq)
		resp    = new(types.NewAdbOrderResp)
		dvc     *types.AdbDevice
		qrtext  string
		qrpng   []byte
		orderID string
		query   bson.M
		err     error
	)

	if err = ctx.Bind(req); err != nil {
		goto END
	}
	if err = req.Valid(); err != nil {
		goto END
	}

	// TODO
	// check the source ip allowed
	// get source ip, if it is allowed
	// if !sourceAllow() {
	// err = errors.New("source deny")
	// goto END
	// }

	// check the fixed paygate token
	if ctx.Req.Header.Get("ADB-PAYGATE-SECRET") != payGateSecret {
		err = errors.New("adbpay secret deny")
		goto END
	}

	// ensure we have corresponding adb device avaliable through
	// once smart pickup
	dvc, err = scheduler.SmartPickupAdbDevice(req)
	if err != nil {
		err = fmt.Errorf("pick up adb device error: %s", err.Error())
		goto END
	}

	// check duplication on out order id
	query = bson.M{"out_order_id": req.OutOrderID}
	if orders, _ := store.DB().ListAdbOrders(nil, query); len(orders) > 0 {
		err = errors.New("duplicated out order id")
		goto END
	}

	// save db adb order
	orderID = s.newOrderID()
	if err = store.DB().AddAdbOrder(&types.AdbOrder{
		ID:              orderID,
		Status:          types.AdbOrderStatusPending, // init status: pending
		NodeID:          dvc.NodeID,
		DeviceID:        dvc.ID,
		NewAdbOrderReq:  *req, // never be nil
		Response:        nil,
		Callback:        nil,
		CallbackHistory: []string{},
		CallbackStatus:  types.AdbOrderCallbackStatusNone, // init status: none
		CreatedAt:       time.Now(),
		PaidAt:          time.Time{},
	}); err != nil {
		goto END
	}

	// ask generate qrcode and return response
	qrpng, qrtext, err = scheduler.GenAdbpayQrCode(dvc.ID, req.QRType, req.Fee, orderID)
	if err != nil {
		// note: after db order created, if we met error while generating qrcode,
		// we should remove the newly db order and tell outside to retry.
		err = fmt.Errorf("generate adbpay qrcode error: [%v], pls try again later", err)
		store.DB().RemoveAdbOrder(orderID) // note: remove the newly db adb order
		goto END
	}

END:
	// fill the response
	if err != nil {
		resp.Code = 0
		resp.Message = err.Error()
	} else {
		resp.Code = 1
		resp.QRText = qrtext
		resp.QRImage = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(qrpng))
	}
	resp.Time = time.Now() // add time at to track order timeline

	// if db order created, save adb order response
	if orderID != "" {
		scheduler.MemoAdbOrderResponse(orderID, resp)

		// subscribe wait the adb order's callback and return our callback to merchant
		go scheduler.SubscribeAdbOrderAndSendCallback(orderID)
	}

	// always 200
	ctx.JSON(200, resp)
}

// adb events
//
func (s *Server) receiveAdbEvents(ctx *httpmux.Context) {
	var (
		ev = new(adbot.AdbEvent)
	)

	if err := ctx.Bind(ev); err != nil {
		ctx.BadRequest(err)
		return
	}

	if err := ev.Valid(); err != nil {
		ctx.BadRequest(err)
		return
	}

	scheduler.PublishAdbDeviceEvent(ev)

	ctx.Status(200)
}

func (s *Server) watchAdbEvents(ctx *httpmux.Context) {
	notifier, ok := ctx.Res.(http.CloseNotifier)
	if !ok {
		ctx.InternalServerError("not a http close notifier")
		return
	}

	flusher, ok := ctx.Res.(http.Flusher)
	if !ok {
		ctx.InternalServerError("not a http flusher")
		return
	}

	// obtain adb device subscriber
	sub := scheduler.SubscribeAdbDeviceEvents()

	// must: evict the subscriber befor page exit
	go func() {
		<-notifier.CloseNotify()
		scheduler.EvictAdbDeviceEvents(sub)
	}()

	// write response header firstly
	ctx.Res.WriteHeader(200)
	ctx.Res.Header().Set("Content-Type", "text/event-stream")
	ctx.Res.Header().Set("Cache-Control", "no-cache")
	ctx.Res.Write(nil)
	flusher.Flush()

	// write adb device event to the client with sse format
	for ev := range sub {
		ctx.Res.Write(ev.(*adbot.AdbEvent).Format())
		flusher.Flush()
	}
}

// utils
//
func (s *Server) wrapAdbNode(node *types.Node) *types.AdbNode {
	if !unmaskSensitive {
		node.Hidden()
	}
	devices, _ := store.DB().ListAdbDevices(nil, bson.M{"node_id": node.ID})
	onlineN, offlineN := s.countAdbDevicesStatus(devices)
	return &types.AdbNode{
		Node:       s.wrapNode(node),
		NumDevices: int64(len(devices)),
		NumOnline:  onlineN,
		NumOffline: offlineN,
	}
}

func (s *Server) countAdbDevicesStatus(dvcs []*types.AdbDevice) (int64, int64) {
	var n1, n2 int64
	for _, dvc := range dvcs {
		switch dvc.Status {
		case types.AdbDeviceStatusOnline:
			n1++
		case types.AdbDeviceStatusOffline:
			n2++
		}
	}
	return n1, n2
}

func (s *Server) wrapAdbDevice(dvc *types.AdbDevice) *types.AdbDeviceWrapper {
	num, fee := scheduler.CountAdbDeviceThisDayOrders(dvc.ID)
	return &types.AdbDeviceWrapper{
		AdbDevice:       dvc,
		TodayBill:       num,
		TodayAmount:     fee,
		TodayAmountYuan: float64(fee) / float64(100),
		MaxAmountYuan:   float64(dvc.MaxAmount) / float64(100),
	}
}

func (s *Server) wrapAdbOrder(o *types.AdbOrder) *types.AdbOrderWrapper {
	return &types.AdbOrderWrapper{
		AdbOrder: o,
		FeeYuan:  float64(o.Fee) / float64(100),
	}
}

func (s *Server) newOrderID() string {
	now := time.Now()
	suffix := utils.RandomString(6)
	return fmt.Sprintf("%d%d%d%d%d%d-%s", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), suffix)
}

func (s *Server) listAllAdbDevicesBrief(ctx *httpmux.Context) {
	dvcs, err := store.DB().ListAdbDevices(nil, nil)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ret := make([]string, 0, 0)
	for _, dvc := range dvcs {
		ret = append(ret, dvc.ID)
	}
	ctx.JSON(200, ret)
}
