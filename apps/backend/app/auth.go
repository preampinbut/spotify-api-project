package app

import (
	"fmt"
	"net"
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

func CallbackHandler(w http.ResponseWriter, r *http.Request, s *Session, done chan int) {
	tok, err := s.auth.Token(r.Context(), state, r,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		logrus.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		logrus.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	s.client = spotify.New(s.auth.Client(r.Context(), tok))
	done <- 1
}

func NewOAuth2Server(session *Session, done chan int) {
	lis, _ := net.Listen("tcp", fmt.Sprintf(":%d", 8888))
	http.HandleFunc("/api/callback", func(w http.ResponseWriter, r *http.Request) {
		CallbackHandler(w, r, session, done)
		lis.Close()
	})
	go http.Serve(lis, nil)
}
