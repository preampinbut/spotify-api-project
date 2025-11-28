package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// eventT is a receiver for event-related methods.
type eventT struct{}

// Event is the global instance for event methods.
var Event eventT

// WritePlayerState serializes the server's player state and writes it as an SSE message.
func (eventT) WritePlayerState(w http.ResponseWriter, server *Server) {
	resp, err := json.Marshal(server.playerState)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal player state")
		return
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", resp); err != nil {
		logrus.WithError(err).Error("failed to write player state to response")
		return
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	} else {
		logrus.Error("response writer does not support flushing")
	}
}
