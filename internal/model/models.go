package model

// generate:reset
type SetURLJsonRequest struct {
	URL string `json:"url"`
}

// generate:reset
type SetURLJsonResponse struct {
	URL string `json:"result"`
}

// generate:reset
type SetArrayURLRequest struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string
}

// generate:reset
type SetArrayURLResponse struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

// generate:reset
type GetArrayURLResponse struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}
