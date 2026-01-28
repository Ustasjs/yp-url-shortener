package handler

import (
	"Ustasjs/yp-url-shortener/internal/repository"
)

type Handler struct {
	store *repository.MemStorage
}

func NewHandler(store *repository.MemStorage) *Handler {
	return &Handler{store: store}
}
