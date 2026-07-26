package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetInternalStats(t *testing.T) {
	m := newMockShortener()
	m.urls["abc123"] = "https://example.com"
	m.urls["def456"] = "https://example.org"

	h := handler.NewHandler(m, newMockPingerOk(), newNoopAuditor())

	r := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	rr := httptest.NewRecorder()

	h.GetInternalStats(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var stats model.StatsResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&stats))
	assert.Equal(t, 2, stats.URLs)
	assert.Equal(t, 1, stats.Users)
}
