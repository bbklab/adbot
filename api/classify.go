package api

import (
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/utils"
)

// handlers category
var (
	catePublic             = "public"
	cateNonForward         = "non-forward"
	cateLicenseFree        = "license-free"
	cateLicenseExpiredDeny = "license-expired-deny"
)

// classify the http handlers once on startup
// mainly be used by http midwares
func (s *Server) classify() {
	s.classifyOnce.Do(func() {
		s.classified = map[string][]httpmux.HandleFunc{
			catePublic: {
				s.ping,               // ping pong (for node join)
				s.queryLeader,        // query current leader (for node join)
				s.checkNodeJoin,      // node join chec  (for node join)k
				s.version,            // query version
				s.anyUser,            // query if any user
				s.addUser,            // add first admin user
				s.userAuthLogin,      // user login
				s.payGateNewAdbOrder, // adb paygate: new ordek (protected by secret header)
				s.receiveAdbEvents,   // used by adb nodes to report adb device events
			},
			cateNonForward: {
				s.ping,          // ping pong (for node join)
				s.queryLeader,   // client detect leader (for node join)
				s.checkNodeJoin, // node join chec  (for node join)k
				s.debugDump,     // for trouble shooting
				s.info,          // for quering each member info's Role, and s.info is safe readonly on store
				s.version,       // query each member version
			},
			cateLicenseFree: {
				s.ping,          // ping pong (for node join)
				s.queryLeader,   // client detect leader (for node join)
				s.checkNodeJoin, // node join chec  (for node join)k
				s.version,       // query version
				s.debugDump,     // for trouble shooting
				s.anyUser,       // query if any user
				s.addUser,       // add first admin user
				s.userProfile,   // user profile
				s.userAuthLogin, // user login
				s.upsertLicense, // user upsert license
				s.licenseInfo,   // show license info
				s.rmLicense,     // remove license
			},
			cateLicenseExpiredDeny: { // license expired deny
				s.payGateNewAdbOrder, // mostly create new objects handlers
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
func (s *Server) isLicenseFreeHandler(hfs []httpmux.HandleFunc) bool {
	return s.containsInCategory(cateLicenseFree, hfs)
}
func (s *Server) isLicenseExpiredDenyHandler(hfs []httpmux.HandleFunc) bool {
	return s.containsInCategory(cateLicenseExpiredDeny, hfs)
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
