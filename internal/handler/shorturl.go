package handler

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/model"
	"Ustasjs/yp-url-shortener/internal/repository"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}

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
		validURL, err := parseURL(body)
		if err != nil {
			writeJSONError(w, "invalid url", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		shortURL, err := h.shortener.CreateShortURL(ctx, validURL)
		var status int
		hasConflict := errors.Is(err, repository.ErrConflict)

		if err != nil && !hasConflict {
			writeJSONError(w, "bad request", http.StatusBadRequest)
			return
		}

		if hasConflict {
			status = http.StatusConflict
		} else {
			status = http.StatusCreated
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			logger.Log.Error("internal server error")
			return
		}
		return
	} else {
		writeJSONError(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}

func (h *Handler) GetShortURLByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx := r.Context()
		id := r.PathValue("id")
		originalURL, err := h.shortener.GetOriginalURL(ctx, id)
		if err != nil {
			writeJSONError(w, "url not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	} else {
		writeJSONError(w, "Only GET requests are allowed", http.StatusBadRequest)
	}
}

func (h *Handler) CreateShortURLJSONApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)

		var request model.CreateShortURLRequest
		if err := decoder.Decode(&request); err != nil {
			logger.Log.Error(err.Error())
			writeJSONError(w, "cannot decode request JSON body", http.StatusBadRequest)
			return
		}

		validURL, err := parseURL(request.URL)
		if err != nil {
			writeJSONError(w, "invalid url", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		shortURL, err := h.shortener.CreateShortURL(ctx, validURL)
		var status int
		hasConflict := errors.Is(err, repository.ErrConflict)

		if err != nil && !hasConflict {
			writeJSONError(w, "bad request", http.StatusBadRequest)
			return
		}

		if hasConflict {
			status = http.StatusConflict
		} else {
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

func parseURL(raw string) (string, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid url")
	}
	return parsed.String(), nil
}
