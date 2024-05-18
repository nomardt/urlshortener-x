package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
	"go.uber.org/zap"
)

type requestShortenURL struct {
	URL string `json:"url"`
}

type requestShortenBatchURLs struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type responseShortenURL struct {
	Result string `json:"result"`
}

type responseShortenBatchURLs struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func shortenURL(urlInput string, h *Handler, correlationID string) (string, error) {
	if urlInput == "" {
		return "", errors.New("no URL provided")
	}

	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	var u *urls.URL
	var err error
	// If a key is predefined in config then store all shortened URLs at that path
	if h.Configuration.Path == "" {
		u, err = urls.NewURLWithoutKey(urlInput, correlationID)
	} else {
		u, err = urls.NewURL(urlInput, h.Configuration.Path, correlationID)
	}
	if err != nil {
		return "", err
	}

	err = h.SaveURL(u)
	if err != nil {
		return "", err
	}
	logger.Log.Info("Shortened new URI", zap.String("data", urlInput), zap.String("date", time.Now().Format("2006/01/02")), zap.String("time", time.Now().Format("15:04:05")))

	return u.ID(), nil
}

func (h *Handler) JSONPostBatch(w http.ResponseWriter, r *http.Request) {
	var clientInput []requestShortenBatchURLs
	if err := json.NewDecoder(r.Body).Decode(&clientInput); err != nil {
		http.Error(w, "You provided invalid JSON! Please specify correlation_id and original_url", http.StatusBadRequest)
		return
	}

	var shortURLs []responseShortenBatchURLs
	for _, url := range clientInput {
		var shortURL string

		key, err := shortenURL(url.OriginalURL, h, url.CorrelationID)
		var errURINotUnique *urlsInfra.ErrURINotUnique
		if errors.As(err, &errURINotUnique) {
			shortURL = "http://" + h.Configuration.ListenAddress + "/" + errURINotUnique.ExistingKey
		} else if errors.Is(err, urlsInfra.ErrCorIDNotUnique) {
			http.Error(w, "The specified correlation ID is already present on the server! "+url.CorrelationID, http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			logger.Log.Info("Couldn't save URL sent in batch", zap.Error(err))
			return
		} else {
			shortURL = "http://" + h.Configuration.ListenAddress + "/" + key
		}

		newShortenedURL := responseShortenBatchURLs{
			CorrelationID: url.CorrelationID,
			ShortURL:      shortURL,
		}

		shortURLs = append(shortURLs, newShortenedURL)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	jsonResp, err := json.MarshalIndent(shortURLs, "", "	")
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		logger.Log.Info("Couldn't create JSON", zap.String("error", err.Error()))
		return
	}
	_, err = w.Write([]byte(jsonResp))
	if err != nil {
		logger.Log.Info("Couldn't send the response with shortened URL address", zap.String("error", err.Error()))
		return
	}
}

func (h *Handler) JSONPostURI(w http.ResponseWriter, r *http.Request) {
	var clientInput requestShortenURL
	if err := json.NewDecoder(r.Body).Decode(&clientInput); err != nil {
		http.Error(w, "You provided invalid JSON! Please specify url", http.StatusBadRequest)
		return
	}

	var shortURL string
	responseCode := http.StatusCreated

	id, err := shortenURL(clientInput.URL, h, "")
	var errURINotUnique *urlsInfra.ErrURINotUnique
	if errors.As(err, &errURINotUnique) {
		shortURL = "http://" + h.Configuration.ListenAddress + "/" + errURINotUnique.ExistingKey
		responseCode = http.StatusConflict
	} else if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	} else {
		shortURL = "http://" + h.Configuration.ListenAddress + "/" + id
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)

	serverOutput := &responseShortenURL{
		Result: shortURL,
	}
	jsonResp, err := json.MarshalIndent(serverOutput, "", "	")
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		logger.Log.Info("Couldn't create JSON", zap.String("error", err.Error()))
		return
	}
	_, err = w.Write([]byte(jsonResp))
	if err != nil {
		logger.Log.Info("Couldn't send the response with shortened URL address", zap.String("error", err.Error()))
		return
	}
}

func (h *Handler) PostURI(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id, err := shortenURL(string(body), h, "")
	var errURINotUnique *urlsInfra.ErrURINotUnique
	if errors.As(err, &errURINotUnique) {
		shortURL := "http://" + h.Configuration.ListenAddress + "/" + errURINotUnique.ExistingKey
		http.Error(w, shortURL, http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, "Please provide a valid URI in request body", http.StatusBadRequest)
		logger.Log.Info("Couldn't shorten URL", zap.Error(err))
		return
	}
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte("http://" + h.Configuration.ListenAddress + "/" + id))
	if err != nil {
		logger.Log.Info("Couldn't send the response with shortened URL address", zap.String("error", err.Error()))
		return
	}
}
