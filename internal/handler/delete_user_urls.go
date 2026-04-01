package handler

import (
	"Ustasjs/yp-url-shortener/internal/middleware"
	"encoding/json"
	"net/http"
)

func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var shortIDs []string
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(shortIDs) == 0 {
		http.Error(w, "Empty request", http.StatusBadRequest)
		return
	}

	h.shortener.DeleteURLsAsync(userID, shortIDs)

	w.WriteHeader(http.StatusAccepted)
}

