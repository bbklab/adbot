package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
)

func (s *Server) showGeoMetadata(ctx *httpmux.Context) {
	ctx.JSON(200, scheduler.CurrentGeoMetaData())
}

func (s *Server) updateGeoData(ctx *httpmux.Context) {
	prev := scheduler.CurrentGeoMetaData()

	cost, err := scheduler.UpdateGeoData()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, map[string]interface{}{
		"previous": prev,
		"current":  scheduler.CurrentGeoMetaData(),
		"cost":     cost,
	})
}
