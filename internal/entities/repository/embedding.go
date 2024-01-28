package repository

import (
	"context"
	"time"
)

// Embedding is an embedding entity.
type Embedding struct {
	Scope     string     `json:"scope"`
	Combined  string     `json:"combined"`
	Embedding []float64  `json:"embedding"`
	NTokens   int        `json:"n_tokens"`
	CreatedAt *time.Time `json:"created_at"`
}

type EmbeddingRepo interface {
	ListEmbeddingByScope(ctx context.Context, scope string) ([]Embedding, error)
}
