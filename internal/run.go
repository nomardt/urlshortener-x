package internal

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
)

func Run(config conf.Configuration) error {
	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("text/plain"))

	urlsRepo := urlsInfra.NewInMemoryRepo()
	urls.Setup(router, urlsRepo, config.Path)

	fmt.Println("Server started at:", config.ListenAddress)
	if err := http.ListenAndServe(config.ListenAddress, router); err != nil {
		return err
	}

	return nil
}
