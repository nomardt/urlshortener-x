package config

import (
	"flag"
	"fmt"
	"os"
)

type Configuration struct {
	Socket string
	Path   string
}

var config = Configuration{
	Socket: "127.0.0.1:8080",
	Path:   "",
}

func LoadConfig() Configuration {
	flag.Func("a", "Specify the IP:PORT you want to start the server at (e.g. 127.0.0.1:8888)", setSocket)
	flag.Func("b", "Specify the full URI you want to keep the shortened URIs at (e.g. http://localhost:8080/defaultPath)", setURL)
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		err := setSocket(envServerAddress)
		if err != nil {
			fmt.Println(err)
		}
	} else if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		err := setURL(envBaseURL)
		if err != nil {
			fmt.Println(err)
		}
	}

	return config
}
