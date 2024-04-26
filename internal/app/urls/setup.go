package urls

import (
	"github.com/go-chi/chi/v5"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
)

func Setup(router *chi.Mux, urlsRepo handlers.Repository, config conf.Configuration) {
	handler := handlers.NewHandler(urlsRepo, config)

	router.Post("/", logger.WithLogging(handler.PostURI))
	router.Get("/{id}", logger.WithLogging(handler.GetURI))
}
