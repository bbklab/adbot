package unixsock

import (
	"net"
	"os"
	"syscall"
)

// New initialize a local unix net.Listener at given path
func New(path string) (net.Listener, error) {
	err := syscall.Unlink(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, err
	}

	return l, nil
}
