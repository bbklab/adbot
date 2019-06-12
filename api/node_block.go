package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/store"
)

// GET /api/nodes/blocked?node_id=xxxx
func (s *Server) queryNodesBlocked(ctx *httpmux.Context) {
	var (
		id = ctx.Query["node_id"]
	)

	// list all of blocked nodes
	if id == "" {
		nodes, err := store.DB().ListBlockedNodes(getPager(ctx))
		if err != nil {
			ctx.AutoError(err)
			return
		}

		ctx.JSON(200, nodes)
		return
	}

	// query specified node blocked or not
	node, _ := store.DB().GetBlockedNode(id)
	if node != nil {
		ctx.Status(403) // blocked: use http.StatusForbidden to identify blocked
		return
	}

	ctx.Status(202) // not blocked: use http.StatusAccepted to identify not blocked
}
