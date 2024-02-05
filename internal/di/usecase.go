package di

import "github.com/yonisaka/similarity/internal/usecases"

// GetSearchUsecase returns SearchUsecase instance.
func GetSearchUsecase() usecases.SearchUsecase {
	return usecases.NewSearchUsecase(
		GetOpenAIClient(),
		GetQdrantClient(),
		GetEmbeddingRepo(),
		GetESClient(),
		GetLogger(),
	)
}

// GetImportUsecase returns ImportUsecase instance.
func GetImportUsecase() usecases.ImportUsecase {
	return usecases.NewImportUsecase(
		GetOpenAIClient(),
		GetQdrantClient(),
		GetEmbeddingRepo(),
		GetESClient(),
		GetLogger(),
	)
}
