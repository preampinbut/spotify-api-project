package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type eventT struct{}

var Event eventT

func (eventT) WritePlayerState(w http.ResponseWriter, server *Server) {
	resp, err := json.Marshal(server.playerState)
	if err != nil {
		logrus.WithError(err).Errorf("failed to marshal player state")
		return
	}
	fmt.Fprintf(w, string(resp))
	w.(http.Flusher).Flush()
}
