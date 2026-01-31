package router

import (
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"net/http"
)

const port = ":8080"

func StartServer() {
	mux := http.NewServeMux()
	initRoutes(mux)

	err := http.ListenAndServe(port, mux)
	if err != nil {
		panic(err)
	}
}

func initRoutes(mux *http.ServeMux) {
	store := repository.NewMemStorage()
	shortener := shortener.NewShortener()
	h := handler.NewHandler(store, shortener)

	mux.HandleFunc("/", h.CreateShortURL)
	mux.HandleFunc("/{id}", h.GetShortURLByID)
}
