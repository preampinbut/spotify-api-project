package main

import (
	"backend/app"
	"backend/config"
	"backend/db"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	databaseURL := os.Getenv("DATABASE_URL")
	if len(strings.TrimSpace(databaseURL)) == 0 {
		logrus.Fatalf("DATABASE_URL did not set")
	}

	dbClient := db.NewClient(
		db.WithDatasourceURL(databaseURL),
	)
	if err := dbClient.Connect(); err != nil {
		logrus.WithError(err).Fatalf("could not connect to database")
	}
	defer func() {
		if err := dbClient.Disconnect(); err != nil {
			logrus.WithError(err).Fatalf("failed when try to disconnect from database")
		}
	}()

	appConfig, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to load config")
	}

	token := config.LoadCredentials(dbClient)

	var session *app.Session
	cfg := app.NewConfig(appConfig)
	if token != nil {
		logrus.Infof("credentials existed skip login")
		session = app.NewSessionWithToken(cfg, dbClient, token)
	} else {
		logrus.Warnf("credentials not existed please login")
		session = app.NewSession(cfg, dbClient)
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
		fmt.Printf("%s", url)
		<-done
	}

	logrus.Infof("Login Success")

	go func() {
		// sleep for 3 seconds to make sure that player state have data before starting
		time.Sleep(3 * time.Second)
		if err = server.StartServer(listener); err != nil {
			logrus.WithError(err).Fatalf("failed to start server")
		}
	}()

	server.StartFetchingSpotify()
}
