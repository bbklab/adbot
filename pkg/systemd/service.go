package systemd

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	units "github.com/docker/go-units"

	"github.com/bbklab/adbot/pkg/cmd"
	"github.com/bbklab/adbot/pkg/utils"
)

// GetServiceStatus is exported
func GetServiceStatus(name string) (*ServiceStatus, error) {
	so, se, err := cmd.RunCmd(map[string]string{"LANG": "en_US"}, "sh", "-c", fmt.Sprintf("systemctl show %s", name))
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, se)
	}

	var (
		buf     = bytes.NewBufferString(string(so))
		scanner = bufio.NewScanner(buf)
		ret     = new(ServiceStatus)
	)

	// detect systemctl field
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ActiveState=") {
			ret.State = strings.TrimPrefix(line, "ActiveState=")
			continue
		}
		if strings.HasPrefix(line, "MainPID=") {
			ret.MainPID, _ = strconv.Atoi(strings.TrimPrefix(line, "MainPID="))
			continue
		}
		if strings.HasPrefix(line, "ActiveEnterTimestamp=") {
			ret.ActiveAt, _ = time.ParseInLocation("Mon 2006-01-02 15:04:05 MST", strings.TrimPrefix(line, "ActiveEnterTimestamp="), time.UTC)
			continue
		}
		if strings.HasPrefix(line, "InactiveEnterTimestamp=") {
			ret.DeadAt, _ = time.ParseInLocation("Mon 2006-01-02 15:04:05 MST", strings.TrimPrefix(line, "InactiveEnterTimestamp="), time.UTC)
			continue
		}
	}

	// detect running
	ret.Running = false
	psexec := new(sigar.ProcExe)
	if err = psexec.Get(ret.MainPID); err == nil {
		ret.Running = true
	}

	// detect fd count
	ret.NumFd = utils.NumFdOf(ret.MainPID)

	return ret, nil
}

// ServiceStatus is exported
type ServiceStatus struct {
	Running  bool      `json:"running"`
	NumFd    int       `json:"num_fd"`
	State    string    `json:"state"`     // ActiveState
	MainPID  int       `json:"main_pid"`  // MainPID
	ActiveAt time.Time `json:"active_at"` // ActiveEnterTimestamp
	DeadAt   time.Time `json:"dead_at"`   // InactiveEnterTimestamp
}

func (s *ServiceStatus) String() string {
	if s.Running {
		dur := units.HumanDuration(time.Since(s.ActiveAt))
		return fmt.Sprintf("running - %s, pid:%d, fd:%d (%s)", s.State, s.MainPID, s.NumFd, dur)
	}

	if !s.DeadAt.IsZero() {
		dur := units.HumanDuration(time.Since(s.DeadAt))
		return fmt.Sprintf("dead - %s (%s)", s.State, dur)
	}

	return fmt.Sprintf("dead - %s", s.State)
}
