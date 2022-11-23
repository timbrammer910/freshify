package spotify

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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
	log.Infoln("you are logged in as:", user.ID)

	s.client = client
	s.user = user.ID

	return s
}

func (s *Spotify) Freshify(playlists []string, age int, min int) error {
	playlistIDs, err := s.filterPlaylists(playlists, min)
	if err != nil {
		return err
	}

	if err := s.processPlaylists(playlistIDs, age, min); err != nil {
		return err
	}

	return nil
}

func (s *Spotify) filterPlaylists(playlists []string, min int) ([]spotify.ID, error) {
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
				if int(e.Tracks.Total) <= min {
					log.Infof("playlist - %s contains too few tracks, skipping...\n", strings.ToLower(e.Name))
					found = true
					break
				}

				if e.Owner.ID == s.user {
					userPlaylists = append(userPlaylists, e.ID)
					found = true
					break
				}
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

func (s *Spotify) processPlaylists(playlistIDs []spotify.ID, age int, min int) error {
	log.Infoln("processing playlists...")

	for _, playlist := range playlistIDs {
		tracks, err := s.getTracks(playlist)
		if err != nil {
			return err
		}

		err = s.processTracks(tracks, playlist, age, min)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Spotify) getTracks(playlistID spotify.ID) ([]spotify.PlaylistItem, error) {
	var tracks []spotify.PlaylistItem

	offset := 0
	limit := 50
	for {
		page, err := s.client.GetPlaylistItems(context.Background(), playlistID, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, page.Items...)
		offset = offset + limit

		if page.Next == "" {
			break
		}
	}

	return tracks, nil
}

func (s *Spotify) processTracks(tracks []spotify.PlaylistItem, playlist spotify.ID, age int, min int) error {
	sort.Slice(tracks, func(l, g int) bool {
		lesser, _ := time.Parse(spotify.TimestampLayout, tracks[l].AddedAt)
		greater, _ := time.Parse(spotify.TimestampLayout, tracks[g].AddedAt)

		return lesser.Unix() < greater.Unix()
	})

	var deletes []spotify.ID
	for _, track := range tracks {
		if tooOld(track, age) {
			log.Infof("track '%s' is too old\n", strings.ToLower(track.Track.Track.Name))
			deletes = append(deletes, track.Track.Track.ID)
			if len(deletes) >= len(tracks)-min {
				log.Infof("min playlist length reached")
				break
			}
		}
	}

	_, err := s.client.RemoveTracksFromPlaylist(context.Background(), playlist, deletes...)
	if err != nil {
		return err
	}

	return nil
}

func tooOld(track spotify.PlaylistItem, age int) bool {
	days, _ := time.ParseDuration(fmt.Sprintf("%dh", age*24))
	added, _ := time.Parse(spotify.TimestampLayout, track.AddedAt)
	now := time.Now()

	if now.Unix()-added.Unix() > int64(days.Seconds()) {
		return true
	}

	return false
}
