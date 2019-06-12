package adbot

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	goadb "github.com/zach-klippenstein/goadb"
)

//
//  AdbHandler Implemention
//

// NewAdb new a AdbHandler
func NewAdb() (AdbHandler, error) {
	adb, err := goadb.New()
	if err != nil {
		return nil, err
	}
	return &Adb{h: adb}, nil
}

// Adb is an AdbHandler implemention
type Adb struct {
	h *goadb.Adb
}

// ListAdbDevices implement AdbHandler
func (a *Adb) ListAdbDevices() ([]string, error) {
	return a.h.ListDeviceSerials()
}

// NewDevice implement AdbHandler
func (a *Adb) NewDevice(serial string) AdbDeviceHandler {
	return &AdbDevice{
		h: a.h.Device(goadb.DeviceWithSerial(serial)),
	}
}

//
//  AdbDeviceHandler Implemention
//

// AdbDevice is an AdbDeviceHandler implemention
type AdbDevice struct {
	h        *goadb.Device
	cmdmutex sync.Mutex // synchronized Run(cmd, args)

	sysNotifyClearButtonXY          [2]int // 下拉 -> 清除所有通知按钮 XY
	alipayChargingButtonXY          [2]int // 首页 -> 收钱 XY
	alipayChargingAmountXY          [2]int // 首页 -> 收钱 -> 设置金额 XY
	alipayChargingQrCodeAddBeiZhuXY [2]int // 首页 -> 收钱 -> 设置金额 -> 添加收款理由 XY
	alipayChargingQrCodeNextBtnXY   [2]int // 首页 -> 收钱 -> 设置金额 -> 确定 XY
	alipayOrderButtonXY             [2]int // 我的 -> 账单 XY
	alipayOrderSearchButtonXY       [2]int // 我的 -> 账单 -> 搜索 XY
	alipayOrderSearchEmitXY         [2]int // 我的 -> 账单 -> 搜索 -> 搜索 XY
}

// Serial implement AdbDeviceHandler
func (dvc *AdbDevice) Serial() (string, error) {
	return dvc.h.Serial()
}

// Online implement AdbDeviceHandler
func (dvc *AdbDevice) Online() bool {
	state, err := dvc.h.State()
	if err != nil {
		log.Warnln("Online() error:", err)
		return false
	}
	return state == goadb.StateOnline
}

// SysInfo implement AdbDeviceHandler
func (dvc *AdbDevice) SysInfo() (*AndroidSysInfo, error) {
	bs, err := dvc.Run("getprop")
	if err != nil {
		log.Warnln("SysInfo() error:", err)
		return nil, err
	}
	sysinfo := parseAndroidSysinfo(string(bs))

	battery, err := dvc.BatteryInfo()
	if err != nil {
		return nil, err
	}

	sysinfo.Battery = battery
	return sysinfo, nil
}

// BatteryInfo implement AdbDeviceHandler
func (dvc *AdbDevice) BatteryInfo() (*AndroidBatteryInfo, error) {
	bs, err := dvc.Run("dumpsys", "battery")
	if err != nil {
		log.Warnln("BatteryInfo() error:", err)
		return nil, err
	}
	return parseAndroidBatteryInfo(string(bs)), nil
}

// Run implement AdbDeviceHandler
func (dvc *AdbDevice) Run(cmd string, args ...string) (string, error) {
	dvc.cmdmutex.Lock()
	defer dvc.cmdmutex.Unlock()
	log.Debugln("Run()", cmd, args)
	return dvc.h.RunCommand(cmd, args...)
}

// IsAwake implement AdbDeviceHandler
func (dvc *AdbDevice) IsAwake() bool {
	bs, err := dvc.Run("dumpsys", "window", "policy")
	if err != nil {
		log.Warnln("IsAwake() error:", err)
	}
	return strings.Contains(string(bs), "mScreenOnEarly=true")
}

// AwakenScreen implement AdbDeviceHandler
func (dvc *AdbDevice) AwakenScreen() error {
	dvc.Run("input", "keyevent", "224")
	if dvc.IsAwake() {
		return nil
	}
	for i := 1; i <= 3; i++ {
		if dvc.IsAwake() {
			return nil
		}
		bs, err := dvc.Run("input", "keyevent", "26")
		if err != nil {
			log.Warnln("AwakenScreen() error:", err, "-", string(bs))
		}
	}

	return errors.New("failed to awaken screen")
}

