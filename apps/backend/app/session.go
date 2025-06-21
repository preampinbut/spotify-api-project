package app

import (
	"backend/config"
	"context"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/oauth2"
)

type Session struct {
	cfg          *oauth2.Config
	auth         Auth
	clients      map[string]bool
	clientsMutex *sync.RWMutex
	token        *oauth2.Token
	tokenMutex   *sync.RWMutex
	dbClient     *mongo.Client
	collection   *mongo.Collection
}

func NewSession(cfg *oauth2.Config, dbClient *mongo.Client, collection *mongo.Collection) *Session {
	return &Session{
		cfg:          cfg,
		auth:         Auth{},
		clients:      make(map[string]bool),
		clientsMutex: &sync.RWMutex{},
		token:        nil,
		tokenMutex:   &sync.RWMutex{},
		dbClient:     dbClient,
		collection:   collection,
	}
}

func NewSessionWithToken(cfg *oauth2.Config, dbClient *mongo.Client, collection *mongo.Collection, token *oauth2.Token) (*Session, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tokenSource := cfg.TokenSource(ctx, token)
	token, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	return &Session{
		cfg:          cfg,
		auth:         Auth{},
		clients:      make(map[string]bool),
		clientsMutex: &sync.RWMutex{},
		token:        token,
		tokenMutex:   &sync.RWMutex{},
		dbClient:     dbClient,
		collection:   collection,
	}, nil
}

func (s *Session) WithClient(fn func(ctx context.Context, client *http.Client) error) error {
	s.tokenMutex.Lock()

	ctx, cancel := context.WithCancel(context.Background())
	tokenSource := s.cfg.TokenSource(ctx, s.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		logrus.WithError(err).Errorf("failed to refresh token")
		cancel()
		return err
	}

	if s.token.AccessToken != newToken.AccessToken {
		s.token = newToken
		if err = config.SaveCredentials(s.collection, s.token); err != nil {
			logrus.WithError(err).Fatalf("failed to save credentials")
		}
		logrus.Info("credentials updated")
	}

	client := s.cfg.Client(ctx, s.token)

	defer func() {
		s.tokenMutex.Unlock()
		cancel()
	}()
	return fn(ctx, client)
}
