package handler

type Storage interface {
	Save(id string, url string)
	Get(id string) (string, error)
}

type Shortener interface {
	ShortenURL(url string) string
}

type Handler struct {
	store     Storage
	shortener Shortener
}

func NewHandler(store Storage, shortener Shortener) *Handler {
	return &Handler{store: store, shortener: shortener}
}
