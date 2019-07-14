// package flock ...
// this file is most borrowed from https://github.com/gofrs/flock --> flock.go, flock_winapi.go, flock_windows.go
// +build windows

package flock

import (
	"os"
	"sync"
	"syscall"
	"unsafe"
)

var (
	kernel32, _         = syscall.LoadLibrary("kernel32.dll")
	procLockFileEx, _   = syscall.GetProcAddress(kernel32, "LockFileEx")
	procUnlockFileEx, _ = syscall.GetProcAddress(kernel32, "UnlockFileEx")
)

const (
	winLockfileExclusiveLock = 0x00000002
)

// Flock is the struct type to handle file locking. All fields are unexported,
// with access to some of the fields provided by getter methods (Path() and Locked()).
type Flock struct {
	path string
	m    sync.RWMutex
	fh   *os.File
	l    bool
	r    bool
}

func newFlock(path string) (*Flock, error) {
	return &Flock{path: path}, nil
}

// Close is equivalent to calling Unlock.
//
// This will release the lock and close the underlying file descriptor.
// It will not remove the file from disk, that's up to your application.
func (f *Flock) Close() error {
	return f.Unlock()
}

// ErrorLockViolation is the error code returned from the Windows syscall when a
// lock would block and you ask to fail immediately.
const ErrorLockViolation syscall.Errno = 0x21 // 33

// Lock is a blocking call to try and take an exclusive file lock. It will wait
// until it is able to obtain the exclusive file lock. It's recommended that
// TryLock() be used over this function. This function may block the ability to
// query the current Locked() or RLocked() status due to a RW-mutex lock.
//
// If we are already locked, this function short-circuits and returns
// immediately assuming it can take the mutex lock.
func (f *Flock) Lock() error {
	return f.lock(&f.l, winLockfileExclusiveLock)
}

func (f *Flock) lock(locked *bool, flag uint32) error {
	f.m.Lock()
	defer f.m.Unlock()

	if *locked {
		return nil
	}

	if f.fh == nil {
		if err := f.setFh(); err != nil {
			return err
		}
	}

	if _, errNo := lockFileEx(syscall.Handle(f.fh.Fd()), flag, 0, 1, 0, &syscall.Overlapped{}); errNo > 0 {
		return errNo
	}

	*locked = true
	return nil
}

// Unlock is a function to unlock the file. This file takes a RW-mutex lock, so
// while it is running the Locked() and RLocked() functions will be blocked.
//
// This function short-circuits if we are unlocked already. If not, it calls
// UnlockFileEx() on the file and closes the file descriptor. It does not remove
// the file from disk. It's up to your application to do.
func (f *Flock) Unlock() error {
	f.m.Lock()
	defer f.m.Unlock()

	// if we aren't locked or if the lockfile instance is nil
	// just return a nil error because we are unlocked
	if (!f.l && !f.r) || f.fh == nil {
		return nil
	}

	// mark the file as unlocked
	if _, errNo := unlockFileEx(syscall.Handle(f.fh.Fd()), 0, 1, 0, &syscall.Overlapped{}); errNo > 0 {
		return errNo
	}

	f.fh.Close()

	f.l = false
	f.r = false
	f.fh = nil

	return nil
}

func (f *Flock) setFh() error {
	// open a new os.File instance
	// create it if it doesn't exist, and open the file read-only.
	fh, err := os.OpenFile(f.path, os.O_CREATE|os.O_RDONLY, os.FileMode(0600))
	if err != nil {
		return err
	}

	// set the filehandle on the struct
	f.fh = fh
	return nil
}

// Use of 0x00000000 for the shared lock is a guess based on some the MS Windows
// `LockFileEX` docs, which document the `LOCKFILE_EXCLUSIVE_LOCK` flag as:
//
// > The function requests an exclusive lock. Otherwise, it requests a shared
// > lock.
//
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365203(v=vs.85).aspx
func lockFileEx(handle syscall.Handle, flags uint32, reserved uint32, numberOfBytesToLockLow uint32, numberOfBytesToLockHigh uint32, offset *syscall.Overlapped) (bool, syscall.Errno) {
	r1, _, errNo := syscall.Syscall6(
		uintptr(procLockFileEx),
		6,
		uintptr(handle),
		uintptr(flags),
		uintptr(reserved),
		uintptr(numberOfBytesToLockLow),
		uintptr(numberOfBytesToLockHigh),
		uintptr(unsafe.Pointer(offset)))

	if r1 != 1 {
		if errNo == 0 {
			return false, syscall.EINVAL
		}

		return false, errNo
	}

	return true, 0
}

func unlockFileEx(handle syscall.Handle, reserved uint32, numberOfBytesToLockLow uint32, numberOfBytesToLockHigh uint32, offset *syscall.Overlapped) (bool, syscall.Errno) {
	r1, _, errNo := syscall.Syscall6(
		uintptr(procUnlockFileEx),
		5,
		uintptr(handle),
		uintptr(reserved),
		uintptr(numberOfBytesToLockLow),
		uintptr(numberOfBytesToLockHigh),
		uintptr(unsafe.Pointer(offset)),
		0)

	if r1 != 1 {
		if errNo == 0 {
			return false, syscall.EINVAL
		}

		return false, errNo
	}

	return true, 0
}
