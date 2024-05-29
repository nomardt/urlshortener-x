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

	router.Get("/{id}", handler.GetURI)
	router.Post("/", middlewares.OnlyPlaintextBody(handler.PostURI))

	router.Route("/api", func(router chi.Router) {
		router.Route("/shorten", func(router chi.Router) {
			router.Post("/", middlewares.OnlyJSONBody(auth.WithAuth(handler.JSONPostURI, config.Secret)))
			router.Post("/batch", middlewares.OnlyJSONBody(auth.WithAuth(handler.JSONPostBatch, config.Secret)))
		})

		router.Route("/user", func(router chi.Router) {
			router.Get("/urls", auth.TokenNecessary(handler.GetUserURLs, config.Secret))
		})
	})
}
