package main

import (
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal"
)

func main() {
	config, err := conf.LoadConfig()
	if err != nil {
		panic(err)
	}

	if err := internal.Run(config); err != nil {
		panic(err)
	}
}
