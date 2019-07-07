package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bbklab/adbot/i18n"
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/session"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
	lictypes "github.com/bbklab/adbot/types/lic"
)

func init() {
	// disable HTTP/2 server side support, because when `Chrome/Firefox` visit `https://`,
	// http.ResponseWriter is actually implemented by *http.http2responseWriter which
	// does NOT implemented http.Hijacker
	// See: https://github.com/golang/go/issues/14797
	os.Setenv("GODEBUG", "http2server=0")
}

// global midware to verify current license
//
//
func (s *Server) checkLicenseMW(ctx *httpmux.Context) {
	// skip this midware
	// if current request hit handlers is license free handlers
	handlers := ctx.MatchedHandlers()
	if s.isLicenseFreeHandler(handlers) {
		return
	}

	// if should ignore license verify, mainly for CI env
	if scheduler.IgnoreLicense() {
		return
	}

	// if no license provided, ask for a new license
	if scheduler.IsEmptyLicense() {
		ctx.PaymentRequired(lictypes.ErrLicenseNotProvided)
		ctx.Abort()
		return
	}

	var (
		lic = scheduler.RuntimeLicense()
	)

	// if license expired, no more new objects are allowed to be created
	if lic.IsExpired() && s.isLicenseExpiredDenyHandler(handlers) {
		ctx.PaymentRequired(lictypes.ErrLicenseExpired)
		ctx.Abort()
		return
	}

	// pass!
}

// global midware to add cors http headers
//
//
func (s *Server) corsMW(ctx *httpmux.Context) {
	orig := ctx.Req.Header.Get("Origin")
	if orig == "" {
		orig = "*"
	}

	ctx.Res.Header().Add("Access-Control-Allow-Origin", orig)
	ctx.Res.Header().Add("Access-Control-Allow-Credentials", "true")
	ctx.Res.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-Registry-Auth, Cache-Control, Authorization, Total-Records, Admin-Access-Token, OrderID, Fee, FeeYuan, Total-Fee-Yuan")
	ctx.Res.Header().Add("Access-Control-Allow-Methods", "HEAD, GET, POST, DELETE, PUT, PATCH, OPTIONS")
	ctx.Res.Header().Add("Access-Control-Expose-Headers", "Total-Records, Admin-Access-Token, OrderID, Fee, FeeYuan, Total-Fee-Yuan")
}

// global midware to bypass and end some special http requests
//
//
func (s *Server) bypassMW(ctx *httpmux.Context) {
	if ctx.Req.Method == "OPTIONS" {
		ctx.Res.Header().Add("Allow", "HEAD, GET, POST, DELETE, PUT, PATCH, OPTIONS")
		ctx.Abort()
	}
}

// global midware for audit logging
//
// Note: this midware is used as httpmux.AuditLog handler which is
// called within the defer of httpmux.ServeHTTP, so this handler
// won't be called by any ctx.Abort() invoked by any previous midwares
// or http handlers.
func (s *Server) auditMW(ctx *httpmux.Context) {
	var (
		verb       = strings.ToUpper(ctx.Req.Method)
		uri        = ctx.Req.URL.Path
		cost       = fmt.Sprintf("%0.4fs", time.Now().Sub(ctx.StartAt()).Seconds())
		code       = ctx.Res.(*httpmux.Response).StatusCode() // note: 0 if Hijack-ed
		size       = ctx.Res.(*httpmux.Response).Size()       // note: 0 if Hijack-ed
		errmsg     = ctx.Res.(*httpmux.Response).ErrMsg()
		sourceIP   string
		verbStatus = types.VerbStatusUnkn
	)

	if fields := strings.SplitN(ctx.Req.RemoteAddr, ":", 2); len(fields) == 2 {
		sourceIP = fields[0]
	}

	switch {
	case code >= 400:
		verbStatus = types.VerbStatusFail // 4xx, 5xx
	case code >= 200 && code <= 300:
		verbStatus = types.VerbStatusSucc // 2xx
	default:
		verbStatus = types.VerbStatusUnkn // 3xx
	}

	entry := &types.AuditEntry{
		Verb:           verb,
		VerbStatus:     verbStatus,
		RequestURI:     uri,
		Source:         sourceIP,
		ResponseCode:   code,
		ResponseSize:   size,
		Cost:           cost,
		Time:           time.Now(),
		Annotations:    nil,
		ResponseErrMsg: errmsg,
	}

	// -> audit file log
	scheduler.LogAuditEntry(entry)

	// -> stdout
	log.Println("HTTP", code, verb, sourceIP, uri, size, cost)
}

