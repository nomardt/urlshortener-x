package config

import (
	"flag"
	"os"
)

type DB struct {
	User     string
	Password string
	Host     string
	Port     string
	DBname   string
	SSLmode  string
}

type Configuration struct {
	ListenAddress string
	Path          string
	StorageFile   string
	DB            DB
	Secret        string
}

const hardcodedSecret = "CHANGEMEPLZ!"

var config = Configuration{
	ListenAddress: "127.0.0.1:8080",
	Path:          "",
	StorageFile:   "",
	DB:            DB{SSLmode: "disable"},
	Secret:        hardcodedSecret,
}

func LoadConfig() (Configuration, error) {
	flag.Func("a", "Specify the IP:PORT you want to start the server at (e.g. 127.0.0.1:8888)", setListenAddress)
	flag.Func("b", "[DEBUG] Specify the full URI you want to access the shortened URIs at (e.g. http://localhost:8888/defaultPath)", setURL)
	flag.Func("d", "Specify the Database Source Name (e.g. postgres://username:password@localhost:5432/mydatabase?sslmode=disable)", parseDSN)
	flag.StringVar(&config.StorageFile, "f", "/tmp/short-url-db.json", "Specify the file where shortened URLs will be stored (default: /tmp/short-url-db.json)")
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

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		err := parseDSN(envDSN)
		if err != nil {
			return config, err
		}
	}

	if envStorageFile, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		config.StorageFile = envStorageFile
	}

	return config, nil
}
