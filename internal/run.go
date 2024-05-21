package internal

import (
	"context"
	"net/http"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	"github.com/nomardt/urlshortener-x/internal/app/urls"
	"github.com/nomardt/urlshortener-x/internal/app/urls/handlers"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	urlsInfra "github.com/nomardt/urlshortener-x/internal/infra/urls"
)

func Run(config conf.Configuration) error {
	if err := logger.Initialize("info"); err != nil {
		return err
	}

	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("text/plain", "application/json", "application/x-gzip"))
	router.Use(middleware.Compress(3))

	var urlsRepo handlers.Repository
	if config.DB.Host != "" {
		urlsRepo = urlsInfra.NewPostgresRepo(config)
	} else {
		urlsRepo = urlsInfra.NewInMemoryRepo(config)
	}

	router.Get("/ping", logger.WithLogging(func(w http.ResponseWriter, r *http.Request) {
		if err := urlsRepo.Ping(context.TODO()); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		} else {
			http.Error(w, "OK", http.StatusOK)
			return
		}
	}))

	urls.Setup(router, urlsRepo, config)

	logger.Log.Info("The server has started", zap.String("address", config.ListenAddress))
	if err := http.ListenAndServe(config.ListenAddress, router); err != nil {
		return err
	}

	return nil
}
