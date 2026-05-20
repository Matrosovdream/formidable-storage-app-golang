package http

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type RestV1FieldsController struct {
	queue *gw.QueueProducer
}

func NewRestV1FieldsController(queue *gw.QueueProducer) *RestV1FieldsController {
	return &RestV1FieldsController{queue: queue}
}

func (h *RestV1FieldsController) UpdateAll(c *fiber.Ctx) error {
	site := middleware.SiteFromCtx(c)
	if site == nil {
		return usecase.ErrUnauthorized
	}
	var req model.UpdateFieldsRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	if err := h.queue.Enqueue(c.Context(), gw.JobUpdateFrmFields, site.ID, req.Data); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"queued": true, "count": len(req.Data)})
}
