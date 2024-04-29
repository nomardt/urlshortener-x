package internal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
)

func Run(config conf.Configuration) error {
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("text/plain", "application/json"))

	urlsRepo := urlsInfra.NewInMemoryRepo()
	urls.Setup(router, urlsRepo, config)

	logger.Log.Info("The server has started", zap.String("address", config.ListenAddress))
	if err := http.ListenAndServe(config.ListenAddress, router); err != nil {
		return err
	}

	return nil
}
