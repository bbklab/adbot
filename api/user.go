package api

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/bbklab/adbot/i18n"
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/session"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

func (s *Server) listUsers(ctx *httpmux.Context) {
	users, err := store.DB().ListUsers(getPager(ctx))
	if err != nil {
		ctx.AutoError(err)
		return
	}

	for _, user := range users {
		user.Hidden()
	}

	ctx.JSON(200, users)
}

func (s *Server) addUser(ctx *httpmux.Context) {
	var (
		i18np = i18nPrinter(ctx)
	)

	// ensure we have only one super user
	has, err := s.hasUser()
	if err != nil {
		ctx.AutoError(err)
		return
	}
	if has {
		ctx.Forbidden(i18np.Sprintf(i18n.MsgAlreadyHasAdmin))
		return
	}

	// obtain new user
	var user = new(types.User)
	if err := ctx.Bind(user); err != nil {
		ctx.BadRequest(err)
		return
	}

	if err := user.Valid(); err != nil {
		ctx.BadRequest(err)
		return
	}

	// encrypt password
	passwd, err := bcrypt.GenerateFromPassword(user.Password.Bytes(), 11)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	user.ID = utils.RandomString(16)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Password = types.Password(passwd)

	if err := store.DB().AddUser(user); err != nil {
		ctx.AutoError(err)
		return
	}

	user.Hidden()
	ctx.JSON(201, user)
}

func (s *Server) anyUser(ctx *httpmux.Context) {
	has, err := s.hasUser()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if has {
		ctx.JSON(200, map[string]bool{"result": true})
		return
	}

	ctx.JSON(200, map[string]bool{"result": false})
}

func (s *Server) userProfile(ctx *httpmux.Context) {
	var (
		userID = ctx.GetKey("USER_ID").(string)
	)

	user, err := store.DB().GetUser(userID)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	user.Hidden()
	ctx.JSON(200, user)
}

func (s *Server) userSessions(ctx *httpmux.Context) {
	var (
		userID = ctx.GetKey("USER_ID").(string)
		sessID = ctx.GetKey("USER_SESSID").(string)
	)

	sesses, err := store.DB().ListUserSessions(userID)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.UserSessionWrapper, len(sesses))
	for idx, sess := range sesses {
		wrap := s.wrapUserSession(sess, sessID)
		wraps[idx] = wrap
	}

	n := store.DB().CountUserSessions(userID)
	ctx.Res.Header().Set("Total-Records", strconv.Itoa(n))
	ctx.JSON(200, wraps)
}

func (s *Server) kickUserSession(ctx *httpmux.Context) {
	var (
		userID     = ctx.GetKey("USER_ID").(string)
		sessID     = ctx.GetKey("USER_SESSID").(string)
		kickSessID = ctx.Path["session_id"]
		i18np      = i18nPrinter(ctx)
	)

	if kickSessID == "" {
		ctx.BadRequest(i18np.Sprintf(i18n.MsgParamRequired) + ": `session_id`")
		return
	}

	if kickSessID == sessID {
		ctx.Forbidden(i18np.Sprintf(i18n.MsgCantKickCurrentSession))
		return
	}

	sess, err := store.DB().GetUserSession(kickSessID)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	if sess.UserID != userID {
		ctx.Forbidden(i18np.Sprintf(i18n.MsgNotThisUserSession))
		return
	}

	store.DB().RemoveUserSession(kickSessID)

	ctx.Status(200)
}

func (s *Server) changeUserPassword(ctx *httpmux.Context) {
	var (
		userID = ctx.GetKey("USER_ID").(string)
		req    = new(types.ReqChangePassword)
		i18np  = i18nPrinter(ctx)
	)

	if err := ctx.Bind(req); err != nil {
		ctx.BadRequest(err)
		return
	}

	if err := req.Valid(); err != nil {
		ctx.BadRequest(i18np.Sprintf(err.Error()))
		return
	}

	// firstly verify the old password
	if err := s.userAuth(userID, req.Old); err != nil {
		ctx.Forbidden(i18np.Sprintf(err.Error()))
		return
	}

	// encrypt new password
	newPasswd, err := bcrypt.GenerateFromPassword(req.New.Bytes(), 11)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	// update db user password
	err = scheduler.MemoUserPassword(userID, types.Password(newPasswd))
	if err != nil {
		ctx.AutoError(err)
		return
	}

	// kick out all exists sessions for this user to forcely request relogin
	scheduler.KickoutUserAllSessions(userID)

	ctx.Status(200)
}

