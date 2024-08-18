package app

import (
	"backend/config"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	state         = "state"
	codeVerifier  = ""
	codeChallenge = ""
)

func NewAuth(opts ...spotifyauth.AuthenticatorOption) *spotifyauth.Authenticator {
	return spotifyauth.New(opts...)
}

func AuthURL(session *Session) string {
	codeVerifier = oauth2.GenerateVerifier()
	codeChallenge = oauth2.S256ChallengeFromVerifier(codeVerifier)

	return session.auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
	)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request, s *Session, done chan struct{}) {
	tok, err := s.auth.Token(r.Context(), state, r,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		logrus.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		logrus.Fatalf("State mismatch: %s != %s", st, state)
	}

	s.client = spotify.New(s.auth.Client(r.Context(), tok))
	config.SaveCredentials(tok)
	close(done)

	_, _ = w.Write([]byte("success"))
}
