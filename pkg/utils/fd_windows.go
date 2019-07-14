package utils

import (
	"os"
)

// NumFd return current process's fd count
func NumFd() int {
	return NumFdOf(os.Getpid())
}

// NumFdOf return given process's fd count
func NumFdOf(pid int) int {
	return 0
}
