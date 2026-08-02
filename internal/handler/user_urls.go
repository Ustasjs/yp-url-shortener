package handler

import (
	"encoding/json"
	"net/http"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/middleware"

	"go.uber.org/zap"
)

// GetUserURLs handles GET /api/user/urls and returns, as JSON, every URL created
// by the authenticated user. It responds with 401 Unauthorized when the request
// carries no user identity, 204 No Content when the user has no URLs, and 200 OK
// otherwise.
func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.urls.ListUserURLs(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		logger.Log.Error("encode user urls response failed", zap.Error(err))
	}
}
