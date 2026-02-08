package router

import (
	"Ustasjs/yp-url-shortener/internal/config/flags"
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func StartServer() {
	flags := flags.InitFlags()

	log.Println("Starting server on:", flags.ServerAddress)

	r := chi.NewRouter()
	initMiddleware(r)
	initRoutes(r, flags)

	srv := &http.Server{
		Addr:              string(flags.ServerAddress),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func initRoutes(r *chi.Mux, f *flags.Flags) {
	store := repository.NewMemStorage()
	shortener := shortener.NewShortener()
	h := handler.NewHandler(store, shortener, f)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetShortURLByID)
}

func initMiddleware(r *chi.Mux) {
	r.Use(middleware.Logger)
}
