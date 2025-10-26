package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// Data models matching Spotify API responses

type Images struct {
	URL string `json:"url"`
}

type Artist struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Images []Images `json:"images"`
}

type ResponseArtist struct {
	Artists []Artist `json:"artists"`
}

type Album struct {
	ID     string   `json:"id"`
	Images []Images `json:"images"`
}

type TrackObject struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	Album      Album    `json:"album"`
	DurationMs int      `json:"duration_ms"`
}

type PlaybackState struct {
	IsPlaying  bool        `json:"is_playing"`
	ProgressMs int         `json:"progress_ms"`
	Item       TrackObject `json:"item"`
}

// fetchPlayerState retrieves the current playback state from the Spotify API.
// It initializes a default state if none exists and optionally fetches data,
// even if no clients are listening (if force is true).
func fetchPlayerState(server *Server, force bool) error {
	// server.session.WithClient wraps the HTTP request with the necessary
	// logic to ensure an authenticated and fresh client is used (e.g., handling token refresh).
	return server.session.WithClient(func(ctx context.Context, client *http.Client) error {
		// Initialize a default state if the server's player state is nil.
		// This provides a fallback/placeholder display when no data is available.
		if server.playerState == nil {
			playerState := PlaybackState{
				IsPlaying: false,
				Item: TrackObject{
					ID:   "",
					Name: "Pream Pinbut", // Placeholder name
					Album: Album{
						Images: []Images{
							{URL: ""},
						},
					},
					Artists: []Artist{
						{
							ID:   "",
							Name: "Pream Pinbut", // Placeholder artist
							Images: []Images{
								{URL: ""},
							},
						},
					},
				},
			}
			server.playerState = &playerState
		}

		// Check the number of connected clients. If there are none and
		// the fetch is not forced, we stop playback locally to save API calls.
		server.session.clientsMutex.Lock()
		clientCount := len(server.session.clients)
		server.session.clientsMutex.Unlock()

		if clientCount <= 0 && force == false {
			server.playerState.IsPlaying = false
			return nil
		}

		// 1. Fetch the current player state
		resp, err := client.Get("https://api.spotify.com/v1/me/player")
		if err != nil {
			logrus.WithError(err).Errorf("failed to get player state")
			return err
		}
		defer func() { resp.Body.Close() }()

		var respState PlaybackState
		err = json.NewDecoder(resp.Body).Decode(&respState)
		if err != nil {
			// If decoding fails (e.g., empty body or bad JSON), assume playback is stopped.
			server.playerState.IsPlaying = false
			return err
		}

		// Log if the song has changed.
		if server.playerState.Item.ID != respState.Item.ID {
			logrus.Infof("update song: %s", respState.Item.Name)
		}

		// Update basic state properties.
		server.playerState.IsPlaying = respState.IsPlaying
		server.playerState.ProgressMs = respState.ProgressMs

		// Update track item details.
		server.playerState.Item.ID = respState.Item.ID
		server.playerState.Item.Name = respState.Item.Name

		// Album images array is assumed to exist and have at least one element.
		server.playerState.Item.Album.ID = respState.Item.Album.ID
		server.playerState.Item.Album.Images[0].URL = respState.Item.Album.Images[0].URL
		server.playerState.Item.DurationMs = respState.Item.DurationMs

		// Clear existing artists before fetching detailed information.
		server.playerState.Item.Artists = server.playerState.Item.Artists[:0]

		// 2. Fetch detailed artist information (including images).
		// The player state endpoint often lacks full image details for artists.
		var ids []string
		for _, artist := range respState.Item.Artists {
			ids = append(ids, artist.ID)
		}

		req, err := http.NewRequest("GET", "https://api.spotify.com/v1/artists", nil)
		if err != nil {
			logrus.WithError(err).Errorf("failed to create request for artists")
			return err
		}

		// Batch artist request: set the query parameter 'ids' with comma-separated IDs.
		q := req.URL.Query()
		q.Set("ids", strings.Join(ids, ","))
		req.URL.RawQuery = q.Encode()

		resp, err = client.Do(req)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get artists")
			return err
		}
		defer func() { resp.Body.Close() }()

		var artists ResponseArtist
		err = json.NewDecoder(resp.Body).Decode(&artists)
		if err != nil {
			logrus.WithError(err).Errorf("failed to decode artists response body")
			return err
		}

		// Update the player state with the detailed artist information (especially images).
		for _, artist := range artists.Artists {
			var imageURL string
			if len(artist.Images) > 0 {
				imageURL = artist.Images[0].URL
			}
			server.playerState.Item.Artists = append(server.playerState.Item.Artists, Artist{
				ID:   artist.ID,
				Name: artist.Name,
				Images: []Images{
					{URL: imageURL},
				},
			})
		}

		return nil
	})
}
