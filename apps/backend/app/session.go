package app

import (
	"backend/config"
	"context"
	"sync"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Session struct {
	auth *spotifyauth.Authenticator

	client      *spotify.Client
	clientMutex *sync.RWMutex
}

func NewSession(auth *spotifyauth.Authenticator) *Session {
	return &Session{
		auth:        auth,
		client:      nil,
		clientMutex: &sync.RWMutex{},
	}
}

func NewSessionWithToken(auth *spotifyauth.Authenticator, token *oauth2.Token) *Session {
	ctx := context.Background()
	return &Session{
		auth:        auth,
		client:      spotify.New(auth.Client(ctx, token)),
		clientMutex: &sync.RWMutex{},
	}
}

func (s *Session) WithClient(fn func(ctx context.Context, client *spotify.Client) error) error {
	s.clientMutex.Lock()
	defer func() { s.clientMutex.Unlock() }()
	token, _ := s.client.Token()
	_ = config.SaveCredentials(token)
	return fn(context.Background(), s.client)
}
