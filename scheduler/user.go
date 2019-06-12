package scheduler

import (
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/ua"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

// user sessions
//

// CleanExpiredUserSessionsLoop is exported
func CleanExpiredUserSessionsLoop() {
	RegisterGoroutine("user_sessions_cleaner", "system")
	defer DeRegisterGoroutine("user_sessions_cleaner", "system")

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	for range ticker.C {
		sesses, err := store.DB().ListUserSessions("")
		if err != nil {
			log.Warnln("user sessions cleaner list db sessions error:", err)
			continue
		}

		var n int
		for _, sess := range sesses {
			if !sess.LastActiveAt.Before(time.Now().Add(-time.Hour)) {
				continue
			}

			store.DB().RemoveUserSession(sess.ID)
			n++
		}

		if n > 0 {
			log.Infof("user sessions cleaner clean up %d expired sessions", n)
		}
	}
}

// RenewUserSession is exported
func RenewUserSession(userID, sessID string, req *http.Request) error {
	sess := genUserSession(userID, sessID, req)
	return store.DB().UpsertUserSession(sess)
}

// CreateUserSession create a new db user session while user login
func CreateUserSession(userID string, req *http.Request) (string, error) {
	sess := genUserSession(userID, "", req)
	return sess.ID, store.DB().UpsertUserSession(sess)
}

// KickoutUserAllSessions remove all sessions for given user
// mostly called while user change its password
func KickoutUserAllSessions(userID string) error {
	sesses, err := store.DB().ListUserSessions(userID)
	if err != nil {
		return err
	}

	for _, sess := range sesses {
		store.DB().RemoveUserSession(sess.ID)
	}
	return nil
}

func genUserSession(userID, sessID string, req *http.Request) *types.UserSession {
	var (
		dev, os, browser = ua.ParseUA(req)
		remoteIP, _, _   = net.SplitHostPort(req.RemoteAddr)
	)

	if sessID == "" { // gen a new ID if not provided
		sessID = utils.RandomString(16)
	}

	return &types.UserSession{
		ID:           sessID,
		UserID:       userID,
		Remote:       remoteIP,
		GeoInfo:      GetGeoInfoEn(remoteIP),
		GeoInfoZh:    GetGeoInfoZh(remoteIP),
		Device:       dev,
		OS:           os,
		Browser:      browser,
		LastActiveAt: time.Now(),
	}
}
