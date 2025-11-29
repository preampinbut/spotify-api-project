package app

import (
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type StreamTicker struct {
	active uint
}

type PlayerTicker struct {
	ticker   *time.Ticker
	status   bool
	active   uint
	inactive uint
}

type Ticker struct {
	stream StreamTicker
	player PlayerTicker
}

type Server struct {
	session     *Session
	s           *http.Server
	playerState *PlaybackState

	ticker Ticker
}

func NewServer(session *Session) *Server {
	return &Server{
		session: session,
		s:       &http.Server{},

		ticker: Ticker{
			stream: StreamTicker{
				active: 1,
			},
			player: PlayerTicker{
				active:   3,
				inactive: 10,
			},
		},
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
			logrus.WithError(err).Fatal("failed to start OAuth2 server")
		}
	}()
}

func (server *Server) StartServer(listener net.Listener) error {
	api := http.NewServeMux()
	api.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_ = fetchPlayerState(server)
		Event.WritePlayerState(w, server)
	})
	api.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Connection", "keep-alive")

		clientKey := r.RemoteAddr
		logrus.Infof("connection from %s", clientKey)
		ctx := r.Context()

		defer func() {
			server.session.clientsMutex.Lock()
			delete(server.session.clients, clientKey)
			server.session.clientsMutex.Unlock()
		}()

		server.session.clientsMutex.Lock()
		server.session.clients[clientKey] = true
		server.session.clientsMutex.Unlock()

		prevStatus := server.ticker.player.status
		if !prevStatus {
			server.ticker.player.ticker.Reset(time.Duration(server.ticker.player.active) * time.Second)
			server.ticker.player.status = true
			logrus.WithFields(logrus.Fields{
				"clientCount": len(server.session.clients),
				"newDuration": time.Duration(server.ticker.player.active),
			}).Info("Ticker duration changed")
		}

		Event.WritePlayerState(w, server)
		for {
			select {
			case <-time.After(time.Duration(server.ticker.stream.active) * time.Second):
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
	getTicker := func(clientCount int) (time.Duration, bool) {
		if clientCount > 0 {
			return time.Duration(server.ticker.player.active), true
		}
		return time.Duration(server.ticker.player.inactive), false
	}

	clientCount := len(server.session.clients)
	duration, _ := getTicker(clientCount)

	server.ticker.player.ticker = time.NewTicker(duration * time.Second)
	server.ticker.player.status = false
	defer server.ticker.player.ticker.Stop()

	logrus.WithFields(logrus.Fields{
		"clientCount": clientCount,
		"newDuration": duration,
	}).Info("Ticker duration initialized")

	_ = fetchPlayerState(server)
	for {
		<-server.ticker.player.ticker.C

		clientCount := len(server.session.clients)
		duration, active := getTicker(clientCount)

		prevActive := server.ticker.player.status
		if prevActive != active {
			server.ticker.player.ticker.Reset(duration * time.Second)
			server.ticker.player.status = active

			logrus.WithFields(logrus.Fields{
				"clientCount": clientCount,
				"newDuration": duration,
			}).Info("Ticker duration changed")
		}

		_ = fetchPlayerState(server)
	}
}
