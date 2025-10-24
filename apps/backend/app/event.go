package app

import (
	"encoding/json" // Package to encode Go data structures into JSON
	"fmt"           // Package implementing formatted I/O (used for writing the SSE message)
	"net/http"      // Package for HTTP client and server implementations

	"github.com/sirupsen/logrus" // Logging library
)

// eventT is an empty struct used as a receiver for event-related methods.
type eventT struct{}

// Event is the globally accessible instance for event-related methods.
var Event eventT

// WritePlayerState serializes the server's player state and writes it to the
// HTTP response writer as a Server-Sent Event (SSE) message.
func (eventT) WritePlayerState(w http.ResponseWriter, server *Server) {
	// Marshal the server's player state into JSON bytes.
	resp, err := json.Marshal(server.playerState)
	if err != nil {
		// Log the error if JSON marshaling fails and exit the function.
		logrus.WithError(err).Errorf("failed to marshal player state")
		return
	}

	// Write the JSON data in the standard Server-Sent Events (SSE) format:
	// "data: [json_payload]\n\n"
	fmt.Fprintf(w, "data: %s\n\n", string(resp))

	// Flush the response writer to ensure the data is immediately sent to the client.
	// This is crucial for real-time updates like SSE.
	w.(http.Flusher).Flush()
}
