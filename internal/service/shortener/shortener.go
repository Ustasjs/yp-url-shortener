// Package shortener contains the core business logic of the URL shortener:
// generating short IDs, delegating persistence to a Storage backend and
// scheduling asynchronous deletions.
package shortener

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/model"
	"context"
	"strings"
)

// Storage abstracts the persistence layer used by the Shortener service.
type Storage interface {
	// Save stores a single short ID together with its original URL and owner.
	Save(ctx context.Context, id string, url string, userID string) error
	// Get returns the original URL stored under id.
	Get(ctx context.Context, id string) (string, error)
	// SaveListUrls stores a batch of records owned by userID.
	SaveListUrls(ctx context.Context, records []model.ShortURLRecord, userID string) error
	// CreateUser creates a new user and returns its identifier.
	CreateUser(ctx context.Context) (string, error)
	// GetUserURLs returns every URL owned by userID.
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
	// DeleteURLsBatch marks the given items as deleted.
	DeleteURLsBatch(ctx context.Context, items []model.DeleteItem) error
}

// Shortener implements the URL shortening business logic on top of a Storage
// backend, building absolute short URLs from a configured base URL.
type Shortener struct {
	repo    Storage
	baseURL settings.BaseURL
	deleter *URLDeleter
}

// NewShortener returns a Shortener that persists URLs through repo, builds short
// URLs using baseURL and offloads deletions to deleter. deleter may be nil, in
// which case DeleteURLsAsync is a no-op.
func NewShortener(repo Storage, baseURL settings.BaseURL, deleter *URLDeleter) *Shortener {
	return &Shortener{
		repo:    repo,
		baseURL: baseURL,
		deleter: deleter,
	}
}

// CreateShortURL generates a short ID for originalURL, persists it for userID and
// returns the absolute short URL. An error wrapping repository.ErrConflict means
// the URL was already shortened; the returned short URL is still valid then.
func (s *Shortener) CreateShortURL(ctx context.Context, originalURL string, userID string) (string, error) {
	id := shortenURL(originalURL)

	err := s.repo.Save(ctx, id, originalURL, userID)

	base := strings.TrimSuffix(string(s.baseURL), "/")
	return base + "/" + id, err
}

// GetOriginalURL returns the original URL associated with the given short id.
func (s *Shortener) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.repo.Get(ctx, id)
}

// CreateShortURLsBatch generates and stores short IDs for every item and returns
// the resulting short URLs, preserving each item's correlation ID.
func (s *Shortener) CreateShortURLsBatch(ctx context.Context, items []model.BatchShortURLRequestItem, userID string) ([]model.BatchShortURLResponseItem, error) {
	records := make([]model.ShortURLRecord, 0, len(items))

	for _, item := range items {
		id := shortenURL(item.OriginalURL)
		records = append(records, model.ShortURLRecord{ID: id, URL: item.OriginalURL})
	}

	if err := s.repo.SaveListUrls(ctx, records, userID); err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(string(s.baseURL), "/")
	response := make([]model.BatchShortURLResponseItem, 0, len(items))
	for i, item := range items {
		response = append(response, model.BatchShortURLResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      base + "/" + records[i].ID,
		})
	}
	return response, nil
}

// GetUserURLs returns every URL owned by userID, with absolute short URLs.
func (s *Shortener) GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error) {
	items, err := s.repo.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(string(s.baseURL), "/")
	for i := range items {
		items[i].ShortURL = base + "/" + items[i].ShortURL
	}

	return items, nil
}

// DeleteURLsAsync schedules deletion of the given short IDs owned by userID. It
// returns ErrServiceOverloaded when the deletion queue is full, and nil when no
// deleter is configured.
func (s *Shortener) DeleteURLsAsync(userID string, shortIDs []string) error {
	if s.deleter != nil {
		return s.deleter.DeleteURLsAsync(userID, shortIDs)
	}
	return nil
}
