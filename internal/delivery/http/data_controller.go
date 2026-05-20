package http

import (
	"strconv"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type DataController struct {
	data *usecase.DataUseCase
}

func NewDataController(data *usecase.DataUseCase) *DataController {
	return &DataController{data: data}
}

func (h *DataController) Entries(c *fiber.Ctx) error {
	siteID, err := strconv.ParseInt(c.Params("site_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	var req model.DataEntriesRequest
	if err := c.QueryParser(&req); err != nil {
		return err
	}
	res, err := h.data.ListEntries(c.Context(), siteID, req)
	if err != nil {
		return err
	}
	return c.JSON(model.BuildPaginated(c.OriginalURL(), res.Data, res.Total, res.Page, res.PerPage, res.LastPage))
}

func (h *DataController) EntryUpdates(c *fiber.Ctx) error {
	siteID, err := strconv.ParseInt(c.Params("site_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	entryID, err := strconv.ParseInt(c.Params("entry_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	resp, err := h.data.EntryUpdates(c.Context(), siteID, entryID)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

func (h *DataController) EntryEmails(c *fiber.Ctx) error {
	siteID, err := strconv.ParseInt(c.Params("site_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	entryID, err := strconv.ParseInt(c.Params("entry_id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	items, err := h.data.EntryEmails(c.Context(), siteID, entryID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": items})
}
