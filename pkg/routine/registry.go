package routine

import "sync"

// Registry is an ongoing goroutines temporarily registry manager
//
type Registry struct {
	sync.RWMutex
	m map[string][]string // routine type -> routine names
}

// NewRegistry  new a routine registry
func NewRegistry() *Registry {
	return &Registry{
		m: make(map[string][]string),
	}
}

// index get the index of given type/name goroutine
// note: unsafe, must be called under the protection of mutex
func (r *Registry) index(typ, name string) int {
	gs, ok := r.m[typ]
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

// All get the whole map of routines
func (r *Registry) All() map[string][]string {
	r.RLock()
	defer r.RUnlock()
	return r.m
}

// GetType get given type of routines list
func (r *Registry) GetType(typ string) []string {
	r.RLock()
	defer r.RUnlock()
	return r.m[typ]
}

// ExistsRoutine check if a routine name & type registered
func (r *Registry) ExistsRoutine(typ, name string) bool {
	r.RLock()
	defer r.RUnlock()
	return r.index(typ, name) >= 0
}

// AddRoutine register a routine name for given type
func (r *Registry) AddRoutine(typ, name string) {
	r.Lock()
	defer r.Unlock()
	r.m[typ] = append(r.m[typ], name)
}

// DelRoutine de-register a goroutine name for given type
func (r *Registry) DelRoutine(typ, name string) {
	r.Lock()
	defer r.Unlock()
	idx := r.index(typ, name)
	if idx < 0 {
		return
	}
	r.m[typ] = append(r.m[typ][:idx], r.m[typ][idx+1:]...)
}
