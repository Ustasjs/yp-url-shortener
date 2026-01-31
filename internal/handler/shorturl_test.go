package handler_test

import (
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testShorteUrlId = "test-id"

type mockShortener struct{}

func (mockShortener) ShortenURL(url string) string { return testShorteUrlId }

func NewMockShortener() *mockShortener {
	return &mockShortener{}
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
				response:    fmt.Sprintf("http://localhost:8080/%s", testShorteUrlId),
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := repository.NewMemStorage()
			mux := http.NewServeMux()
			shortener := NewMockShortener()
			h := handler.NewHandler(store, shortener)
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

	store := repository.NewMemStorage()
	store.Save("123", "https://example.com")

	tests := []struct {
		name  string // description of this test case
		store *repository.MemStorage
		r     *http.Request
		want  want
	}{
		{
			name:  "check get request with invalid method",
			store: repository.NewMemStorage(),
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
			h := handler.NewHandler(tt.store, shortener)
			mux.HandleFunc("/{id}", h.GetShortURLByID)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, tt.r)

			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, tt.want.response, rr.Body.String())
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
		})
	}
}
