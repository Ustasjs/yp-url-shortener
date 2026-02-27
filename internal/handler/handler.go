package handler

type Shortener interface {
	CreateShortURL(originalURL string) string
	GetOriginalURL(id string) (string, error)
}

type Handler struct {
	shortener Shortener
}

func NewHandler(shortener Shortener) *Handler {
	return &Handler{shortener: shortener}
}
