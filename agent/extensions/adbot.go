package extensions

import (
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/adbot"
)

var (
	adb adbot.AdbHandler
)

func setupAdbotHandler() error {
	var err error
	if adb == nil {
		adb, err = adbot.NewAdb()
		if err != nil {
			return err
		}
	}
	return nil
}

// ListAdbDevices return the adb devices list
func ListAdbDevices() (map[string]*adbot.AndroidSysInfo, error) {
	if err := setupAdbotHandler(); err != nil {
		return nil, err
	}

	ids, err := adb.ListAdbDevices()
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

			dvc := adb.NewDevice(id)
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
