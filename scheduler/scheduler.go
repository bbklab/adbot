package scheduler

import (
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/robfig/cron"

	"github.com/bbklab/adbot/pkg/geoip"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/pkg/pubsub"
	"github.com/bbklab/adbot/pkg/routine"
)

var (
	sched *Scheduler // package scope global reference
)

// Scheduler is a runtime cluster scheduler
type Scheduler struct {
	master       *mole.Master      // cluster mole master reference
	routineMgr   *routine.Registry // goroutine registry manager
	joinMgr      *joinMgr          // node join notifier manager
	refreshMgr   *refreshMgr       // node refresh notifier manager
	arefreshMgr  *refreshMgr       // adb node refresh notifier manager (similar as above but for adbnode)
	auditLogger  *auditLogger      // audit logger
	cron         *cron.Cron        // cron
	adbcbpub     *pubsub.Publisher // adbpay order callback event publisher
	limitMgr     *rateLimiterMgr   // event rate limiter
	pubKeyData   string            // public key file path or text (for verify license data)
	tgbot        *tgbot            // telegram bot
	geo          geoip.Handler     // geo data
	startAt      time.Time         // started at time
	sync.RWMutex                   // protect leader flag
	leader       bool              // if elected as leader
}

// Init initilize the package scope scheduler reference,
// called while master boot up, fatal exit if met any errors
func Init(m *mole.Master, pubKeyData string) {
	if m == nil {
		log.Fatalln("nil cluster master")
	}

	geoReader, err := geoip.NewGeo(resGeoCity, resGeoAsn)
	if err != nil {
		log.Fatalln("initialize GeoIP2 data reader error", err)
	}

	sched = &Scheduler{
		master:      m,
		routineMgr:  routine.NewRegistry(),
		joinMgr:     newJoinMgr(),
		refreshMgr:  newRefreshMgr(),
		arefreshMgr: newRefreshMgr(),
		auditLogger: newRollingAuditLogger(),
		cron:        cron.New(),
		adbcbpub:    pubsub.NewPublisher(time.Second*5, 1024),
		limitMgr:    newRateLimiter(),
		pubKeyData:  pubKeyData,
		tgbot:       newRuntimeTGBot(),
		geo:         geoReader,
		leader:      false,
		startAt:     time.Now(),
	}

	// start cron daemon
	// mark all of adb devices .OverQuota == false
	sched.cron.AddFunc("0 0 0 * * *", func() { ResetAllAdbDevicesOverQuotaFlag() })
	sched.cron.Start()

	// register node join/die call back
	m.RegisterNodeJoinCallBack(NodeJoinCallBack)
	m.RegisterNodeDieCallBack(NodeDieCallBack)
}

// Leader
//

// SetLeader update current leader flag
func SetLeader(flag bool) {
	sched.Lock()
	sched.leader = flag
	sched.Unlock()
}

func isLeader() bool {
	sched.RLock()
	defer sched.RUnlock()
	return sched.leader
}

// Pubsub Adb Order Events
//

// PublishAdbOrderCallbackEvent  is exported
func PublishAdbOrderCallbackEvent(orderID string) {
	sched.adbcbpub.Publish(orderID)
}

func subscribeAdbOrderCallbackEvent(orderID string, timeout time.Duration) error {
	sub := sched.adbcbpub.Subcribe(func(v interface{}) bool {
		if vv, ok := v.(string); ok {
			return vv == orderID
		}
		return false
	})
	defer sched.adbcbpub.Evict(sub)

	// hanging wait until timeout
	select {
	case <-sub:
		return nil

	case <-time.After(timeout):
		return errors.New("timeout while waitting for backend adb callback")
	}
}

// Geo
//

// GetGeoInfoZh is exported
func GetGeoInfoZh(addr string) *geoip.GeoInfo {
	return sched.geo.GetGeoInfo(addr, "zh-CN")
}

// GetGeoInfoEn is exported
func GetGeoInfoEn(addr string) *geoip.GeoInfo {
	return sched.geo.GetGeoInfo(addr, "en")
}

// CurrentGeoMetaData is exported
func CurrentGeoMetaData() map[string]maxminddb.Metadata {
	return sched.geo.Metadata()
}

// UpdateGeoData is exported
func UpdateGeoData() (time.Duration, error) {
	startAt := time.Now()
	err := sched.geo.Update()
	return time.Since(startAt), err
}

// Routines
//

// AllGoroutines show all of registered routines
func AllGoroutines() map[string][]string {
	return sched.routineMgr.All()
}

// Goroutines get given type of routines list
func Goroutines(typ string) []string {
	return sched.routineMgr.GetType(typ)
}

// IsRegisteredGoRoutine check if given type/name goroutine registered
func IsRegisteredGoRoutine(typ, name string) bool {
	return sched.routineMgr.ExistsRoutine(typ, name)
}

// RegisterGoroutine register a goroutine name for given type
func RegisterGoroutine(typ, name string) {
	sched.routineMgr.AddRoutine(typ, name)
}

// DeRegisterGoroutine de-register a goroutine name for given type
func DeRegisterGoroutine(typ, name string) {
	sched.routineMgr.DelRoutine(typ, name)
}
