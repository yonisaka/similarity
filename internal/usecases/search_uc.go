package usecases

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/internal/entities"
	"github.com/yonisaka/similarity/internal/entities/repository"
	"github.com/yonisaka/similarity/pkg/qdrant"
)

type searchUsecase struct {
	client        openai.Client
	qdrantClient  qdrant.QdrantClient
	embeddingRepo repository.EmbeddingRepo
}

func NewSearchUsecase(client openai.Client, qdrantClient qdrant.QdrantClient, embeddingRepo repository.EmbeddingRepo) SearchUsecase {
	return &searchUsecase{
		client:        client,
		qdrantClient:  qdrantClient,
		embeddingRepo: embeddingRepo,
	}
}

type SearchUsecase interface {
	Search(ctx context.Context, query string) (string, error)
	StringsRankedByRelatedness(query string, records []repository.Embedding, topN int) ([]entities.StringAndRelatedness, error)
	EmbeddingQuery(query string) ([]float64, error)
	NumTokens(text string) int
	QueryMessage(query string, records []entities.StringAndRelatedness, tokenBudget int) string
	Ask(ctx context.Context, query string, records []entities.StringAndRelatedness, tokenBudget int) (string, error)
	LoadJSONDataSources(path string) ([]repository.Embedding, error)
}
