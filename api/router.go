package api

import (
	"net/http/pprof"

	"github.com/bbklab/adbot/pkg/httpmux"
)

func (s *Server) setupRoutes(mux *httpmux.Mux) {
	mux.GET("", s.ping)
	mux.GET("/", s.ping)
	mux.GET("/ping", s.ping)
	mux.GET("/query_leader", s.queryLeader)
	mux.GET("/version", s.version)
	// summary info
	mux.GET("/info", s.info)

	// panic handler (mainly for integration test)
	// note: only allow to visit with a secret token
	mux.GET("/panic", s.panic)

	// get telegram bot status
	mux.GET("/tgbot", s.telegramBotStatus)

	// prometheus metrics exporter
	mux.GET("/metrics", s.exportMetrics)

	// debug
	mux.GET("/debug/dump", s.debugDump)

	// profiling
	// output the profiling datas that maybe scraped by `go tool pprof` or directly http request
	// See: https://github.com/moby/moby/pull/32453
	mux.GET("/pprof/cmdline", s.adaptHandlerFunc(pprof.Cmdline))
	mux.GET("/pprof/profile", s.adaptHandlerFunc(pprof.Profile))
	mux.GET("/pprof/symbol", s.adaptHandlerFunc(pprof.Symbol))
	mux.GET("/pprof/trace", s.adaptHandlerFunc(pprof.Trace))
	mux.GET("/pprof/:pname", s.handlePProfName) // See: https://godoc.org/runtime/pprof#Profile  - goroutine,heap,allocs,threadcreate,block,mutex

	// users
	mux.GET("/users/any", s.anyUser) // public
	mux.POST("/users", s.addUser)    // public
	mux.GET("/users", s.listUsers)
	mux.GET("/users/profile", s.userProfile)
	mux.GET("/users/sessions", s.userSessions)
	mux.DELETE("/users/sessions/:session_id", s.kickUserSession)
	mux.PATCH("/users/change_password", s.changeUserPassword)
	mux.POST("/users/login", s.userAuthLogin) //public, ip allow
	mux.DELETE("/users/logout", s.userLogout)

	// nodes
	mux.GET("/nodes", s.listNodes) // support filtered by labels
	mux.GET("/nodes/:node_id", s.getNode)
	mux.GET("/nodes/:node_id/events", s.watchNodeEvents)
	mux.GET("/nodes/:node_id/stats", s.watchNodeStats)
	mux.POST("/nodes/:node_id/exec", s.runNodeCmd)
	mux.DELETE("/nodes/:node_id/close", s.closeNode)
	mux.PUT("/nodes/:node_id/hostname", s.setNodeHostname)
	// node labels
	mux.PUT("/nodes/:node_id/labels", s.upsertNodeLabels)
	mux.DELETE("/nodes/:node_id/labels", s.rmNodeLabels)
	// nodes terminal
	mux.ANY("/nodes/:node_id/terminal", s.openNodeTerminal) // Legacy, only for cli node terminal
	mux.ANY("/nodes/:node_id/terminal_ng", s.openNodeTerminalNG)

	// geo
	mux.GET("/geo/metadata", s.showGeoMetadata)
	mux.PATCH("/geo/update", s.updateGeoData)

	// adb nodes
	mux.GET("/adb_nodes", s.listAdbNodes)
	mux.GET("/adb_nodes/:node_id", s.getAdbNode)
	// adb devices
	mux.GET("/adb_devices", s.listAdbDevices)
	mux.GET("/adb_devices/:device_id", s.getAdbDevice)
	mux.PUT("/adb_devices/:device_id/bill", s.setAdbDeviceBill)
	mux.PUT("/adb_devices/:device_id/amount", s.setAdbDeviceAmount)
	mux.PUT("/adb_devices/:device_id/weight", s.setAdbDeviceWeight)
	mux.PUT("/adb_devices/:device_id/alipay", s.bindAdbDeviceAlipay)
	mux.DELETE("/adb_devices/:device_id/alipay", s.revokeAdbDeviceAlipay)
	// mux.DELETE("/adb_devices/:device_id", s.rmAdbDevice) // TODO
	// adb orders
	mux.GET("/adb_orders", s.listAdbOrders)
	// adb order hook, receive adb order status hook
	//  - called by adb nodes (TODO how to identify & autheticated ?)
	mux.POST("/adb_orders/hook", s.hookAdbOrder)
	// adb paygate
	//  - called by out side pay system, only autheticated by secret header (TODO more authetications)
	mux.POST("/adb_paygate/new", s.newAdbOrder)
	mux.POST("/adb_paygate/check", s.checkAdbOrder)

	// settings
	mux.GET("/settings", s.getSettings)
	mux.PATCH("/settings", s.updateSettings)                             // update all setting fields except `Attrs`
	mux.PUT("/settings/reset", s.resetSettings)                          // reset all settings to initilial default values
	mux.GET("/settings/advertise_addr/qrcode", s.genAdvertiseAddrQrCode) // gen advertise address QrCode image
	// global attrs (similar to node labels)
	mux.PUT("/settings/attrs", s.setGlobalAttrs)
	mux.DELETE("/settings/attrs", s.rmGlobalAttrs)

	// register web ui terminal as global fallback
	// mux.SetNotFound(s.webui)  // TODO
}
