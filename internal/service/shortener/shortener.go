package shortener

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/model"
	"context"
	"fmt"
	"strings"
)

type Storage interface {
	Save(ctx context.Context, id string, url string) error
	Get(ctx context.Context, id string) (string, error)
	SaveListUrls(ctx context.Context, records []model.ShortURLRecord) error
}

type Shortener struct {
	repo    Storage
	baseURL settings.BaseURL
}

func NewShortener(repo Storage, baseURL settings.BaseURL) *Shortener {
	return &Shortener{
		repo,
		baseURL,
	}
}

func (s *Shortener) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	id := shortenURL(originalURL)
	err := s.repo.Save(ctx, id, originalURL)

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

	if err := s.repo.SaveListUrls(ctx, records); err != nil {
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
