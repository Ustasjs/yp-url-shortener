package handler_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/handler"
	customMiddleware "Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testShortURL = "http://localhost:8080/test-id"

// newPingerOk returns a Pinger that reports a healthy database. The handlers
// under test never ping, so the expectation is optional.
func newPingerOk(t *testing.T) *handler.MockPinger {
	t.Helper()

	pinger := handler.NewMockPinger(t)
	pinger.EXPECT().PingContext(mock.Anything).Return(nil).Maybe()
	return pinger
}

// newNoopAuditor returns an audit publisher that drops every event, for tests
// that do not assert on the audit trail.
func newNoopAuditor(t *testing.T) *handler.MockAuditPublisher {
	t.Helper()

	auditor := handler.NewMockAuditPublisher(t)
	auditor.EXPECT().Publish(mock.Anything).Maybe()
	return auditor
}

// newAuditRecorder returns an audit publisher that appends every published event
// to the returned slice. Publish runs in the request goroutine, so the slice is
// safe to read once the handler has returned.
func newAuditRecorder(t *testing.T) (*handler.MockAuditPublisher, *[]audit.Event) {
	t.Helper()

	var events []audit.Event
	auditor := handler.NewMockAuditPublisher(t)
	auditor.EXPECT().Publish(mock.Anything).Run(func(event audit.Event) {
		events = append(events, event)
	}).Maybe()
	return auditor, &events
}

