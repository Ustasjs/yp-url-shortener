package handler

import (
	"context"
	"net/http"
	"time"
)

func (h *Handler) GetDBPing(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := h.pinger.PingContext(ctx); err != nil {
			http.Error(w, "No connection to DB", http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Only GET requests are allowed", http.StatusBadRequest)
	}
}
