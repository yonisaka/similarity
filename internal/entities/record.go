package entities

type Record struct {
	Combined  string    `json:"combined"`
	Embedding []float64 `json:"embedding"`
}
