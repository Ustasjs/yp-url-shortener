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

type mockShortenerOverloaded struct {
	*mockShortener
	overloaded bool
}

func (m *mockShortenerOverloaded) DeleteURLsAsync(userID string, shortIDs []string) error {
	if m.overloaded {
		return shortener.ErrServiceOverloaded
	}
	return nil
}

func TestHandler_DeleteUserURLs(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		body           interface{}
		overloaded     bool
		expectedStatus int
	}{
		{
			name: "successful deletion - returns 202",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body:           []string{"short1", "short2", "short3"},
			overloaded:     false,
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "service overloaded - returns 503",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body:           []string{"short1", "short2"},
			overloaded:     true,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "no user context - returns 401",
			setupContext: func() context.Context {
				return context.Background()
			},
			body:           []string{"short1"},
			overloaded:     false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "empty array - returns 400",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body:           []string{},
			overloaded:     false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid json - returns 400",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDContextKey, "user-123")
			},
			body:           "invalid",
			overloaded:     false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(bodyBytes))
			r = r.WithContext(tt.setupContext())

			shortenerMock := &mockShortenerOverloaded{
				mockShortener: newMockShortener(),
				overloaded:    tt.overloaded,
			}

			h := handler.NewHandler(shortenerMock, newMockPingerOk(), newNoopAuditor())
			rr := httptest.NewRecorder()

			h.DeleteUserURLs(rr, r)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
