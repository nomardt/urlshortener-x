package urls

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
)

type InMemoryRepo struct {
	urls map[string]string
	file string
	mu   sync.Mutex
}

type urlInFile struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
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

	if file, err := os.OpenFile(r.file, os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		jsonURL := &urlInFile{
			UUID:        uuid.New(),
			ShortURL:    url.ID(),
			OriginalURL: url.LongURL(),
		}
		data, err := json.Marshal(jsonURL)
		if err != nil {
			logger.Log.Info("Couldn't store the shortened URL in the file", zap.String("error", err.Error()))
			return nil
		}
		data = append(data, '\n')

		file.Write(data)
		file.Close()
	}

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

// Load previously shortened URLs from the file specified in config
func (r *InMemoryRepo) LoadStoredURLs(config conf.Configuration) error {
	r.file = config.StorageFile

	file, err := os.Open(config.StorageFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		url := &urlInFile{}
		if err := json.Unmarshal(line, url); err != nil {
			return err
		}
		r.urls[url.ShortURL] = url.OriginalURL
	}

	return nil
}
