package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
)

func (h *Handler) GetURI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if url, err := h.GetURL(&id); !errors.Is(err, urlsInfra.ErrNotFoundURL) {
		w.Header().Set("Location", url)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL with the specified ID was not found on the server!", http.StatusBadRequest)
	}
}
