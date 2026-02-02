package router

import (
	"Ustasjs/yp-url-shortener/internal/config/flags"
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func StartServer() {
	flags := flags.InitFlags()

	host := flags.ServerAddress.Host
	port := flags.ServerAddress.Port
	address := fmt.Sprintf("%s:%s", host, port)

	fmt.Println("Starting server on:", address)

	r := chi.NewRouter()
	initMiddleware(r)
	initRoutes(r)

	err := http.ListenAndServe(address, r)
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
