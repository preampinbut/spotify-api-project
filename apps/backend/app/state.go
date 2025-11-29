package app

import (
	"context"
	"encoding/json"
	"io"
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
func fetchPlayerState(server *Server) error {
	return server.session.WithClient(func(ctx context.Context, client *http.Client) error {
		if server.playerState == nil {
			server.playerState = &PlaybackState{
				IsPlaying: false,
				Item: TrackObject{
					ID:   "",
					Name: "Pream Pinbut",
					Album: Album{
						Images: []Images{{URL: ""}},
					},
					Artists: []Artist{{
						ID:   "",
						Name: "Pream Pinbut",
						Images: []Images{{URL: ""}},
					}},
				},
			}
		}

		resp, err := client.Get("https://api.spotify.com/v1/me/player")
		if err != nil {
			logrus.WithError(err).Error("failed to get player state")
			return err
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				logrus.WithError(cerr).Warn("failed to close player state response body")
			}
		}()

		var respState PlaybackState
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&respState); err != nil {
			if err == io.EOF {
				server.playerState.IsPlaying = false
				return nil
			}
			server.playerState.IsPlaying = false
			logrus.WithError(err).Error("failed to decode player state")
			return err
		}

		if server.playerState.Item.ID != respState.Item.ID {
			logrus.Infof("update song: %s", respState.Item.Name)
		}

		*server.playerState = respState
		server.playerState.Item.Artists = server.playerState.Item.Artists[:0]

		var ids []string
		for _, artist := range respState.Item.Artists {
			ids = append(ids, artist.ID)
		}

		req, err := http.NewRequest("GET", "https://api.spotify.com/v1/artists", nil)
		if err != nil {
			logrus.WithError(err).Error("failed to create request for artists")
			return err
		}

		q := req.URL.Query()
		q.Set("ids", strings.Join(ids, ","))
		req.URL.RawQuery = q.Encode()

		resp, err = client.Do(req)
		if err != nil {
			logrus.WithError(err).Error("failed to get artists")
			return err
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				logrus.WithError(cerr).Warn("failed to close artists response body")
			}
		}()
		var artists ResponseArtist
		if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
			logrus.WithError(err).Error("failed to decode artists response body")
			return err
		}

		for _, artist := range artists.Artists {
			imageURL := ""
			if len(artist.Images) > 0 {
				imageURL = artist.Images[0].URL
			}
			server.playerState.Item.Artists = append(server.playerState.Item.Artists, Artist{
				ID:   artist.ID,
				Name: artist.Name,
				Images: []Images{{URL: imageURL}},
			})
		}

		return nil
	})
}
