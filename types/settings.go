package types

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/label"
)

var (
	// SensitiveHolder represents some sensitive datas that should not exposed to public
	// eg: user password, ssh password, ssh private key etc
	SensitiveHolder = "******"
)

var (
	// GlobalDefaultSettings define the default settings
	// applied on the first startup or settings reset
	GlobalDefaultSettings = &Settings{
		LogLevel:           "info",
		EnableHTTPMuxDebug: false,
		UnmarkSensitive:    false,
		TGBotToken:         "",
		GlobalAttrs:        label.New(nil),
		UpdatedAt:          time.Time{},
		Initial:            true,
	}
)

// Settings is a db setting
type Settings struct {
	LogLevel           string       `json:"log_level" bson:"log_level"`                       // logrus log level
	EnableHTTPMuxDebug bool         `json:"enable_httpmux_debug" bson:"enable_httpmux_debug"` // enable httpmux debug or not
	UnmarkSensitive    bool         `json:"unmask_sensitive" bson:"unmask_sensitive"`         // uncover the sensitive fields, eg: ssh password, access key, etc
	TGBotToken         string       `json:"tg_bot_token" bson:"tg_bot_token"`                 // telegram bot token
	GlobalAttrs        label.Labels `json:"global_attrs" bson:"global_attrs"`                 // user customized kv, we just treat it as general label kv
	UpdatedAt          time.Time    `json:"updated_at" bson:"updated_at"`
	Initial            bool         `json:"initial" bson:"initial"`
}

// Hidden set the smtpd config sensitive fields as invisible
func (s *Settings) Hidden() {
	if s.TGBotToken != "" {
		s.TGBotToken = SensitiveHolder
	}
}

// UpdateSettingsReq is similar to types.Settings, but all changable fields are pointer type
type UpdateSettingsReq struct {
	LogLevel           *string `json:"log_level"`
	EnableHTTPMuxDebug *bool   `json:"enable_httpmux_debug"`
	UnmarkSensitive    *bool   `json:"unmask_sensitive"`
	TGBotToken         *string `json:"tg_bot_token"`
}

// Valid verify the UpdateSettingsReq
func (req *UpdateSettingsReq) Valid() error {
	if req.LogLevel != nil {
		if _, err := log.ParseLevel(*req.LogLevel); err != nil {
			return err
		}
	}
	if req.TGBotToken != nil {
		if *req.TGBotToken == "" {
			return fmt.Errorf("tg bot token required")
		}
	}
	return nil
}
