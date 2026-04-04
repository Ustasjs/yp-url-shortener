package handler

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/model"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func (h *Handler) CreateShortURLSByBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))

		var request []model.BatchShortURLRequestItem
		if err := decoder.Decode(&request); err != nil {
			logger.Log.Error(err.Error())
			writeJSONError(w, "cannot decode request JSON body", http.StatusBadRequest)
			return
		}

		for i := range request {
			parsedURL, err := url.ParseRequestURI(request[i].OriginalURL)
			if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
				writeJSONError(w, "invalid url", http.StatusBadRequest)
				return
			}
			request[i].OriginalURL = parsedURL.String()
		}

		ctx := r.Context()
		userID, _ := middleware.GetUserIDFromContext(ctx)
		response, err := h.shortener.CreateShortURLsBatch(ctx, request, userID)
		if err != nil {
			logger.Log.Error(err.Error())
			writeJSONError(w, "error saving batch", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(response); err != nil {
			logger.Log.Error(err.Error())
			writeJSONError(w, "error encoding response", http.StatusInternalServerError)
			return
		}
		return
	} else {
		writeJSONError(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}
