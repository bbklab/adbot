package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
)

func (s *Server) telegramBotStatus(ctx *httpmux.Context) {
	status, err := scheduler.TGBotStatus()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, status)
}
