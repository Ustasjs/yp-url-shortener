package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"Ustasjs/yp-url-shortener/internal/middleware"

	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnet(t *testing.T) {
	tests := []struct {
		name           string
		cidr           string
		realIP         string
		expectedStatus int
	}{
		{
			name:           "empty subnet forbids any request",
			cidr:           "",
			realIP:         "192.168.1.10",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "ip inside subnet is allowed",
			cidr:           "192.168.1.0/24",
			realIP:         "192.168.1.10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ip outside subnet is forbidden",
			cidr:           "192.168.1.0/24",
			realIP:         "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing X-Real-IP is forbidden",
			cidr:           "192.168.1.0/24",
			realIP:         "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "malformed X-Real-IP is forbidden",
			cidr:           "192.168.1.0/24",
			realIP:         "not-an-ip",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := middleware.TrustedSubnet(tt.cidr)(next)

			r := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			if tt.realIP != "" {
				r.Header.Set(middleware.RealIPHeader, tt.realIP)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, r)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
