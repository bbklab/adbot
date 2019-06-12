package extensions

import (
	"io/ioutil"
	"os"
)

var (
	fProc = "/proc/sys/kernel/hostname"
	fEtc  = "/etc/hostname"
)

// SetHostname is exported
func SetHostname(name string) error {
	var (
		content = append([]byte(name), '\n')
	)

	previous, err := os.Hostname()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fProc, content, os.FileMode(0644))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fEtc, content, os.FileMode(0644))
	if err != nil {
		ioutil.WriteFile(fProc, append([]byte(previous), '\n'), os.FileMode(0644)) // recover previous hostname
		return err
	}

	return nil
}
