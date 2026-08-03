// Package router wires together configuration, logging, storage, services,
// middleware and HTTP routes, and builds the HTTP and gRPC servers. The caller
// owns their lifecycle: starting them and shutting down the dependencies.
package router

import (
	"compress/gzip"
	"crypto/tls"
	"database/sql"
	"net/http"
	"time"

	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/grpcserver"
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/logger"
	customMiddleware "Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"Ustasjs/yp-url-shortener/internal/service/urlservice"
	"Ustasjs/yp-url-shortener/migrations"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// App bundles the initialized HTTP and gRPC servers together with the
// dependencies whose lifecycle must be shut down by the caller after the servers
// stop. Shutting the components down in the right order — the servers first,
// then the dependencies they use — is the caller's responsibility.
type App struct {
	Server     *http.Server
	GRPCServer *grpcserver.Server
	Deleter    *shortener.URLDeleter
	Notifier   *audit.Notifier
	DB         *sql.DB

	useHTTPS bool
}

// Setup loads configuration, initializes logging, storage and the audit
// subsystem, and registers all routes and middleware. It returns an App holding
// both servers and the dependencies whose lifecycle the caller is responsible
// for shutting down. It panics on a fatal startup error.
func Setup() *App {
	settingsMap, settingsErr := settings.InitSettings()

	loggerErr := logger.Initialize(settingsMap.LogLevel)
	if loggerErr != nil {
		panic(loggerErr)
	}

	if settingsErr != nil {
		logger.Log.Fatal("Invalid configuration", zap.Error(settingsErr))
	}

	var db *sql.DB
	if settingsMap.DatabaseDSN != "" {
		var dbErr error
		db, dbErr = sql.Open("pgx", string(settingsMap.DatabaseDSN))

		logger.Log.Info("Connect to database")

		if dbErr != nil {
			panic(dbErr)
		}

		migrationsErr := migrations.RunMigrations(db)
		if migrationsErr != nil {
			panic(migrationsErr)
		}
	}

	logger.Log.Info("Starting server on:",
		zap.String("address", string(settingsMap.ServerAddress)),
		zap.String("grpcAddress", string(settingsMap.GRPCAddress)))

	var store shortener.Storage
	if settingsMap.DatabaseDSN != "" {
		store = repository.NewPostgresStorage(db)
	} else {
		store = repository.NewMemStorage(string(settingsMap.FileStoragePath))
	}

	notifier := initAudit(settingsMap)

	deleter := shortener.NewURLDeleter(store, 100, 5*time.Second)
	deleter.Start(3)

	shortenerService := shortener.NewShortener(store, settingsMap.BaseURL, deleter)

	r := chi.NewRouter()
	initMiddleware(r, store)
	initRoutes(r, settingsMap, handler.NewHandler(shortenerService, db, notifier))

	srv := &http.Server{
		Addr:              string(settingsMap.ServerAddress),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// The TLS config is built outside the server so that gRPC can reuse it.
	var tlsConfig *tls.Config
	if settingsMap.EnableHTTPS {
		var tlsErr error
		if settingsMap.TLSCertFile != "" && settingsMap.TLSKeyFile != "" {
			logger.Log.Info("Loading TLS certificate from files",
				zap.String("cert", string(settingsMap.TLSCertFile)),
				zap.String("key", string(settingsMap.TLSKeyFile)))
			tlsConfig, tlsErr = tlsConfigFromFiles(string(settingsMap.TLSCertFile), string(settingsMap.TLSKeyFile))
		} else {
			logger.Log.Info("Generating in-memory self-signed TLS certificate")
			tlsConfig, tlsErr = generateTLSConfig()
		}
		if tlsErr != nil {
			panic(tlsErr)
		}
		srv.TLSConfig = tlsConfig
	}

	// Both transports share one urlservice, so the business rules and the audit
	// events cannot drift apart between HTTP and gRPC.
	urls := urlservice.New(shortenerService, notifier)
	grpcSrv := grpcserver.New(string(settingsMap.GRPCAddress), urls, store, tlsConfig)

	return &App{
		Server:     srv,
		GRPCServer: grpcSrv,
		Deleter:    deleter,
		Notifier:   notifier,
		DB:         db,
		useHTTPS:   settingsMap.EnableHTTPS,
	}
}

// ListenAndServe starts the HTTP (or HTTPS) server and blocks until it stops.
// It returns http.ErrServerClosed after a graceful Shutdown.
func (a *App) ListenAndServe() error {
	if a.useHTTPS {
		logger.Log.Info("HTTPS enabled")
		return a.Server.ListenAndServeTLS("", "")
	}
	return a.Server.ListenAndServe()
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

func initRoutes(r *chi.Mux, s *settings.Settings, h *handler.Handler) {
	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetShortURLByID)
	r.Get("/ping", h.GetDBPing)

	r.Post("/api/shorten", h.CreateShortURLJSONApi)
	r.Post("/api/shorten/batch", h.CreateShortURLSByBatch)

	r.With(customMiddleware.TrustedSubnet(string(s.TrustedSubnet))).
		Get("/api/internal/stats", h.GetInternalStats)

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
