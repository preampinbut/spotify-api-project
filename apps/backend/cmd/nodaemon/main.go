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
		logrus.WithError(err).Fatalf("failed to load config")
	}

	cs, err := connstring.ParseAndValidate(appConfig.ConnectionString)
	if err != nil {
		logrus.WithError(err).Fatalf("connection string is invalid")
	}
	dbClient, err := mongo.Connect(options.Client().ApplyURI(cs.Original))

	if err != nil {
		logrus.WithError(err).Fatalf("could not connect to database")
	}
	defer func() {
		if err := dbClient.Disconnect(context.Background()); err != nil {
			logrus.WithError(err).Fatalf("failed when try to disconnect from database")
		}
	}()

	collection := dbClient.Database(cs.Database).Collection(appConfig.Collection)
	token := config.LoadCredentials(collection)

	var session *app.Session
	var token_err error = nil
	cfg := app.NewConfig(appConfig)

	if token != nil {
		session, token_err = app.NewSessionWithToken(cfg, dbClient, collection, token)
	}

	if token == nil || token_err != nil {
		session = app.NewSession(cfg, dbClient, collection)
	}

	server := app.NewServer(session)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Port))
	if err != nil {
		logrus.WithError(err).Fatalf("failed to create listener")
	}

	// if no token existed we halt here login first
	if token == nil || token_err != nil {
		logrus.Warnf("credentials not existed or revoked please login")
		done := make(chan struct{})
		server.StartOAuth2Server(listener, done)
		url := session.AuthURL()
		fmt.Printf("%s\n", url)
		<-done
		logrus.Infof("Login Success")
	} else {
		logrus.Infof("credentials existed skip login")
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
