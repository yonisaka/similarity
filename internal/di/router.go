package di

import "github.com/gofiber/fiber/v2"

func GetRouter(app *fiber.App) {
	// API Group
	api := app.Group("/api")
	v1 := api.Group("/v1")

	importHandler := GetImportHandler()
	v1.Post("/import", importHandler.Import)

	searchHandler := GetSearchHandler()
	v1.Post("/search", searchHandler.Search)
}
