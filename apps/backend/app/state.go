package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type PlayerStateItemImage struct {
	URL string `json:"url"`
}

type PlayerStateItemArtist struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Images []PlayerStateItemImage `json:"images"`
}

type ResponseTypeArtists struct {
	Artists []PlayerStateItemArtist `json:"artists"`
}

type PlayerStateItemAlbum struct {
	Images []PlayerStateItemImage `json:"images"`
}

type PlayerStateItem struct {
	Name    string                  `json:"name"`
	Artists []PlayerStateItemArtist `json:"artists"`
	Album   PlayerStateItemAlbum    `json:"album"`
}

type PlayerState struct {
	IsPlaying bool            `json:"is_playing"`
	Item      PlayerStateItem `json:"item"`
}

func fetchPlayerState(server *Server) error {
	return server.session.WithClient(func(ctx context.Context, client *http.Client) error {
		resp, err := client.Get("https://api.spotify.com/v1/me/player")
		if err != nil {
			logrus.WithError(err).Errorf("failed to get player state")
			return err
		}
		defer func() { resp.Body.Close() }()

		if server.playerState == nil {
			var playerState PlayerState
			playerState = PlayerState{
				IsPlaying: false,
				Item: PlayerStateItem{
					Name: "Pream Pinbut",
					Album: PlayerStateItemAlbum{
						Images: []PlayerStateItemImage{
							{
								URL: "",
							},
						},
					},
					Artists: []PlayerStateItemArtist{
						{
							Name: "Pream Pinbut",
							Images: []PlayerStateItemImage{
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

		var respState PlayerState
		err = json.NewDecoder(resp.Body).Decode(&respState)
		if err != nil {
			server.playerState.IsPlaying = false
			return err
		}

		server.playerState.IsPlaying = respState.IsPlaying
		server.playerState.Item.Name = respState.Item.Name
		server.playerState.Item.Album.Images[0].URL = respState.Item.Album.Images[0].URL

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

		var artists ResponseTypeArtists
		err = json.NewDecoder(resp.Body).Decode(&artists)
		if err != nil {
			logrus.WithError(err).Errorf("failed to decode response body")
			return err
		}

		for _, artist := range artists.Artists {
			server.playerState.Item.Artists = append(server.playerState.Item.Artists, PlayerStateItemArtist{
				Name: artist.Name,
				Images: []PlayerStateItemImage{
					{
						URL: artist.Images[0].URL,
					},
				},
			})
		}

		return nil
	})
}
