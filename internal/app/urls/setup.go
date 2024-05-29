package urls

import (
	"github.com/go-chi/chi/v5"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
	"github.com/nomardt/urlshortener-x/internal/app/urls/middlewares"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

func Setup(router *chi.Mux, urlsRepo urlsDomain.Repository, config conf.Configuration) {
	handler := handlers.NewHandler(urlsRepo, config)

	router.Get("/{id}", handler.GetURI)
	router.Get("/api/user/urls", handler.GetUserURLs)

	router.Post("/", middlewares.OnlyPlaintextBody(handler.PostURI))
	router.Post("/api/shorten", middlewares.OnlyJSONBody(handler.JSONPostURI))
	router.Post("/api/shorten/batch", middlewares.OnlyJSONBody(handler.JSONPostBatch))
}
