package adbot

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/utils"
	units "github.com/docker/go-units"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

// AndroidSysInfo is exported
type AndroidSysInfo struct {
	SerialNo           string              `json:"serial_no"`            // ro.serialno
	DeviceName         string              `json:"device_name"`          // persist.sys.device_name: xxx redmi
	Manufacturer       string              `json:"manufacturer"`         // ro.product.manufacturer: Xiaomi
	ProductBrand       string              `json:"product_brand"`        // ro.product.brand: Xiaomi
	ProductModel       string              `json:"product_model"`        // ro.product.model: Redmi 4A
	ProductName        string              `json:"product_name"`         // ro.product.name: rolex
	ProductLocale      string              `json:"product_locale"`       // ro.product.locale: zh-CN
	ReleaseVersion     string              `json:"release_version"`      // ro.build.version.release: 6.0.1
	SDKVersion         string              `json:"sdk_version"`          // ro.build.version.sdk: 23
	BuildDateUTC       string              `json:"build_date_utc"`       // ro.build.date.utc: 1490366979
	TimeZone           string              `json:"time_zone"`            // persist.sys.timezone: Asia/Shanghai
	GsmOperatorAlpha   string              `json:"gsm_operator_alpha"`   // gsm.operator.alpha: 中国移动
	GsmOperatorCountry string              `json:"gsm_operator_country"` // gsm.operator.iso-country: cn
	GsmSerial          string              `json:"gsm_serial"`           // gsm.serial: 2Z721F215568
	GsmSimState        string              `json:"gsm_sim_state"`        // gsm.sim.state: READY,ABSENT
	GsmNitzTime        string              `json:"gsm_nitz_time"`        // gsm.nitz.time: 1559783581005
	GsmNitzTimeAt      time.Time           `json:"gsm_nitz_time_at"`     // parsed from above
	BootTime           string              `json:"boot_time"`            // ro.runtime.firstboot: 1557559953185
	BootTimeAt         time.Time           `json:"boot_time_at"`         // parsed from above
	BootTimeFor        string              `json:"boot_time_for"`        // parsed from above
	Battery            *AndroidBatteryInfo `json:"battery"`              // battery info
}

