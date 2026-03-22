package handler

import (
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/service"
	"encoding/json"
	"net/http"
)

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.AuthCookieName)
	if err == nil && cookie.Value != "" {
		if _, err := service.GetUserID(cookie.Value); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.shortener.GetUserURLs(r.Context(), userID)
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
	json.NewEncoder(w).Encode(urls)
}
