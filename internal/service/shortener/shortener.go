package shortener

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"fmt"
	"strings"
)

type Storage interface {
	Save(id string, url string)
	Get(id string) (string, error)
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

func (s *Shortener) CreateShortURL(originalURL string) string {
	id := shortenURL(originalURL)
	s.repo.Save(id, originalURL)

	base := strings.TrimSuffix(string(s.baseURL), "/")
	return fmt.Sprintf("%s/%s", base, id)
}

func (s *Shortener) GetOriginalURL(id string) (string, error) {
	return s.repo.Get(id)
}
