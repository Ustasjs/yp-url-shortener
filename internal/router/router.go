// Package router wires together configuration, logging, storage, services,
// middleware and HTTP routes, and runs the server.
package router

import (
	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/logger"
	customMiddleware "Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"Ustasjs/yp-url-shortener/migrations"
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// StartServer loads configuration, initializes logging, storage and the audit
// subsystem, registers all routes and middleware, and starts the HTTP server.
// It blocks until the server stops and panics on a fatal startup error.
func StartServer() {
	settingsMap := settings.InitSettings()
	loggerErr := logger.Initialize(settingsMap.LogLevel)
	if loggerErr != nil {
		panic(loggerErr)
	}

	var db *sql.DB
	if settingsMap.DatabaseDSN != "" {
		var dbErr error
		db, dbErr = sql.Open("pgx", string(settingsMap.DatabaseDSN))

		logger.Log.Info("Connect to database")

		if dbErr != nil {
			panic(dbErr)
		}
		defer db.Close()

		migrationsErr := migrations.RunMigrations(db)
		if migrationsErr != nil {
			panic(migrationsErr)
		}
	}

	logger.Log.Info("Starting server on:", zap.String("address", string(settingsMap.ServerAddress)))

	var store shortener.Storage
	if settingsMap.DatabaseDSN != "" {
		store = repository.NewPostgresStorage(db)
	} else {
		store = repository.NewMemStorage(string(settingsMap.FileStoragePath))
	}

	notifier := initAudit(settingsMap)

	r := chi.NewRouter()
	initMiddleware(r, store)
	initRoutes(r, settingsMap, db, store, notifier)

	srv := &http.Server{
		Addr:              string(settingsMap.ServerAddress),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	case <-ctx.Done():
		stop()
		logger.Log.Info("Shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("Server shutdown failed", zap.Error(err))
		}
	}

	notifier.Close()
}

func initAudit(s *settings.Settings) *audit.Notifier {
	notifier := audit.NewNotifier()

	if s.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(string(s.AuditFile))
		if err != nil {
			panic(err)
		}
		notifier.Attach(fileObserver)
		logger.Log.Info("Audit file sink enabled", zap.String("path", string(s.AuditFile)))
	}

	if s.AuditURL != "" {
		notifier.Attach(audit.NewHTTPObserver(string(s.AuditURL)))
		logger.Log.Info("Audit HTTP sink enabled", zap.String("url", string(s.AuditURL)))
	}

	return notifier
}

func initRoutes(r *chi.Mux, s *settings.Settings, db *sql.DB, store shortener.Storage, notifier *audit.Notifier) {
	deleter := shortener.NewURLDeleter(store, 100, 5*time.Second)
	deleter.Start(3)

	shortenerService := shortener.NewShortener(store, s.BaseURL, deleter)
	h := handler.NewHandler(shortenerService, db, notifier)

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetShortURLByID)
	r.Get("/ping", h.GetDBPing)

	r.Post("/api/shorten", h.CreateShortURLJSONApi)
	r.Post("/api/shorten/batch", h.CreateShortURLSByBatch)

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.RequireAuth())
		r.Get("/api/user/urls", h.GetUserURLs)
		r.Delete("/api/user/urls", h.DeleteUserURLs)
	})
}

func initMiddleware(r *chi.Mux, store customMiddleware.UserRepository) {
	r.Use(logger.LoggerMiddleware)
	r.Use(customMiddleware.GzipDecompress)
	r.Use(middleware.Compress(gzip.DefaultCompression))
	r.Use(customMiddleware.Auth(store))
}
