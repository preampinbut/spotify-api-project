package main

import (
	"backend/app"
	"backend/config"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	appConfig, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to load config")
	}

	token, _ := config.LoadCredentials()

	var session *app.Session
	cfg := app.NewConfig(appConfig)
	if token != nil {
		logrus.Infof("credentials existed skip login")
		session = app.NewSessionWithToken(cfg, token)
	} else {
		logrus.Warnf("credentials not existed please login")
		session = app.NewSession(cfg)
	}
	server := app.NewServer(session)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Port))
	if err != nil {
		logrus.WithError(err).Fatalf("failed to create listener")
	}

	// if no token existed we halt here login first
	if token == nil {
		done := make(chan struct{})
		server.StartOAuth2Server(listener, done)
		url := session.AuthURL()
		logrus.Infof("%s", url)
		<-done
	}

	go func() {
		// sleep for 3 seconds to make sure that player state have data before starting
		time.Sleep(3 * time.Second)
		if err = server.StartServer(listener); err != nil {
			logrus.WithError(err).Fatalf("failed to start server")
		}
	}()

	server.StartFetchingSpotify()
}
