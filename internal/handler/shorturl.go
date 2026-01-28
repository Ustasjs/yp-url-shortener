package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
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
	} else {
		http.Error(w, "Only POST requests are allowed", http.StatusBadRequest)
	}
}

func GetShortURLById(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

	} else {
		http.Error(w, "Only GET requests are allowed", http.StatusBadRequest)
	}
}
