package scheduler

import (
	"errors"
	"sync"
	"time"

	"github.com/bbklab/adbot/i18n"
	"github.com/bbklab/adbot/pkg/rate"
)

// rateLimiterMgr is a event limiter runtime manager
type rateLimiterMgr struct {
	sync.RWMutex
	m map[string]rate.Limiter // event key -> event rate limiter (the limiter will be GCed while Taken() == 0)
}

func newRateLimiter() *rateLimiterMgr {
	mgr := &rateLimiterMgr{
		m: make(map[string]rate.Limiter),
	}
	go mgr.gc()
	return mgr
}

func (mgr *rateLimiterMgr) gc() {
	for ; ; time.Sleep(time.Minute) {
		mgr.Lock()
		for evkey, limiter := range mgr.m {
			if limiter.Taken() == 0 {
				delete(mgr.m, evkey)
			}
		}
		mgr.Unlock()
	}
}

func (mgr *rateLimiterMgr) list() map[string]string {
	mgr.RLock()
	defer mgr.RUnlock()

	ret := make(map[string]string)
	for evkey, limiter := range mgr.m {
		ret[evkey] = limiter.String()
	}
	return ret
}

// ListEventLimiters list current all of event limiters
func ListEventLimiters() map[string]string {
	return sched.limitMgr.list()
}

// CheckEventLimiter check the rate limiter if the given event key reached limitation
func CheckEventLimiter(evkey string) error {
	sched.limitMgr.Lock()
	defer sched.limitMgr.Unlock()

	l := sched.limitMgr.m[evkey]
	if l == nil { // if previous not exists, directly pass the rate check
		return nil
	}

	remains := l.Remains()
	if remains == 0 {
		return errors.New(i18n.MsgRateLimited)
	}
	return nil
}

// IncrEventLimiter increase the counter for given event key, if the previous of
// corresponding rate limiter for the given event key not exists, register a new one
func IncrEventLimiter(evkey string, dur time.Duration, limit int) {
	sched.limitMgr.Lock()
	defer sched.limitMgr.Unlock()

	l := sched.limitMgr.m[evkey]
	if l == nil { // if previous not exists, put a new event limiter
		l = rate.NewLimiter(dur, limit)
		sched.limitMgr.m[evkey] = l
	}

	l.Take() // take one token
}

// ClearEventLimiter remove the given event key limiter
func ClearEventLimiter(evkey string) {
	sched.limitMgr.Lock()
	delete(sched.limitMgr.m, evkey)
	sched.limitMgr.Unlock()
}
