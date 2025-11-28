package main

import (
	"backend/app"
	"backend/config"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	appConfig, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatal("failed to load config")
	}

	cs, err := connstring.ParseAndValidate(appConfig.ConnectionString)
	if err != nil {
		logrus.WithError(err).Fatal("connection string is invalid")
	}
	dbClient, err := mongo.Connect(options.Client().ApplyURI(cs.Original))
	if err != nil {
		logrus.WithError(err).Fatal("could not connect to database")
	}
	defer func() {
		if err := dbClient.Disconnect(context.Background()); err != nil {
			logrus.WithError(err).Fatal("failed to disconnect from database")
		}
	}()

	collection := dbClient.Database(cs.Database).Collection("config")
	token := config.LoadCredentials(collection)
	cfg := app.NewConfig(appConfig)

	var session *app.Session
	var tokenErr error

	if token != nil {
		session, tokenErr = app.NewSessionWithToken(cfg, dbClient, collection, token)
	}
	if token == nil || tokenErr != nil {
		session = app.NewSession(cfg, dbClient, collection)
		logrus.Warn("credentials not existed or revoked, please complete OAuth2 login")
		done := make(chan struct{})
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Port))
		if err != nil {
			logrus.WithError(err).Fatal("failed to create listener")
		}
		server := app.NewServer(session)
		server.StartOAuth2Server(listener, done)
		fmt.Printf("Authorize this application by visiting: %s\n", session.AuthURL())
		<-done
		logrus.Info("Login Success: credentials saved")
		// Start main server after OAuth2 flow
		go func() {
			time.Sleep(3 * time.Second)
			if err = server.StartServer(listener); err != nil {
				logrus.WithError(err).Fatal("failed to start main server")
			}
		}()
		server.StartFetchingSpotify()
		return
	}

	logrus.Info("credentials existed, skipping login")
	server := app.NewServer(session)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Port))
	if err != nil {
		logrus.WithError(err).Fatal("failed to create listener")
	}

	go func() {
		time.Sleep(3 * time.Second)
		if err = server.StartServer(listener); err != nil {
			logrus.WithError(err).Fatal("failed to start main server")
		}
	}()
	server.StartFetchingSpotify()
}
