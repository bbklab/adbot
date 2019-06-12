package scheduler

import "sync"

// routineMgr is an ongoing goroutines temporarily registry manager
//
type routineMgr struct {
	sync.RWMutex
	m map[string][]string // routine type -> routine names
}

func newRoutineMgr() *routineMgr {
	return &routineMgr{
		m: make(map[string][]string),
	}
}

// index get the index of given type/name goroutine
// note: unsafe, must be called under the protection of mutex
func (m *routineMgr) index(typ, name string) int {
	gs, ok := m.m[typ]
	if !ok {
		return -1
	}

	for i, g := range gs {
		if g == name {
			return i
		}
	}

	return -1
}

// AllGoroutines show all of registered routines
func AllGoroutines() map[string][]string {
	sched.routineMgr.RLock()
	defer sched.routineMgr.RUnlock()
	return sched.routineMgr.m
}

// Goroutines get given type of routines list
func Goroutines(typ string) []string {
	sched.routineMgr.RLock()
	defer sched.routineMgr.RUnlock()
	return sched.routineMgr.m[typ]
}

// IsRegisteredGoRoutine check if given type/name goroutine registered
func IsRegisteredGoRoutine(typ, name string) bool {
	sched.routineMgr.RLock()
	defer sched.routineMgr.RUnlock()
	return sched.routineMgr.index(typ, name) >= 0
}

// RegisterGoroutine register a goroutine name for given type
func RegisterGoroutine(typ, name string) {
	sched.routineMgr.Lock()
	defer sched.routineMgr.Unlock()
	sched.routineMgr.m[typ] = append(sched.routineMgr.m[typ], name)
}

// DeRegisterGoroutine de-register a goroutine name for given type
func DeRegisterGoroutine(typ, name string) {
	sched.routineMgr.Lock()
	defer sched.routineMgr.Unlock()
	idx := sched.routineMgr.index(typ, name)
	if idx < 0 {
		return
	}
	sched.routineMgr.m[typ] = append(sched.routineMgr.m[typ][:idx], sched.routineMgr.m[typ][idx+1:]...)
}
