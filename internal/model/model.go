package model

type CreateShortURLRequest struct {
	URL string `json:"url"`
}

type CreateShortURLResponce struct {
	Result string `json:"result"`
}

type BatchShortURLRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortURLResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ShortURLRecord struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
