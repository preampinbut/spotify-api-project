package app

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
)

type PlayerStateItemArtist struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type PlayerStateItem struct {
	Name    string                  `json:"name"`
	Image   string                  `json:"image"`
	Artists []PlayerStateItemArtist `json:"artists"`
}

type PlayerState struct {
	IsPlaying bool            `json:"is_playing"`
	Item      PlayerStateItem `json:"item"`
}

func fetchPlayerState(server *Server) {
	server.session.WithClient(func(ctx context.Context, client *spotify.Client) error {
		respState, err := client.PlayerState(ctx)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get player state")
			return err
		}

		if server.playerState == nil {
			var playerState PlayerState
			playerState = PlayerState{
				IsPlaying: false,
				Item: PlayerStateItem{
					Name:  "Pream Pinbut",
					Image: "",
					Artists: []PlayerStateItemArtist{
						{
							Name:  "Pream Pinbut",
							Image: "",
						},
					},
				},
			}
			server.playerState = &playerState
		}

		if respState.Item == nil {
			server.playerState.IsPlaying = false
			return nil
		}

		// Reuse or clear existing data
		server.playerState.IsPlaying = respState.CurrentlyPlaying.Playing
		server.playerState.Item.Name = respState.CurrentlyPlaying.Item.Name
		server.playerState.Item.Image = respState.CurrentlyPlaying.Item.Album.Images[0].URL

		// Clear existing artists slice
		server.playerState.Item.Artists = server.playerState.Item.Artists[:0]

		var ids []spotify.ID
		for _, artist := range respState.CurrentlyPlaying.Item.Artists {
			ids = append(ids, artist.ID)
		}

		respArtists, err := client.GetArtists(ctx, ids...)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get artists")
			return err
		}

		for _, artist := range respArtists {
			server.playerState.Item.Artists = append(server.playerState.Item.Artists, PlayerStateItemArtist{
				Name:  artist.Name,
				Image: artist.Images[0].URL,
			})
		}

		return nil
	})
}
