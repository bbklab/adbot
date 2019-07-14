package agent

import (
	"github.com/bbklab/adbot/pkg/httpmux"
)

func (agent *Agent) terminal(ctx *httpmux.Context) {
	ctx.BadRequest("node terminal not supported under windows")
	return
}
