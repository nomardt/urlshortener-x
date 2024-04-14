package urls

import (
	"github.com/go-chi/chi/v5"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
)

func Setup(router *chi.Mux, urlsRepo handlers.Repository, config conf.Configuration) {
	handler := handlers.NewHandler(urlsRepo, config)

	router.Post("/", handler.PostURI)
	router.Get("/{id}", handler.GetURI)
}
