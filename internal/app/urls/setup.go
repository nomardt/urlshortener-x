package urls

import (
	"github.com/go-chi/chi/v5"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
)

func Setup(router *chi.Mux, urlsRepo handlers.Repository, defaultRoute string) {
	handler := handlers.NewHandler(urlsRepo, defaultRoute)

	router.Post("/", handler.PostURI)
	router.Get("/{id}", handler.GetURI)
}
