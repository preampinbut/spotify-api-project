package app

import (
	"backend/config"
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Session struct {
	auth *spotifyauth.Authenticator

	ctx          context.Context
	client       *spotify.Client
	clientMutex  *sync.RWMutex
	clients      map[*websocket.Conn]bool
	clientsMutex *sync.RWMutex
}

func NewSession(auth *spotifyauth.Authenticator) *Session {
	return &Session{
		ctx:          context.Background(),
		auth:         auth,
		client:       nil,
		clientMutex:  &sync.RWMutex{},
		clients:      make(map[*websocket.Conn]bool),
		clientsMutex: &sync.RWMutex{},
	}
}

func NewSessionWithToken(auth *spotifyauth.Authenticator, token *oauth2.Token) *Session {
	ctx := context.Background()
	return &Session{
		ctx:          ctx,
		auth:         auth,
		client:       spotify.New(auth.Client(ctx, token)),
		clientMutex:  &sync.RWMutex{},
		clients:      make(map[*websocket.Conn]bool),
		clientsMutex: &sync.RWMutex{},
	}
}

func (s *Session) WithClient(fn func(ctx context.Context, client *spotify.Client) error) error {
	s.clientMutex.Lock()
	defer func() { s.clientMutex.Unlock() }()
	token, _ := s.client.Token()
	config.SaveCredentials(token)
	return fn(s.ctx, s.client)
}

func (s *Session) WithClients(fn func(clients map[*websocket.Conn]bool) error) error {
	s.clientsMutex.Lock()
	defer func() { s.clientsMutex.Unlock() }()
	return fn(s.clients)
}
