package handler_test

import (
	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/handler"
	customMiddleware "Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testShorteURLID = "test-id"
var testFilePath = "./test.json"

type mockShortener struct{}

func (mockShortener) ShortenURL(url string) string { return testShorteURLID }

func NewMockShortener() *mockShortener {
	return &mockShortener{}
}

func NewMockSettings() *settings.Settings {
	return &settings.Settings{
		ServerAddress: settings.ServerAddress("localhost:8080"),
		BaseURL:       settings.BaseURL("http://localhost:8080"),
	}
}

func TestHandler_CreateShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		r    *http.Request
		want want
	}{
		{
			name: "check request with invalid method",
			r:    httptest.NewRequest(http.MethodGet, "/", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "Only POST requests are allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "check post request with no body",
			r:    httptest.NewRequest(http.MethodPost, "/", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "url is required\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "check post request with invalid body",
			r:    httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid body")),
			want: want{
				code:        http.StatusBadRequest,
				response:    "invalid url\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "check post request with valid body",
			r:    httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com")),
			want: want{
				code:        http.StatusCreated,
				response:    fmt.Sprintf("http://localhost:8080/%s", testShorteURLID),
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage(testFilePath)
			mux := http.NewServeMux()
			shortener := NewMockShortener()
			settings := NewMockSettings()
			h := handler.NewHandler(store, shortener, settings)
			mux.HandleFunc("/", h.CreateShortURL)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, tt.r)

			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, tt.want.response, rr.Body.String())
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
		})
	}
}

func TestHandler_GetShortURLByID(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	store := repository.NewMemStorage(testFilePath)
	store.Save("123", "https://example.com")

	tests := []struct {
		name  string // description of this test case
		store *repository.MemStorage
		r     *http.Request
		want  want
	}{
		{
			name:  "check get request with invalid method",
			store: repository.NewMemStorage(testFilePath),
			r:     httptest.NewRequest(http.MethodPost, "/123", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "Only GET requests are allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:  "check get request with valid id",
			store: store,
			r:     httptest.NewRequest(http.MethodGet, "/123", nil),
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "<a href=\"https://example.com\">Temporary Redirect</a>.\n\n",
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name:  "check get request with invalid id",
			store: store,
			r:     httptest.NewRequest(http.MethodGet, "/456", nil),
			want: want{
				code:        http.StatusNotFound,
				response:    "url not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			shortener := NewMockShortener()
			settings := NewMockSettings()
			h := handler.NewHandler(tt.store, shortener, settings)
			mux.HandleFunc("/{id}", h.GetShortURLByID)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, tt.r)

			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, tt.want.response, rr.Body.String())
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
		})
	}
}

func TestHandler_CreateShortURLJSONApi(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		r    *http.Request
		want want
	}{
		{
			name: "check request with invalid method",
			r:    httptest.NewRequest(http.MethodGet, "/api/shorten", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "Only POST requests are allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "check post request with no body",
			r:    httptest.NewRequest(http.MethodPost, "/api/shorten", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "cannot decode request JSON body\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "check post request with invalid body",
			r:    httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("invalid body")),
			want: want{
				code:        http.StatusBadRequest,
				response:    "cannot decode request JSON body\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "check post request with valid body",
			r:    httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("{ \"url\": \"https://practicum.yandex.ru/\"}")),
			want: want{
				code:        http.StatusCreated,
				response:    fmt.Sprintf("{\"result\":\"http://localhost:8080/%s\"}\n", testShorteURLID),
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage(testFilePath)
			mux := http.NewServeMux()
			shortener := NewMockShortener()
			settings := NewMockSettings()
			h := handler.NewHandler(store, shortener, settings)
			mux.HandleFunc("/api/shorten", h.CreateShortURLJSONApi)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, tt.r)

			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, tt.want.response, rr.Body.String())
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
		})
	}
}

func TestHandler_CreateShortURL_Gzip(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("https://example.com"))
	assert.NoError(t, err)
	assert.NoError(t, gz.Close())

	r := httptest.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Encoding", "gzip")

	store := repository.NewMemStorage(testFilePath)
	mux := http.NewServeMux()
	shortener := NewMockShortener()
	settings := NewMockSettings()
	h := handler.NewHandler(store, shortener, settings)

	wrapped := customMiddleware.GzipDecompress(http.HandlerFunc(h.CreateShortURL))
	mux.Handle("/", wrapped)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, fmt.Sprintf("http://localhost:8080/%s", testShorteURLID), rr.Body.String())
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
}
