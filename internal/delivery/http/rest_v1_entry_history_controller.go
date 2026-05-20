package http

import (
	"strconv"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type RestV1EntryHistoryController struct {
	history *usecase.FrmEntryHistoryUseCase
	queue   *gw.QueueProducer
}

func NewRestV1EntryHistoryController(history *usecase.FrmEntryHistoryUseCase, queue *gw.QueueProducer) *RestV1EntryHistoryController {
	return &RestV1EntryHistoryController{history: history, queue: queue}
}

func (h *RestV1EntryHistoryController) Update(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.UpdateEntryHistoryRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	if err := h.queue.Enqueue(c.Context(), gw.JobUpdateFrmEntryHistory, site.ID, req.Data); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"queued": true, "count": len(req.Data)})
}

func (h *RestV1EntryHistoryController) GetEntryHistory(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	entryID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return usecase.ErrNotFound
	}
	resp, err := h.history.GetByEntry(c.Context(), site.ID, entryID)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}
