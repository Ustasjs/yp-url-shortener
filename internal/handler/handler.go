package handler

import "context"

type Shortener interface {
	CreateShortURL(originalURL string) string
	GetOriginalURL(id string) (string, error)
}

type Pinger interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	shortener Shortener
	pinger    Pinger
}

func NewHandler(shortener Shortener, pinger Pinger) *Handler {
	return &Handler{shortener: shortener, pinger: pinger}
}
