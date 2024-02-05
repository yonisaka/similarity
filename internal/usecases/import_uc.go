package usecases

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/internal/entities/repository"
	"github.com/yonisaka/similarity/pkg/elasticsearch"
	"github.com/yonisaka/similarity/pkg/logger"
	"github.com/yonisaka/similarity/pkg/qdrant"
	"mime/multipart"
)

type importUsecase struct {
	client        openai.Client
	qdrantClient  qdrant.QdrantClient
	embeddingRepo repository.EmbeddingRepo
	esClient      elasticsearch.ESClient
	logger        logger.Logger
}

func NewImportUsecase(
	client openai.Client,
	qdrantClient qdrant.QdrantClient,
	embeddingRepo repository.EmbeddingRepo,
	esClient elasticsearch.ESClient,
	logger logger.Logger,
) ImportUsecase {
	return &importUsecase{
		client:        client,
		qdrantClient:  qdrantClient,
		embeddingRepo: embeddingRepo,
		esClient:      esClient,
		logger:        logger,
	}
}

type ImportUsecase interface {
	Import(ctx context.Context, fileHeader *multipart.FileHeader, filename string) error
	MigrateToQdrant(ctx context.Context) error
	MigrateToElasticsearch(ctx context.Context) error
	ReadUploadedCSV(fileHeader *multipart.FileHeader) ([]string, []string, error)
	ReadCSV(filename string) ([]string, []string, error)
}
