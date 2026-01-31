package handler_test

import (
	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_CreateShortURL(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		store *repository.MemStorage
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewHandler(tt.store)
			h.CreateShortURL(tt.w, tt.r)
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
			h := handler.NewHandler(tt.store)
			mux.HandleFunc("/{id}", h.GetShortURLByID)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, tt.r)

			assert.Equal(t, tt.want.code, rr.Code)
			assert.Equal(t, tt.want.response, rr.Body.String())
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
		})
	}
}
