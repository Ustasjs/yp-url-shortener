package router

import (
	"Ustasjs/yp-url-shortener/internal/handler"
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
	mux.HandleFunc("/", handler.CreateShortURL)
	mux.HandleFunc("/{id}", handler.GetShortURLById)
}
