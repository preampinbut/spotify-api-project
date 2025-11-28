package app

import (
	"context"
	"fmt"
	"net/http"

	"backend/config" // Package for application configuration and constants.
	"backend/util"   // Package for utility functions, including random string generation.

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify" // The Spotify-specific OAuth2 endpoint configuration.
)

// Auth holds the necessary state for the OAuth 2.0 Authorization Code Flow
// with Proof Key for Code Exchange (PKCE).
type Auth struct {
	state        string
	codeVerifier string // Used for PKCE to prove the client requesting the token is the one that started the flow.
}

// NewAuth initializes a new Auth struct with a default state value.
func NewAuth() *Auth {
	state, err := util.GenerateRandomString(16)
	if err != nil {
		logrus.WithError(err).Fatalf("failed to generate random state string")
	}
	// The 'state' parameter is used to maintain state between the request and the callback
	// and to prevent cross-site request forgery (CSRF).
	return &Auth{
		state: state,
	}
}

// NewConfig creates and returns an *oauth2.Config, pre-configured for the Spotify API.
func NewConfig(cfg *config.ServerConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID: cfg.ClientID,
		// RedirectURL is the endpoint the authorization server (Spotify) sends the user back to
		// after they grant or deny permission.
		RedirectURL: fmt.Sprintf("%s%s", cfg.BaseURL, config.CallbackPath),
		Endpoint:    spotify.Endpoint,
		// Scopes define the permissions the application is requesting access to.
		Scopes: []string{
			"user-read-private",
			"user-read-playback-state",
		},
	}
}

// AuthURL generates the URL to which the user is redirected to initiate the OAuth flow.
func (s *Session) AuthURL() string {
	// Generate a new code verifier for PKCE before generating the URL.
	s.auth.codeVerifier = oauth2.GenerateVerifier()

	return s.cfg.AuthCodeURL(
		s.auth.state,
		// S256ChallengeOption is required for PKCE, providing the code challenge derived from the verifier.
		oauth2.S256ChallengeOption(s.auth.codeVerifier),
	)
}

// CallbackHandler handles the redirect from the authorization server (Spotify).
// It exchanges the authorization code for an access token.
func CallbackHandler(w http.ResponseWriter, r *http.Request, s *Session, done chan struct{}) {
	// Extract the authorization code from the query parameters.
	code := r.URL.Query().Get("code")

	// Use a context with cancellation for the token exchange network request.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Exchange the authorization code for an OAuth2 token,
	// verifying the request with the stored code verifier (PKCE).
	token, err := s.cfg.Exchange(ctx, code, oauth2.VerifierOption(s.auth.codeVerifier))
	if err != nil {
		http.Error(w, "Couldn't exchange for token", http.StatusForbidden)
		logrus.Fatal(err) // Treat failure to exchange as a fatal application error.
	}

	// Store the successfully retrieved token in the session.
	s.token = token
	// Persist the credentials (token) for future use.
	if err = config.SaveCredentials(s.collection, s.token); err != nil {
		logrus.WithError(err).Fatalf("failed to save credentials")
	}

	// Signal that the authentication process is complete.
	close(done)

	// Write a simple success message to the client.
	_, _ = w.Write([]byte("success"))
}
