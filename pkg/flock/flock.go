package flock

import (
	"os"
	"syscall"
)

// FileLock wraps os.File to be used as a lock using syscall.Flock
type FileLock struct {
	f *os.File
}

// New opens file/dir at path and returns unlocked FileLock object
func New(path string) (*FileLock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &FileLock{f}, nil
}

// Close closes underlying file
func (l *FileLock) Close() error {
	return l.f.Close()
}

// Lock acquires an exclusive lock
func (l *FileLock) Lock() error {
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_EX)
}

// Unlock releases the lock
func (l *FileLock) Unlock() error {
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}
