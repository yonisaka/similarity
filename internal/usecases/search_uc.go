package usecases

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/internal/entities"
)

type searchUsecase struct {
	client *openai.Client
}

func NewSearchUsecase(client *openai.Client) SearchUsecase {
	return &searchUsecase{client: client}
}

type SearchUsecase interface {
	Search(ctx context.Context, query string) (string, error)
	CosineSimilarity(vecA, vecB []float64) float64
	StringsRankedByRelatedness(query string, records []entities.Record, topN int) ([]entities.StringAndRelatedness, error)
	EmbeddingQuery(query string) ([]float64, error)
	NumTokens(text string) int
	QueryMessage(query string, records []entities.StringAndRelatedness, tokenBudget int) string
	Ask(ctx context.Context, query string, records []entities.StringAndRelatedness, tokenBudget int) (string, error)
	LoadDataSources(path string) ([]entities.Record, error)
}
