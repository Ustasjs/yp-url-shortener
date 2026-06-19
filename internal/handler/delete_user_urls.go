package handler

import (
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
	"encoding/json"
	"errors"
	"net/http"
)

// DeleteUserURLs handles DELETE /api/user/urls. It accepts a JSON array of short
// IDs and schedules their asynchronous deletion for the authenticated user,
// responding with 202 Accepted. It returns 401 Unauthorized without a user
// identity, 400 Bad Request for an invalid or empty body, and 503 Service
// Unavailable when the deletion queue is overloaded.
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

	err := h.shortener.DeleteURLsAsync(userID, shortIDs)
	if errors.Is(err, shortener.ErrServiceOverloaded) {
		http.Error(w, "Service is overloaded, try again later", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
