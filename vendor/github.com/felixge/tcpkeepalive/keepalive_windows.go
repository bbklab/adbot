package tcpkeepalive

import (
	"os"
	"syscall"
)

func setIdle(fd int, secs int) error {
	return nil
}

func setCount(fd int, n int) error {
	return nil
}

func setInterval(fd int, secs int) error {
	return nil
}

func setNonblock(fd int) error {
	return os.NewSyscallError("setsockopt", syscall.SetNonblock(syscall.Handle(uintptr(fd)), true))
}