func TestHandler_CreateShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name           string
		setupShortener func(*handler.MockShortener)
		r              *http.Request
		want           want
	}{
		{
			name: "check request with invalid method",
			r:    httptest.NewRequest(http.MethodGet, "/", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"Only POST requests are allowed\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check post request with no body",
			r:    httptest.NewRequest(http.MethodPost, "/", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"url is required\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check post request with invalid body",
			r:    httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid body")),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"invalid url\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check post request with valid body",
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					CreateShortURL(mock.Anything, "https://example.com", mock.Anything).
					Return(testShortURL, nil)
			},
			r: httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com")),
			want: want{
				code:        http.StatusCreated,
				response:    testShortURL,
				contentType: "text/plain",
			},
		},
		{
			name: "check post request when url already exists (conflict)",
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					CreateShortURL(mock.Anything, "https://example.com", mock.Anything).
					Return(testShortURL, repository.ErrConflict)
			},
			r: httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com")),
			want: want{
				code:        http.StatusConflict,
				response:    testShortURL,
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			shortener := handler.NewMockShortener(t)
			if tt.setupShortener != nil {
				tt.setupShortener(shortener)
			}
			h := handler.NewHandler(shortener, newPingerOk(t), newNoopAuditor(t))
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

	tests := []struct {
		name           string
		setupShortener func(*handler.MockShortener)
		r              *http.Request
		want           want
	}{
		{
			name: "check get request with invalid method",
			r:    httptest.NewRequest(http.MethodPost, "/123", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"Only GET requests are allowed\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check get request with valid id",
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					GetOriginalURL(mock.Anything, "123").
					Return("https://example.com", nil)
			},
			r: httptest.NewRequest(http.MethodGet, "/123", nil),
			want: want{
				code:        http.StatusTemporaryRedirect,
				response:    "<a href=\"https://example.com\">Temporary Redirect</a>.\n\n",
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name: "check get request with invalid id",
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					GetOriginalURL(mock.Anything, "456").
					Return("", repository.ErrRecordNotFound)
			},
			r: httptest.NewRequest(http.MethodGet, "/456", nil),
			want: want{
				code:        http.StatusNotFound,
				response:    "{\"error\":\"url not found\"}\n",
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			shortener := handler.NewMockShortener(t)
			if tt.setupShortener != nil {
				tt.setupShortener(shortener)
			}
			h := handler.NewHandler(shortener, newPingerOk(t), newNoopAuditor(t))
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
		name           string
		setupShortener func(*handler.MockShortener)
		r              *http.Request
		want           want
	}{
		{
			name: "check request with invalid method",
			r:    httptest.NewRequest(http.MethodGet, "/api/shorten", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"Only POST requests are allowed\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check post request with no body",
			r:    httptest.NewRequest(http.MethodPost, "/api/shorten", nil),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"cannot decode request JSON body\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check post request with invalid body",
			r:    httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("invalid body")),
			want: want{
				code:        http.StatusBadRequest,
				response:    "{\"error\":\"cannot decode request JSON body\"}\n",
				contentType: "application/json",
			},
		},
		{
			name: "check post request with valid body",
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					CreateShortURL(mock.Anything, "https://practicum.yandex.ru/", mock.Anything).
					Return(testShortURL, nil)
			},
			r: httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("{ \"url\": \"https://practicum.yandex.ru/\"}")),
			want: want{
				code:        http.StatusCreated,
				response:    fmt.Sprintf("{\"result\":\"%s\"}\n", testShortURL),
				contentType: "application/json",
			},
		},
		{
			name: "check post request when url already exists (conflict)",
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					CreateShortURL(mock.Anything, "https://practicum.yandex.ru/", mock.Anything).
					Return(testShortURL, repository.ErrConflict)
			},
			r: httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://practicum.yandex.ru/"}`)),
			want: want{
				code:        http.StatusConflict,
				response:    fmt.Sprintf("{\"result\":\"%s\"}\n", testShortURL),
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			shortener := handler.NewMockShortener(t)
			if tt.setupShortener != nil {
				tt.setupShortener(shortener)
			}
			h := handler.NewHandler(shortener, newPingerOk(t), newNoopAuditor(t))
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

	mux := http.NewServeMux()
	shortener := handler.NewMockShortener(t)
	shortener.EXPECT().
		CreateShortURL(mock.Anything, "https://example.com", mock.Anything).
		Return(testShortURL, nil)
	h := handler.NewHandler(shortener, newPingerOk(t), newNoopAuditor(t))

	wrapped := customMiddleware.GzipDecompress(http.HandlerFunc(h.CreateShortURL))
	mux.Handle("/", wrapped)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, testShortURL, rr.Body.String())
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
}

func TestHandler_CreateShortURL_PublishesAuditEvent(t *testing.T) {
	shortener := handler.NewMockShortener(t)
	shortener.EXPECT().
		CreateShortURL(mock.Anything, "https://example.com", "user-42").
		Return(testShortURL, nil)
	auditor, events := newAuditRecorder(t)
	h := handler.NewHandler(shortener, newPingerOk(t), auditor)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	ctx := context.WithValue(r.Context(), customMiddleware.UserIDContextKey, "user-42")
	r = r.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.CreateShortURL(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	assert.Len(t, *events, 1)
	assert.Equal(t, audit.ActionShorten, (*events)[0].Action)
	assert.Equal(t, "user-42", (*events)[0].UserID)
	assert.Equal(t, "https://example.com", (*events)[0].URL)
	assert.NotZero(t, (*events)[0].Timestamp)
}

func TestHandler_CreateShortURL_NoAuditOnError(t *testing.T) {
	auditor := handler.NewMockAuditPublisher(t)
	h := handler.NewHandler(handler.NewMockShortener(t), newPingerOk(t), auditor)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-valid-url"))
	rr := httptest.NewRecorder()
	h.CreateShortURL(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	auditor.AssertNotCalled(t, "Publish")
}

func TestHandler_CreateShortURLJSONApi_PublishesAuditEvent(t *testing.T) {
	shortener := handler.NewMockShortener(t)
	shortener.EXPECT().
		CreateShortURL(mock.Anything, "https://practicum.yandex.ru/", "user-7").
		Return(testShortURL, nil)
	auditor, events := newAuditRecorder(t)
	h := handler.NewHandler(shortener, newPingerOk(t), auditor)

	r := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://practicum.yandex.ru/"}`))
	ctx := context.WithValue(r.Context(), customMiddleware.UserIDContextKey, "user-7")
	r = r.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.CreateShortURLJSONApi(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	assert.Len(t, *events, 1)
	assert.Equal(t, audit.ActionShorten, (*events)[0].Action)
	assert.Equal(t, "user-7", (*events)[0].UserID)
	assert.Equal(t, "https://practicum.yandex.ru/", (*events)[0].URL)
}

func TestHandler_GetShortURLByID_PublishesAuditEvent(t *testing.T) {
	shortener := handler.NewMockShortener(t)
	shortener.EXPECT().
		GetOriginalURL(mock.Anything, "abc").
		Return("https://example.com/path", nil)
	auditor, events := newAuditRecorder(t)
	h := handler.NewHandler(shortener, newPingerOk(t), auditor)

	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", h.GetShortURLByID)

	r := httptest.NewRequest(http.MethodGet, "/abc", nil)
	ctx := context.WithValue(r.Context(), customMiddleware.UserIDContextKey, "user-9")
	r = r.WithContext(ctx)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)

	assert.Len(t, *events, 1)
	assert.Equal(t, audit.ActionFollow, (*events)[0].Action)
	assert.Equal(t, "user-9", (*events)[0].UserID)
	assert.Equal(t, "https://example.com/path", (*events)[0].URL)
}

func TestHandler_GetShortURLByID_NoAuditOnNotFound(t *testing.T) {
	shortener := handler.NewMockShortener(t)
	shortener.EXPECT().
		GetOriginalURL(mock.Anything, "missing").
		Return("", repository.ErrRecordNotFound)
	auditor := handler.NewMockAuditPublisher(t)
	h := handler.NewHandler(shortener, newPingerOk(t), auditor)

	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", h.GetShortURLByID)

	r := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, r)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	auditor.AssertNotCalled(t, "Publish")
}
