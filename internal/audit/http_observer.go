package audit

import (
	"encoding/json"
	"net/http"
	"time"

	"Ustasjs/yp-url-shortener/internal/logger"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

// HTTPObserver is an audit Observer that delivers events as JSON to a remote
// HTTP endpoint.
type HTTPObserver struct {
	url    string
	client *retryablehttp.Client
}

// NewHTTPObserver returns an HTTPObserver that POSTs events to url using a
// retrying HTTP client.
func NewHTTPObserver(url string) *HTTPObserver {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 100 * time.Millisecond
	client.RetryWaitMax = 1 * time.Second
	client.HTTPClient.Timeout = 5 * time.Second

	return &HTTPObserver{
		url:    url,
		client: client,
	}
}

// Notify sends event as a JSON POST request to the configured URL, retrying on
// transient failures. Errors and non-2xx responses are logged rather than
// returned.
func (h *HTTPObserver) Notify(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit: marshal event failed", zap.Error(err))
		return
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, h.url, data)
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
