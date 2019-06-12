package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
)

// summary informations
func (s *Server) info(ctx *httpmux.Context) {
	info, err := scheduler.SummaryInfo()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	for _, l := range s.ls {
		info.Listens = append(info.Listens, l.Addr().String())
	}

	ctx.JSON(200, info)
}
