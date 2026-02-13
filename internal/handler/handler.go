package handler

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
)

type Storage interface {
	Save(id string, url string)
	Get(id string) (string, error)
}

type Shortener interface {
	ShortenURL(url string) string
}

type Handler struct {
	store     Storage
	shortener Shortener
	settings  *settings.Settings
}

func NewHandler(store Storage, shortener Shortener, settings *settings.Settings) *Handler {
	return &Handler{store: store, shortener: shortener, settings: settings}
}
