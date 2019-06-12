package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/bbklab/adbot/i18n"
	"github.com/bbklab/adbot/pkg/geoip"
	"github.com/bbklab/adbot/pkg/validator"
)

// User is a db user
type User struct {
	ID          string    `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"` // uniq
	Password    Password  `json:"password" bson:""`
	Desc        string    `json:"desc" bson:"desc"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at" bson:"last_login_at"` // last auth login
}

// Valid is exported
func (u *User) Valid() error {
	if err := validator.String(u.Name, 2, 64, validator.NormalCharacters); err != nil {
		return fmt.Errorf("user name %v", err)
	}
	if u.Name == "any" {
		return errors.New("user name `any` is reserved")
	}
	if err := validator.String(u.Desc, -1, 1024, nil); err != nil {
		return fmt.Errorf("user desc %v", err)
	}
	return u.Password.Valid()
}

// Hidden set the field Password as invisible
func (u *User) Hidden() {
	u.Password = Password(SensitiveHolder)
}

// Password is a plain password user provided
type Password string

// Valid is exported
func (p Password) Valid() error {
	if err := validator.String(string(p), 4, 64, nil); err != nil {
		return fmt.Errorf(i18n.MsgPasswordInvalid) // fixme, show err detail messages
	}
	return nil
}

// Bytes is exported
func (p Password) Bytes() []byte {
	return []byte(string(p))
}

// ReqChangePassword is a request user provided to change user's password
type ReqChangePassword struct {
	Old Password `json:"old"`
	New Password `json:"new"`
}

// Valid is exported
func (req *ReqChangePassword) Valid() error {
	if err := req.Old.Valid(); err != nil {
		return err
	}
	if err := req.New.Valid(); err != nil {
		return err
	}
	if req.New == req.Old {
		return errors.New("new password should not be the same as original")
	}
	return nil
}

// ReqLoginWithCaptcha is a ReqLogin with Captcha
type ReqLoginWithCaptcha struct {
	ReqLogin
	CapKey   string `json:"cap_key"`
	CapValue string `json:"cap_value"`
}

// Valid is exported
func (req *ReqLoginWithCaptcha) Valid() error {
	if req.CapKey == "" || req.CapValue == "" {
		return fmt.Errorf(i18n.MsgCaptchaInvalid)
	}
	return req.ReqLogin.Valid()
}

// ReqLogin is a request user provided to login the system
type ReqLogin struct {
	UserName string   `json:"username"`
	Password Password `json:"password"`
}

// Valid is exporte
func (req *ReqLogin) Valid() error {
	if err := validator.String(req.UserName, 2, 64, nil); err != nil {
		return fmt.Errorf("user name %v", err)
	}
	return req.Password.Valid()
}

// UserSessionWrapper is exported
type UserSessionWrapper struct {
	*UserSession
	Current bool `json:"current"`
}

// UserSession is a db user session
type UserSession struct {
	ID           string         `json:"id" bson:"id"`           // session id
	UserID       string         `json:"user_id" bson:"user_id"` // ref: related user id
	Remote       string         `json:"remote" bson:"remote"`   // remote login address
	GeoInfo      *geoip.GeoInfo `json:"geoinfo" bson:"geoinfo"` // detected GEO info
	GeoInfoZh    *geoip.GeoInfo `json:"geoinfo_zh" bson:"geoinfo_zh"`
	Device       string         `json:"device" bson:"device"` // detected UA info
	OS           string         `json:"os" bson:"os"`
	Browser      string         `json:"browser" bson:"browser"`
	LastActiveAt time.Time      `json:"last_active_at" bson:"last_active_at"`
}
