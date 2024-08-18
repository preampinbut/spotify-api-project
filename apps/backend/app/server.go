package app

import (
	"context"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
)

type Server struct {
	session *Session
	s       *http.Server
}

func NewServer(session *Session) *Server {
	return &Server{
		session: session,
		s:       &http.Server{},
	}
}

func (server *Server) StartOAuth2Server(listener net.Listener, session *Session, done chan int) {
	api := http.NewServeMux()
	api.HandleFunc("/api/callback", func(w http.ResponseWriter, r *http.Request) {
		CallbackHandler(w, r, session, done)
	})
	go func() {
		server.s.Handler = api
		if err := server.s.Serve(listener); err != nil {
			logrus.WithError(err).Fatalf("failed to start server")
		}
	}()
}

func (server *Server) StartServer(listener net.Listener) error {
	api := http.NewServeMux()
	api.HandleFunc("/api/state",
		func(w http.ResponseWriter, r *http.Request) {
			server.session.WithClient(func(ctx context.Context, client *spotify.Client) error {
				user, err := client.CurrentUser(ctx)
				if err != nil {
					logrus.WithError(err).Errorf("failed to get current user")
					return err
				}
				_, _ = w.Write([]byte(user.DisplayName))
				return nil
			})
		},
	)

	server.s.Handler = api
	return server.s.Serve(listener)
}
