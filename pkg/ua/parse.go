package ua

import (
	"net/http"

	"github.com/varstr/uaparser"
)

// ParseUA parse http request User-Agent Header to obtain the correspoding
// device, os, browser informations
func ParseUA(r *http.Request) (dev, os, browser string) {
	if r == nil {
		return
	}

	info := uaparser.Parse(r.UserAgent())
	if info == nil {
		return
	}

	if info.DeviceType != nil {
		dev = info.DeviceType.Name
	}

	if info.OS != nil {
		os = info.OS.Name
	}

	if info.Browser != nil {
		browser = info.Browser.Name
	}

	return
}
