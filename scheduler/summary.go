package scheduler

import (
	"time"

	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
	"github.com/bbklab/adbot/version"
)

// SummaryInfo is exported
func SummaryInfo() (*types.SummaryInfo, error) {
	info := &types.SummaryInfo{
		Version:  version.GetVersion() + "-" + version.GetGitCommit(),
		Uptime:   time.Since(sched.startAt).String(),
		StoreTyp: store.DB().Type(),
		Listens:  make([]string, 0, 0),
	}

	if isLeader() {
		info.Role = types.RoleLeader
	} else {
		info.Role = types.RoleCandidate
	}

	nodes, err := store.DB().ListNodes(nil)
	if err != nil {
		return nil, err
	}
	info.NumNodes = make(map[string]int)
	for _, node := range nodes {
		info.NumNodes[node.Status]++
	}

	return info, nil
}
