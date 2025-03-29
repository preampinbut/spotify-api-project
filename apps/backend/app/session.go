package app

import (
	"backend/config"
	"backend/db"
	"context"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Session struct {
	cfg          *oauth2.Config
	auth         Auth
	clients      map[string]bool
	clientsMutex *sync.RWMutex
	token        *oauth2.Token
	tokenMutex   *sync.RWMutex
	dbClient     *db.PrismaClient
}

func NewSession(cfg *oauth2.Config, dbClient *db.PrismaClient) *Session {
	return &Session{
		cfg:          cfg,
		auth:         Auth{},
		clients:      make(map[string]bool),
		clientsMutex: &sync.RWMutex{},
		token:        nil,
		tokenMutex:   &sync.RWMutex{},
		dbClient:     dbClient,
	}
}

func NewSessionWithToken(cfg *oauth2.Config, dbClient *db.PrismaClient, token *oauth2.Token) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tokenSource := cfg.TokenSource(ctx, token)
	token, err := tokenSource.Token()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to get token from token source")
	}
	return &Session{
		cfg:          cfg,
		auth:         Auth{},
		clients:      make(map[string]bool),
		clientsMutex: &sync.RWMutex{},
		token:        token,
		tokenMutex:   &sync.RWMutex{},
		dbClient:     dbClient,
	}
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
		if err = config.SaveCredentials(s.dbClient, s.token); err != nil {
			logrus.WithError(err).Fatalf("failed to save credentials")
		}
	}

	client := s.cfg.Client(ctx, s.token)

	defer func() {
		s.tokenMutex.Unlock()
		cancel()
	}()
	return fn(ctx, client)
}
