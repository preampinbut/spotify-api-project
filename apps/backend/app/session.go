package app

import (
	"context"
	"sync"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type Session struct {
	auth *spotifyauth.Authenticator

	ctx        context.Context
	clientLock *sync.RWMutex
	client     *spotify.Client
}

func NewSession(auth *spotifyauth.Authenticator) *Session {
	return &Session{
		ctx:        context.Background(),
		auth:       auth,
		clientLock: &sync.RWMutex{},
	}
}

func (s *Session) WithClient(fn func(ctx context.Context, client *spotify.Client) error) error {
	s.clientLock.Lock()
	defer func() { s.clientLock.Unlock() }()
	return fn(s.ctx, s.client)
}
