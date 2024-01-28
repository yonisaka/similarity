package di

import "github.com/yonisaka/similarity/internal/usecases"

// GetSearchUsecase returns SearchUsecase instance.
func GetSearchUsecase() usecases.SearchUsecase {
	return usecases.NewSearchUsecase(
		GetOpenAIClient(),
		GetEmbeddingRepo(),
	)
}
