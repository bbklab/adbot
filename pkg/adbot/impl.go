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

// WatchAdbEvents implement AdbHandler
func (a *Adb) WatchAdbEvents() (<-chan *AdbEvent, chan struct{}) {
	ch := make(chan *AdbEvent, 1024)
	stopch := make(chan struct{})

	go func() {
		w := a.h.NewDeviceWatcher()
		defer w.Shutdown()

		for ev := range w.C() {
			var (
				serial = ev.Serial
				event  string
				msg    = fmt.Sprintf("%s->%s", ev.OldState, ev.NewState)
			)

			switch {
			case ev.CameOnline():
				event = AdbEventDeviceAlive
			case ev.WentOffline():
				event = AdbEventDeviceDie
			}

			select {
			case ch <- &AdbEvent{Serial: serial, Type: event, Message: msg, Time: time.Now()}:
			case <-stopch:
				return
			default:
			}
		}
	}()

	return ch, stopch
}

// NewDevice implement AdbHandler
func (a *Adb) NewDevice(serial string) (AdbDeviceHandler, error) {
	dvc := &AdbDevice{
		h: a.h.Device(goadb.DeviceWithSerial(serial)),
	}
	_, err := dvc.h.Serial()
	return dvc, err
}

//
//  AdbDeviceHandler Implemention
//

// AdbDevice is an AdbDeviceHandler implemention
type AdbDevice struct {
	h *goadb.Device
	l sync.Mutex // synchronized Adb Ops including any combined or single ops, mostly are `input` ops

	// alipay button XY
	alipayBillListSearchButton [2]int // 我的->账单->搜索
	alipayBillSearchEmitButton [2]int // 我的->账单->搜索->搜索
}

// Serial implement AdbDeviceHandler
func (dvc *AdbDevice) Serial() (string, error) {
	return dvc.h.Serial()
}

