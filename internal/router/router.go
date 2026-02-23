package router

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"compress/gzip"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func StartServer() {
	settings := settings.InitSettings()
	logger.Initialize(settings.LogLevel)

	logger.Log.Info("Starting server on:", zap.String("address", string(settings.ServerAddress)))

	r := chi.NewRouter()
	initMiddleware(r)
	initRoutes(r, settings)

	srv := &http.Server{
		Addr:              string(settings.ServerAddress),
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

func initRoutes(r *chi.Mux, s *settings.Settings) {
	store := repository.NewMemStorage()
	shortener := shortener.NewShortener()
	h := handler.NewHandler(store, shortener, s)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetShortURLByID)

	r.Post("/api/shorten", h.CreateShortURLJSONApi)
}

func initMiddleware(r *chi.Mux) {
	r.Use(logger.LoggerMiddleware)
	r.Use(middleware.Compress(gzip.DefaultCompression))
}
