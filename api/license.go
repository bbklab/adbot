package api

import (
	"io/ioutil"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/scheduler"
)

func (s *Server) upsertLicense(ctx *httpmux.Context) {
	licbs, err := ioutil.ReadAll(ctx.Req.Body)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	err = scheduler.RenewLicense(string(licbs))
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

func (s *Server) licenseInfo(ctx *httpmux.Context) {
	lic := scheduler.RuntimeLicense()
	ctx.JSON(200, lic)
}

func (s *Server) rmLicense(ctx *httpmux.Context) {
	err := scheduler.RemoveLicense()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(204)
}
