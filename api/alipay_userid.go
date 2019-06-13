package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/qrcode"
)

var (
	alipayUserIDURL = "https://render.alipay.com/p/f/fd-ixpo7iia/index.html"
)

func (s *Server) getAlipayUserIDQrCode(ctx *httpmux.Context) {
	qrpng, err := qrcode.Encode(alipayUserIDURL)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Res.Header().Set("Content-Type", "image/png")
	ctx.Res.Write(qrpng)
}
