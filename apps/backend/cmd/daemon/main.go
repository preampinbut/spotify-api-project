// This example demonstrates how to authenticate with Spotify using the authorization code flow with PKCE.
// In order to run this example yourself, you'll need to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
package main

import (
	"backend/app"
	"backend/config"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	auth := app.NewAuth(spotifyauth.WithClientID(cfg.ClientId), spotifyauth.WithRedirectURL(config.RedirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadPlaybackState))
	session := app.NewSession(cfg, auth)
	done := make(chan int)
	app.NewOAuth2Server(session, done)
	url := app.AuthURL(session)
	fmt.Println(url)

	<-done

	session.WithClient(func(ctx context.Context, client *spotify.Client) error {
		user, err := client.CurrentUser(ctx)
		if err != nil {
			logrus.WithError(err).Errorf("could not get current user")
			return err
		}
		logrus.Infof("Current User: %s\n", user.DisplayName)
		return nil
	})
}
