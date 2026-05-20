package http

import (
	"strconv"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type SiteController struct {
	site *usecase.SiteUseCase
}

func NewSiteController(site *usecase.SiteUseCase) *SiteController {
	return &SiteController{site: site}
}

func (h *SiteController) List(c *fiber.Ctx) error {
	out, err := h.site.List(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": out})
}

func (h *SiteController) View(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("site_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	resp, err := h.site.View(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

func (h *SiteController) Create(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"form": fiber.Map{
			"fields": []fiber.Map{
				{"name": "name", "required": true, "type": "string"},
				{"name": "url", "required": true, "type": "url"},
			},
		},
	})
}

func (h *SiteController) Store(c *fiber.Ctx) error {
	var req model.CreateSiteRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	resp, err := h.site.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *SiteController) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("site_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	if err := h.site.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true})
}
