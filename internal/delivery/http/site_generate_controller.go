package http

import (
	"strconv"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// SiteGenerateController exposes the /api/sites/generate/* endpoints used to
// bulk-create dummy emails, fields, and entry-history rows for a site.
type SiteGenerateController struct {
	gen *usecase.SiteGenerateUseCase
}

func NewSiteGenerateController(gen *usecase.SiteGenerateUseCase) *SiteGenerateController {
	return &SiteGenerateController{gen: gen}
}

func parseSiteID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("site_id"), 10, 64)
	if err != nil {
		return 0, usecase.ErrNotFound
	}
	return id, nil
}

func parseAmount(c *fiber.Ctx) int {
	n, _ := strconv.Atoi(c.Query("amount"))
	return n
}

func parseLength(c *fiber.Ctx) int {
	n, _ := strconv.Atoi(c.Query("length"))
	return n
}

func (h *SiteGenerateController) Emails(c *fiber.Ctx) error {
	siteID, err := parseSiteID(c)
	if err != nil {
		return err
	}
	resp, err := h.gen.GenerateEmails(c.Context(), siteID, parseAmount(c), parseLength(c))
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

func (h *SiteGenerateController) Fields(c *fiber.Ctx) error {
	siteID, err := parseSiteID(c)
	if err != nil {
		return err
	}
	resp, err := h.gen.GenerateFields(c.Context(), siteID, parseAmount(c))
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

func (h *SiteGenerateController) EntryUpdates(c *fiber.Ctx) error {
	siteID, err := parseSiteID(c)
	if err != nil {
		return err
	}
	resp, err := h.gen.GenerateEntryUpdates(c.Context(), siteID, parseAmount(c))
	if err != nil {
		return err
	}
	return c.JSON(resp)
}
