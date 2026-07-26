// Package handler implements the HTTP handlers of the URL shortener service.
package handler

import (
	"context"

	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/model"
)

// Shortener is the business-logic dependency used by the HTTP handlers to create
// and resolve short URLs and to manage a user's URLs.
type Shortener interface {
	// CreateShortURL stores originalURL with userID and returns its short
	// URL. It reports repository.ErrConflict if the URL was already shortened.
	CreateShortURL(ctx context.Context, originalURL string, userID string) (string, error)
	// GetOriginalURL returns the original URL previously stored under id.
	GetOriginalURL(ctx context.Context, id string) (string, error)
	// CreateShortURLsBatch stores a batch of URLs and returns their short URLs,
	// preserving the correlation IDs supplied by the caller.
	CreateShortURLsBatch(ctx context.Context, items []model.BatchShortURLRequestItem, userID string) ([]model.BatchShortURLResponseItem, error)
	// GetUserURLs returns every URL created by userID.
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
	// DeleteURLsAsync schedules the asynchronous deletion of the given short IDs
	// owned by userID.
	DeleteURLsAsync(userID string, shortIDs []string) error
	// GetStats returns the service-wide number of shortened URLs and users.
	GetStats(ctx context.Context) (model.StatsResponse, error)
}

// Pinger reports whether the underlying storage (typically the database) is
// reachable. It is satisfied by *sql.DB.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// AuditPublisher publishes audit events produced while handling requests.
type AuditPublisher interface {
	Publish(event audit.Event)
}

// Handler holds the dependencies shared by all HTTP handlers of the service.
type Handler struct {
	shortener Shortener
	pinger    Pinger
	auditor   AuditPublisher
}

// NewHandler returns a Handler that uses the given shortener service, pinger and
// audit publisher.
//
// The pinger and auditor dependencies may be nil: GetDBPing then reports that
// the database is not configured, and audit events are silently dropped.
func NewHandler(shortener Shortener, pinger Pinger, auditor AuditPublisher) *Handler {
	return &Handler{shortener: shortener, pinger: pinger, auditor: auditor}
}
