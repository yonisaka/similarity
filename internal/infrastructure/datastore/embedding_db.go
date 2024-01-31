package datastore

import (
	"context"
	"errors"
	"fmt"
	"github.com/yonisaka/similarity/internal/entities/repository"
	"strconv"
	"strings"
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
	query := `SELECT id, combined, translate(embeddings, '[]', '{}')::float[], n_tokens, created_at
				FROM embeddings
					WHERE scope = $1 `

	rows, err := r.dbSlave.Query(ctx, query, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var embeddings []repository.Embedding

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for rows.Next() {
		var embedding repository.Embedding
		if err := rows.Scan(&embedding.ID, &embedding.Combined, &embedding.Embedding, &embedding.NTokens, &embedding.CreatedAt); err != nil {
			return nil, err
		}
		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

func (r *embeddingRepo) CountEmbeddingByScope(ctx context.Context, scope string) (int, error) {
	query := `SELECT COUNT(*)
				FROM embeddings
					WHERE scope = $1 `

	var count int
	if err := r.dbSlave.QueryRow(ctx, query, scope).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *embeddingRepo) CreateEmbedding(ctx context.Context, embedding *repository.Embedding) error {
	query := `INSERT INTO embeddings(scope, combined, embeddings, n_tokens, created_at)
				VALUES($1, $2, $3, $4, NOW())`

	if _, err := r.dbMaster.Exec(ctx, query, embedding.Scope, embedding.Combined, floatArrayToString(embedding.Embedding, ", "), embedding.NTokens); err != nil {
		return err
	}

	return nil
}

func floatArrayToString(arr []float64, delimiter string) string {
	var strArr []string
	for _, num := range arr {
		strArr = append(strArr, strconv.FormatFloat(num, 'f', -1, 64))
	}
	return fmt.Sprintf("[%s]", strings.Join(strArr, delimiter))
}
