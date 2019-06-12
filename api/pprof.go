package api

import (
	"net/http"
	"net/http/pprof"

	"github.com/bbklab/adbot/pkg/httpmux"
)

func (s *Server) adaptHandlerFunc(handler http.HandlerFunc) func(ctx *httpmux.Context) {
	return func(ctx *httpmux.Context) {
		handler(ctx.Res, ctx.Req)
	}
}

// the parameter name could be:
//  goroutine    - stack traces of all current goroutines
//  heap         - a sampling of memory allocations of live objects
//  allocs       - a sampling of all past memory allocations
//  threadcreate - stack traces that led to the creation of new OS threads
//  block        - stack traces that led to blocking on synchronization primitives
//  mutex        - stack traces of holders of contended mutexes
//
// See:
//  https://godoc.org/runtime/pprof#Profile
func (s *Server) handlePProfName(ctx *httpmux.Context) {
	var profileName = ctx.Path["pname"]
	pprof.Handler(profileName).ServeHTTP(ctx.Res, ctx.Req)
}
