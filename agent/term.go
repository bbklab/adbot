package agent

import (
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/docker/docker/pkg/term"
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

func (t *termFDStore) resize(id string, width, height int) error {
	t.Lock()
	defer t.Unlock()

	fd, ok := t.m[id]
	if !ok {
		return nil
	}

	ws := &term.Winsize{Height: uint16(height), Width: uint16(width)}
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)

	if err == 0 {
		return nil
	}
	return err
}
