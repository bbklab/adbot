package helpers

import (
	"os"
	"path"
)

// nolint
var (
	ParamAll = "all"
)

// nolint
var (
	homedir              = path.Join(GetHomeDir(), ".adbot")
	LocalConfigFile      = path.Join(homedir, "config")
	LocalConfigFileLock  = path.Join(homedir, ".config.lock")
	LocalSSHKeyTempDir   = path.Join(homedir, "sshkeys.temp")
	LocalSSUserTokenFile = path.Join(homedir, "ssuser.token")
)

// nolint
var (
	AdbotServiceName  = "adbot-master" // should be same as the systemd service name of: contrib/rpm/files/systemd/adbot-master.service
	MongodServiceName = "mongod"
	PromServiceName   = "prometheus"
)

// nolint
var (
	resBaseDir              = "/usr/share/adbot"
	resDepDir               = path.Join(resBaseDir, "dependency")
	ResDepMongod            = path.Join(resDepDir, "mongod.pkg")
	ResDepPrometheus        = path.Join(resDepDir, "prometheus.pkg")
	ResDepMongodPackageName = "mongodb-org-server" // rpm package name
	ResDepPromPackageName   = "prometheus2"        // rpm package name
)

func init() {
	os.MkdirAll(homedir, os.FileMode(0755))
}
