package shortener

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/model"
	"context"
	"fmt"
	"strings"
)

type Storage interface {
	Save(ctx context.Context, id string, url string, userID string) error
	Get(ctx context.Context, id string) (string, error)
	SaveListUrls(ctx context.Context, records []model.ShortURLRecord, userID string) error
	CreateUser(ctx context.Context) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
	DeleteURLsBatch(ctx context.Context, items []model.DeleteItem) error
}

type Shortener struct {
	repo    Storage
	baseURL settings.BaseURL
	deleter *URLDeleter
}

func NewShortener(repo Storage, baseURL settings.BaseURL, deleter *URLDeleter) *Shortener {
	return &Shortener{
		repo:    repo,
		baseURL: baseURL,
		deleter: deleter,
	}
}

func (s *Shortener) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	id := shortenURL(originalURL)
	
	userID, _ := middleware.GetUserIDFromContext(ctx)
	err := s.repo.Save(ctx, id, originalURL, userID)

	base := strings.TrimSuffix(string(s.baseURL), "/")
	return fmt.Sprintf("%s/%s", base, id), err
}

func (s *Shortener) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.repo.Get(ctx, id)
}

func (s *Shortener) CreateShortURLsBatch(ctx context.Context, items []model.BatchShortURLRequestItem) ([]model.BatchShortURLResponseItem, error) {
	records := make([]model.ShortURLRecord, 0, len(items))

	for _, item := range items {
		id := shortenURL(item.OriginalURL)
		records = append(records, model.ShortURLRecord{ID: id, URL: item.OriginalURL})
	}

	userID, _ := middleware.GetUserIDFromContext(ctx)
	if err := s.repo.SaveListUrls(ctx, records, userID); err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(string(s.baseURL), "/")
	response := make([]model.BatchShortURLResponseItem, 0, len(items))
	for i, item := range items {
		response = append(response, model.BatchShortURLResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", base, records[i].ID),
		})
	}
	return response, nil
}

func (s *Shortener) GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error) {
	items, err := s.repo.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(string(s.baseURL), "/")
	for i := range items {
		items[i].ShortURL = fmt.Sprintf("%s/%s", base, items[i].ShortURL)
	}

	return items, nil
}

func (s *Shortener) DeleteURLsAsync(userID string, shortIDs []string) error {
	if s.deleter != nil {
		return s.deleter.DeleteURLsAsync(userID, shortIDs)
	}
	return nil
}
