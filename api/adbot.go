package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

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
	nodes, err := scheduler.FilterNodes(nil, nil, false)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.AdbNode, len(nodes))
	for idx, node := range nodes {
		wraps[idx] = s.wrapAdbNode(node)
	}

	ctx.Res.Header().Set("Total-Records", strconv.Itoa(len(wraps)))
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
		search    = ctx.Query["search"] // id or node_id
		status    = ctx.Query["status"]
		overquota = ctx.Query["over_quota"]
		query     = bson.M{}
	)

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

//
// adb orders
//

func (s *Server) listAdbOrders(ctx *httpmux.Context) {
	var (
		status  = ctx.Query["status"]
		nid     = ctx.Query["node_id"]
		device  = ctx.Query["device_id"]
		startAt = ctx.Query["start_at"]
		endAt   = ctx.Query["end_at"]
		query   = bson.M{}
	)

	// build query
	if status != "" {
		query["status"] = status
	}
	if nid != "" {
		query["node_id"] = nid
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

//
// adb payhook
//  - public apis (visit by each adb nodes)
//  - protected by https (TODO)
//  - protected by fixed access token: ADB-HOOK-Secret
//
var (
	AdbPayHookHeaderKey    = "ADB-HOOK-Secret"
	AdbPayHookAccessSecret = "98093efz97ca47db7a2em402a0608696eebe3a4z"
)

func (s *Server) hookAdbOrder(ctx *httpmux.Context) {
	// check the fixed payhook token
	if ctx.Req.Header.Get(AdbPayHookHeaderKey) != AdbPayHookAccessSecret {
		ctx.Forbidden("adbhook secret deny")
		return
	}

	// get the request's order notify event message
	var (
		params  = ctx.Query
		orderID = ctx.Query["order_id"] // order_id
		code    = ctx.Query["code"]
	)

	// get the db adbpay order firstly
	order, err := store.DB().GetAdbOrder(orderID)
	if err != nil {
		ctx.AutoError(fmt.Sprintf("get db adb order %s error: %v", orderID, err))
		return
	}

	// memo update the order's inner callback
	// convert query params to expect data type
	cb := map[string]interface{}{}
	for key, val := range params {
		cb[key] = val
	}
	scheduler.MemoAdbOrderInnerCallback(orderID, &types.InnerPayCallback{Callback: cb, WaitError: "", Time: time.Now()})

	// memo update the order's status
	if code != "1" {
		ctx.BadRequest("hook adb order code invalid, only `1` acceptable")
		return
	}
	err = scheduler.MemoAdbOrderStatus(orderID, types.AdbOrderStatusPaid)
	if err != nil {
		ctx.AutoError(fmt.Sprintf("update db adb order status error: %v", err))
		return
	}

	// publish order callback event
	scheduler.PublishAdbOrderCallbackEvent(order.ID)

	ctx.Status(200)
}

//
// adb paygate
//  - public apis (visit by out side pay system)
//  - protected by https (TODO)
//  - protected by ip range check (TODO)
//  - protected by fixed access token: ADB-PAYGATE-Secret
//
var (
	AdbPayGateHeaderKey    = "ADB-PAYGATE-Secret"
	AdbPayGateAccessSecret = "4a99amf2c272fb99911e1002bd7d9c387acd705x"
)

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
	if ctx.Req.Header.Get(AdbPayGateHeaderKey) != AdbPayGateAccessSecret {
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
		InnerCallback:   nil,
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
		err = fmt.Errorf("generate adbpay qrcode error: %v, pls try again later")
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

func (s *Server) checkAdbOrder(ctx *httpmux.Context) {
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
		AdbDevice:   dvc,
		TodayBill:   num,
		TodayAmount: fee,
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
