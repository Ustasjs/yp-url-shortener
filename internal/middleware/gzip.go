package middleware

import (
	"compress/gzip"
	"net/http"

	"Ustasjs/yp-url-shortener/internal/logger"

	"go.uber.org/zap"
)

// GzipDecompress is HTTP middleware that transparently decompresses request
// bodies sent with Content-Encoding: gzip.
func GzipDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					logger.Log.Error("close gzip reader failed", zap.Error(err))
				}
			}()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
