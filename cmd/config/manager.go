package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidAddressPair = errors.New("please specify a valid IP:Port pair! Example: 127.0.0.1:8888")
	ErrInvalidPort        = errors.New("please specify a valid port! Example: 127.0.0.1:8888")
	ErrInvalidURL         = errors.New("please enter a valid URL! Input example: http://localhost:8888/qsd54gFg")
	ErrInvalidPath        = errors.New("please specify a valid path! It can only contain alphanumeric characters, '.', '_', '~' and '-'")
	ErrNestedDirs         = errors.New("unsupported URL structure! No nested directories are allowed. Valid input example: http://localhost:8888/qsd54gFg")
)

func setListenAddress(addr string) error {
	hp := strings.Split(addr, ":")
	if len(hp) != 2 {
		return ErrInvalidAddressPair
	}

	var host string
	if parsed := net.ParseIP(hp[0]); parsed == nil && hp[0] != "localhost" && hp[0] != "" {
		return ErrInvalidAddressPair
	} else if hp[0] == "localhost" || hp[0] == "" {
		host = "127.0.0.1"
	} else {
		host = parsed.String()
	}

	if port, err := strconv.Atoi(hp[1]); err != nil || port < 1 || port > 65535 {
		return ErrInvalidPort
	}

	config.ListenAddress = host + ":" + hp[1]
	return nil
}

func setURL(urlRaw string) error {
	u, err := url.ParseRequestURI(urlRaw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || string(u.Host[0]) == "." || string(u.Host[len(u.Host)-1]) == "." {
		return ErrInvalidURL
	}
	urlSplit := strings.Split(urlRaw, "/")

	err = setListenAddress(urlSplit[2])
	if err != nil {
		fmt.Println(err)
	}

	if len(urlSplit) == 4 {
		re := regexp.MustCompile("[^A-Za-z0-9_.~-]+")
		if re.MatchString(urlSplit[3]) {
			return ErrInvalidPath
		}

		config.Path = urlSplit[3]
	} else if len(urlSplit) > 4 {
		return ErrNestedDirs
	}

	return nil
}
