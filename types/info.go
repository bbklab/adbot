package types

import (
	"io"

	"github.com/bbklab/adbot/pkg/template"
)

var summaryInfoTemplate = ` Version:       {{.Version}}
 Listens:       {{.Listens}}
 Uptime:        {{.Uptime}}
 Role:          {{if eq .Role "leader"}}{{green .Role}}{{else}}{{cyan .Role}}{{end}}
 Store:         {{.StoreTyp}}
 Nodes:         {{range $key, $val := .NumNodes}}{{$key}}:{{$val}} {{end}}
 BlockedNodes:  {{.NumBlockedNodes}}
`

// SummaryInfo is exported
type SummaryInfo struct {
	Version         string   `json:"version"`
	Listens         []string `json:"listens"`
	Uptime          string   `json:"uptime"`
	Role            Role     `json:"role"` // leader, candidate ...
	StoreTyp        string   `json:"store_type"`
	NumNodes        NodeInfo `json:"num_nodes"`
	NumBlockedNodes int      `json:"num_blocked_nodes"`
}

// Role is exported
type Role string

var (
	// RoleLeader is exported
	RoleLeader = Role("leader")
	// RoleCandidate is exported
	RoleCandidate = Role("candidate")
)

// NodeInfo is exported
type NodeInfo map[string]int // online|offline -> num

// WriteTo is exported
func (info *SummaryInfo) WriteTo(w io.Writer) (int64, error) {
	parser, err := template.NewParser(summaryInfoTemplate)
	if err != nil {
		return -1, err
	}
	return -1, parser.Execute(w, info) // just make pass govet
}
