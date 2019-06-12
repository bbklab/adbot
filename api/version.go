package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/version"
)

func (s *Server) version(ctx *httpmux.Context) {
	ctx.JSON(200, version.Version())
}
