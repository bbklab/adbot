package types

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/pkg/validator"
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
		AdvertiseAddr:       "",
		LogLevel:            "info",
		EnableHTTPMuxDebug:  false,
		MetricsAuthUser:     "adbot",
		MetricsAuthPassword: "adbot",
		UnmarkSensitive:     false,
		TGBotToken:          "",
		GlobalAttrs:         label.New(nil),
		UpdatedAt:           time.Time{},
		Initial:             true,
	}
)

// Settings is a db setting
type Settings struct {
	AdvertiseAddr       string       `json:"advertise_addr" bson:"advertise_addr"`               // agent join addrs, multi splited by comma ','
	LogLevel            string       `json:"log_level" bson:"log_level"`                         // logrus log level
	EnableHTTPMuxDebug  bool         `json:"enable_httpmux_debug" bson:"enable_httpmux_debug"`   // enable httpmux debug or not
	MetricsAuthUser     string       `json:"metrics_auth_user" bson:"metrics_auth_user"`         // metrics basic auth user (for prometheus scrapers)
	MetricsAuthPassword string       `json:"metrics_auth_password" bson:"metrics_auth_password"` // metrics basic auth password (for prometheus scrapers)
	UnmarkSensitive     bool         `json:"unmask_sensitive" bson:"unmask_sensitive"`           // uncover the sensitive fields, eg: ssh password, access key, etc
	TGBotToken          string       `json:"tg_bot_token" bson:"tg_bot_token"`                   // telegram bot token
	GlobalAttrs         label.Labels `json:"global_attrs" bson:"global_attrs"`                   // user customized kv, we just treat it as general label kv
	UpdatedAt           time.Time    `json:"updated_at" bson:"updated_at"`
	Initial             bool         `json:"initial" bson:"initial"`
}

// Hidden set the smtpd config sensitive fields as invisible
func (s *Settings) Hidden() {
	if s.TGBotToken != "" {
		s.TGBotToken = SensitiveHolder
	}
}

// UpdateSettingsReq is similar to types.Settings, but all changable fields are pointer type
type UpdateSettingsReq struct {
	AdvertiseAddr       *string `json:"advertise_addr"`
	LogLevel            *string `json:"log_level"`
	EnableHTTPMuxDebug  *bool   `json:"enable_httpmux_debug"`
	MetricsAuthUser     *string `json:"metrics_auth_user"`
	MetricsAuthPassword *string `json:"metrics_auth_password"`
	UnmarkSensitive     *bool   `json:"unmask_sensitive"`
	TGBotToken          *string `json:"tg_bot_token"`
}

// Valid verify the UpdateSettingsReq
func (req *UpdateSettingsReq) Valid() error {
	if req.AdvertiseAddr != nil {
		if err := validator.String(*req.AdvertiseAddr, 1, 4096, nil); err != nil {
			return fmt.Errorf("advertise addr %v", err)
		}
	}
	if req.AdvertiseAddr != nil {
		for _, addr := range strings.Split(*req.AdvertiseAddr, ",") {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return fmt.Errorf("advertise addr %s should be the format host:port", addr)
			}
			// ensure port must be numberic, because golang http.Client doesn't support naming port: invalid URL port
			if _, err = strconv.Atoi(port); err != nil {
				return fmt.Errorf("advertise addr %s port invalid: %v", addr, err)
			}
		}
	}
	if req.LogLevel != nil {
		if _, err := log.ParseLevel(*req.LogLevel); err != nil {
			return err
		}
	}
	if req.MetricsAuthUser != nil || req.MetricsAuthPassword != nil {
		if ptype.StringV(req.MetricsAuthUser) == "" || ptype.StringV(req.MetricsAuthPassword) == "" {
			return fmt.Errorf("metrics basic authentication user and password required")
		}
	}
	if req.TGBotToken != nil {
		if *req.TGBotToken == "" {
			return fmt.Errorf("tg bot token required")
		}
	}
	return nil
}
