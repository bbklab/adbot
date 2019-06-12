package scheduler

import (
	"path"
)

var (
	resBaseDir = "/usr/share/adbot"
	resGeoCity = path.Join(resBaseDir, "geo/GeoLite2-City.mmdb")
	resGeoAsn  = path.Join(resBaseDir, "geo/GeoLite2-ASN.mmdb")
)
