package handlers

import (
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
)

type Handler struct {
	urlsDomain.Repository
	conf.Configuration
}

func NewHandler(repo urlsDomain.Repository, config conf.Configuration) *Handler {
	return &Handler{
		Repository:    repo,
		Configuration: config,
	}
}
