package handler

import (
	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/model"
	"context"
)

type Shortener interface {
	CreateShortURL(ctx context.Context, originalURL string, userID string) (string, error)
	GetOriginalURL(ctx context.Context, id string) (string, error)
	CreateShortURLsBatch(ctx context.Context, items []model.BatchShortURLRequestItem, userID string) ([]model.BatchShortURLResponseItem, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
	DeleteURLsAsync(userID string, shortIDs []string) error
}

type Pinger interface {
	PingContext(ctx context.Context) error
}

type AuditPublisher interface {
	Publish(event audit.Event)
}

type Handler struct {
	shortener Shortener
	pinger    Pinger
	auditor   AuditPublisher
}

func NewHandler(shortener Shortener, pinger Pinger, auditor AuditPublisher) *Handler {
	return &Handler{shortener: shortener, pinger: pinger, auditor: auditor}
}
