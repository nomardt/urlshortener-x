package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/nomardt/urlshortener-x/internal/infra/auth"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
)

type responseUserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Redirects to the URL shortened with the key specified in query path
func (h *Handler) GetURI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if url, err := h.GetURL(id); !errors.Is(err, urlsInfra.ErrNotFoundURL) {
		w.Header().Set("Location", url)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL with the specified ID:"+id+" was not found on the server!", http.StatusBadRequest)
	}
}

// Get all URLs associated with the current user
func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	// Handle the session cookie
	jwtCookie := r.Header.Get("Authorization") // No need to check if cookie is present because it is done by middleware
	jwtCookie, _ = strings.CutPrefix(jwtCookie, "Bearer ")
	userID, err := auth.GetUserID(jwtCookie, h.Secret)
	if err != nil {
		logger.Log.Info("Couldn't decrypt the cookie", zap.String("jwt_session", jwtCookie), zap.Error(err))
		http.Error(w, "Unathorized", http.StatusUnauthorized)
	}

	allURLsMap, err := h.GetAllUserURLs(userID)
	if err != nil {
		http.Error(w, "Unathorized", http.StatusUnauthorized)
		return
	}

	// The user has not shortened any URLs; return 204
	if len(allURLsMap) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	allURLs := []responseUserURL{}
	for shortURL, originalURL := range allURLsMap {
		shortURL = "http://" + h.ListenAddress + "/" + shortURL
		allURLs = append(allURLs, responseUserURL{ShortURL: shortURL, OriginalURL: originalURL})
	}

	allURLsJSON, err := json.MarshalIndent(allURLs, "", "	")
	if err != nil {
		logger.Log.Info("Couldn't create a JSON response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(allURLsJSON) //nolint:all
}
