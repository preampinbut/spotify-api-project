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

func (server *Server) PlayerState() PlayerState {
	return *server.playerState
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
			logrus.WithError(err).Fatalf("failed to start server\n")
		}
	}()
}

func (server *Server) StartServer(listener net.Listener) error {
	api := http.NewServeMux()
	api.HandleFunc("/api/state",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			Event.WritePlayerState(w, server)
		},
	)
	api.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Connection", "keep-alive")
		// handleWebSocket(w, r, server)
		tick := time.NewTicker(3 * time.Second)
		defer func() { tick.Stop() }()

		Event.WritePlayerState(w, server)
		for {
			select {
			case <-tick.C:
				Event.WritePlayerState(w, server)
			}
		}
	})

	server.s.Handler = api
	return server.s.Serve(listener)
}

func (server *Server) StartFetchingSpotify() {
	tick := time.NewTicker(6 * time.Second)
	defer func() { tick.Stop() }()

	fetchPlayerState(server)
	for {
		select {
		case <-tick.C:
			fetchPlayerState(server)
		}
	}
}
