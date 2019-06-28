package main

import (
	"math/rand"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	icli "github.com/bbklab/adbot/cli"
	_ "github.com/bbklab/adbot/debug"
	"github.com/bbklab/adbot/version"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	globalFlags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug mode",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "log-file",
			Usage:  "The log file path",
			EnvVar: "LOG_FILE",
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "adbot"
	app.Author = "Coding Bot"
	app.Email = "codingbot@gmail.com"
	app.Version = version.GetVersion()
	if gitCommit := version.GetGitCommit(); gitCommit != "" {
		app.Version += "-" + gitCommit
	}

	app.Flags = globalFlags

	app.Before = func(c *cli.Context) error {
		var (
			debug   = c.Bool("debug")
			logFile = c.String("log-file")
		)

		log.SetLevel(log.InfoLevel)
		if debug {
			log.SetLevel(log.DebugLevel)
		}

		if os.Getenv("IN_CONTAINER") != "" { // docker-compose doesn't have any log system like journald
			log.SetFormatter(&log.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			})
		} else {
			log.SetFormatter(&log.TextFormatter{
				DisableTimestamp: true, // already logged by journald system
			})
		}

		if logFile == "" {
			log.SetOutput(os.Stdout)
		} else {
			fd, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			log.SetOutput(fd)
		}
		return nil
	}

	app.Commands = []cli.Command{
		// start master
		icli.ServeCommand(),
		// start agent
		icli.JoinCommand(),
		// cli hosts setup
		icli.HostCommand(),
		// CLI client
		icli.LoginCommand(),
		icli.LogoutCommand(),
		icli.VersionCommand(),
		icli.DumpCommand(),
		icli.PProfCommand(),
		icli.InfoCommand(),
		icli.UserCommand(),
		icli.NodeCommand(),
		icli.SettingsCommand(),
		icli.GeoCommand(),
		icli.AdbNodeCommand(),
		icli.AdbDeviceCommand(),
	}

	app.RunAndExitOnError()
}
