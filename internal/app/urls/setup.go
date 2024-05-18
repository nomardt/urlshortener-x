package urls

import (
	"github.com/go-chi/chi/v5"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
	"github.com/nomardt/urlshortener-x/internal/app/urls/middlewares"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
)

func Setup(router *chi.Mux, urlsRepo handlers.Repository, config conf.Configuration) {
	handler := handlers.NewHandler(urlsRepo, config)

	router.Post("/", logger.WithLogging(middlewares.OnlyPlaintextBody(handler.PostURI)))
	router.Get("/{id}", logger.WithLogging(handler.GetURI))

	router.Post("/api/shorten", logger.WithLogging(middlewares.OnlyJSONBody(handler.JSONPostURI)))
	router.Post("/api/shorten/batch", logger.WithLogging(middlewares.OnlyJSONBody(handler.JSONPostBatch)))
}
