package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/model"
	"Ustasjs/yp-url-shortener/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetUserURLs(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		setupCookie    func(*http.Request)
		setupShortener func(*handler.MockShortener)
		expectedStatus int
		expectedBody   string
		checkJSON      bool
	}{
		{
			name: "no cookie - new user created - no urls - 204",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "new-user-id")
			},
			setupCookie: func(r *http.Request) {
			},
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().GetUserURLs(mock.Anything, "new-user-id").Return(nil, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
			checkJSON:      false,
		},
		{
			name: "valid cookie - user has no urls - 204",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			setupCookie: func(r *http.Request) {
				validToken, _ := service.BuildJWTString("user-123")
				r.AddCookie(&http.Cookie{
					Name:  middleware.AuthCookieName,
					Value: validToken,
				})
			},
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().GetUserURLs(mock.Anything, "user-123").Return(nil, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
			checkJSON:      false,
		},
		{
			name: "valid cookie - user has urls - 200 with json",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			setupCookie: func(r *http.Request) {
				validToken, _ := service.BuildJWTString("user-123")
				r.AddCookie(&http.Cookie{
					Name:  middleware.AuthCookieName,
					Value: validToken,
				})
			},
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().GetUserURLs(mock.Anything, "user-123").Return([]model.UserURLItem{
					{ShortURL: testShortURL, OriginalURL: "https://example.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name: "invalid cookie - 401",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  middleware.AuthCookieName,
					Value: "invalid-token",
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
			checkJSON:      false,
		},
		{
			name: "no user id in context - 401",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupCookie: func(r *http.Request) {
				validToken, _ := service.BuildJWTString("user-123")
				r.AddCookie(&http.Cookie{
					Name:  middleware.AuthCookieName,
					Value: validToken,
				})
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
			checkJSON:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			r = r.WithContext(tt.setupContext())
			tt.setupCookie(r)

			shortener := handler.NewMockShortener(t)
			if tt.setupShortener != nil {
				tt.setupShortener(shortener)
			}
			h := handler.NewHandler(shortener, newPingerOk(t), newNoopAuditor(t))

			rr := httptest.NewRecorder()

			handlerFunc := http.HandlerFunc(h.GetUserURLs)
			wrappedHandler := middleware.RequireAuth()(handlerFunc)
			wrappedHandler.ServeHTTP(rr, r)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkJSON {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var urls []model.UserURLItem
				err := json.NewDecoder(rr.Body).Decode(&urls)
				require.NoError(t, err)
				assert.Greater(t, len(urls), 0)

				for _, url := range urls {
					assert.NotEmpty(t, url.ShortURL)
					assert.NotEmpty(t, url.OriginalURL)
					assert.Contains(t, url.ShortURL, "http://")
				}
			} else if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
