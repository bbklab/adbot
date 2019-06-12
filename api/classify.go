package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/utils"
)

// handlers category
var (
	catePublic     = "public"
	cateNonForward = "non-forward"
)

// classify the http handlers once on startup
// mainly be used by http midwares
func (s *Server) classify() {
	s.classifyOnce.Do(func() {
		s.classified = map[string][]httpmux.HandleFunc{
			catePublic: {
				s.ping,          // ping pong (for node join)
				s.queryLeader,   // query current leader (for node join)
				s.version,       // query version
				s.anyUser,       // query if any user
				s.addUser,       // add first admin user
				s.userAuthLogin, // user login
				s.hookAdbOrder,  // adb hook (protected by secret header)
				s.newAdbOrder,   // adb paygate: new ordek & check order (protected by secret header)
				s.checkAdbOrder,
			},
			cateNonForward: {
				s.ping,        // ping pong (for node join)
				s.queryLeader, // client detect leader (for node join)
				s.debugDump,   // for trouble shooting
				s.info,        // for quering each member info's Role, and s.info is safe readonly on store
				s.version,     // query each member version
			},
		}
	})
}

func (s *Server) isPublicHandler(hfs []httpmux.HandleFunc) bool {
	return s.containsInCategory(catePublic, hfs)
}
func (s *Server) isNonForwardHandler(hfs []httpmux.HandleFunc) bool {
	return s.containsInCategory(cateNonForward, hfs)
}

func (s *Server) containsInCategory(category string, hfs []httpmux.HandleFunc) bool {
	var (
		funcs = s.classified[category]
	)
	if len(funcs) == 0 {
		return false
	}

	for _, fun := range funcs {
		for _, hf := range hfs {
			if utils.FuncName(fun) == utils.FuncName(hf) {
				return true
			}
		}
	}
	return false
}
