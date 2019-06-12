package api

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bbklab/adbot/pkg/httpmux"
)

var (
	metricsAuthUser     string
	metricsAuthPassword string
)

func (s *Server) exportMetrics(ctx *httpmux.Context) {
	user, password, ok := ctx.Req.BasicAuth()
	if !ok {
		ctx.Unauthorized("basic auth required")
		return
	}

	if user != metricsAuthUser || password != metricsAuthPassword {
		ctx.Unauthorized("basic auth failed")
		return
	}

	promhttp.Handler().ServeHTTP(ctx.Res, ctx.Req)
}
