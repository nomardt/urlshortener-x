package handlers

import (
	"context"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

type Repository interface {
	SaveURL(*urlsDomain.URL) error
	GetURL(*string) (string, error)
	Ping(ctx context.Context) error
}

type Handler struct {
	Repository
	conf.Configuration
}

func NewHandler(repo Repository, config conf.Configuration) *Handler {
	return &Handler{
		Repository:    repo,
		Configuration: config,
	}
}
