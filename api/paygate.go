package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

// adb paygate
//  - public apis, visit by out side pay system
//  - protected by signature verify (with payGateSecret)
//
func (s *Server) payGateNewAdbOrder(ctx *httpmux.Context) {
	var (
		req     = new(types.NewAdbOrderReq)
		resp    = new(types.NewAdbOrderResp)
		dvc     *types.AdbDevice
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

	// verify signature
	err = scheduler.VerifySignature(payGateSecret, req)
	if err != nil {
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
	qrpng, _, err = scheduler.GenAdbpayQrCode(dvc.ID, req.QRType, req.Fee, orderID)
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
		resp.QRImage = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(qrpng))
	}
	resp.OrderID = orderID
	resp.OutOrderID = req.OutOrderID
	resp.Fee = req.Fee
	resp.FeeYuan = float64(req.Fee) / float64(100)
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
