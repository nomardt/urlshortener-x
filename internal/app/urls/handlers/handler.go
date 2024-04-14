package handlers

import (
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

type Repository interface {
	SaveURL(*urlsDomain.URL) error
	GetURL(*string) (string, error)
}

type Handler struct {
	Repository
	defaultRoute string
}

func NewHandler(repo Repository, defaultPath string) *Handler {
	return &Handler{
		Repository:   repo,
		defaultRoute: defaultPath,
	}
}
