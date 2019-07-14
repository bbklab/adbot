package cmd

import (
	"fmt"
	"io"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

// RunCmdProgress run shell command background and return a stream progress io reader
// if markErrEnd, the final progress stream message will be prefixed with ErrEndFlagPrefix
func RunCmdProgress(envs map[string]string, markErrEnd bool, stopNotify chan struct{}, cmd string, args ...string) (io.ReadCloser, error) {
	var (
		pipeReader, pipeWriter = io.Pipe()
	)

	command := exec.Command(cmd, args...)
	for key, val := range envs {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, val))
	}

	// attach command stderr/stdout to pipe writer
	// when the pipeReader was closed, the command will exit as we expect
	// with error: `signal: broken pipe`
	command.Stdout = pipeWriter
	command.Stderr = pipeWriter

	if stopNotify != nil {
		go func() {
			<-stopNotify
			log.Warnln("Stopping exec ", cmd, args, "...")
			StopCmd(command)
		}()
	}

	go func() {
		err := command.Run()
		if err != nil { // signal: terminated | signal: broken pipe ...
			var errEndMsg = err.Error() + "\r\n"
			if markErrEnd {
				errEndMsg = "\r\n" + ErrEndFlagPrefix + errEndMsg
			}
			pipeWriter.Write([]byte(errEndMsg))
		}

		pipeWriter.Close() // must: then the PipeReader got EOF
	}()

	// when the pipeReader was closed, the above command.Run() will return with error: `signal: broken pipe`
	// NOTE: if the command (eg sleep 10000) do NOT cares about Stdout, Stderr, close the pipeReader won't be effective
	// on this case, the caller should use stopNotify to stop the command
	return io.ReadCloser(pipeReader), nil
}
