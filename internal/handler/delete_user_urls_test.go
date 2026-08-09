package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/service/shortener"

	"github.com/stretchr/testify/assert"
)

func TestHandler_DeleteUserURLs(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		body           interface{}
		setupShortener func(*handler.MockShortener)
		expectedStatus int
	}{
		{
			name: "successful deletion - returns 202",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body: []string{"short1", "short2", "short3"},
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					DeleteURLsAsync("user-123", []string{"short1", "short2", "short3"}).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "service overloaded - returns 503",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body: []string{"short1", "short2"},
			setupShortener: func(m *handler.MockShortener) {
				m.EXPECT().
					DeleteURLsAsync("user-123", []string{"short1", "short2"}).
					Return(shortener.ErrServiceOverloaded)
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "no user context - returns 401",
			setupContext: func() context.Context {
				return context.Background()
			},
			body:           []string{"short1"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "empty array - returns 400",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body:           []string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid json - returns 400",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body:           "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(bodyBytes))
			r = r.WithContext(tt.setupContext())

			shortenerMock := handler.NewMockShortener(t)
			if tt.setupShortener != nil {
				tt.setupShortener(shortenerMock)
			}

			h := handler.NewHandler(shortenerMock, newPingerOk(t), newNoopAuditor(t))
			rr := httptest.NewRecorder()

			h.DeleteUserURLs(rr, r)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
