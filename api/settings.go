package api

import (
	"errors"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/qrcode"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

var (
	unmaskSensitive bool // unmask the sensitive fields when api response
)

func (s *Server) getSettings(ctx *httpmux.Context) {
	settings, err := store.DB().GetSettings()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if !unmaskSensitive {
		settings.Hidden()
	}

	ctx.JSON(200, settings)
}

func (s *Server) updateSettings(ctx *httpmux.Context) {
	var req = new(types.UpdateSettingsReq)
	if err := ctx.Bind(&req); err != nil {
		ctx.BadRequest(err)
		return
	}

	if err := req.Valid(); err != nil {
		ctx.BadRequest(err)
		return
	}

	// force verify the telegram bot token if provided
	if token := req.TGBotToken; token != nil {
		if err := scheduler.VerifyTGBotToken(*token); err != nil {
			ctx.AutoError(err)
			return
		}
	}

	err := scheduler.MemoSettingsSet(req)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	// runtime apply
	s.applyRuntimeSettings()
	current, _ := store.DB().GetSettings()
	if !unmaskSensitive {
		current.Hidden()
	}
	ctx.JSON(200, current)
}

func (s *Server) resetSettings(ctx *httpmux.Context) {
	err := scheduler.MemoSettings(types.GlobalDefaultSettings)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	// runtime apply
	s.applyRuntimeSettings()
	current, _ := store.DB().GetSettings()
	if !unmaskSensitive {
		current.Hidden()
	}
	ctx.JSON(200, current)
}

func (s *Server) genAdvertiseAddrQrCode(ctx *httpmux.Context) {
	settings, err := store.DB().GetSettings()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if settings.AdvertiseAddr == "" {
		ctx.InternalServerError("global advertise addr not set yet")
		return
	}

	png, err := qrcode.Encode(settings.AdvertiseAddr)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.Res.Header().Set("Content-Type", "image/png")
	ctx.Res.WriteHeader(200)
	ctx.Res.Write(png)
}

func (s *Server) setGlobalAttrs(ctx *httpmux.Context) {
	var attrs label.Labels
	if err := ctx.Bind(&attrs); err != nil {
		ctx.BadRequest(err)
		return
	}
	if attrs.Len() == 0 {
		ctx.BadRequest("at least one attr key-value required")
		return
	}

	err := scheduler.UpsertSettingsAttr(attrs)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	current, _ := store.DB().GetSettings()
	ctx.JSON(200, current.GlobalAttrs)
}

func (s *Server) rmGlobalAttrs(ctx *httpmux.Context) {
	var (
		all, _ = strconv.ParseBool(ctx.Query["all"])
		keys   []string
	)
	if !all {
		if err := ctx.Bind(&keys); err != nil {
			ctx.BadRequest(err)
			return
		}
		if len(keys) == 0 {
			ctx.BadRequest("at least one attr key required")
			return
		}
	}

	err := scheduler.RemoveSettingsAttr(all, keys)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	current, _ := store.DB().GetSettings()
	ctx.JSON(200, current.GlobalAttrs)
}

// get db settings & apply runtime
//
func (s *Server) applyRuntimeSettings() error {
	current, _ := store.DB().GetSettings()
	if current == nil {
		return errors.New("db global settings lost")
	}

	l, _ := log.ParseLevel(current.LogLevel)
	log.SetLevel(l)

	s.mux.SetDebug(current.EnableHTTPMuxDebug)

	unmaskSensitive = current.UnmarkSensitive

	scheduler.RenewTGBot(current.TGBotToken)

	return nil
}
