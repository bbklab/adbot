package version

import (
	"runtime"

	"github.com/bbklab/adbot/types"
)

var (
	version   string // set by build LD_FLAGS
	gitCommit string // set by build LD_FLAGS
	buildAt   string // set by build LD_FLAGS
)

// GetVersion is exported
func GetVersion() string {
	return version
}

// GetGitCommit is exported
func GetGitCommit() string {
	return gitCommit
}

// Version is exported
func Version() types.Version {
	return types.Version{
		Version:   version,
		GitCommit: gitCommit,
		BuildTime: buildAt,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}
