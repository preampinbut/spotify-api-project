package app

import (
	"backend/config"
	"context"
	"net/http"
	"sync" // Package for synchronization primitives like RWMutex

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/oauth2"
)

// Session holds all necessary components for interacting with an external service
// using OAuth2, including configuration, database access, and client synchronization.
type Session struct {
	cfg   *oauth2.Config // OAuth2 configuration
	auth  Auth           // A struct (presumably) for authentication-related helper methods

	// clients map tracks active clients/sessions (usage context unclear, but synchronized)
	clients      map[string]bool
	clientsMutex *sync.RWMutex // Mutex to protect access to the clients map

	token *oauth2.Token // The current OAuth2 token
	// tokenMutex protects read/write access to the token, essential for token refresh logic
	tokenMutex *sync.RWMutex

	dbClient   *mongo.Client      // MongoDB client instance
	collection *mongo.Collection // MongoDB collection used for storing credentials
}

// NewSession creates a new Session with nil token, primarily for initial authentication flow.
func NewSession(cfg *oauth2.Config, dbClient *mongo.Client, collection *mongo.Collection) *Session {
	return &Session{
		cfg:          cfg,
		auth:         Auth{},
		clients:      make(map[string]bool),
		clientsMutex: &sync.RWMutex{},
		token:        nil, // Token is initially nil, to be acquired later
		tokenMutex:   &sync.RWMutex{},
		dbClient:     dbClient,
		collection:   collection,
	}
}

// NewSessionWithToken creates a new Session, attempts to refresh the provided token
// immediately, and stores the new token. Useful for loading saved credentials.
func NewSessionWithToken(cfg *oauth2.Config, dbClient *mongo.Client, collection *mongo.Collection, token *oauth2.Token) (*Session, error) {
	// Create a context for the token source
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context cancellation

	// Use the provided token to create a token source, which may automatically refresh the token
	tokenSource := cfg.TokenSource(ctx, token)
	newToken, err := tokenSource.Token() // Get the current token, refreshing if needed
	if err != nil {
		return nil, err
	}

	return &Session{
		cfg:          cfg,
		auth:         Auth{},
		clients:      make(map[string]bool),
		clientsMutex: &sync.RWMutex{},
		token:        newToken,
		tokenMutex:   &sync.RWMutex{},
		dbClient:     dbClient,
		collection:   collection,
	}, nil
}

// WithClient handles token refresh logic and provides an authenticated *http.Client
// to the provided function 'fn'. It ensures that the token is up-to-date.
func (s *Session) WithClient(fn func(ctx context.Context, client *http.Client) error) error {
	s.tokenMutex.Lock() // Lock access to the token for the duration of the refresh/client creation

	// Context for the token source and the client
	ctx, cancel := context.WithCancel(context.Background())
	tokenSource := s.cfg.TokenSource(ctx, s.token)

	// Attempt to get a fresh token. This call handles automatic refreshing.
	newToken, err := tokenSource.Token()
	if err != nil {
		logrus.WithError(err).Errorf("failed to refresh token")
		cancel() // Cancel context before returning error
		s.tokenMutex.Unlock() // Unlock before returning error
		return err
	}

	// Check if the access token was updated by the refresh process
	if s.token.AccessToken != newToken.AccessToken {
		s.token = newToken // Update the session's token
		// Save the new token to the database
		if err = config.SaveCredentials(s.collection, s.token); err != nil {
			// A failure to save is treated as fatal, as the new token may be lost
			logrus.WithError(err).Fatalf("failed to save credentials")
		}
		logrus.Info("credentials updated")
	}

	// Create an authenticated HTTP client using the current token
	client := s.cfg.Client(ctx, s.token)

	// Unlock the token mutex and cancel the context when the function 'fn' finishes
	defer func() {
		s.tokenMutex.Unlock()
		cancel()
	}()

	// Execute the provided function with the authenticated client
	return fn(ctx, client)
}
