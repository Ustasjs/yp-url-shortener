package shortener

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"context"
	"fmt"
	"strings"
)

type Storage interface {
	Save(ctx context.Context, id string, url string) error
	Get(ctx context.Context, id string) (string, error)
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

func (s *Shortener) CreateShortURL(ctx context.Context, originalURL string) string {
	id := shortenURL(originalURL)
	s.repo.Save(ctx, id, originalURL)

	base := strings.TrimSuffix(string(s.baseURL), "/")
	return fmt.Sprintf("%s/%s", base, id)
}

func (s *Shortener) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.repo.Get(ctx, id)
}