// Exists implement AdbDeviceHandler
func (dvc *AdbDevice) Exists() bool {
	_, err := dvc.h.DeviceInfo()
	if err == nil {
		return true
	}
	return !strings.HasPrefix(err.Error(), "DeviceNotFound")
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

// Reboot implement AdbDeviceHandler
func (dvc *AdbDevice) Reboot() error {
	_, err := dvc.Run("reboot")
	return err
}

// Run implement AdbDeviceHandler
func (dvc *AdbDevice) Run(cmd string, args ...string) (string, error) {
	return dvc.h.RunCommand(cmd, args...)
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
	dvc.l.Lock()
	defer dvc.l.Unlock()

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
		time.Sleep(time.Millisecond * 500)
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
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.gotoHomeUnsafe()
}

func (dvc *AdbDevice) gotoHomeUnsafe() error {
	_, err := dvc.Run("input", "keyevent", "3")
	return err
}

// GoBack implement AdbDeviceHandler
func (dvc *AdbDevice) GoBack() error {
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.gobackUnsafe()
}

func (dvc *AdbDevice) gobackUnsafe() error {
	_, err := dvc.Run("input", "keyevent", "4")
	return err
}

// Click implement AdbDeviceHandler
func (dvc *AdbDevice) Click(x, y int) error {
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.clickUnsafe(x, y)
}

func (dvc *AdbDevice) clickUnsafe(x, y int) error { // maybe called within other combined ops
	_, err := dvc.Run("input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	return err
}

// Swipe implement AdbDeviceHandler
func (dvc *AdbDevice) Swipe(x1, y1, x2, y2 int) error {
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.swipeUnsafe(x1, y1, x2, y2)
}

func (dvc *AdbDevice) swipeUnsafe(x1, y1, x2, y2 int) error {
	_, err := dvc.Run("input", "swipe", strconv.Itoa(x1), strconv.Itoa(y1), strconv.Itoa(x2), strconv.Itoa(y2))
	return err
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
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.dumpCurrentUIUnsafe()
}

func (dvc *AdbDevice) dumpCurrentUIUnsafe() ([]*AndroidUINode, error) {
	bs, err := dvc.Run("uiautomator", "dump")
	if err != nil {
		log.Errorln("dumpCurrentUI().ui.dump error:", err)
		return nil, err
	}

	var xmlfile string
	fields := strings.SplitN(bs, "dumped to:", 2)
	if len(fields) == 2 {
		xmlfile = strings.TrimSpace(fields[1])
	}

	reader, err := dvc.h.OpenRead(xmlfile)
	if err != nil {
		log.Errorln("dumpCurrentUI().read.xml.file error:", xmlfile, err)
		return nil, err
	}
	defer reader.Close()

	xmlbs, _ := ioutil.ReadAll(reader)
	return parseAndroidUINodes(xmlbs)
}

var (
	errUINodeNotFound = errors.New("can't find the target UI node")
)

// FindUINodeAndClick implement AdbDeviceHandler
func (dvc *AdbDevice) FindUINodeAndClick(resourceid, resourcetext string) (int, int, error) {
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.findUINodeAndClickUnsafe(resourceid, resourcetext)
}

func (dvc *AdbDevice) findUINodeAndClickUnsafe(resourceid, resourcetext string) (int, int, error) {
	nodes, err := dvc.dumpCurrentUIUnsafe()
	if err != nil {
		log.Errorln("findUINodeAndClick().DumpCurrentUI() error:", err)
		return -1, -1, err
	}

	var targetNode = dvc.findUINode(nodes, resourceid, resourcetext)
	if targetNode == nil {
		return -1, -1, errUINodeNotFound
	}

	x, y, err := targetNode.MiddleXY()
	if err != nil {
		return -1, -1, fmt.Errorf("findUINodeAndClick() can't find the target UI node MiddleXY(): %v", err)
	}

	return x, y, dvc.clickUnsafe(x, y)
}

func (dvc *AdbDevice) findUINode(nodes []*AndroidUINode, resourceid, resourcetext string) *AndroidUINode {
	for _, node := range nodes {
		if node.ResourceID == resourceid {
			if resourcetext == "" {
				return node
			}
			if strings.Contains(node.Text, resourcetext) {
				return node
			}
		}
	}
	return nil
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
		return nil
	}

	var (
		rxTitle = regexp.MustCompile(`^[ \t]*NotificationRecord`)
		keys    = []string{"tickerText"}
	)
	ret, err := parseSectionText(string(bs), rxTitle, keys)
	if err != nil {
		log.Warnln("ListSysNotifies() error:", err)
		return nil
	}

	notifies := []*AndroidSysNotify{}
	for _, lbs := range ret {
		title := lbs.Get("_TITLE_")
		text := lbs.Get("tickerText")
		pkg, id := parseSysNotifyTitle(title)
		notifies = append(notifies, &AndroidSysNotify{
			ID:      id,
			Source:  pkg,
			Message: text,
		})
	}
	return notifies
}

// ClearSysNotifies implement AdbDeviceHandler
func (dvc *AdbDevice) ClearSysNotifies() error {
	dvc.l.Lock()
	defer dvc.l.Unlock()
	return dvc.clearSysNotifiesUnsafe()
}

func (dvc *AdbDevice) clearSysNotifiesUnsafe() error {
	dvc.swipeUnsafe(300, 0, 300, 1200) // pull down from top to show the notifies

	// find the UI node and click it
	var (
		resourceid   = "com.android.systemui:id/dismiss_view"
		resourcetext = ""
	)
	_, _, err := dvc.findUINodeAndClickUnsafe(resourceid, resourcetext)
	if err != nil {
		if err == errUINodeNotFound {
			log.Warnln("ClearSysNotifies() no system notifies need to be cleaned")
			return nil
		}
		log.Errorln("ClearSysNotifies().findUINodeAndClick() error:", err)
		return err
	}

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
	var err error
	for i := 1; i <= 5; i++ {
		if dvc.isAlipayActive() {
			return nil
		}
		_, err = dvc.Run("am", "start", "com.eg.android.AlipayGphone/.AlipayLogin")
		if err != nil {
			log.Warnf("StartAliPay() %d am.start error: %v", i, err)
		}
		time.Sleep(time.Second)
	}
	return err
}

// AlipaySearchOrder implement AdbDeviceHandler
// 我的 -> 账单 -> 搜索 -> 搜索
func (dvc *AdbDevice) AlipaySearchOrder(orderID string) (*AlipayOrder, error) {
	dvc.l.Lock()
	defer dvc.l.Unlock()

	if orderID == "" {
		return nil, errors.New("AlipaySearchOrder() order id required")
	}

	// clear any previous alipay notifies
	if n := dvc.countAlipayNotifies(); n > 3 {
		dvc.clearSysNotifiesUnsafe()
	}

	// ensure current top activity is alipay bill list page
	if !dvc.isAlipayBillListActive() {
		err := dvc.gotoAlipayListOrder()
		if err != nil {
			return nil, err
		}
	}

	var (
		resourceid   string
		resourcetext string
		newX, newY   int
		err          error
	)

	// 点击账单列表页的'搜索'
	// if we know the button XY, directly click it
	if currX, currY := dvc.alipayBillListSearchButton[0], dvc.alipayBillListSearchButton[1]; currX > 0 && currY > 0 {
		dvc.clickUnsafe(currX, currY)
		// ensure we got the right activity, otherwise re-caculate the button XY and click it
		err = dvc.waitAlipayBillSearchActive(5, time.Second)
		if err != nil {
			log.Warnln("AlipaySearchOrder().DirectClick() on `Bill-List-Search-Button` maybe outdated, need fix the button XY")
		} else {
			goto EMITSEARCH
		}
	}

	// caculate the button XY and click it
	resourceid = "com.alipay.mobile.bill.list:id/search_btn"
	resourcetext = "搜索"
	newX, newY, err = dvc.findUINodeAndClickUnsafe(resourceid, resourcetext)
	if err != nil {
		log.Errorln("AlipaySearchOrder().findUINodeAndClick() on `Bill-List-Search-Button` error:", err)
		return nil, err
	}
	// ensure we got the right activity
	err = dvc.waitAlipayBillSearchActive(5, time.Second)
	if err != nil {
		log.Errorln("AlipaySearchOrder().waitAlipayBillSearchActive() error:", err)
		return nil, err
	}
	// now we got the right activity
	if newX > 0 && newY > 0 {
		dvc.alipayBillListSearchButton = [2]int{newX, newY} // renew the new button XY
		newX, newY = 0, 0                                   // clear the value
	}

EMITSEARCH:

	// goback once afterwards to the alipay order list page
	defer func() {
		dvc.gobackUnsafe()
		dvc.waitAlipayBillListActive(5, time.Second)
	}()

	// 输入订单号
	if _, err := dvc.Run("input", "text", orderID); err != nil {
		return nil, fmt.Errorf("AlipaySearchOrder() input text orderID error: %v", err)
	}

	// 点击账单搜索页的'搜索' 进行提交
	// if we know the button XY, directly click it
	if currX, currY := dvc.alipayBillSearchEmitButton[0], dvc.alipayBillSearchEmitButton[1]; currX > 0 && currY > 0 {
		dvc.clickUnsafe(currX, currY)
	} else { // caculate the button XY and click it
		resourceid = ""
		resourcetext = "搜索"
		newX, newY, err = dvc.findUINodeAndClickUnsafe(resourceid, resourcetext)
		if err != nil {
			log.Errorln("AlipaySearchOrder().findUINodeAndClick() on `Bill-Search-Emit-Button` error:", err)
			return nil, err
		}
		if newX > 0 && newY > 0 {
			dvc.alipayBillSearchEmitButton = [2]int{newX, newY} // renew the new button XY
			newX, newY = 0, 0                                   // clear the value
		}
	}

	// parse the search result page
	nodes, err := dvc.dumpCurrentUIUnsafe()
	if err != nil {
		log.Errorln("AlipaySearchOrder().dumpCurrentUI() on search result page error:", err)
		return nil, err
	}
	// retry max 10 times if loading
	for i := 1; i <= 10; i++ {
		loadingNode := dvc.findUINode(nodes, "android:id/progress", "") // 加载中
		if loadingNode != nil {
			time.Sleep(time.Second)
			nodes, _ = dvc.dumpCurrentUIUnsafe()
			continue
		}
		break
	}

	// we assume alipay loaded the search result
	resourceid = "com.alipay.mobile.antui:id/tips"
	resourcetext = "没有找到"
	tipNode := dvc.findUINode(nodes, resourceid, resourcetext)
	if tipNode != nil { // not found this order
		return nil, errors.New("no such order")
	}

	var (
		billNameNode   = dvc.findUINode(nodes, "com.alipay.mobile.bill.list:id/billName", "") // comment-user-username
		billAmountNode = dvc.findUINode(nodes, "com.alipay.mobile.bill.list:id/billAmount", "")
		billTime1Node  = dvc.findUINode(nodes, "com.alipay.mobile.bill.list:id/timeInfo1", "")
		billTime2Node  = dvc.findUINode(nodes, "com.alipay.mobile.bill.list:id/timeInfo2", "")
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

func (dvc *AdbDevice) countAlipayNotifies() int {
	notifies := dvc.ListSysNotifies()
	var n int
	for _, notify := range notifies {
		if notify.Source == "com.eg.android.AlipayGphone" {
			n++
		}
	}
	return n
}

func (dvc *AdbDevice) gotoAlipayListOrder() error {
	if err := dvc.gotoAlipayTabProfile(); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 300)

	var (
		resourceid   = "com.alipay.mobile.antui:id/item_left_text"
		resourcetext = "账单"
	)
	_, _, err := dvc.findUINodeAndClickUnsafe(resourceid, resourcetext)
	if err != nil {
		log.Errorln("gotoAlipayListOrder().findUINodeAndClick() error:", err)
		return err
	}

	// now we expect the bill list activity at top
	return dvc.waitAlipayBillListActive(10, time.Second)
}

func (dvc *AdbDevice) gotoAlipayTabProfile() error {
	if err := dvc.StartAliPay(); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 300)

	var (
		resourceid   = "com.alipay.android.phone.wealth.home:id/tab_description"
		resourcetext = "我的"
		maxRetry     = 10
		retryN       int
		err          error
	)

RETRY:
	retryN++
	if retryN > maxRetry {
		return fmt.Errorf("gotoAlipayTabProfile() failed after %d retries, error: %v", retryN, err)
	}

	dvc.StartAliPay() // ensure every time alipay at top
	_, _, err = dvc.findUINodeAndClickUnsafe(resourceid, resourcetext)
	if err != nil {
		if err == errUINodeNotFound {
			log.Warnln("gotoAlipayTabProfile() goback one step and retry to find the '我的' button ...")
		} else {
			log.Warnln("gotoAlipayTabProfile().findUINodeAndClick error:", err)
		}
		dvc.gobackUnsafe() // go back one step and retry
		goto RETRY
	}

	return nil
}

func (dvc *AdbDevice) isAlipayActive() bool {
	activity, _ := dvc.CurrentTopActivity()
	return strings.Contains(activity, "com.eg.android.AlipayGphone")
}

func (dvc *AdbDevice) isAlipayBillListActive() bool { // 我的->订单
	activity, _ := dvc.CurrentTopActivity()
	return activity == "com.eg.android.AlipayGphone/com.alipay.mobile.bill.list.ui.BillMainListActivity"
}

func (dvc *AdbDevice) waitAlipayBillListActive(maxWait int, interval time.Duration) error {
	for i := 1; i <= maxWait; i++ {
		if dvc.isAlipayBillListActive() {
			return nil
		}
		time.Sleep(interval)
		continue
	}
	return errors.New("failed to wait for the AlipayBillList Activity")
}

func (dvc *AdbDevice) isAlipayBillSearchActive() bool { // 我的->订单->搜索
	activity, _ := dvc.CurrentTopActivity()
	return activity == "com.eg.android.AlipayGphone/com.alipay.mobile.bill.list.ui.BillWordSearchActivity_"
}

func (dvc *AdbDevice) waitAlipayBillSearchActive(maxWait int, interval time.Duration) error {
	for i := 1; i <= maxWait; i++ {
		if dvc.isAlipayBillSearchActive() {
			return nil
		}
		time.Sleep(interval)
		continue
	}
	return errors.New("failed to wait for the AlipayBillSearch Activity")
}
