package handler

import (
	"encoding/json"
	"net/http"

	"Ustasjs/yp-url-shortener/internal/logger"

	"go.uber.org/zap"
)

// GetInternalStats handles GET /api/internal/stats and returns, as JSON, the
// service-wide number of shortened URLs and users. Access is restricted to
// trusted clients by the TrustedSubnet middleware, so this handler assumes the
// caller has already been authorized.
func (h *Handler) GetInternalStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.shortener.GetStats(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logger.Log.Error("encode stats response failed", zap.Error(err))
	}
}
