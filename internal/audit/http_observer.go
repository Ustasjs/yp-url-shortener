package audit

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPObserver is an audit Observer that delivers events as JSON to a remote
// HTTP endpoint.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver returns an HTTPObserver that POSTs events to url with a short
// timeout.
func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Notify sends event as a JSON POST request to the configured URL. Errors and
// non-2xx responses are logged rather than returned.
func (h *HTTPObserver) Notify(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit: marshal event failed", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(data))
	if err != nil {
		logger.Log.Error("audit: build http request failed", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		logger.Log.Error("audit: http send failed", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Log.Error("audit: http sink returned bad status", zap.Int("status", resp.StatusCode))
	}
}
