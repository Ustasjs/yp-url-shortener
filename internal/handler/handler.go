package handler

import (
	"Ustasjs/yp-url-shortener/internal/model"
	"context"
)

type Shortener interface {
	CreateShortURL(ctx context.Context, originalURL string) (string, error)
	GetOriginalURL(ctx context.Context, id string) (string, error)
	CreateShortURLsBatch(ctx context.Context, items []model.BatchShortURLRequestItem) ([]model.BatchShortURLResponseItem, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
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
