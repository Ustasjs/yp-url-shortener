package router

import (
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const port = ":8080"

func StartServer() {
	r := chi.NewRouter()
	initMiddleware(r)
	initRoutes(r)

	err := http.ListenAndServe(port, r)
	if err != nil {
		panic(err)
	}
}

func initRoutes(r *chi.Mux) {
	store := repository.NewMemStorage()
	shortener := shortener.NewShortener()
	h := handler.NewHandler(store, shortener)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetShortURLByID)
}

func initMiddleware(r *chi.Mux) {
	r.Use(middleware.Logger)
}
