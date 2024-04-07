package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/idgenerator"
)

// Declaring the map to store the shortened URIs at
type URIStorage map[string]string

var storage URIStorage
var conf config.Configuration

func newURIHandler(storage URIStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		u, err := url.ParseRequestURI(string(body))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || string(u.Host[0]) == "." || string(u.Host[len(u.Host)-1]) == "." {
			http.Error(w, "Please enter a valid URL", http.StatusBadRequest)
			return
		}

		var key string
		if conf.Path == "" {
			key = idgenerator.GenerateRandomID(8)
		} else {
			key = conf.Path
		}
		storage[key] = u.String()

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte("http://" + conf.Socket + "/" + key))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getURIHandler(storage URIStorage, unitTestID ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		// If it's a unit test then the id is provided as an argument
		if len(unitTestID) > 0 {
			id = unitTestID[0]
		}

		if url, exists := storage[id]; exists {
			w.Header().Set("Location", url)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "URL not found on the server!", http.StatusBadRequest)
		}
	}
}

func main() {
	storage = make(URIStorage)
	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("text/plain"))

	router.Post("/", newURIHandler(storage))
	router.Get("/{id}", getURIHandler(storage))

	conf = config.LoadConfig()

	fmt.Println("Server started at:", conf.Socket)
	log.Fatal(http.ListenAndServe(conf.Socket, router))
}
