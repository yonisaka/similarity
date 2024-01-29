package usecases

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/internal/entities/repository"
	"github.com/yonisaka/similarity/pkg/qdrant"
)

type importUsecase struct {
	client        openai.Client
	qdrantClient  qdrant.QdrantClient
	embeddingRepo repository.EmbeddingRepo
}

func NewImportUsecase(client openai.Client, qdrantClient qdrant.QdrantClient, embeddingRepo repository.EmbeddingRepo) ImportUsecase {
	return &importUsecase{
		client:        client,
		qdrantClient:  qdrantClient,
		embeddingRepo: embeddingRepo,
	}
}

type ImportUsecase interface {
	Import(ctx context.Context, filename string, offset int) error
	MigrateToQdrant(ctx context.Context) error
	ReadCSV(filename string) ([]string, []string, error)
}
