package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sonntuet1997/medical-chain-utils/common"
	"github.com/urfave/cli/v2"
)

const (
	serviceName = "rate-limiter"
)

var logger *logrus.Logger

func main() {

	logger = logrus.New()
	if err := makeApp().Run(os.Args); err != nil {
		logger.WithField("err", err).Error("shutting down due to error")
		_ = os.Stderr.Sync()
		os.Exit(1)
	}
}

func makeApp() *cli.App {
	app := &cli.App{
		Name:                 serviceName,
		Version:              "v1.0.1",
		EnableBashCompletion: true,
		Compiled:             time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "Vu Nguyen",
				Email: "aqaurius6666@gmail.com",
			},
		},
		Copyright: "(c) 2021 SOTANEXT inc.",
		Action:    runMain,
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "run server",
				Action:  runMain,
			},
		},
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:     "redis-uri",
				Required: true,
				EnvVars:  []string{"REDIS_URI"},
				Usage:    "The URI for connecting to redis",
			},
			&cli.StringFlag{
				Name:     "redis-password",
				Required: true,
				EnvVars:  []string{"PASSWORD"},
				Usage:    "Password for redis database",
			},
			&cli.StringFlag{
				Name:     "redis-username",
				Required: true,
				EnvVars:  []string{"USERNAME"},
				Usage:    "Username for redis database",
			},
			&cli.StringFlag{
				Name:     "remote-url",
				Required: true,
				EnvVars:  []string{"REMOTE_URL"},
				Usage:    "Remote url for proxy",
			},
		},
			append(common.CommonGRPCFlag,
				common.LoggerFlag...)...),
	}
	return app
}
