// Package model defines the request and response payloads exchanged by the URL
// shortener HTTP API, together with the records passed between the service and
// storage layers.
package model

// CreateShortURLRequest is the JSON body of POST /api/shorten.
type CreateShortURLRequest struct {
	URL string `json:"url"`
}

// CreateShortURLResponce is the JSON response of POST /api/shorten and carries
// the resulting short URL.
type CreateShortURLResponce struct {
	Result string `json:"result"`
}

// BatchShortURLRequestItem is a single entry in the POST /api/shorten/batch
// request. CorrelationID is echoed back so the caller can match responses to
// requests.
type BatchShortURLRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortURLResponseItem is a single entry in the POST /api/shorten/batch
// response, pairing the caller's CorrelationID with the generated ShortURL.
type BatchShortURLResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ErrorResponse is the JSON body returned for failed API requests.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ShortURLRecord pairs a generated short ID with its original URL when a batch
// is persisted.
type ShortURLRecord struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// UserURLItem is a single short/original URL pair owned by a user, as returned
// by GET /api/user/urls.
type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// StatsResponse is the JSON response of GET /api/internal/stats and reports
// the service-wide number of shortened URLs and users.
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// DeleteItem identifies a single short URL to delete on behalf of its owner.
type DeleteItem struct {
	UserID  string
	ShortID string
}
