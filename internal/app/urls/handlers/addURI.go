package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	"go.uber.org/zap"
)

type request struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

func shortenURL(urlInput string, h *Handler) (string, error) {
	if urlInput == "" {
		return "", errors.New("no URL provided")
	}

	var u *urls.URL
	var err error
	// If a key is predefined in config then store all shortened URLs at that path
	if h.Configuration.Path == "" {
		u, err = urls.NewURLWithoutID(urlInput)
	} else {
		u, err = urls.NewURL(urlInput, h.Configuration.Path)
	}
	if err != nil {
		return "", errors.New("invalid URL")
	}

	err = h.SaveURL(u)
	if err != nil {
		return "", errors.New("couldn't save")
	}
	logger.Log.Info("Shortened new URI", zap.String("data", urlInput), zap.String("date", time.Now().Format("2006/01/02")), zap.String("time", time.Now().Format("15:04:05")))

	return u.ID(), nil
}

func (h *Handler) JSONPostURI(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		http.Error(w, "Please use only \"Content-Type: application/json\" for this endpoint!", http.StatusUnsupportedMediaType)
		return
	}

	var clientInput request
	if err := json.NewDecoder(r.Body).Decode(&clientInput); err != nil {
		http.Error(w, "You provided invalid JSON!", http.StatusBadRequest)
		return
	}

	id, err := shortenURL(clientInput.URL, h)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resultingShortenedURL := "http://" + h.Configuration.ListenAddress + "/" + id
	serverOutput := &response{
		Result: resultingShortenedURL,
	}
	jsonResp, err := json.MarshalIndent(serverOutput, "", "	")
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		logger.Log.Debug("Couldn't create JSON", zap.String("error", err.Error()))
		return
	}
	_, err = w.Write([]byte(jsonResp))
	if err != nil {
		logger.Log.Debug("Couldn't send the response with shortened URL address", zap.String("error", err.Error()))
		return
	}
}

func (h *Handler) PostURI(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		http.Error(w, "Please use only \"Content-Type: text/plain\" for this endpoint!", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id, err := shortenURL(string(body), h)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte("http://" + h.Configuration.ListenAddress + "/" + id))
	if err != nil {
		logger.Log.Debug("Couldn't send the response with shortened URL address", zap.String("error", err.Error()))
		return
	}
}
