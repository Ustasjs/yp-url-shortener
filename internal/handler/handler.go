package handler

import "Ustasjs/yp-url-shortener/internal/config/flags"

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
	flags     *flags.Flags
}

func NewHandler(store Storage, shortener Shortener, flags *flags.Flags) *Handler {
	return &Handler{store: store, shortener: shortener, flags: flags}
}
