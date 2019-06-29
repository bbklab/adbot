package types

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/validator"
)

// nolint
var (
	QRCodeTypeAlipay = "alipay"
	QRCodeTypeWxpay  = "wxpay"
)

// nolint
var (
	AdbDeviceStatusOnline  = "online"
	AdbDeviceStatusOffline = "offline"
)

// AdbNode is a wrapper of db node with ralated adb devices
type AdbNode struct {
	Node       *NodeWrapper `json:"node"`        // db node
	NumDevices int64        `json:"num_devices"` // nb of total devices
	NumOnline  int64        `json:"num_online"`  // nb of online devices
	NumOffline int64        `json:"num_offline"` // nb of offline devices
}

// AdbDeviceWrapper is exported
type AdbDeviceWrapper struct {
	*AdbDevice
	RecentAdbOrders RecentAdbOrders `json:"recent_adb_orders"`
	MaxAmountYuan   float64         `json:"max_amount_yuan"`
	TodayPaidRate   float64         `json:"today_paid_rate"` // today paid order rate
}

// AdbDevice is a db adb device
type AdbDevice struct {
	ID        string                `json:"id" bson:"id"`                 // note: device id
	NodeID    string                `json:"node_id" bson:"node_id"`       // related node id
	SysInfo   *adbot.AndroidSysInfo `json:"sysinfo" bson:"sysinfo"`       // device info, update by refresher
	Desc      string                `json:"desc" bson:"desc"`             // description text
	Status    string                `json:"status" bson:"status"`         // status: online,offline updated by refresher
	Errmsg    string                `json:"error" bson:"error"`           // error message, updated by refresher
	MaxAmount int                   `json:"max_amount" bson:"max_amount"` // max amount per day, 0 means unlimit
	MaxBill   int                   `json:"max_bill" bson:"max_bill"`     // max bill per day, by CNY, 0 means unlimit
	OverQuota bool                  `json:"over_quota" bson:"over_quota"` // over quota (MaxAmount/MaxBill) flag, updated by timer calculator
	Weight    int                   `json:"weight" bson:"weight"`         // weight value, must between [0-100], the higher value means the higher weight
	Alipay    *AlipayAccount        `json:"alipay" bson:"alipay"`         // binded alipay account
	Wxpay     *WxpayAccount         `json:"wxpay" bson:"wxpay"`           // binded wxpay account
}

// Name is exported
func (d *AdbDevice) Name() string {
	if d.SysInfo == nil {
		return d.ID
	}
	return d.SysInfo.DeviceName
}

// WeightN implement balancer.Item
func (d *AdbDevice) WeightN() int {
	return d.Weight
}

// AlipayAccount is exported
type AlipayAccount struct {
	UserID   string `json:"user_id" bson:"user_id"`   // must, alipay scan: https://render.alipay.com/p/f/fd-ixpo7iia/index.html
	Username string `json:"username" bson:"username"` // must,
	Nickname string `json:"nickname" bson:"nickname"` // optional
}

// Valid is exported
func (a *AlipayAccount) Valid() error {
	if a.UserID == "" {
		return errors.New("alipay userid required")
	}
	if a.Username == "" {
		return errors.New("alipay username required")
	}
	return nil
}

// WxpayAccount is exported
type WxpayAccount struct {
}

// nolint
var (
	AdbOrderTimeout = time.Minute * 5
)

// nolint
var (
	AdbOrderStatusPending = "pending" // init status
	AdbOrderStatusPaid    = "paid"    // paid
	AdbOrderStatusTimeout = "timeout" // timeout
)

// nolint
var (
	AdbOrderCallbackStatusNone    = "none"    // init status, not triggered yet
	AdbOrderCallbackStatusOngoing = "ongoing" // ongoing, triggered by adb device callback, sending callback to outerside
	AdbOrderCallbackStatusSucceed = "succeed" // succeed, final state
	AdbOrderCallbackStatusError   = "error"   // error, final state
	AdbOrderCallbackStatusAborted = "aborted" // reset `ongoing` callback on startup initilization
)

// AdbOrderWrapper is exported
type AdbOrderWrapper struct {
	*AdbOrder
	DeviceName string  `json:"device_name"`
	FeeYuan    float64 `json:"fee_yuan"`
}

