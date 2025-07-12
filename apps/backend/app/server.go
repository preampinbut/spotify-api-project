package app

import (
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	session *Session
	s       *http.Server

	playerState *PlayerState
}

func NewServer(session *Session) *Server {
	return &Server{
		session: session,
		s:       &http.Server{},
	}
}

func (server *Server) StartOAuth2Server(listener net.Listener, done chan struct{}) {
	api := http.NewServeMux()
	api.HandleFunc("/api/callback", func(w http.ResponseWriter, r *http.Request) {
		CallbackHandler(w, r, server.session, done)
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
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			_ = fetchPlayerState(server, true)
			Event.WritePlayerState(w, server)
		},
	)
	api.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Connection", "keep-alive")

		tick := time.NewTicker(3 * time.Second)
		clientKey := r.RemoteAddr

		logrus.Infof("connection from %s", clientKey)

		ctx := r.Context()

		defer func() {
			tick.Stop()
			server.session.clientsMutex.Lock()
			delete(server.session.clients, clientKey)
			server.session.clientsMutex.Unlock()
		}()

		server.session.clientsMutex.Lock()
		server.session.clients[clientKey] = true
		server.session.clientsMutex.Unlock()

		Event.WritePlayerState(w, server)
		for {
			select {
			case <-tick.C:
				Event.WritePlayerState(w, server)
			case <-ctx.Done():
				logrus.Infof("%s disconnected", clientKey)
				return
			}
		}
	})

	server.s.Handler = api
	return server.s.Serve(listener)
}

func (server *Server) StartFetchingSpotify() {
	tick := time.NewTicker(6 * time.Second)
	defer tick.Stop()

	_ = fetchPlayerState(server, true)

	for range tick.C {
		_ = fetchPlayerState(server, false)
	}
}
