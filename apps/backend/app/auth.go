package app

import (
	"context"
	"fmt"
	"net/http"

	"backend/config"
	"backend/util"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

// Auth holds state for OAuth 2.0 Authorization Code Flow with PKCE.
type Auth struct {
	state        string
	codeVerifier string
}

// NewAuth initializes Auth with a random state string.
func NewAuth() *Auth {
	state, err := util.GenerateRandomString(16)
	if err != nil {
		logrus.WithError(err).Fatal("failed to generate random state string")
	}
	return &Auth{state: state}
}

// NewConfig returns an *oauth2.Config for Spotify API.
func NewConfig(cfg *config.ServerConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    cfg.ClientID,
		RedirectURL: fmt.Sprintf("%s%s", cfg.BaseURL, config.CallbackPath),
		Endpoint:    spotify.Endpoint,
		Scopes: []string{
			"user-read-private",
			"user-read-playback-state",
		},
	}
}

// AuthURL generates the Spotify OAuth authorization URL.
func (s *Session) AuthURL() string {
	s.auth.codeVerifier = oauth2.GenerateVerifier()
	return s.cfg.AuthCodeURL(
		s.auth.state,
		oauth2.S256ChallengeOption(s.auth.codeVerifier),
	)
}

// CallbackHandler exchanges the authorization code for an access token.
func CallbackHandler(w http.ResponseWriter, r *http.Request, s *Session, done chan struct{}) {
	code := r.URL.Query().Get("code")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	token, err := s.cfg.Exchange(ctx, code, oauth2.VerifierOption(s.auth.codeVerifier))
	if err != nil {
		http.Error(w, "Couldn't exchange for token", http.StatusForbidden)
		logrus.WithError(err).Fatal("failed to exchange token")
	}

	s.token = token
	if err = config.SaveCredentials(s.collection, s.token); err != nil {
		logrus.WithError(err).Fatal("failed to save credentials")
	}

	close(done)
	_, _ = w.Write([]byte("success"))
}
