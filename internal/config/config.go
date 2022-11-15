package config

import (
	"fmt"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	freshly struct {
		conf string `yaml:"conf"`
	}
}

func New(filename string) (*Config, error) {
	var cfg Config

	if err := envconfig.Process("freshly", &cfg); err != nil {
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

	err = authenticate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func authenticate() error {
	fmt.Println("Authenticate with Spotify")

	return nil
}
