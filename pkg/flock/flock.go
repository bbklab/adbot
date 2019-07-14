package flock

// FileLock is exported
type FileLock interface {
	Close() error
	Lock() error
	Unlock() error
}

// New new an FileLock instance
func New(path string) (FileLock, error) {
	return newFlock(path)
}
