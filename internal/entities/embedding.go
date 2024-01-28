package entities

// EmbeddingResponse represents the expected structure of the response from your embedding API
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
