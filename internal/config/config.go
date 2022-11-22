package config

import (
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	// AccessToken   string `yaml:"-" envconfig:"ACCESS_TOKEN" required:"true"`
	RefreshToken  string `yaml:"-" envconfig:"REFRESH_TOKEN" required:"true"`
	SpotifyID     string `yaml:"-" envconfig:"SPOTIFY_ID" required:"true"`
	SpotifySecret string `yaml:"-" envconfig:"SPOTIFY_SECRET" required:"true"`
	Spotify       struct {
		Playlists []string `yaml:"playlists"`
		MaxAge    int      `yaml:"maxAge"`
		MinSongs  int      `yaml:"minSongs"`
	} `yaml:"spotify"`
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

	return &cfg, nil
}
