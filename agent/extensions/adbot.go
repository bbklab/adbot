package extensions

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/routine"
)

var (
	am   *adbMgr // runtime adb managers
	amux sync.Mutex
)

func setupAdbotMgr() error {
	amux.Lock()
	defer amux.Unlock()

	if am != nil {
		return nil
	}

	ah, err := adbot.NewAdb()
	if err != nil {
		return err
	}

	am = &adbMgr{
		ah:  ah,
		dhs: make(map[string]adbot.AdbDeviceHandler),
		reg: routine.NewRegistry(),
	}
	go am.watchAllDeviceEvents()
	go am.watchAllDeviceAlipayActivity()

	return nil
}

type adbMgr struct {
	ah         adbot.AdbHandler                  // adb handler (father adb handler)
	dhs        map[string]adbot.AdbDeviceHandler // map of device handlers
	sync.Mutex                                   // protect above map
	reg        *routine.Registry                 // device watcher goroutine registry
}

// getDevice get or insert given device id handler
func (mgr *adbMgr) getDevice(id string) (adbot.AdbDeviceHandler, error) {
	mgr.Lock()
	defer mgr.Unlock()

	dh, ok := mgr.dhs[id]
	if ok {
		return dh, nil
	}

	newdh, err := mgr.ah.NewDevice(id)
	if err != nil {
		return nil, err
	}

	// add new device handler
	mgr.dhs[id] = newdh

	// launch new device alipay watcher
	if !mgr.reg.ExistsRoutine("adb_device_alipay_watcher", id) {
		go mgr.watchDeviceAlipay(newdh, id)
	}

	return newdh, nil
}

func (mgr *adbMgr) watchDeviceAlipay(dvc adbot.AdbDeviceHandler, id string) {
	var (
		loopName = fmt.Sprintf("device %s alipay order watcher loop", id)
	)

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s", loopName)

	mgr.reg.AddRoutine("adb_device_alipay_watcher", id)
	defer mgr.reg.DelRoutine("adb_device_alipay_watcher", id)

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	ch, stopch := dvc.WatchSysNotifies()
	defer close(stopch)

	for {
		var (
			err error
			msg string
		)

		select {
		case sysNotify := <-ch:
			if sysNotify.Source != "com.eg.android.AlipayGphone" {
				continue // skip
			}

			msg = sysNotify.Message
			ev := &adbot.AdbEvent{
				Serial:  id,
				Type:    adbot.AdbEventAlipayOrder,
				Message: msg,
				Time:    time.Now(),
			}
			err = reportAdbEvent(ev)

		case <-ticker.C:
			msg = "NOOP ALIPAY EVENT IN CASE OF MISSING SYSNOTIFY"
			ev := &adbot.AdbEvent{
				Serial:  id,
				Type:    adbot.AdbEventAlipayOrder,
				Message: msg,
				Time:    time.Now(),
			}
			err = reportAdbEvent(ev)
		}

		if err != nil {
			log.Warnf("report alipay order event to master error: %v - [%s]", err, msg)
		} else {
			log.Infof("report alipay order event to master succeed - [%s]", msg)
		}
	}
}

// watch all device's event
func (mgr *adbMgr) watchAllDeviceEvents() {
	var (
		loopName = fmt.Sprintf("all devices events watcher loop")
	)

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s", loopName)

	mgr.reg.AddRoutine("adb_device_event_watcher", "system")
	defer mgr.reg.DelRoutine("adb_device_event_watcher", "system")

	ch, stopch := mgr.ah.WatchAdbEvents()
	defer close(stopch)

	for ev := range ch {
		if err := reportAdbEvent(ev); err != nil {
			log.Warnf("report adb device event to master error: %v - [%s]", err, ev.Message)
		} else {
			log.Infof("report adb device event to master succeed - [%s]", ev.Message)
		}
	}
}

// ensure all device's alipay is the top activity
func (mgr *adbMgr) watchAllDeviceAlipayActivity() {
	var (
		loopName = fmt.Sprintf("all devices alipay activity watcher loop")
	)

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s", loopName)

	mgr.reg.AddRoutine("adb_alipay_activity_watcher", "system")
	defer mgr.reg.DelRoutine("adb_alipay_activity_watcher", "system")

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		ids, _ := am.ah.ListAdbDevices()

		var wg sync.WaitGroup
		wg.Add(len(ids))
		for _, id := range ids {
			go func(id string) {
				defer wg.Done()

				dvc, err := am.getDevice(id)
				if err != nil {
					log.Warnf("%s getDevice() on %s error: %v", loopName, id, err)
					return
				}

				if !dvc.IsAwake() {
					dvc.AwakenScreen()
					// dvc.SwipeUpUnlock() // note: this require the adb device must disabled the screen lock via USB Debug Options
				}

				if activity, _ := dvc.CurrentTopActivity(); !strings.Contains(activity, "com.eg.android.AlipayGphone") {
					dvc.StartAliPay()
				}
			}(id)
		}
		wg.Wait()
	}
}

func reportAdbEvent(ev *adbot.AdbEvent) error {
	var (
		client = GetMasterAPIClient()
		err    error
	)
	for i := 1; i <= 3; i++ {
		err = client.ReportAdbEvent(ev)
		if err == nil {
			break
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return err
}

// ListAdbDevices return the adb devices list
func ListAdbDevices() (map[string]*adbot.AndroidSysInfo, error) {
	if err := setupAdbotMgr(); err != nil {
		return nil, err
	}

	ids, err := am.ah.ListAdbDevices()
	if err != nil {
		return nil, err
	}

	var (
		ret = make(map[string]*adbot.AndroidSysInfo)
		mux sync.Mutex
		wg  sync.WaitGroup
	)
	wg.Add(len(ids))
	for _, id := range ids {
		go func(id string) {
			defer wg.Done()

			dvc, err := am.getDevice(id)
			if err != nil {
				log.Errorln("get adb device handler error", id, err)
				return
			}

			info, err := dvc.SysInfo()
			if err != nil {
				log.Errorln("get adb device sysinfo error", id, err)
				return
			}

			mux.Lock()
			ret[id] = info
			mux.Unlock()
		}(id)
	}
	wg.Wait()

	return ret, nil
}

// CheckAdbAlipayOrder check one alipay order on given adb device
func CheckAdbAlipayOrder(dvcID, orderID string) (*adbot.AlipayOrder, error) {
	dvc, err := am.getDevice(dvcID)
	if err != nil {
		return nil, err
	}

	if !dvc.IsAwake() {
		err := dvc.AwakenScreen()
		if err != nil {
			return nil, err
		}
		dvc.SwipeUpUnlock()
		dvc.GotoHome()
	}

	return dvc.AlipaySearchOrder(orderID)
}

// AdbDeviceScreenCap take screen cap on given adb device
func AdbDeviceScreenCap(dvcID string) ([]byte, error) {
	dvc, err := am.getDevice(dvcID)
	if err != nil {
		return nil, err
	}
	return dvc.ScreenCap()
}

// AdbDeviceReboot reboot given adb device
func AdbDeviceReboot(dvcID string) error {
	dvc, err := am.getDevice(dvcID)
	if err != nil {
		return err
	}

	// reboot
	err = dvc.Reboot()
	if err != nil {
		return err
	}

	// report device die event
	reportAdbEvent(&adbot.AdbEvent{
		Serial:  dvcID,
		Type:    adbot.AdbEventDeviceDie,
		Message: "device rebooted",
		Time:    time.Now(),
	})

	return nil
}
