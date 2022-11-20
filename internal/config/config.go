package config

import (
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
	"github.com/timbrammer910/freshly/internal/authenticate"
	"github.com/zmb3/spotify/v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AccessToken     string `yaml:"-" envconfig:"ACCESS_TOKEN" required:"true"`
	RefreshToken    string `yaml:"-" envconfig:"REFRESH_TOKEN" required:"true"`
	Token           string `yaml:"-" envconfig:"SPOTIFY_ID" required:"true"`
	SignatureSecret string `yaml:"-" envconfig:"SPOTIFY_SECRET" required:"true"`
	SpotifyClient   *spotify.Client
}

func New(filename string) (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	// Read the config file contents
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Load the file contents into the config struct and return it
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	spotifyClient, err := authenticate.GetAccessToken(cfg.AccessToken, cfg.RefreshToken)
	if err != nil {
		return nil, err
	}

	cfg.SpotifyClient = spotifyClient

	return &cfg, nil
}
