package config

import (
	"errors"
	"flag"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Configuration struct {
	Socket string
	Path   string
}

var config = Configuration{
	Socket: "127.0.0.1:8080",
	Path:   "",
}

func setSocket(addr string) error {
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
}

func InitializeConfig() Configuration {
	flag.Func("a", "Specify the IP:PORT you want to start the server at (e.g. 127.0.0.1:8888)", setSocket)
	flag.Func("b", "Specify the full URI where you want to keep the shortened URIs at (e.g. http://localhost:8080/defaultPath)", func(urlRaw string) error {
		// Check if the flag value is an actual URL
		u, err := url.ParseRequestURI(urlRaw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || string(u.Host[0]) == "." || string(u.Host[len(u.Host)-1]) == "." {
			return errors.New("please enter a valid URL! Input example: http://localhost:8000/qsd54gFg")
		}
		urlSplit := strings.Split(urlRaw, "/")

		setSocket(urlSplit[2])

		if len(urlSplit) == 4 {
			re := regexp.MustCompile("[^A-Za-z0-9_.~-]+")
			if re.MatchString(urlSplit[3]) {
				return errors.New("please specify a valid path! It can only contain alphanumeric characters, '.', '_', '~' and '-'")
			}

			config.Path = urlSplit[3]
		}

		return nil
	})
	flag.Parse()

	return config
}
