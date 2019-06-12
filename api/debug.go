package api

import (
	"runtime/pprof"

	"github.com/bbklab/adbot/debug"
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
)

// dump debug informations
func (s *Server) debugDump(ctx *httpmux.Context) {
	var (
		name = ctx.Query["name"]
	)

	switch name {

	case "goroutine": // dump runtime goroutines stack only
		pprof.Lookup("goroutine").WriteTo(ctx.Res, 2)

	case "general": // dump nb of fds, goroutines, memory, uptime etc
		ctx.JSON(200, debug.NewDebugInfo())

	case "config": // dump runtime master configs
		ctx.JSON(200, s.cfg)

	case "application": // dump more application related infos

		var (
			pingCost, pingErr = scheduler.DBPing()
			pingErrMsg        string
		)
		if pingErr != nil {
			pingErrMsg = pingErr.Error()
		}

		ctx.JSON(200, map[string]interface{}{
			"db_store_latency": map[string]interface{}{
				"error": pingErrMsg,
				"cost":  pingCost.String(),
			},
			"joined_nodes": scheduler.Nodes(),
			"limiters":     scheduler.ListEventLimiters(),
			"routes":       scheduler.AllGoroutines(),
		})

	default:
		ctx.BadRequest("unsupported dump name")
		return
	}
}
