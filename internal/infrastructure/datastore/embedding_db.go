package datastore

import (
	"context"
	"errors"
	"github.com/yonisaka/similarity/internal/entities/repository"
)

var (
	// ErrNotFound is an error for indicates record not found.
	ErrNotFound = errors.New("error not found")
)

type embeddingRepo struct {
	*BaseRepo
}

// NewEmbeddingRepo returns EmbeddingRepo.
func NewEmbeddingRepo(base *BaseRepo) repository.EmbeddingRepo {
	return &embeddingRepo{
		BaseRepo: base,
	}
}

func (r *embeddingRepo) ListEmbeddingByScope(ctx context.Context, scope string) ([]repository.Embedding, error) {
	query := `SELECT combined, translate(embeddings, '[]', '{}')::float[], n_tokens, created_at
				FROM embeddings
					WHERE scope = $1 `

	rows, err := r.dbSlave.Query(ctx, query, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var embeddings []repository.Embedding
	for rows.Next() {
		var embedding repository.Embedding
		if err := rows.Scan(&embedding.Combined, &embedding.Embedding, &embedding.NTokens, &embedding.CreatedAt); err != nil {
			return nil, err
		}
		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}
