package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Configuration struct {
	Socket string
	Path   string
}

func InitializeConfig() Configuration {
	config := Configuration{
		Socket: "127.0.0.1:8080",
	}

	flag.Func("a", "Specify the address you want to start the server at (e.g. 127.0.0.1:8888)", func(addr string) error {
		hp := strings.Split(addr, ":")
		if len(hp) != 2 {
			return errors.New("please specify both IP address and port! Example: 127.0.0.1:8888")
		}

		var host net.IP
		if hp[0] != "localhost" {
			host = net.ParseIP(hp[0])
		} else {
			host = []byte("127.0.0.1")
		}
		port, err := strconv.Atoi(hp[1])

		if err != nil {
			return fmt.Errorf("please enter a valid port! %v", err)
		}

		if host == nil || port < 1 || port > 65535 {
			return errors.New("please specify a valid address! Example: 127.0.0.1:8888")
		}

		config.Socket = addr
		return nil
	})
	flag.StringVar(&config.Path, "b", "", "Specify the path you want to shorten all URLs at")
	flag.Parse()

	return config
}