// global midware to verify the `system admin` auth login by checking access token
//
//
// note: this is a midware handler to verify the request is with
// legal `system admin` access token, if the check passed through, this handler
// will put two Keys `USER_ID` `USER_SESSID` to the current context and
// passed to the following handlers in the chain
func (s *Server) checkAuthLoginMW(ctx *httpmux.Context) {
	var (
		i18np = i18nPrinter(ctx)
	)

	// skip websocket handshake
	// as we can't obtain the request Scheme from server side http.Request.URL which is parsed from RequestURI
	// so we check http Header: `Upgrade: websocket` to detect the websocket handshake
	if ctx.Req.Header.Get("Upgrade") == "websocket" {
		return
	}

	// skip this auth midware
	// if current request hit handlers contains public handlers
	handlers := ctx.MatchedHandlers()
	if s.isPublicHandler(handlers) {
		return
	}

	// no such route
	if len(handlers) == 0 {
		ctx.NotFound("no such route")
		ctx.Abort()
		return
	}

	// get the encoded token
	var (
		token             = ctx.Req.Header.Get("Admin-Access-Token")
		clientIP          = ctx.ClientIP()
		limitEvKeyHour    = fmt.Sprintf("admin:bad_token_per_hour:%s", clientIP) // bad token: client ip
		limitEvKeyDay     = fmt.Sprintf("admin:bad_token_per_day:%s", clientIP)  // bad token: client ip
		incrEvLimiterFunc = func() {
			scheduler.IncrEventLimiter(limitEvKeyHour, time.Hour, 300)   // bad token, per_hour limiter +1
			scheduler.IncrEventLimiter(limitEvKeyDay, time.Hour*24, 900) // bad token, per day limiter +1
		}
	)

	// first check the limiter
	if err := scheduler.CheckEventLimiter(limitEvKeyHour); err != nil {
		ctx.TooManyRequests(i18np.Sprintf(err.Error())) // rate limited
		ctx.Abort()
		return
	}
	if err := scheduler.CheckEventLimiter(limitEvKeyDay); err != nil {
		ctx.TooManyRequests(i18np.Sprintf(err.Error())) // rate limited
		ctx.Abort()
		return
	}

	if token == "" {
		token = ctx.Query["Admin-Access-Token"]
	}

	if token == "" {
		incrEvLimiterFunc()
		ctx.Unauthorized(i18np.Sprintf(i18n.MsgUserLoginRequired)) // 401: without token (not auth login yet)
		ctx.Abort()
		return
	}

	// decode the token and get the session to check the corresponding user exists
	combinedSessID, err := session.Decode(token, "USER_SESSION_ID")
	if err != nil {
		incrEvLimiterFunc()
		ctx.Unauthorized(i18np.Sprintf(i18n.MsgUnRecognizedUserToken)) // 401: invalid token
		ctx.Abort()
		return
	}

	fields := strings.SplitN(combinedSessID, "/", 2)
	userID, sessID := fields[0], fields[1]

	if userID == "" || sessID == "" {
		incrEvLimiterFunc()
		ctx.Unauthorized(i18np.Sprintf(i18n.MsgUnCompleteUserToken)) // 401: uncomplete token
		ctx.Abort()
		return
	}

	// verify the user
	user, err := store.DB().GetUser(userID)
	if err != nil {
		incrEvLimiterFunc()
		ctx.Unauthorized(i18np.Sprintf(i18n.MsgUserNotExists)) // 401: user not exists
		ctx.Abort()
		return
	}

	// verify the user session
	_, err = store.DB().GetUserSession(sessID)
	if err != nil {
		incrEvLimiterFunc()
		ctx.Unauthorized(i18np.Sprintf(i18n.MsgExpiredUserToken)) // 401: user session not exists, maybe cleaned up
		ctx.Abort()
		return
	}

	// passed by!
	scheduler.ClearEventLimiter(limitEvKeyHour) // succeed, remove the limiter
	scheduler.ClearEventLimiter(limitEvKeyDay)  // succeed, remove the limiter

	// renew the session
	scheduler.RenewUserSession(userID, sessID, ctx.Req)

	// fine! set the key
	ctx.SetKey("USER_ID", user.ID)
	ctx.SetKey("USER_SESSID", sessID)
}

//
// change db settings handlers
//

func (s *Server) isChangeSettings(hfs []httpmux.HandleFunc) bool {
	var (
		funcs = []httpmux.HandleFunc{
			s.updateSettings,
			s.resetSettings,
		}
	)

	for _, fun := range funcs {
		for _, hf := range hfs {
			if utils.FuncName(fun) == utils.FuncName(hf) {
				return true
			}
		}
	}
	return false
}
