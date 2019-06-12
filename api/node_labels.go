package api

import (
	"strconv"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
)

func (s *Server) upsertNodeLabels(ctx *httpmux.Context) {
	var (
		id = ctx.Path["node_id"]
	)

	var lbs label.Labels
	if err := ctx.Bind(&lbs); err != nil {
		ctx.BadRequest(err)
		return
	}

	if lbs.Len() == 0 {
		ctx.BadRequest("at least one label key-value required")
		return
	}

	err := scheduler.UpsertNodeLabels(id, lbs)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	node, _ := store.DB().GetNode(id)
	ctx.JSON(200, node.Labels)
}

func (s *Server) rmNodeLabels(ctx *httpmux.Context) {
	var (
		id     = ctx.Path["node_id"]
		all, _ = strconv.ParseBool(ctx.Query["all"])
		keys   []string
	)

	if !all {
		if err := ctx.Bind(&keys); err != nil {
			ctx.BadRequest(err)
			return
		}
		if len(keys) == 0 {
			ctx.BadRequest("at least one label key required")
			return
		}
	}

	err := scheduler.RemoveNodeLabels(id, all, keys)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	node, _ := store.DB().GetNode(id)
	ctx.JSON(200, node.Labels)
}
