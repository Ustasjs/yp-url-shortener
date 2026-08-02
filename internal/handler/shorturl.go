package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/model"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/urlservice"

	"go.uber.org/zap"
)

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(model.ErrorResponse{Error: message}); err != nil {
		logger.Log.Error("encode error response failed", zap.Error(err))
	}
}

// CreateShortURL handles POST / with a plain-text body containing the URL to
// shorten. On success it responds with 201 Created (or 409 Conflict if the URL
// was already shortened) and writes the short URL as text/plain. Invalid input
// yields 400 Bad Request.
func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeJSONError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		body := strings.TrimSpace(string(bodyBytes))
		if body == "" {
			writeJSONError(w, "url is required", http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		userID, _ := middleware.GetUserIDFromContext(ctx)
		shortURL, err := h.urls.Shorten(ctx, body, userID)

		var status int
		switch {
		case errors.Is(err, urlservice.ErrInvalidURL):
			writeJSONError(w, "invalid url", http.StatusBadRequest)
			return
		case errors.Is(err, repository.ErrConflict):
			status = http.StatusConflict
		case err != nil:
			writeJSONError(w, "bad request", http.StatusBadRequest)
			return
		default:
			status = http.StatusCreated
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		if _, err := w.Write([]byte(shortURL)); err != nil {
			logger.Log.Error("internal server error")
		}
		return
	} else {
		writeJSONError(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}

// GetShortURLByID handles GET /{id} and redirects to the original URL with 307
// Temporary Redirect. It responds with 404 Not Found for an unknown id and 410
// Gone if the URL has been deleted.
func (h *Handler) GetShortURLByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx := r.Context()
		id := r.PathValue("id")
		userID, _ := middleware.GetUserIDFromContext(ctx)

		originalURL, err := h.urls.Expand(ctx, id, userID)
		if errors.Is(err, repository.ErrDeleted) {
			http.Error(w, "Gone", http.StatusGone)
			return
		}
		if err != nil {
			writeJSONError(w, "url not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	} else {
		writeJSONError(w, "Only GET requests are allowed", http.StatusBadRequest)
	}
}

// CreateShortURLJSONApi handles POST /api/shorten with a JSON body
// (model.CreateShortURLRequest). It responds with 201 Created (or 409 Conflict
// if the URL already exists) and a model.CreateShortURLResponce JSON body.
// Invalid input yields 400 Bad Request.
func (h *Handler) CreateShortURLJSONApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)

		var request model.CreateShortURLRequest
		if err := decoder.Decode(&request); err != nil {
			logger.Log.Error(err.Error())
			writeJSONError(w, "cannot decode request JSON body", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		userID, _ := middleware.GetUserIDFromContext(ctx)
		shortURL, err := h.urls.Shorten(ctx, request.URL, userID)

		var status int
		switch {
		case errors.Is(err, urlservice.ErrInvalidURL):
			writeJSONError(w, "invalid url", http.StatusBadRequest)
			return
		case errors.Is(err, repository.ErrConflict):
			status = http.StatusConflict
		case err != nil:
			writeJSONError(w, "bad request", http.StatusBadRequest)
			return
		default:
			status = http.StatusCreated
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		responce := model.CreateShortURLResponce{
			Result: shortURL,
		}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(&responce); err != nil {
			logger.Log.Error(err.Error())
			writeJSONError(w, "error encoding response", http.StatusBadRequest)
			return
		}
		return
	} else {
		writeJSONError(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}