func parseAndroidSysinfo(text string) *AndroidSysInfo {
	text = strings.NewReplacer([]string{"[", "", "]", ""}...).Replace(text)

	var (
		buf     = bytes.NewBufferString(text)
		scanner = bufio.NewScanner(buf)
		lbs     = label.Labels{}
	)
	for scanner.Scan() {
		kv := strings.SplitN(scanner.Text(), ":", 2)
		if len(kv) < 2 {
			continue
		}
		lbs.Set(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
	}

	info := &AndroidSysInfo{
		SerialNo:           lbs.Get("ro.serialno"),
		DeviceName:         lbs.Get("persist.sys.device_name"),
		Manufacturer:       lbs.Get("ro.product.manufacturer"),
		ProductBrand:       lbs.Get("ro.product.brand"),
		ProductModel:       lbs.Get("ro.product.model"),
		ProductName:        lbs.Get("ro.product.name"),
		ProductLocale:      lbs.Get("ro.product.locale"),
		ReleaseVersion:     lbs.Get("ro.build.version.release"),
		SDKVersion:         lbs.Get("ro.build.version.sdk"),
		BuildDateUTC:       lbs.Get("ro.build.date.utc"),
		TimeZone:           lbs.Get("persist.sys.timezone"),
		GsmOperatorAlpha:   lbs.Get("gsm.operator.alpha"),
		GsmOperatorCountry: lbs.Get("gsm.operator.iso-country"),
		GsmSerial:          lbs.Get("gsm.serial"),
		GsmSimState:        lbs.Get("gsm.sim.state"),
		GsmNitzTime:        lbs.Get("gsm.nitz.time"),
		BootTime:           lbs.Get("ro.runtime.firstboot"),
	}

	var gsmUnixSecN int64
	if len(info.GsmNitzTime) == 13 {
		gsmUnixSecN, _ = strconv.ParseInt(info.GsmNitzTime, 10, 64)
		gsmUnixSecN /= 1000
	} else {
		gsmUnixSecN, _ = strconv.ParseInt(info.GsmNitzTime, 10, 64)
	}
	info.GsmNitzTimeAt = time.Unix(gsmUnixSecN, 0)

	var bootUnixSecN int64
	if len(info.BootTime) == 13 {
		bootUnixSecN, _ = strconv.ParseInt(info.BootTime, 10, 64)
		bootUnixSecN /= 1000
	} else {
		bootUnixSecN, _ = strconv.ParseInt(info.BootTime, 10, 64)
	}
	info.BootTimeAt = time.Unix(bootUnixSecN, 0)
	info.BootTimeFor = units.HumanDuration(time.Since(info.BootTimeAt))

	return info
}

// AndroidBatteryInfo is exported
type AndroidBatteryInfo struct {
	ACPowered       string `json:"ac_powered"`       // 电源线充电
	USBPowered      string `json:"usb_powered"`      // USB充电
	WireLessPowered string `json:"wireless_powered"` // 无线充电
	Status          string `json:"status"`           // UNKNOWN=1，CHARGING=2，DISCHARGING=3，NOT_CHARGING=4，FULL=5
	Level           int    `json:"level"`            // 当前电量
	Scale           int    `json:"scale"`            // 最大电量
}

func parseAndroidBatteryInfo(text string) *AndroidBatteryInfo {
	var (
		buf     = bytes.NewBufferString(text)
		scanner = bufio.NewScanner(buf)
		lbs     = label.Labels{}
	)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		kv := strings.SplitN(line, ":", 2)
		if len(kv) < 2 {
			continue
		}
		lbs.Set(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
	}

	level, _ := strconv.Atoi(lbs.Get("level"))
	scale, _ := strconv.Atoi(lbs.Get("scale"))
	return &AndroidBatteryInfo{
		ACPowered:       lbs.Get("AC powered"),
		USBPowered:      lbs.Get("USB powered"),
		WireLessPowered: lbs.Get("Wireless powered"),
		Status:          lbs.Get("status"),
		Level:           level,
		Scale:           scale,
	}
}

// AndroidUINode is exported
type AndroidUINode struct {
	Index       string // 4
	Text        string // 手机电池已充满(100%)
	ResourceID  string // com.android.systemui:id/clear_all_button
	Package     string // com.android.systemui
	ContentDesc string // 清除所有通知。
	Bounds      string // [301,1138][419,1256]
}

// MiddleXY is exported
func (n *AndroidUINode) MiddleXY() (int, int, error) {
	bounds := strings.NewReplacer(
		[]string{
			",", " ",
			"[", " ",
			"]", " ",
		}...).Replace(string(n.Bounds))

	locations := strings.Fields(bounds)
	if len(locations) != 4 {
		return -1, -1, errors.New("the UI bounds are invalid")
	}

	x1, y1, x2, y2 := locations[0], locations[1], locations[2], locations[3]
	if x1 == "" || y1 == "" || x2 == "" || y2 == "" {
		return -1, -1, errors.New("can't find the UI bounds value")
	}

	x1N, _ := strconv.Atoi(x1)
	y1N, _ := strconv.Atoi(y1)
	x2N, _ := strconv.Atoi(x2)
	y2N, _ := strconv.Atoi(y2)
	x := (x1N + x2N) / 2
	y := (y1N + y2N) / 2

	return x, y, nil
}

func parseAndroidUINodes(xmldata []byte) ([]*AndroidUINode, error) {
	var (
		nodes        = make([]string, 0, 0)
		noderx       = regexp.MustCompile(`<node ([^>]*)/>`)
		nodesMatched = noderx.FindAllSubmatch(xmldata, -1)
	)

	for _, nodeMatched := range nodesMatched {
		if len(nodeMatched) >= 2 {
			nodes = append(nodes, string(nodeMatched[1]))
		}
	}

	var (
		nodekvrx = regexp.MustCompile(`([^ \t]+)="([^"]+)"`)
		ret      = []*AndroidUINode{}
	)
	for _, node := range nodes {
		nodekvsMatched := nodekvrx.FindAllSubmatch([]byte(node), -1)
		lbs := label.Labels{}
		for _, nodekvMatched := range nodekvsMatched {
			if len(nodekvMatched) >= 3 {
				key := string(nodekvMatched[1])
				val := string(nodekvMatched[2])
				lbs.Set(key, val)
			}
		}
		ret = append(ret, &AndroidUINode{ // here we only pick up the fields we want
			Index:       lbs.Get("index"),
			Text:        lbs.Get("text"),
			ResourceID:  lbs.Get("resource-id"),
			Package:     lbs.Get("package"),
			ContentDesc: lbs.Get("content-desc"),
			Bounds:      lbs.Get("bounds"),
		})
	}

	return ret, nil
}

// AndroidSysNotify is exported
type AndroidSysNotify struct {
	Source  string `json:"source"`  // source: om.android.mms, com.android.calendar
	Message string `json:"message"` // ayy text messages ...
}

// EqualsTo is exported
func (n *AndroidSysNotify) EqualsTo(new *AndroidSysNotify) bool {
	return n.Source == new.Source && n.Message == new.Message
}

// AlipayOrder is exported
type AlipayOrder struct {
	Comment string `json:"comment"`
	Account string `json:"account"`
	Amount  string `json:"amount"`
	Time    string `json:"time"`
}

// AlipayChargingQrCode is exported
type AlipayChargingQrCode struct {
	Image []byte `json:"image"`
}

func parseSectionText(text string, rxTitle *regexp.Regexp, keys []string) ([]label.Labels, error) {
	var (
		ret []label.Labels
	)

	var (
		buf    = bytes.NewBuffer([]byte(text))
		reader = bufio.NewReader(buf)
		item   = label.Labels{}
	)

	keys = utils.MakeUniq(keys)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if item.Len() > 1 && item.Get("_TITLE_") != "" {
					ret = append(ret, item.Clone())
					item = label.Labels{}
				}
				break
			}
			return nil, err
		}

		line = bytes.TrimSpace(line)

		if rxTitle.Match(line) {
			if item.Len() > 1 && item.Get("_TITLE_") != "" {
				ret = append(ret, item.Clone()) // save current parsed repo source
				item = label.Labels{}
			}

			item.Set("_TITLE_", string(line))
			continue
		}

		for _, key := range keys {
			prefix := []byte(fmt.Sprintf("%s=", key))
			if bytes.HasPrefix(line, prefix) {
				val := string(bytes.TrimPrefix(line, prefix))
				if val != "null" && val != "String" { // ignore null values
					item.Set(key, val)
					break
				}
			}
		}
	}

	return label.Uniq(ret), nil
}

// parse notification title from string like:
//   NotificationRecord(0x4270cd28: pkg=com.tencent.mm user=UserHandle{0} id=1363572628 tag=null score=10: Notification(pri=1 contentView=com.tencent.mm/0x1090065 ...
func parseSysNotifyFromPKG(sysNotifyTitle string) string {
	for _, kv := range strings.Fields(sysNotifyTitle) {
		if strings.HasPrefix(kv, "pkg=") {
			return strings.TrimPrefix(kv, "pkg=")
		}
	}
	return ""
}
