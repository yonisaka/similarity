package di

import (
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/yonisaka/similarity/pkg/elasticsearch"
	"github.com/yonisaka/similarity/pkg/qdrant"
	"os"
	"strconv"
)

// GetOpenAIClient returns OpenAI client instance.
func GetOpenAIClient() openai.Client {
	return *openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

// GetQdrantClient returns Qdrant client instance.
func GetQdrantClient() qdrant.QdrantClient {
	size, err := strconv.Atoi(os.Getenv("QDRANT_SIZE"))
	if err != nil {
		panic(err)
	}

	memmapThreshold, err := strconv.Atoi(os.Getenv("QDRANT_MEMMAP_THRESHOLD"))
	if err != nil {
		panic(err)
	}

	hnswOndisk, err := strconv.ParseBool(os.Getenv("QDRANT_HNSW_ONDISK"))
	if err != nil {
		panic(err)
	}

	hnswM, err := strconv.Atoi(os.Getenv("QDRANT_HNSW_M"))
	if err != nil {
		panic(err)
	}

	hnswEFConstruct, err := strconv.Atoi(os.Getenv("QDRANT_HNSW_EF_CONSTRUCT"))
	if err != nil {
		panic(err)
	}

	return *qdrant.NewQdrantClient(
		fmt.Sprintf("%s:%s", os.Getenv("QDRANT_HOST"), os.Getenv("QDRANT_PORT")),
		os.Getenv("QDRANT_COLLECTION_NAME"),
		uint64(size),
		uint64(memmapThreshold),
		hnswOndisk,
		uint64(hnswM),
		uint64(hnswEFConstruct),
	)
}

func GetESClient() elasticsearch.ESClient {
	return *elasticsearch.NewElasticsearch()
}
