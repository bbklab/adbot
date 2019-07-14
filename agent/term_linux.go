package agent

import (
	"syscall"
	"unsafe"

	"github.com/docker/docker/pkg/term"
)

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
