package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
)

func (s *Server) queryLeader(ctx *httpmux.Context) {
	if s.isLeader() {
		ctx.Text(200, "I'm the leader")
		return
	}

	ctx.Text(410, "TODO: ask ha.Campaigner.CurrentLeader() who is the current leader")
}
