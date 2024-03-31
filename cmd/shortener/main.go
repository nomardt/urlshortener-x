package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/nomardt/urlshortener-x/internal/idgenerator"
)

type uriStorage map[string]string

var storage uriStorage

func main() {
	storage = make(uriStorage)

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)

	return http.ListenAndServe(`:8081`, mux)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]

	if r.Method == http.MethodPost && id == "" {
		// Check the Content-Type value
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(w, "Invalid Content-Type! Please use text/plain", http.StatusBadRequest)
			return
		}

		// Get the body value
		body, err := io.ReadAll(r.Body)
		if err != nil {
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
	} else if r.Method == http.MethodGet {
		if url, exists := storage[id]; exists {
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "URL not found on the server!", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}
