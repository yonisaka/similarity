package httphandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/yonisaka/similarity/internal/types"
	"github.com/yonisaka/similarity/internal/usecases"
)

type searchHandler struct {
	searchUsecase usecases.SearchUsecase
}

func NewSearchHandler(searchUsecase usecases.SearchUsecase) SearchHandler {
	return &searchHandler{
		searchUsecase: searchUsecase,
	}
}

type SearchHandler interface {
	Search(c *fiber.Ctx) error
}

func (h *searchHandler) Search(c *fiber.Ctx) error {
	prompt := c.FormValue("prompt")
	answer, err := h.searchUsecase.Search(c.Context(), prompt)
	if err != nil {
		log.Warn(err)
		return c.JSON(fiber.ErrInternalServerError)
	}

	result := types.SearchResponse{
		Question: prompt,
		Answer:   answer,
	}

	return c.JSON(
		types.Http{
			Code:    fiber.StatusOK,
			Message: "Success",
			Data:    result,
		},
	)
}
