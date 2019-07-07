package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	lictypes "github.com/bbklab/adbot/types/lic"
)

func (s *Server) checkNodeJoin(ctx *httpmux.Context) {
	var (
		id  = ctx.Query["node_id"]
		lic = scheduler.RuntimeLicense()
		msg string
	)

	// query blocked or not
	node, _ := store.DB().GetBlockedNode(id)
	if node != nil {
		msg = "node has been blocked"
		goto DENY
	}

	// check license max nodes allowed
	if store.DB().CountNodes(nil) >= lic.MaxNodes {
		msg = lictypes.ErrLicenseNodesOverQuota.Error()
		goto DENY
	}

	// passed!
	ctx.Status(202) // pass: use http.StatusAccepted to identify passed through
	return

DENY:
	ctx.Text(403, msg) // deny: use http.StatusForbidden to identify deny join
	return
}
