package agent

import (
	"os"
	"sync"
)

// terminal fds runtime store
//
//
type termFDStore struct {
	sync.Mutex
	m map[string]*os.File
}

var termFDs = termFDStore{m: make(map[string]*os.File)}

func (t *termFDStore) exists(id string) bool {
	t.Lock()
	defer t.Unlock()
	_, ok := t.m[id]
	return ok
}

func (t *termFDStore) add(id string, fd *os.File) {
	t.Lock()
	t.m[id] = fd
	t.Unlock()
}

func (t *termFDStore) del(id string) {
	t.Lock()
	defer t.Unlock()

	fd, ok := t.m[id]
	if !ok {
		return
	}

	delete(t.m, id)
	fd.Close()
}
