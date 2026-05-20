package http

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type RestV1EpShipmentHistoryController struct {
	history *usecase.FrmEpShipmentHistoryUseCase
	queue   *gw.QueueProducer
}

func NewRestV1EpShipmentHistoryController(history *usecase.FrmEpShipmentHistoryUseCase, queue *gw.QueueProducer) *RestV1EpShipmentHistoryController {
	return &RestV1EpShipmentHistoryController{history: history, queue: queue}
}

func (h *RestV1EpShipmentHistoryController) UpdateAll(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.UpdateEpShipmentHistoryRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	if err := h.queue.Enqueue(c.Context(), gw.JobUpdateEpShipmentHistory, site.ID, req.Data); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"queued": true, "count": len(req.Data)})
}

func (h *RestV1EpShipmentHistoryController) List(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.ListEpShipmentHistoryRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	res, err := h.history.List(c.Context(), site.ID, req)
	if err != nil {
		return err
	}
	return c.JSON(model.BuildPaginated(c.OriginalURL(), res.Data, res.Total, res.Page, res.PerPage, res.LastPage))
}
