package scheduler

import (
	"errors"
	"io/ioutil"
	"path"
)

var (
	resBaseDir      = "/usr/share/adbot"
	resPublicAPIDoc = path.Join(resBaseDir, "public_api.pdf")
	resGeoCity      = path.Join(resBaseDir, "geo/GeoLite2-City.mmdb")
	resGeoAsn       = path.Join(resBaseDir, "geo/GeoLite2-ASN.mmdb")
)

var (
	resFileMap = map[string][2]string{
		"public-api": {resPublicAPIDoc, ""},
	}
)

// GetResFile is exported
func GetResFile(name string) (string, string, error) {
	filehash, ok := resFileMap[name]
	if !ok {
		return "", "", errors.New("resource file name undefined")
	}

	file, hash := filehash[0], filehash[1]
	if hash == "" {
		return file, "", nil
	}

	sha1sum, _ := ioutil.ReadFile(hash)
	return file, string(sha1sum), nil
}
