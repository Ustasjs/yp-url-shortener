package handler

import "context"

type Shortener interface {
	CreateShortURL(ctx context.Context, originalURL string) string
	GetOriginalURL(ctx context.Context, id string) (string, error)
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
