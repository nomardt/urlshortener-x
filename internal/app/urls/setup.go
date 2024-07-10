package urls

import (
	"github.com/go-chi/chi/v5"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
	"github.com/nomardt/urlshortener-x/internal/app/urls/middlewares"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/auth"
)

func Setup(router *chi.Mux, urlsRepo urlsDomain.Repository, config conf.Configuration) {
	handler := handlers.NewHandler(urlsRepo, config)

	router.Group(func(r chi.Router) {
		r.Use(auth.WithAuth(config.Secret))

		r.Post("/", middlewares.OnlyPlaintextBody(handler.PostURI))

		r.Post("/api/shorten", middlewares.OnlyJSONBody(handler.JSONPostURI))
		r.Post("/api/shorten/batch", middlewares.OnlyJSONBody(handler.JSONPostBatch))

		r.Delete("/api/user/urls", middlewares.OnlyJSONBody(handler.JSONDeleteBatch))
	})

	router.Group(func(r chi.Router) {
		r.Use(auth.TokenNecessary(config.Secret))

		r.Get("/api/user/urls", handler.GetUserURLs)
	})

	router.Get("/{id}", handler.GetURI)
}
