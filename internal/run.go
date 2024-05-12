package internal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"

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
		return err
	}

	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("text/plain", "application/json", "application/x-gzip"))
	router.Use(middleware.Compress(3))

	if config.DB.Host != "" {
		ps := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.DB.Host, config.DB.Port, config.DB.User, config.DB.Password, config.DB.DBname, config.DB.SSLmode)
		db, err := sql.Open("pgx", ps)
		if err != nil {
			return err
		}
		defer db.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err = db.PingContext(ctx); err != nil {
			return err
		}

		router.Get("/ping", logger.WithLogging(func(w http.ResponseWriter, r *http.Request) {
			if err := db.PingContext(ctx); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			} else {
				http.Error(w, "OK", http.StatusOK)
				return
			}
		}))

		logger.Log.Info("PostgreSQL initiated succesfully")
	}

	urlsRepo := urlsInfra.NewInMemoryRepo()
	if err := urlsRepo.LoadStoredURLs(config); err != nil {
		logger.Log.Info("Couldn't recover any previously shortened URLs!", zap.String("error", err.Error()))

		if _, err := os.Create(config.StorageFile); err != nil {
			logger.Log.Info("No file will be created to store shortened URLs")
		}
	}
	urls.Setup(router, urlsRepo, config)

	logger.Log.Info("The server has started", zap.String("address", config.ListenAddress))
	if err := http.ListenAndServe(config.ListenAddress, router); err != nil {
		return err
	}

	return nil
}
