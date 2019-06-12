package utils

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

var (
	sysProc = "/proc" // linux only
)

// NumFd return current process's fd count
func NumFd() int {
	return NumFdOf(os.Getpid())
}

// NumFdOf return given process's fd count
func NumFdOf(pid int) int {
	files, _ := ioutil.ReadDir(path.Join(sysProc, strconv.Itoa(pid), "fd"))
	return len(files)
}
