package main

import (
	"backend/app"
	"backend/config"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload" // Automatically loads .env files for local development
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
)

func main() {
	// Configure logging output.
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	// 1. Load configuration from environment variables (using config.LoadConfig from the Canvas).
	appConfig, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to load config")
	}

	// 2. Connect to MongoDB.
	cs, err := connstring.ParseAndValidate(appConfig.ConnectionString)
	if err != nil {
		logrus.WithError(err).Fatalf("connection string is invalid")
	}
	dbClient, err := mongo.Connect(options.Client().ApplyURI(cs.Original))

	if err != nil {
		logrus.WithError(err).Fatalf("could not connect to database")
	}
	// Ensure the database connection is closed when main exits.
	defer func() {
		if err := dbClient.Disconnect(context.Background()); err != nil {
			logrus.WithError(err).Fatalf("failed when try to disconnect from database")
		}
	}()

	// Get the specific collection for credential storage and attempt to load existing token.
	collection := dbClient.Database(cs.Database).Collection("config")
	token := config.LoadCredentials(collection)

	var session *app.Session
	var token_err error = nil
	cfg := app.NewConfig(appConfig)

	// 3. Initialize application session based on token status.
	if token != nil {
		// Attempt to create a session with an existing token (which handles token refresh/validation internally).
		session, token_err = app.NewSessionWithToken(cfg, dbClient, collection, token)
	}

	if token == nil || token_err != nil {
		// If no token exists or the existing token failed validation, start a new session for fresh login.
		session = app.NewSession(cfg, dbClient, collection)
	}

	server := app.NewServer(session)

	// Set up the network listener on the configured port.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", appConfig.Port))
	if err != nil {
		logrus.WithError(err).Fatalf("failed to create listener")
	}

	// 4. Handle OAuth2 login flow if no valid token exists.
	if token == nil || token_err != nil {
		logrus.Warnf("credentials not existed or revoked, please complete OAuth2 login")
		done := make(chan struct{})
		// Start a temporary HTTP server for the OAuth2 redirect/callback.
		server.StartOAuth2Server(listener, done)
		url := session.AuthURL()
		fmt.Printf("Authorize this application by visiting: %s\n", url)
		<-done // Block until the OAuth2 flow is complete and the token is saved.
		logrus.Infof("Login Success: credentials saved")
	} else {
		logrus.Infof("credentials existed, skipping login")
	}

	// 5. Start the main application server in a goroutine.
	go func() {
		// Wait briefly to ensure any initial data fetching (like user profile or playback state) is complete.
		time.Sleep(3 * time.Second)
		if err = server.StartServer(listener); err != nil {
			logrus.WithError(err).Fatalf("failed to start main server")
		}
	}()

	// 6. Start the recurring background process (e.g., fetching Spotify playback state).
	server.StartFetchingSpotify()
}
