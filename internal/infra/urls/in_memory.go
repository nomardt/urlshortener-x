package urls

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"

	"go.uber.org/zap"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
)

type InMemoryRepo struct {
	urls []urlInFile
	file string
	mu   sync.Mutex
}

type urlInFile struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
}

// Create a new Repo which consists of urls map[string]string
func NewInMemoryRepo(config conf.Configuration) *InMemoryRepo {
	inMemoryRepo := &InMemoryRepo{
		urls: make([]urlInFile, 0),
	}
	if err := inMemoryRepo.loadStoredURLs(config); err != nil {
		logger.Log.Info("Couldn't recover any previously shortened URLs!", zap.String("error", err.Error()))

		if _, err := os.Create(config.StorageFile); err != nil {
			logger.Log.Info("No file will be created to store shortened URLs")
		}
	}

	return inMemoryRepo
}

// Add the specified URL to the Repo
func (r *InMemoryRepo) SaveURL(url *urlsDomain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Checking if the provided correlation ID is unique
	for _, savedURL := range r.urls {
		if savedURL.CorrelationID == url.CorrelationID() {
			logger.Log.Info("The specified correlation ID already exists", zap.String("correlation_id", url.CorrelationID()))
			return ErrCorIDNotUnique
		}
	}

	r.urls = append(r.urls, urlInFile{url.CorrelationID(), url.ID(), url.LongURL()})

	// Saving the new URL on the hard drive
	if file, err := os.OpenFile(r.file, os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		jsonURL := &urlInFile{
			CorrelationID: url.CorrelationID(),
			ShortURL:      url.ID(),
			OriginalURL:   url.LongURL(),
		}
		data, err := json.Marshal(jsonURL)
		if err != nil {
			logger.Log.Info("Couldn't store the shortened URL in the file", zap.String("error", err.Error()))
			return nil
		}
		data = append(data, '\n')

		_, _ = file.Write(data)
		file.Close()
	}

	return nil
}

// Check if there is a URL stored in the Repo with the specified ID
func (r *InMemoryRepo) GetURL(id *string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range r.urls {
		if url.ShortURL == *id && url.OriginalURL != "" {
			return url.OriginalURL, nil
		}
	}

	return "", ErrNotFoundURL
}

func (r *InMemoryRepo) Ping(_ context.Context) error {
	if _, err := os.Stat(r.file); err != nil {
		return err
	} else {
		return nil
	}
}

// Load previously shortened URLs from the file specified in config
func (r *InMemoryRepo) loadStoredURLs(config conf.Configuration) error {
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

		r.urls = append(r.urls, *url)
	}

	return nil
}
