// Package handler implements the HTTP handlers of the URL shortener service.
package handler

import (
	"context"

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

// URLService is the transport-agnostic business logic behind the handlers that
// shorten, resolve and list URLs
type URLService interface {
	// Shorten validates rawURL, stores it for userID and returns the short URL.
	// It reports repository.ErrConflict if the URL was already shortened, and
	// urlservice.ErrInvalidURL if rawURL is not an absolute URL.
	Shorten(ctx context.Context, rawURL string, userID string) (string, error)
	// Expand returns the original URL stored under id.
	Expand(ctx context.Context, id string, userID string) (string, error)
	// ListUserURLs returns every URL created by userID.
	ListUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
}

// Handler holds the dependencies shared by all HTTP handlers of the service.
type Handler struct {
	shortener Shortener
	urls      URLService
	pinger    Pinger
}

// NewHandler returns a Handler that uses the given shortener service, URL
// service and pinger.
//
// The pinger may be nil: GetDBPing then reports that the database is not
// configured.
func NewHandler(shortener Shortener, urls URLService, pinger Pinger) *Handler {
	return &Handler{
		shortener: shortener,
		urls:      urls,
		pinger:    pinger,
	}
}
