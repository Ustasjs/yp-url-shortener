package router

import (
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
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
	h := handler.NewHandler(store)

	mux.HandleFunc("/", h.CreateShortURL)
	mux.HandleFunc("/{id}", h.GetShortURLByID)
}
