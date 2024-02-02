package httphandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.ErrBadRequest
	}

	err = h.importUsecase.Import(c.Context(), file.Filename)
	if err != nil {
		log.Warn(err)
		return fiber.ErrInternalServerError
	}

	return c.SendStatus(fiber.StatusCreated)
}
