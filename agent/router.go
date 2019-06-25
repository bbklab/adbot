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

	// collect node sysinfo
	mux.GET("/sysinfo", agent.sysinfo)
	mux.GET("/stats", agent.stats) // live stream of sysinfo

	// exec node command
	mux.POST("/exec", agent.runCmd)

	// node terminal
	mux.GET("/terminal", agent.terminal)
	mux.HEAD("/terminal", agent.terminalQuery)
	mux.PATCH("/terminal", agent.terminalResize)

	// adb bot
	mux.GET("/adbot/devices", agent.listAdbDevices)
	mux.GET("/adbot/alipay_order", agent.checkAdbAlipayOrder)
	mux.GET("/adbot/device/screencap", agent.screenCapAdbDevice)
	mux.GET("/adbot/device/uinodes", agent.dumpAdbDeviceUINodes)
	mux.PATCH("/adbot/device/click", agent.clickAdbDevice)
	mux.PATCH("/adbot/device/reboot", agent.rebootAdbDevice)
	mux.POST("/adbot/device/exec", agent.runAdbDeviceCmd)
}
