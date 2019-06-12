package agent

import (
	"github.com/bbklab/adbot/pkg/httpmux"
)

func (agent *Agent) setupRoutes(mux *httpmux.Mux) {
	// ping -> pong
	mux.GET("", agent.ping)
	mux.GET("/", agent.ping)
	mux.GET("/ping", agent.ping)

	// version
	mux.GET("/version", agent.version)

	// purge this node resources
	mux.DELETE("/purge", agent.purgeAll)

	// collect node sysinfo
	mux.GET("/sysinfo", agent.sysinfo)
	mux.GET("/stats", agent.stats) // live stream of sysinfo

	// exec node command
	mux.POST("/exec", agent.runCmd)
	// set os hostname
	mux.PUT("/hostname", agent.setHostname)

	// node terminal
	mux.GET("/terminal", agent.terminal)
	mux.HEAD("/terminal", agent.terminalQuery)
	mux.PATCH("/terminal", agent.terminalResize)

	// adb bot
	mux.GET("/adbot/devices", agent.listAdbDevices)
}