func (s *Server) userAuthLogin(ctx *httpmux.Context) {
	var (
		req   = new(types.ReqLogin)
		i18np = i18nPrinter(ctx)
	)

	if err := ctx.Bind(req); err != nil {
		ctx.BadRequest(err)
		return
	}

	if err := req.Valid(); err != nil {
		ctx.BadRequest(i18np.Sprintf(err.Error()))
		return
	}

	var (
		userName       = string(req.UserName)
		clientIP       = ctx.ClientIP()
		limitEvKeyHour = fmt.Sprintf("admin:login_fail_per_hour:%s:%s", userName, clientIP) // login fail: user + client ip
		limitEvKeyDay  = fmt.Sprintf("admin:login_fail_per_day:%s:%s", userName, clientIP)  // login fail: user + client ip
	)

	// first check the limiter
	if err := scheduler.CheckEventLimiter(limitEvKeyHour); err != nil {
		ctx.TooManyRequests(i18np.Sprintf(err.Error())) // rate limited
		return
	}
	if err := scheduler.CheckEventLimiter(limitEvKeyDay); err != nil {
		ctx.TooManyRequests(i18np.Sprintf(err.Error())) // rate limited
		return
	}

	// auth login check
	if err := s.userAuth(req.UserName, req.Password); err != nil {
		scheduler.IncrEventLimiter(limitEvKeyHour, time.Hour, 10)   // auth failed, per_hour limiter +1
		scheduler.IncrEventLimiter(limitEvKeyDay, time.Hour*24, 50) // auth failed, per day limiter +1
		ctx.Forbidden(i18np.Sprintf(err.Error()))
		return
	}
	scheduler.ClearEventLimiter(limitEvKeyHour) // auth succeed, remove the limiter
	scheduler.ClearEventLimiter(limitEvKeyDay)  // auth succeed, remove the limiter

	// add new user session for this user
	user, _ := store.DB().GetUser(userName)
	sessID, err := scheduler.CreateUserSession(user.ID, ctx.Req)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	// setup token
	// note: here we use securecookie to encode `user id + sess id` to generate the accesss token
	// and we can decode the original `user id + sess id` from the token that client provided
	token, err := session.Encode("USER_SESSION_ID", fmt.Sprintf("%s/%s", user.ID, sessID))
	if err != nil {
		ctx.AutoError(err)
		return
	}
	ctx.Res.Header().Set("Admin-Access-Token", token)

	// update user last login at
	scheduler.MemoUserLastLoginAt(user.ID)

	ctx.Status(202)
}

func (s *Server) userLogout(ctx *httpmux.Context) {
	var (
		sessID = ctx.GetKey("USER_SESSID").(string)
	)

	// destroy db user session
	store.DB().RemoveUserSession(sessID)

	// set the token empty, the client should clear the previous stored token
	ctx.Res.Header().Set("Admin-Access-Token", "")
	ctx.Status(204)
}

// utils
//
func (s *Server) hasUser() (bool, error) {
	users, err := store.DB().ListUsers(nil)
	return len(users) != 0, err
}

// note: we should use fixed error message instead of
// original error message which are vulnerable to be attacked
func (s *Server) userAuth(idOrName string, provided types.Password) error {
	var (
		errAuthFailed = errors.New(i18n.MsgAuthenticationFailed)
	)

	user, err := store.DB().GetUser(idOrName)
	if err != nil {
		if store.DB().ErrNotFound(err) {
			return errAuthFailed // use fixed message
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword(user.Password.Bytes(), provided.Bytes())
	if err != nil {
		return errAuthFailed // use fixed message
	}

	return nil
}

func (s *Server) wrapUserSession(sess *types.UserSession, currSessID string) *types.UserSessionWrapper {
	return &types.UserSessionWrapper{
		UserSession: sess,
		Current:     sess.ID == currSessID,
	}
}
