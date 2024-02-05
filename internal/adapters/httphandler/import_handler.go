package httphandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/yonisaka/similarity/internal/types"
	"github.com/yonisaka/similarity/internal/usecases"
)

type importHandler struct {
	importUsecase usecases.ImportUsecase
}

func NewImportHandler(importUsecase usecases.ImportUsecase) ImportHandler {
	return &importHandler{
		importUsecase: importUsecase,
	}
}

type ImportHandler interface {
	Import(c *fiber.Ctx) error
}

func (h *importHandler) Import(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(fiber.ErrBadRequest)
	}

	err = h.importUsecase.Import(c.Context(), fileHeader, "")
	if err != nil {
		log.Warn(err)
		return c.JSON(fiber.ErrInternalServerError)
	}

	return c.JSON(types.Http{
		Code:    fiber.StatusCreated,
		Message: "Success importing",
	})
}
