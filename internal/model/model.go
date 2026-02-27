package model

type CreateShortURLRequest struct {
	URL string `json:"url"`
}

type CreateShortURLResponce struct {
	Result string `json:"result"`
}
