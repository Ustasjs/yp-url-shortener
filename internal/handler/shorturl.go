package handler

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func makeShortURL(id string, baseURL settings.BaseURL) string {
	base := strings.TrimSuffix(string(baseURL), "/")
	return fmt.Sprintf("%s/%s", base, id)
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		body := strings.TrimSpace(string(bodyBytes))
		if body == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}
		parsedURL, err := url.ParseRequestURI(body)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}

		id := h.shortener.ShortenURL(parsedURL.String())
		h.store.Save(id, parsedURL.String())

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(makeShortURL(id, h.settings.BaseURL)))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		return
	} else {
		http.Error(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}

func (h *Handler) GetShortURLByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		id := r.PathValue("id")
		url, err := h.store.Get(id)
		if err != nil {
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Only GET requests are allowed", http.StatusBadRequest)
	}
}

func (h *Handler) CreateShortURLJSONApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)

		var request model.CreateShortURLRequest
		if err := decoder.Decode(&request); err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
			return
		}

		parsedURL, err := url.ParseRequestURI(request.URL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}

		id := h.shortener.ShortenURL(parsedURL.String())
		h.store.Save(id, parsedURL.String())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		responce := model.CreateShortURLResponce{
			Result: makeShortURL(id, h.settings.BaseURL),
		}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(&responce); err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, "error encoding response", http.StatusBadRequest)
			return
		}
		return
	} else {
		http.Error(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}
