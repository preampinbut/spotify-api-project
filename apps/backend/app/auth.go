package app

import (
	"backend/config"
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

type Auth struct {
	state        string
	codeVerifier string
}

func NewAuth() *Auth {
	return &Auth{
		state: "state",
	}
}

func NewConfig(cfg *config.ConfigType) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    cfg.ClientId,
		RedirectURL: fmt.Sprintf("%s%s", cfg.BaseURL, config.CallbackPath),
		Endpoint:    spotify.Endpoint,
		Scopes: []string{
			"user-read-private",
			"user-read-playback-state",
		},
	}
}

func (session *Session) AuthURL() string {
	session.auth.codeVerifier = oauth2.GenerateVerifier()

	return session.cfg.AuthCodeURL(session.auth.state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(session.auth.codeVerifier),
	)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request, s *Session, done chan struct{}) {
	code := r.URL.Query().Get("code")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	token, err := s.cfg.Exchange(ctx, code, oauth2.VerifierOption(s.auth.codeVerifier))
	if err != nil {
		http.Error(w, "Couldn't exchange for token", http.StatusForbidden)
		logrus.Fatal(err)
	}

	s.token = token
	config.SaveCredentials(s.token)

	close(done)

	_, _ = w.Write([]byte("success"))
}
