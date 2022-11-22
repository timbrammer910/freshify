package spotify

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/timbrammer910/freshly/internal/authenticate"
	"github.com/timbrammer910/freshly/internal/config"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type Spotify struct {
	client *spotify.Client
	user   string
}

func New(cfg *config.Config) *Spotify {
	s := &Spotify{}

	token, err := authenticate.RefreshToken(cfg.RefreshToken, cfg.SpotifyID, cfg.SpotifySecret)
	if err != nil {
		log.Fatalf("token refresh failed: %v\n", err)
	}

	tok := &oauth2.Token{
		AccessToken: token,
	}

	// build client with refresh token and test
	client := spotify.New(authenticate.SpotifyAuth.Client(context.Background(), tok))
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	s.client = client
	s.user = user.ID

	return s
}

func (s *Spotify) Freshify(playlists []string, age int, min int) error {
	playlistIDs, err := s.filterPlaylists(playlists)
	if err != nil {
		return err
	}

	fmt.Printf("Playlist IDs: %v", playlistIDs)

	// if err := s.processPlaylists(playlistIDs, age, min); err != nil {
	// 	return err
	// }

	return nil
}

func (s *Spotify) filterPlaylists(playlists []string) ([]spotify.ID, error) {
	var userPlaylists []spotify.ID

	list, err := s.getPlaylists()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no playlists found")
	}

	for _, playlist := range playlists {
		var found bool
		for _, e := range list {
			if playlist == e.Name {
				if e.Owner.ID == s.user {
					userPlaylists = append(userPlaylists, e.ID)
					found = true
				}
				break
			}
		}
		if !found {
			log.Errorf("unable to find playlist: %s\n", playlist)
		}
	}

	return userPlaylists, nil
}

func (s *Spotify) getPlaylists() ([]spotify.SimplePlaylist, error) {
	var playlists []spotify.SimplePlaylist

	offset := 0
	limit := 50
	for {
		page, err := s.client.GetPlaylistsForUser(context.Background(), s.user, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		playlists = append(playlists, page.Playlists...)
		offset = offset + limit

		if page.Next == "" {
			break
		}
	}

	return playlists, nil
}
