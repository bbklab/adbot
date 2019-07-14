package cmd

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	// ErrEndFlagPrefix is an uniq prefix identify the final end errmsg
	ErrEndFlagPrefix = "86aed754032446b7f46e598f4840e3f9ed35cf82: "
)

// RunCmd execute given cmd & args and wait until finished
// and return the stdout/stderr text
func RunCmd(envs map[string]string, cmd string, args ...string) (string, string, error) {
	var (
		outbuf bytes.Buffer
		errbuf bytes.Buffer
	)

	command := exec.Command(cmd, args...)
	for key, val := range envs {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, val))
	}
	command.Stdout = &outbuf
	command.Stderr = &errbuf

	err := command.Run()
	return outbuf.String(), errbuf.String(), err

}

// RunCmdTimeout run shell command and wait for `maxWait` to terminate it if not finished.
// FIXME: if using sh -c "cmd", the timeout won't works, eg:
//    sh -c 'sleep 5s; echo done' with max 1s, this method still haning for 5s...
func RunCmdTimeout(envs map[string]string, maxWait time.Duration, cmd string, args ...string) (string, string, error) {
	var (
		outbuf bytes.Buffer
		errbuf bytes.Buffer
	)

	command := exec.Command(cmd, args...)
	for key, val := range envs {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, val))
	}
	command.Stdout = &outbuf
	command.Stderr = &errbuf

	timer := time.AfterFunc(maxWait, func() {
		log.Warnln("Response from exec ", cmd, args, " time out. Stopping process ...")
		StopCmd(command)
	})

	err := command.Run()
	timer.Stop()

	return outbuf.String(), errbuf.String(), err
}

// DetectExitCode is exported
func DetectExitCode(err error) (code int) {
	if err == nil {
		return 0
	}

	defer func() {
		if r := recover(); r != nil {
			code = int(math.MinInt32)
		}
	}()

	return err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus()
}

// StopCmd try best to stop a running process started by exec.Command
func StopCmd(cmd *exec.Cmd) error {
	proc := cmd.Process

	if proc == nil {
		return nil
	}

	err := proc.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	return proc.Kill()
}