// ScreenCap implement AdbDeviceHandler
func (dvc *AdbDevice) ScreenCap() ([]byte, error) {
	tmpfile := "/sdcard/.adb.screen.temp.png"
	out, err := dvc.Run("screencap", "-p", tmpfile)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, out)
	}

	reader, err := dvc.h.OpenRead(tmpfile)
	if err != nil {
		log.Errorln("ScreenCap().read.screencap.file error:", tmpfile, err)
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

// GotoHome implement AdbDeviceHandler
func (dvc *AdbDevice) GotoHome() error {
	_, err := dvc.Run("input", "keyevent", "3")
	return err
}

// GoBack implement AdbDeviceHandler
func (dvc *AdbDevice) GoBack() error {
	_, err := dvc.Run("input", "keyevent", "4")
	return err
}

// Click implement AdbDeviceHandler
func (dvc *AdbDevice) Click(x, y int) error {
	_, err := dvc.Run("input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	return err
}

// Swipe implement AdbDeviceHandler
func (dvc *AdbDevice) Swipe(x1, y1, x2, y2 int) error {
	_, err := dvc.Run("input", "swipe", strconv.Itoa(x1), strconv.Itoa(y1), strconv.Itoa(x2), strconv.Itoa(y2))
	return err
}

// SwipeUpUnlock implement AdbDeviceHandler
func (dvc *AdbDevice) SwipeUpUnlock() error { // wipe up to unlock
	return dvc.Swipe(300, 1200, 300, 0) // FIXME bottom hardcoded = 1200
}

// SwipeDownShowNotify implement AdbDeviceHandler
func (dvc *AdbDevice) SwipeDownShowNotify() error { // pull down from top to show the notifies
	return dvc.Swipe(300, 0, 300, 1200)
}

// CurrentTopActivity implement AdbDeviceHandler
func (dvc *AdbDevice) CurrentTopActivity() (string, error) {
	out, err := dvc.Run("dumpsys", "activity", "top")
	if err != nil {
		return "", err
	}

	var (
		buf     = bytes.NewBufferString(out)
		scanner = bufio.NewScanner(buf)
	)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 2 && fields[0] == "ACTIVITY" && fields[1] != "" {
			return fields[1], nil
		}
	}
	return "", errors.New("current top activity not found")
}

// DumpCurrentUI implement AdbDeviceHandler
func (dvc *AdbDevice) DumpCurrentUI() ([]*AndroidUINode, error) {
	bs, err := dvc.Run("uiautomator", "dump")
	if err != nil {
		log.Errorln("DumpCurrentUI().ui.dump error:", err)
		return nil, err
	}

	var xmlfile string
	fields := strings.SplitN(bs, "dumped to:", 2)
	if len(fields) == 2 {
		xmlfile = strings.TrimSpace(fields[1])
	}

	reader, err := dvc.h.OpenRead(xmlfile)
	if err != nil {
		log.Errorln("DumpCurrentUI().read.xml.file error:", xmlfile, err)
		return nil, err
	}
	defer reader.Close()

	xmlbs, _ := ioutil.ReadAll(reader)
	return parseAndroidUINodes(xmlbs)
}

// FindUINode implement AdbDeviceHandler
func (dvc *AdbDevice) FindUINode(nodes []*AndroidUINode, resourceid, resourcetext string) *AndroidUINode {
	for _, node := range nodes {
		if node.ResourceID == resourceid {
			if resourcetext == "" {
				return node
			}
			if node.Text == resourcetext {
				return node
			}
		}
	}
	return nil
}

var (
	errUINodeNotFound = errors.New("can't find the target UI node")
)

// FindUINodeAndTapMiddleXY implement AdbDeviceHandler
func (dvc *AdbDevice) FindUINodeAndTapMiddleXY(resourceid, resourcetext string) (int, int, error) {
	nodes, err := dvc.DumpCurrentUI()
	if err != nil {
		log.Errorln("FindUINodeAndTapMiddleXY().DumpCurrentUI() error:", err)
		return -1, -1, err
	}

	var targetNode = dvc.FindUINode(nodes, resourceid, resourcetext)
	if targetNode == nil {
		return -1, -1, errUINodeNotFound
	}

	x, y, err := targetNode.MiddleXY()
	if err != nil {
		return -1, -1, fmt.Errorf("FindUINodeAndTapMiddleXY() can't find the target UI node MiddleXY(): %v", err)
	}

	return x, y, dvc.Click(x, y)
}

// TailSysLogs implement AdbDeviceHandler
func (dvc *AdbDevice) TailSysLogs() (<-chan string, chan struct{}) {
	var (
		ch     = make(chan string, 10240)
		stopch = make(chan struct{})
	)

	go func() {
		log.Println("tail follow system logs loop started")
		defer log.Println("tail follow system logs loop stopped")

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		sendFunc := func(e string) {
			select {
			case ch <- e:
			default:
			}
		}

		for {
			select {

			case <-ticker.C:
				out, err := dvc.Run("logcat", "-d") // dump the buffered logs
				if err != nil {
					log.Warnln("TailSysLogs().logcat.dump error:", err)
					continue
				}

				scanner := bufio.NewScanner(bytes.NewBufferString(out))
				for scanner.Scan() {
					sendFunc(scanner.Text())
				}

				dvc.Run("logcat", "-c") // clear the buffer logs

			case <-stopch:
				close(ch)
				return
			}
		}
	}()

	return ch, stopch
}

// WatchSysEvents implement AdbDeviceHandler
func (dvc *AdbDevice) WatchSysEvents(keywords []string) (<-chan string, chan struct{}) {
	var (
		ch     = make(chan string, 10240)
		stopch = make(chan struct{})
	)

	go func() {
		log.Println("watch system events loop started")
		defer log.Println("watch system events loop stopped")

		sendFunc := func(e string) {
			select {
			case ch <- e:
			default:
			}
		}

		logch, logstopch := dvc.TailSysLogs()

		for {
			select {
			case log := <-logch:
				for _, keyword := range keywords {
					if strings.Contains(log, keyword) {
						sendFunc(log)
					}
				}
			case <-stopch:
				close(logstopch)
				close(ch)
				return
			}
		}
	}()

	return ch, stopch
}

// ListSysNotifies implement AdbDeviceHandler
func (dvc *AdbDevice) ListSysNotifies() []*AndroidSysNotify {
	bs, err := dvc.Run("dumpsys", "notification")
	if err != nil {
		log.Warnln("ListSysNotifies() error:", err)
	}

	var (
		rxTitle = regexp.MustCompile(`^[ \t]*NotificationRecord`)
		keys    = []string{"tickerText"}
	)
	ret, err := parseSectionText(string(bs), rxTitle, keys)
	if err != nil {
		log.Warnln("ListSysNotifies() error:", err)
	}

	notifies := []*AndroidSysNotify{}
	for _, lbs := range ret {
		title := lbs.Get("_TITLE_")
		text := lbs.Get("tickerText")
		notifies = append(notifies, &AndroidSysNotify{
			Source:  parseSysNotifyFromPKG(title),
			Message: text,
		})
	}
	return notifies
}

// ClearSysNotifies implement AdbDeviceHandler
func (dvc *AdbDevice) ClearSysNotifies() error {
	// swipe down list notifies
	dvc.SwipeDownShowNotify()

	// if we have already knows the clear button location, directly click it
	if x, y := dvc.sysNotifyClearButtonXY[0], dvc.sysNotifyClearButtonXY[1]; x > 0 && y > 0 {
		return dvc.Click(x, y)
	}

	// find the UI node and click it
	var (
		resourceid   = "com.android.systemui:id/clear_all_button"
		resourcetext = ""
	)
	x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
	if err != nil {
		if err == errUINodeNotFound {
			log.Warnln("ClearSysNotifies() no system notifies need to be cleaned")
			return nil
		}
		log.Errorln("ClearSysNotifies().FindUINodeAndTapMiddleXY() error:", err)
		return err
	}

	// save the clear button XY
	dvc.sysNotifyClearButtonXY = [2]int{x, y}
	return nil
}

// WatchSysNotifies implement AdbDeviceHandler
func (dvc *AdbDevice) WatchSysNotifies() (<-chan *AndroidSysNotify, chan struct{}) {
	var (
		ch     = make(chan *AndroidSysNotify, 10240)
		stopch = make(chan struct{})
	)

	go func() {
		log.Println("watch system notifies loop started")
		defer log.Println("watch system notifies loop stopped")

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		sendFunc := func(e *AndroidSysNotify) {
			select {
			case ch <- e:
			default:
			}
		}

		var prevNotifies = []*AndroidSysNotify{}

		preExistsFunc := func(n *AndroidSysNotify) bool {
			for _, notify := range prevNotifies {
				if n.EqualsTo(notify) {
					return true
				}
			}
			return false
		}

		for {
			select {
			case <-ticker.C:
				// dump curent notifies
				currentNotifies := dvc.ListSysNotifies()

				// compare to get the newly notifies
				var n int
				for _, notify := range currentNotifies {
					if !preExistsFunc(notify) {
						sendFunc(notify)
						n++
					}
				}
				// FIXME clean while idle
				// if n > 0 {
				// dvc.ClearSysNotifies() // clear sys notifies
				// }

				// update previous notifies
				prevNotifies = currentNotifies
			case <-stopch:
				close(ch)
				return
			}
		}
	}()

	return ch, stopch
}

//
// Alipay App
//

// StartAliPay implement AdbDeviceHandler
func (dvc *AdbDevice) StartAliPay() error {
	_, err := dvc.Run("am", "start", "com.eg.android.AlipayGphone/.AlipayLogin")
	if err != nil {
		log.Errorln("StartAliPay() am.start error:", err)
		return err
	}
	return nil
}

// GotoAlipayTabHome implement AdbDeviceHandler
// note: depends on StartAliPay()
func (dvc *AdbDevice) GotoAlipayTabHome() error {
	var (
		resourceid   = "com.alipay.android.phone.openplatform:id/tab_description"
		resourcetext = "首页"
		retryN       int
	)

RETRY:
	retryN++

	_, _, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
	if err != nil {
		if err == errUINodeNotFound {
			if retryN >= 7 {
				log.Errorln("GotoAlipayTabHome() can't find the alipay wealth.home button")
				return errors.New("GotoAlipayTabHome() can't find the alipay wealth.home button")
			}
			log.Warnln("GotoAlipayTabHome() goback one step and retry to find the alipay wealth.home button ...")
			dvc.GoBack()
			goto RETRY
		} else {
			log.Errorln("GotoAlipayTabHome().FindUINodeAndTapMiddleXY error:", err)
			return fmt.Errorf("GotoAlipayTabHome().FindUINodeAndTapMiddleXY error: %v", err)
		}
	}

	// TODO save the X,Y
	return nil
}

// GotoAlipayCharging implement AdbDeviceHandler
// note: depends on GotoAlipayTabHome()
func (dvc *AdbDevice) GotoAlipayCharging() error {
	// we knows it, directly click it
	if x, y := dvc.alipayChargingButtonXY[0], dvc.alipayChargingButtonXY[1]; x > 0 && y > 0 {
		return dvc.Click(x, y)
	}

	// find the UI node and click it
	var (
		resourceid   = "com.alipay.android.phone.openplatform:id/collect_tv"
		resourcetext = "收钱"
	)
	x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
	if err != nil {
		log.Errorln("GotoAlipayCharging().FindUINodeAndTapMiddleXY() error:", err)
		return err
	}

	// save the order button XY
	dvc.alipayChargingButtonXY = [2]int{x, y}
	return nil
}

// GotoAlipayChargingAmount implement AdbDeviceHandler
// note: depends on GotoAlipayCharging()
func (dvc *AdbDevice) GotoAlipayChargingAmount() error {
	// we knows it, directly click it
	if x, y := dvc.alipayChargingAmountXY[0], dvc.alipayChargingAmountXY[1]; x > 0 && y > 0 {
		return dvc.Click(x, y)
	}

	// find the UI node and click it
	var (
		resourceid   = "com.alipay.mobile.payee:id/payee_QRCodePayModifyMoney"
		resourcetext = "设置金额"
	)
	x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
	if err != nil {
		log.Errorln("GotoAlipayChargingAmount().FindUINodeAndTapMiddleXY() error:", err)
		return err
	}

	// save the order button XY
	dvc.alipayChargingAmountXY = [2]int{x, y}
	return nil
}

// AlipayGenerateChargingAmountQrCode implement AdbDeviceHandler
func (dvc *AdbDevice) AlipayGenerateChargingAmountQrCode(orderID string, fee int) (*AlipayChargingQrCode, error) {
	// input the fee
	_, err := dvc.Run("input", "text", fmt.Sprintf("%0.2f", float64(fee)/float64(100)))
	if err != nil {
		return nil, fmt.Errorf("AlipayGenerateChargingAmountQrCode() input text fee amount: %v", err)
	}

	// click 添加收款理由
	if x, y := dvc.alipayChargingQrCodeAddBeiZhuXY[0], dvc.alipayChargingQrCodeAddBeiZhuXY[1]; x > 0 && y > 0 {
		err = dvc.Click(x, y)
		if err != nil {
			log.Errorln("AlipayGenerateChargingAmountQrCode() directly click1 error:", err)
			return nil, err
		}
	} else {
		// find the UI node and click it
		resourceid := "com.alipay.mobile.payee:id/payee_QRAddBeiZhuLink"
		resourcetext := "添加收款理由"
		x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
		if err != nil {
			log.Errorln("AlipayGenerateChargingAmountQrCode().FindUINodeAndTapMiddleXY().1 error:", err)
			return nil, err
		}
		// save the button XY
		dvc.alipayChargingQrCodeAddBeiZhuXY = [2]int{x, y}
	}

	// input the order id
	_, err = dvc.Run("input", "text", orderID)
	if err != nil {
		return nil, fmt.Errorf("AlipayGenerateChargingAmountQrCode() input text orderID: %v", err)
	}

	// click 确定
	if x, y := dvc.alipayChargingQrCodeNextBtnXY[0], dvc.alipayChargingQrCodeNextBtnXY[1]; x > 0 && y > 0 {
		err = dvc.Click(x, y)
		if err != nil {
			log.Errorln("AlipayGenerateChargingAmountQrCode() directly click2 error:", err)
			return nil, err
		}
	} else {
		resourceid := "com.alipay.mobile.payee:id/payee_NextBtn"
		resourcetext := "确定"
		x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
		if err != nil {
			log.Errorln("AlipayGenerateChargingAmountQrCode().FindUINodeAndTapMiddleXY().2 error:", err)
			return nil, err
		}
		// save the button XY
		dvc.alipayChargingQrCodeNextBtnXY = [2]int{x, y}
	}

	// screen cap
	imgdata, err := dvc.ScreenCap()
	if err != nil {
		log.Errorln("AlipayGenerateChargingAmountQrCode().ScreenCap() error:", err)
		return nil, err
	}

	return &AlipayChargingQrCode{Image: imgdata}, nil
}

// GotoAlipayTabProfile implement AdbDeviceHandler
// note: depends on StartAliPay()
func (dvc *AdbDevice) GotoAlipayTabProfile() error {
	var (
		resourceid   = "com.alipay.android.phone.wealth.home:id/tab_description"
		resourcetext = "我的"
		retryN       int
	)

RETRY:
	retryN++

	// ensure alipay is the top activity
	if activity, _ := dvc.CurrentTopActivity(); !strings.Contains(activity, "com.eg.android.AlipayGphone") {
		dvc.StartAliPay()
	}

	_, _, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
	if err != nil {
		if err == errUINodeNotFound {
			if retryN >= 5 {
				log.Errorln("GotoAlipayTabProfile() can't find the alipay wealth.home button")
				return errors.New("GotoAlipayTabProfile() can't find the alipay wealth.home button")
			}
			log.Warnln("GotoAlipayTabProfile() goback one step and retry to find the alipay wealth.home button ...")
			dvc.GoBack()
			goto RETRY
		} else {
			log.Errorln("GotoAlipayTabProfile().FindUINodeAndTapMiddleXY error:", err)
			return fmt.Errorf("GotoAlipayTabProfile().FindUINodeAndTapMiddleXY error: %v", err)
		}
	}

	return nil
}

// GotoAlipayListOrder implement AdbDeviceHandler
// note: depends on GotoAlipayTabProfile()
func (dvc *AdbDevice) GotoAlipayListOrder() error {
	// we knows it, directly click it
	if x, y := dvc.alipayOrderButtonXY[0], dvc.alipayOrderButtonXY[1]; x > 0 && y > 0 {
		return dvc.Click(x, y)
	}

	// find the UI node and click it
	var (
		resourceid   = "com.alipay.mobile.antui:id/item_left_text"
		resourcetext = "账单"
	)
	x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
	if err != nil {
		log.Errorln("GotoAlipayListOrder().FindUINodeAndTapMiddleXY() error:", err)
		return err
	}

	// save the order button XY
	dvc.alipayOrderButtonXY = [2]int{x, y}
	return nil
}

// AlipaySearchOrder implement AdbDeviceHandler
// note: depends on GotoAlipayListOrder()
func (dvc *AdbDevice) AlipaySearchOrder(orderID string) (*AlipayOrder, error) {
	if orderID == "" {
		return nil, errors.New("AlipaySearchOrder() order id required")
	}

	// ensure alipay order list is the top activity  (分页: 我的 -> 账单)
	if activity, _ := dvc.CurrentTopActivity(); !strings.Contains(activity, "com.eg.android.AlipayGphone/com.alipay.mobile.bill.list.ui.BillMainListActivity") {
		dvc.GotoAlipayTabProfile()
		dvc.GotoAlipayListOrder()
	}

	// we knows it, directly click it
	if x, y := dvc.alipayOrderSearchButtonXY[0], dvc.alipayOrderSearchButtonXY[1]; x > 0 && y > 0 {
		err := dvc.Click(x, y)
		if err != nil {
			log.Errorln("AlipaySearchOrder() directly click1 error:", err)
			return nil, err
		}
	} else { // find the UI node and click it
		var (
			resourceid   = "com.alipay.mobile.bill.list:id/search_btn"
			resourcetext = "搜索"
		)
		x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
		if err != nil {
			log.Errorln("AlipaySearchOrder().FindUINodeAndTapMiddleXY().1 error:", err)
			return nil, err
		}
		// save the search order button XY
		dvc.alipayOrderSearchButtonXY = [2]int{x, y}
	}

	// goback once afterwards to the alipay Order List Page
	defer dvc.GoBack()

	// input the order id
	if _, err := dvc.Run("input", "text", orderID); err != nil {
		return nil, fmt.Errorf("AlipaySearchOrder() input text orderID error: %v", err)
	}

	// we known it, directly click it
	if x, y := dvc.alipayOrderSearchEmitXY[0], dvc.alipayOrderSearchEmitXY[1]; x > 0 && y > 0 {
		err := dvc.Click(x, y)
		if err != nil {
			log.Errorln("AlipaySearchOrder() directly click2 error:", err)
			return nil, err
		}
	} else { // find the emit button and go search it!
		var (
			resourceid   = ""
			resourcetext = "搜索"
		)
		x, y, err := dvc.FindUINodeAndTapMiddleXY(resourceid, resourcetext)
		if err != nil {
			log.Errorln("AlipaySearchOrder().FindUINodeAndTapMiddleXY().2 error:", err)
			return nil, err
		}
		// save the search order button XY
		dvc.alipayOrderSearchEmitXY = [2]int{x, y}
	}

	// parse the search result page
	nodes, err := dvc.DumpCurrentUI()
	if err != nil {
		log.Errorln("AlipaySearchOrder().DumpCurrentUI() on search result page error:", err)
		return nil, err
	}

	// retry max 10 times if loading
	for i := 1; i <= 10; i++ {
		loadingNode := dvc.FindUINode(nodes, "android:id/progress", "") // 加载中
		if loadingNode != nil {
			time.Sleep(time.Second)
			nodes, _ = dvc.DumpCurrentUI()
			continue
		}
		break
	}

	// assume alipay loaded the search result
	var (
		resourceid   = "com.alipay.mobile.antui:id/tips"
		resourcetext = "没有找到近1年的相关记录"
	)
	tipNode := dvc.FindUINode(nodes, resourceid, resourcetext)
	if tipNode != nil { // not found this order
		return nil, errors.New("no such order")
	}

	var (
		billNameNode   = dvc.FindUINode(nodes, "com.alipay.mobile.bill.list:id/billName", "") // comment-user-username
		billAmountNode = dvc.FindUINode(nodes, "com.alipay.mobile.bill.list:id/billAmount", "")
		billTime1Node  = dvc.FindUINode(nodes, "com.alipay.mobile.bill.list:id/timeInfo1", "")
		billTime2Node  = dvc.FindUINode(nodes, "com.alipay.mobile.bill.list:id/timeInfo2", "")
		billComment    string // comment
		billAccount    string // user-username
		billAmount     string // +0.01
		billTime1      string // 昨天
		billTime2      string // 11:42
	)

	if billNameNode == nil {
		return nil, errors.New("unexpected search result page: the `billName` UI node not found")
	}
	billNameText := billNameNode.Text // 111111111-bbk-bbk
	fields := strings.SplitN(billNameText, "-", 2)
	billComment = fields[0]
	billAccount = fields[1]
	if billComment == "" {
		return nil, errors.New("unexpected search result page: the `billName` UI node missing order comment text")
	}

	if billAmountNode == nil {
		return nil, errors.New("unexpected search result page: the `billAmount` UI node not found")
	}
	billAmount = billAmountNode.Text // +0.01
	billAmount = strings.TrimPrefix(billAmount, "+")

	if billTime1Node != nil {
		billTime1 = billTime1Node.Text // 昨天
	}
	if billTime2Node != nil {
		billTime2 = billTime2Node.Text // 11:42
	}

	return &AlipayOrder{
		Comment: billComment,
		Account: billAccount,
		Amount:  billAmount,
		Time:    billTime1 + "-" + billTime2,
	}, nil
}
