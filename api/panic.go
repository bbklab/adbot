package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
)

var (
	panicToken = "7fe64e8c4d89a7a2d204c2f9df9ef5345d95d9fa"
)

func (s *Server) panic(ctx *httpmux.Context) {
	var (
		token = ctx.Req.Header.Get("Panic-Secret-Token")
	)

	if token != panicToken {
		ctx.Forbidden("go away")
		return
	}

	panic("the panic sucks")
}
