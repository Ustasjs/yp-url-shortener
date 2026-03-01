package router

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/logger"
	customMiddleware "Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"compress/gzip"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func StartServer() {
	settingsMap := settings.InitSettings()
	loggerErr := logger.Initialize(settingsMap.LogLevel)
	if loggerErr != nil {
		panic(loggerErr)
	}

	db, dbErr := sql.Open("pgx", string(settingsMap.DatabaseDSN))

	logger.Log.Info("Connect to database")

	if dbErr != nil {
		panic(dbErr)
	}
	defer db.Close()

	logger.Log.Info("Starting server on:", zap.String("address", string(settingsMap.ServerAddress)))

	r := chi.NewRouter()
	initMiddleware(r)
	initRoutes(r, settingsMap, db)

	srv := &http.Server{
		Addr:              string(settingsMap.ServerAddress),
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

func initRoutes(r *chi.Mux, s *settings.Settings, db *sql.DB) {
	store := repository.NewMemStorage(string(s.FileStoragePath))
	shortenerService := shortener.NewShortener(store, s.BaseURL)
	h := handler.NewHandler(shortenerService, db)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetShortURLByID)
	r.Get("/ping", h.GetDBPing)

	r.Post("/api/shorten", h.CreateShortURLJSONApi)
}

func initMiddleware(r *chi.Mux) {
	r.Use(logger.LoggerMiddleware)
	r.Use(customMiddleware.GzipDecompress)
	r.Use(middleware.Compress(gzip.DefaultCompression))
}