// AdbOrder is a db adb order
type AdbOrder struct {
	ID              string                          `json:"id" bson:"id"`               // order id, uniq
	Status          string                          `json:"status" bson:"status"`       // pending, paid, timeout
	NodeID          string                          `json:"node_id" bson:"node_id"`     // ref: adb device node id
	DeviceID        string                          `json:"device_id" bson:"device_id"` // ref: adb device id
	NewAdbOrderReq  `json:",inline" bson:",inline"` // step1: order request <- from merchant
	Response        *NewAdbOrderResp                `json:"response" bson:"response"`                 // step2: order response -> to out side
	Callback        *NewAdbOrderCallback            `json:"callback" bson:"callback"`                 // step4: order callback -> to out side
	CallbackStatus  string                          `json:"callback_status" bson:"callback_status"`   // callback status: none, ongoing, succeed, error
	CallbackHistory []string                        `json:"callback_history" bson:"callback_history"` // callback history with all failure retries
	CreatedAt       time.Time                       `json:"created_at" bson:"created_at"`
	PaidAt          time.Time                       `json:"paid_at" bson:"paid_at"`
}

// NewAdbOrderReq is a new adb order request
type NewAdbOrderReq struct {
	OutOrderID string `json:"out_order_id" bson:"out_order_id"` // must: out side order id, [1-64], [a-zA-Z0-9.-_]
	QRType     string `json:"qrtype" bson:"qrtype"`             // must: qrcode type [alipay,wxpay]
	Fee        int    `json:"fee" bson:"fee"`                   // must: order fee [1,10000000000]
	Attach     string `json:"attach" bson:"attach"`             // optional: out side custom data, [0-128]
	NotifyURL  string `json:"notify_url" bson:"notify_url"`     // optional: call back url, [0-128]
	Sign       string `json:"sign" bson:"sign"`                 // must: signature, [1-64]
}

// StringToSign return **uniq** string to be signed
func (r *NewAdbOrderReq) StringToSign() string {
	return fmt.Sprintf("out_order_id=%s&fee=%d", r.OutOrderID, r.Fee)
}

// Valid is exported
func (r *NewAdbOrderReq) Valid() error {
	if err := validator.String(r.OutOrderID, 1, 64, validator.NormalCharacters); err != nil {
		return fmt.Errorf("out order id %v", err)
	}

	switch r.QRType {
	case QRCodeTypeAlipay, QRCodeTypeWxpay:
	default:
		return fmt.Errorf("qrcode type unrecoginized")
	}

	if err := validator.Int(r.Fee, 1, 1000000000); err != nil {
		return fmt.Errorf("fee %v", err)
	}

	if err := validator.String(r.Attach, -1, 128, nil); err != nil {
		return fmt.Errorf("attach %v", err)
	}

	if err := validator.String(r.NotifyURL, -1, 128, nil); err != nil {
		return fmt.Errorf("notify url %v", err)
	}
	if r.NotifyURL != "" {
		u, err := url.Parse(r.NotifyURL)
		if err != nil {
			return fmt.Errorf("notify url %v", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.New("notify url only support scheme: http|https")
		}
	}

	if err := validator.String(r.Sign, 1, 64, nil); err != nil {
		return fmt.Errorf("signature %v", err)
	}

	return nil
}

// NewAdbOrderResp is a new adb order response
type NewAdbOrderResp struct {
	Code       int       `json:"code" bson:"code"`                 // 1:success  0:error
	Message    string    `json:"message" bson:"message"`           // error message while Code==1
	QRImage    string    `json:"qrimage" bson:"-"`                 // qrcode png base64: data:image/png;base64,{xxxx}  (note: ignore mgo db save)
	OrderID    string    `json:"order_id" bson:"order_id"`         // adbot order id
	OutOrderID string    `json:"out_order_id" bson:"out_order_id"` // copy from Req.OutOrderID
	Fee        int       `json:"fee" bson:"fee"`                   // copy from Req.Fee
	FeeYuan    float64   `json:"fee_yuan" bson:"fee_yuan"`         // copy from Req, Req.Fee/100
	Time       time.Time `json:"time" bson:"time"`                 // set by us, only used for tracking order steps time line
}

// NewAdbOrderCallback is exported
type NewAdbOrderCallback struct {
	Code       int       `json:"code" bson:"code"`                 // always = 1:success
	OutOrderID string    `json:"out_order_id" bson:"out_order_id"` // out side order id
	Fee        int       `json:"fee" bson:"fee"`                   // order fee
	Attach     string    `json:"attach" bson:"attach"`             // out side custom data, return unchanged
	Sign       string    `json:"sign" bson:"sign"`                 // set by us, used for receiver verify this callback to prevent any fake callbacks
	Time       time.Time `json:"time" bson:"time"`                 // set by us, only used for tracking order steps time line
}

// AdbDeviceCmd is exported
type AdbDeviceCmd struct {
	Command string `json:"command"`
}
