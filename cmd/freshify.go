package main

import (
	"log"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/creasty/defaults"
	"github.com/timbrammer910/freshly/internal/authenticate"
	"github.com/timbrammer910/freshly/internal/config"
	"github.com/timbrammer910/freshly/internal/spotify"
)

var args struct {
	ConfigFilename string `arg:"--config" default:"conf/freshify.yml" help:"path to config file"`
	Auth           bool   `help:"run Spotify OAuth2 authorizer"`
}

func main() {
	if err := defaults.Set(&args); err != nil {
		panic(err)
	}
	arg.MustParse(&args)

	if args.Auth {
		if err := authenticate.Authenticate(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	cfg, err := config.New(args.ConfigFilename)
	if err != nil {
		log.Fatal(err)
	}

	s := spotify.New(cfg)

	if err := s.Freshify(cfg.Spotify.Playlists, cfg.Spotify.MaxAge, cfg.Spotify.MinSongs); err != nil {
		log.Fatal(err)
	}

}
