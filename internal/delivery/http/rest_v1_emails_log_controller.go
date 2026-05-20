package http

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type RestV1EmailsLogController struct {
	emails *usecase.FrmEmailLogUseCase
	queue  *gw.QueueProducer
}

func NewRestV1EmailsLogController(emails *usecase.FrmEmailLogUseCase, queue *gw.QueueProducer) *RestV1EmailsLogController {
	return &RestV1EmailsLogController{emails: emails, queue: queue}
}

// UpdateAll enqueues the rows for async processing.
func (h *RestV1EmailsLogController) UpdateAll(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.UpdateEmailsLogRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	if err := h.queue.Enqueue(c.Context(), gw.JobUpdateEmailsLog, site.ID, req.Data); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"queued": true, "count": len(req.Data)})
}

// UpdateAllRaw runs the upsert synchronously and returns the affected count.
func (h *RestV1EmailsLogController) UpdateAllRaw(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.UpdateEmailsLogRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	n, err := h.emails.UpdateAllRaw(c.Context(), site.ID, req.Data)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "count": n})
}

func (h *RestV1EmailsLogController) List(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.ListEmailsLogRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	res, err := h.emails.List(c.Context(), site.ID, req)
	if err != nil {
		return err
	}
	return c.JSON(model.BuildPaginated(c.OriginalURL(), res.Data, res.Total, res.Page, res.PerPage, res.LastPage))
}
