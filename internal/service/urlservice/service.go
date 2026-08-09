// Package urlservice holds the business logic shared by the HTTP handlers and
// the gRPC server: shorten a URL, expand a short ID, list the URLs of a user.
// It also checks the input and publishes audit events. It knows nothing about
// the transport.
//
// Storage errors are passed through as they are, so callers can check them with
// errors.Is against the repository sentinels.
package urlservice

import (
	"context"
	"errors"
	"net/url"
	"time"

	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/model"
	"Ustasjs/yp-url-shortener/internal/repository"
)

// Shortener is the domain service under the use cases. *shortener.Shortener
// implements it.
type Shortener interface {
	// CreateShortURL saves originalURL for userID and returns the short URL.
	// It returns repository.ErrConflict if the URL was shortened before.
	CreateShortURL(ctx context.Context, originalURL string, userID string) (string, error)
	// GetOriginalURL returns the original URL saved under id.
	GetOriginalURL(ctx context.Context, id string) (string, error)
	// GetUserURLs returns all URLs made by userID.
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
}

// AuditPublisher publishes the audit events made by the use cases.
type AuditPublisher interface {
	Publish(event audit.Event)
}

// Service groups the use cases shared by all transports. It has no state and is
// safe for concurrent use.
type Service struct {
	shortener Shortener
	auditor   AuditPublisher
}

// New returns a Service backed by shortener. auditor can be nil; then audit
// events are dropped.
func New(shortener Shortener, auditor AuditPublisher) *Service {
	return &Service{shortener: shortener, auditor: auditor}
}

// Shorten checks rawURL, saves it for userID and publishes a shorten audit
// event. It returns ErrInvalidURL if rawURL is not an absolute URL.
//
// repository.ErrConflict means the URL was shortened before. The short URL is
// still good, so callers should return it together with the error.
func (s *Service) Shorten(ctx context.Context, rawURL string, userID string) (string, error) {
	validURL, err := parseURL(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	shortURL, err := s.shortener.CreateShortURL(ctx, validURL, userID)
	if err != nil && !errors.Is(err, repository.ErrConflict) {
		return "", err
	}

	s.publish(audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    audit.ActionShorten,
		UserID:    userID,
		URL:       validURL,
	})

	// err is nil or repository.ErrConflict; shortURL works in both cases.
	return shortURL, err
}

// Expand returns the original URL saved under id and publishes a follow audit
// event. It returns repository.ErrDeleted for a deleted URL and
// repository.ErrRecordNotFound for an unknown id.
func (s *Service) Expand(ctx context.Context, id string, userID string) (string, error) {
	originalURL, err := s.shortener.GetOriginalURL(ctx, id)
	if err != nil {
		return "", err
	}

	s.publish(audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    audit.ActionFollow,
		UserID:    userID,
		URL:       originalURL,
	})

	return originalURL, nil
}

// ListUserURLs returns all URLs of userID, with absolute short URLs.
func (s *Service) ListUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error) {
	return s.shortener.GetUserURLs(ctx, userID)
}

func (s *Service) publish(event audit.Event) {
	if s.auditor == nil {
		return
	}
	s.auditor.Publish(event)
}

// parseURL takes only absolute URLs. It returns the normalized form, because
// that is what we save and audit.
func parseURL(raw string) (string, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}
	return parsed.String(), nil
}
