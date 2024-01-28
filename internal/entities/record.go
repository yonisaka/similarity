package entities

// Record represents a record in the data source.
type Record struct {
	Combined  string    `json:"combined"`
	Embedding []float64 `json:"embedding"`
}
