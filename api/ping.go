package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
)

func (s *Server) ping(ctx *httpmux.Context) {
	ctx.Res.Write([]byte{'O', 'K'})
}
