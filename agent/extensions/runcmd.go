package extensions

import (
	"io"

	"github.com/bbklab/adbot/pkg/cmd"
)

// RunCmd is exported ...
func RunCmd(command string, stopCh chan struct{}) (io.ReadCloser, error) {
	envs := map[string]string{"TERM": "xterm"}
	return cmd.RunCmdProgress(envs, false, stopCh, "/bin/sh", "-c", command)
}
