package urls

import (
	"errors"
	"sync"

	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

type InMemoryRepo struct {
	urls map[string]string
	mu   sync.RWMutex
}

var (
	ErrNotFoundURL = errors.New("the URL with the specified id was not found")
)

// Create a new Repo which consists of urls map[string]string
func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		urls: make(map[string]string),
	}
}

// Add the specified URL to the Repo
func (r *InMemoryRepo) SaveURL(url *urlsDomain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[url.ID()] = url.LongURL()

	return nil
}

// Check if there is a URL stored in the Repo with the specified ID
func (r *InMemoryRepo) GetURL(id *string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if url := r.urls[*id]; url != "" {
		return url, nil
	} else {
		return "", ErrNotFoundURL
	}
}
