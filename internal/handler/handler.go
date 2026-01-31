package handler

type Storage interface {
	Save(id string, url string)
	Get(id string) (string, error)
}

type Handler struct {
	store Storage
}

func NewHandler(store Storage) *Handler {
	return &Handler{store: store}
}
