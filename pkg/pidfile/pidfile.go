package pidfile

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

var (
	// ErrConflictPidFile represents activity pid file conflicts
	ErrConflictPidFile = errors.New("conflicts, activity pid file, ensure previous process isn't running")
)

// New creates a pidfile at given path
func New(path string) error {

	// ensure previous pidfile not exists or pid not running
	if err := EnsurePreviousPidNotRunning(path); err != nil {
		return err
	}

	// create necessary parent directory
	basedir := filepath.Dir(path)
	if _, err := os.Stat(basedir); os.IsNotExist(err) {
		err = os.MkdirAll(basedir, os.FileMode(0755))
		if err != nil {
			return err
		}
	}

	// directly overwrite new pidfile to the given path
	return ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// EnsurePreviousPidNotRunning check if one specified pid file contains a running process pid
func EnsurePreviousPidNotRunning(path string) error {
	pidbs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	_, err = strconv.Atoi(string(pidbs))
	if err != nil {
		return nil
	}

	if _, err := os.Stat(filepath.Join("/proc", string(pidbs))); err != nil {
		return nil
	}

	return ErrConflictPidFile
}
