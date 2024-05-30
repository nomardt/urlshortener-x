package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/auth"
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

// Shortens url on the backend with the specified correlation ID.
// If correlation ID is set to "" then a random uuid will be used
func shortenURL(urlInput string, userID string, h *Handler, correlationID string) (string, error) {
	if urlInput == "" {
		return "", errors.New("no URL provided")
	}

	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	var u *urls.URL
	var err error
	// If a key is predefined then store all shortened URLs at that query path
	if h.Configuration.Path == "" {
		u, err = urls.NewURLWithoutKey(urlInput, correlationID)
	} else {
		u, err = urls.NewURL(urlInput, h.Configuration.Path, correlationID)
	}
	if err != nil {
		return "", err
	}

	// Stores the urls.URL object on the backend
	err = h.SaveURL(*u, userID)
	if err != nil {
		return "", err
	}
	logger.Log.Info("Shortened new URI", zap.String("data", urlInput), zap.String("date", time.Now().Format("2006/01/02")), zap.String("time", time.Now().Format("15:04:05")))

	return u.ID(), nil
}

// Shortens all urls received in JSON objects in request body
func (h *Handler) JSONPostBatch(w http.ResponseWriter, r *http.Request) {
	var clientInput []requestShortenBatchURLs
	if err := json.NewDecoder(r.Body).Decode(&clientInput); err != nil {
		http.Error(w, "You provided invalid JSON! Please specify correlation_id and original_url", http.StatusBadRequest)
		return
	}

	// Handle the session cookie
	jwtCookie := r.Header.Get("Authorization") // No need to check if cookie is present because it is done by middleware
	jwtCookie, _ = strings.CutPrefix(jwtCookie, "Bearer ")
	userID, err := auth.GetUserID(jwtCookie, h.Secret)
	if err != nil {
		logger.Log.Info("Couldn't decrypt the cookie", zap.String("jwt_session", jwtCookie), zap.Error(err))
		http.Error(w, "Unathorized", http.StatusUnauthorized)
	}

	var shortURLs []responseShortenBatchURLs
	for _, url := range clientInput {
		var shortURL string

		// Saving every JSON object with correlation_id and original_url fields on the backend separately
		key, err := shortenURL(url.OriginalURL, userID, h, url.CorrelationID)
		var errURINotUnique *urlsInfra.ErrURINotUnique
		if errors.As(err, &errURINotUnique) {
			// The JSON object contains original_url that's already been shortened, use the existing query path
			shortURL = "http://" + h.Configuration.ListenAddress + "/" + errURINotUnique.ExistingKey
		} else if errors.Is(err, urlsInfra.ErrCorIDNotUnique) {
			// The JSON object contains correlation_id that's already been taken
			http.Error(w, "The specified correlation ID is already present on the server! "+url.CorrelationID, http.StatusBadRequest)
			return
		} else if err != nil {
			// General error, abort
			http.Error(w, "Bad Request", http.StatusBadRequest)
			logger.Log.Info("Couldn't save URL sent in batch", zap.Error(err))
			return
		} else {
			// No error, construct the shortened URL with the returned key (query path)
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

// Shortens url received in JSON url field
func (h *Handler) JSONPostURI(w http.ResponseWriter, r *http.Request) {
	var clientInput requestShortenURL
	if err := json.NewDecoder(r.Body).Decode(&clientInput); err != nil {
		http.Error(w, "You provided invalid JSON! Please specify url", http.StatusBadRequest)
		return
	}

	// Handle the session cookie
	jwtCookie := r.Header.Get("Authorization") // No need to check if cookie is present because it is done by middleware
	jwtCookie, _ = strings.CutPrefix(jwtCookie, "Bearer ")
	userID, err := auth.GetUserID(jwtCookie, h.Secret)
	if err != nil {
		logger.Log.Info("Couldn't decrypt the cookie", zap.String("jwt_session", jwtCookie), zap.Error(err))
		http.Error(w, "Unathorized", http.StatusUnauthorized)
	}

	var shortURL string
	responseCode := http.StatusCreated

	// Saving the JSON object with url field on the backend
	key, err := shortenURL(clientInput.URL, userID, h, "")
	var errURINotUnique *urlsInfra.ErrURINotUnique
	if errors.As(err, &errURINotUnique) {
		// If the url's already been saved then use the existing (old) query path
		shortURL = "http://" + h.Configuration.ListenAddress + "/" + errURINotUnique.ExistingKey
		responseCode = http.StatusConflict
	} else if err != nil {
		// General error, abort
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	} else {
		// No error, construct the shortened URL with the returned key (query path)
		shortURL = "http://" + h.Configuration.ListenAddress + "/" + key
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

// Shortens url received in plaintext request body
func (h *Handler) PostURI(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Handle the session cookie
	jwtCookie := r.Header.Get("Authorization") // No need to check if cookie is present because it is done by middleware
	jwtCookie, _ = strings.CutPrefix(jwtCookie, "Bearer ")
	userID, err := auth.GetUserID(jwtCookie, h.Secret)
	if err != nil {
		logger.Log.Info("Couldn't decrypt the cookie", zap.String("Authorization", jwtCookie), zap.Error(err))
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}

	// Shorten the body on the backend
	id, err := shortenURL(string(body), userID, h, "")
	var errURINotUnique *urlsInfra.ErrURINotUnique
	if errors.As(err, &errURINotUnique) {
		w.WriteHeader(http.StatusConflict)
		shortURL := "http://" + h.Configuration.ListenAddress + "/" + errURINotUnique.ExistingKey
		w.Write([]byte(shortURL)) //nolint:all
		return
	} else if err != nil {
		http.Error(w, "Please provide a valid URI in request body", http.StatusBadRequest)
		logger.Log.Info("Couldn't shorten URL", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://" + h.Configuration.ListenAddress + "/" + id)) //nolint:all
}
