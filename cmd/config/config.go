package config

import (
	"errors"
	"flag"
	"net"
	"regexp"
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
		Path:   "",
	}

	flag.Func("a", "Specify the address you want to start the server at (e.g. 127.0.0.1:8888)", func(addr string) error {
		hp := strings.Split(addr, ":")
		if len(hp) != 2 {
			return errors.New("please specify both the IP address and a port! Example: 127.0.0.1:8888")
		}

		var host string
		if parsed := net.ParseIP(hp[0]); parsed == nil && hp[0] != "localhost" && hp[0] != "" {
			return errors.New("please specify a valid IP:Port pair! Example: 127.0.0.1:8888")
		} else if hp[0] != "localhost" && hp[0] != "" {
			host = parsed.String()
		} else {
			host = "127.0.0.1"
		}

		if port, err := strconv.Atoi(hp[1]); err != nil || port < 1 || port > 65535 {
			return errors.New("please specify a valid port! Example: 127.0.0.1:8888")
		}

		config.Socket = host + ":" + hp[1]
		return nil
	})
	flag.Func("b", "Specify the path you want to save all the shortened URIs at", func(path string) error {
		re := regexp.MustCompile("[^A-Za-z0-9_.~-]+")
		if re.MatchString(path) {
			return errors.New("please specify a valid path! It can only contain alphanumeric characters, '.', '_', '~' and '-'")
		}

		config.Path = path
		return nil
	})
	flag.Parse()

	return config
}
