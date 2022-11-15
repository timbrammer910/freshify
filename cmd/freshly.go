package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	"github.com/creasty/defaults"
	"github.com/timbrammer910/freshly/internal/config"
)

var args struct {
	ConfigFilename string `arg:"--config" default:"conf/freshly.yml" help:"path to config file"`
}

func main() {
	if err := defaults.Set(&args); err != nil {
		panic(err)
	}
	arg.MustParse(&args)

	_, err := config.New(args.ConfigFilename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("played make believe")
}
