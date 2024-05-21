package main

import (
	"log"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal"
)

func main() {
	config, err := conf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := internal.Run(config); err != nil {
		log.Fatal(err)
	}
}
