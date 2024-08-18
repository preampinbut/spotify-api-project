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
		if respState.Item == nil {
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
			return nil
		}
		var ids []spotify.ID
		for _, artist := range respState.CurrentlyPlaying.Item.Artists {
			ids = append(ids, artist.ID)
		}
		respArtists, err := client.GetArtists(ctx, ids...)
		if err != nil {
			logrus.WithError(err).Errorf("failed to get artists")
			return err
		}

		artists := []PlayerStateItemArtist{}
		for _, artist := range respArtists {
			artists = append(artists, PlayerStateItemArtist{
				Name:  artist.Name,
				Image: artist.Images[0].URL,
			})
		}
		var playerState PlayerState
		playerState = PlayerState{
			IsPlaying: respState.CurrentlyPlaying.Playing,
			Item: PlayerStateItem{
				Name:    respState.CurrentlyPlaying.Item.Name,
				Image:   respState.CurrentlyPlaying.Item.Album.Images[0].URL,
				Artists: artists,
			},
		}

		server.playerState = &playerState
		return nil
	})
}
