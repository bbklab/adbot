// Package lic ...
//
// note: this file will be copied to product codes to report the license infos
// we introduce this into product side by copying this file instead of importing this package
// because of security concern about personal identifiers
package lic

import (
	"time"

	"github.com/bbklab/adbot/pkg/geoip"
)

// Report is a db license report
type Report struct {
	ID         string                 `bson:"id" json:"id"`                   // ref: license id
	ExtraInfo  map[string]interface{} `bson:"extra_info" json:"extra_info"`   // set by product side, collected datas
	ReportedAt time.Time              `bson:"reported_at" json:"reported_at"` // set by product side
	RemoteAddr string                 `bson:"remote_addr" json:"remote_addr"` // set by license server side
	GeoInfo    *geoip.GeoInfo         `json:"geoinfo" bson:"geoinfo"`         // set by license server side, detected GEO info
	GeoInfoZh  *geoip.GeoInfo         `json:"geoinfo_zh" bson:"geoinfo_zh"`   // set by license server side
	ReceivedAt time.Time              `bson:"received_at" json:"received_at"` // set by license server side
}
