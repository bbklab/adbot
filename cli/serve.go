package cli

import (
	"github.com/urfave/cli"

	"github.com/bbklab/adbot/master"
	"github.com/bbklab/adbot/types"
)

var (
	serveFlags = []cli.Flag{
		cli.StringFlag{
			Name:   "listen",
			Usage:  "The address that API service listens on, eg: :80",
			EnvVar: "LISTEN_ADDR",
			Value:  ":80",
		},
		cli.StringFlag{
			Name:   "tls-cert",
			Usage:  "The TLS certificate file over http serving",
			EnvVar: "TLS_CERT_FILE",
		},
		cli.StringFlag{
			Name:   "tls-key",
			Usage:  "The TLS key file over http serving",
			EnvVar: "TLS_KEY_FILE",
		},
		cli.StringFlag{
			Name:   "db-type",
			Usage:  "The database store type, [mongodb]",
			Value:  "mongodb",
			EnvVar: "DB_TYPE",
		},
		cli.StringFlag{
			Name:   "mgo-url",
			Usage:  "The mongodb url address",
			EnvVar: "MGO_URL",
			Value:  "mongodb://127.0.0.1:27017/adbot",
		},
		cli.StringFlag{
			Name:   "unix-sock",
			Usage:  "The unix socket file path",
			Value:  "/var/run/adbot/adbot.sock",
			EnvVar: "UNIX_SOCK",
		},
		cli.StringFlag{
			Name:   "pid-file",
			Usage:  "The pid file path",
			Value:  "/var/run/adbot/adbot.pid",
			EnvVar: "PID_FILE",
		},
	}
)

// ServeCommand is exported
func ServeCommand() cli.Command {
	return cli.Command{
		Name:      "serve",
		ShortName: "m",
		Usage:     "start adbot master and serve API",
		Flags:     serveFlags,
		Action:    runMaster,
	}
}

func runMaster(c *cli.Context) error {
	cfg := &types.MasterConfig{
		Listen:   c.String("listen"),
		TLSCert:  c.String("tls-cert"),
		TLSKey:   c.String("tls-key"),
		UnixSock: c.String("unix-sock"),
		PidFile:  c.String("pid-file"),
		Store: &types.StoreConfig{
			Type: c.String("db-type"),
			MongodbConfig: &types.MongodbConfig{
				MgoURL: c.String("mgo-url"),
			},
		},
	}
	if err := cfg.Valid(); err != nil {
		return err
	}

	master := master.New(cfg)
	master.Run() // fatal -> exit

	return nil
}
