package flock

import (
	"os"
	"syscall"
)

// Flock wraps os.File to be used as a lock using syscall.Flock under linux
type Flock struct {
	f *os.File
}

func newFlock(path string) (*Flock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &Flock{f}, nil
}

// Close closes underlying file
func (l *Flock) Close() error {
	return l.f.Close()
}

// Lock acquires an exclusive lock
func (l *Flock) Lock() error {
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_EX)
}

// Unlock releases the lock
func (l *Flock) Unlock() error {
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}
