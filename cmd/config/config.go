package config

import (
	"flag"
	"os"
)

type Configuration struct {
	ListenAddress string
	Path          string
}

var config = Configuration{
	ListenAddress: "127.0.0.1:8080",
	Path:          "",
}

func LoadConfig() (Configuration, error) {
	flag.Func("a", "Specify the IP:PORT you want to start the server at (e.g. 127.0.0.1:8888)", setListenAddress)
	flag.Func("b", "Specify the full URI you want to keep the shortened URIs at (e.g. http://localhost:8888/defaultPath)", setURL)
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		err := setListenAddress(envServerAddress)
		if err != nil {
			return config, err
		}
	} else if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		err := setURL(envBaseURL)
		if err != nil {
			return config, err
		}
	}

	return config, nil
}
