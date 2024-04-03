package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nomardt/urlshortener-x/internal/idgenerator"
)

// Declaring the map to store shortened URLs in
type URIStorage map[string]string

var storage URIStorage

func newURIHandler(storage URIStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the body value
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Check if the body value is an actual URL
		u, err := url.ParseRequestURI(string(body))
		hostname := u.Host
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || !strings.Contains(hostname, ".") || string(hostname[0]) == "." || string(hostname[len(hostname)-1]) == "." {
			http.Error(w, "Please enter a valid URL", http.StatusBadRequest)
			return
		}

		// Add the received URL to storage
		randomID := idgenerator.GenerateRandomID(8)
		storage[randomID] = u.String()

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + randomID))
	}
}

func getURIHandler(storage URIStorage, unitTestID ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		// This is necessary because chi.URLParam doesn't parse IDs from unit tests for some reason
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
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("text/plain"))

	r.Post("/", newURIHandler(storage))
	r.Get("/{id}", getURIHandler(storage))

	log.Fatal(http.ListenAndServe(`:8080`, r))
}
