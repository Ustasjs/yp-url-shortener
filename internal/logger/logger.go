// Package logger provides a process-wide zap logger and HTTP middleware that
// logs each request and its response.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log is the shared application logger. Until Initialize is called it is a no-op
// logger, so it is always safe to use.
var Log *zap.Logger = zap.NewNop()

// Initialize configures the shared Log at the given level using zap's production
// configuration.
func Initialize(level zap.AtomicLevel) error {
	cfg := zap.NewProductionConfig()
	cfg.Level = level
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

// LoggerMiddleware is HTTP middleware that logs the URI, method and duration of
// each request together with the status and size of the response.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		Log.Info("request", zap.String("uri", uri), zap.String("method", method), zap.Duration("duration", duration))
		Log.Info("responce", zap.Int("status", responseData.status), zap.Int("size", responseData.size))
	})
}
