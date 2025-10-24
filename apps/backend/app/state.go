package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

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

func fetchPlayerState(server *Server, force bool) error {
	return server.session.WithClient(func(ctx context.Context, client *http.Client) error {
		if server.playerState == nil {
			var playerState PlaybackState
			playerState = PlaybackState{
				IsPlaying: false,
				Item: TrackObject{
					ID:   "",
					Name: "Pream Pinbut",
					Album: Album{
						Images: []Images{
							{
								URL: "",
							},
						},
					},
					Artists: []Artist{
						{
							ID:   "",
							Name: "Pream Pinbut",
							Images: []Images{
								{
									URL: "",
								},
							},
						},
					},
				},
			}
			server.playerState = &playerState
		}

		server.session.clientsMutex.Lock()
		clientCount := len(server.session.clients)
		server.session.clientsMutex.Unlock()

		if clientCount <= 0 && force == false {
			server.playerState.IsPlaying = false
			return nil
		}

		resp, err := client.Get("https://api.spotify.com/v1/me/player")
		if err != nil {
			logrus.WithError(err).Errorf("failed to get player state")
			return err
		}
		defer func() { resp.Body.Close() }()

		var respState PlaybackState
		err = json.NewDecoder(resp.Body).Decode(&respState)
		if err != nil {
			server.playerState.IsPlaying = false
			return err
		}

		if server.playerState.Item.ID != respState.Item.ID {
			logrus.Infof("update song: %s", respState.Item.Name)
		}

		server.playerState.IsPlaying = respState.IsPlaying
		server.playerState.ProgressMs = respState.ProgressMs

		server.playerState.Item.ID = respState.Item.ID
		server.playerState.Item.Name = respState.Item.Name
		server.playerState.Item.Album.Images[0].URL = respState.Item.Album.Images[0].URL
		server.playerState.Item.DurationMs = respState.Item.DurationMs

		server.playerState.Item.Artists = server.playerState.Item.Artists[:0]

		var ids []string
		for _, artist := range respState.Item.Artists {
			ids = append(ids, artist.ID)
		}

		req, err := http.NewRequest("GET", "https://api.spotify.com/v1/artists", nil)
		if err != nil {
			logrus.WithError(err).Errorf("failed to create request")
			return err
		}

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
			logrus.WithError(err).Errorf("failed to decode response body")
			return err
		}

		for _, artist := range artists.Artists {
			var imageURL string
			if len(artist.Images) > 0 {
				imageURL = artist.Images[0].URL
			}
			server.playerState.Item.Artists = append(server.playerState.Item.Artists, Artist{
				ID:   artist.ID,
				Name: artist.Name,
				Images: []Images{
					{
						URL: imageURL,
					},
				},
			})
		}

		return nil
	})
}
