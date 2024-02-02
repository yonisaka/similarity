package di

import "github.com/yonisaka/similarity/internal/adapters/httphandler"

// GetImportHandler is a function to get http openAI handler
func GetImportHandler() httphandler.ImportHandler {
	return httphandler.NewImportHandler(
		GetImportUsecase(),
	)
}

// GetSearchHandler is a function to get http openAI handler
func GetSearchHandler() httphandler.SearchHandler {
	return httphandler.NewSearchHandler(
		GetSearchUsecase(),
	)
}
