package messaging

import (
	"context"
	"encoding/json"

	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
)

// HandlerDeps bundles the use cases needed by job handlers.
type HandlerDeps struct {
	Fields           *usecase.FrmFieldUseCase
	EntryHistory     *usecase.FrmEntryHistoryUseCase
	EmailLog         *usecase.FrmEmailLogUseCase
	ShipmentHistory  *usecase.FrmEpShipmentHistoryUseCase
}

func RegisterHandlers(d *Dispatcher, deps HandlerDeps) {
	d.Register(gw.JobUpdateFrmFields, updateFrmFieldsHandler(deps))
	d.Register(gw.JobUpdateFrmEntryHistory, updateFrmEntryHistoryHandler(deps))
	d.Register(gw.JobUpdateEmailsLog, updateEmailsLogHandler(deps))
	d.Register(gw.JobUpdateEpShipmentHistory, updateEpShipmentHistoryHandler(deps))
}

func updateFrmFieldsHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, env gw.JobEnvelope) error {
		var payload []model.FrmFieldInput
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return deps.Fields.UpdateAll(ctx, env.SiteID, payload)
	}
}

func updateFrmEntryHistoryHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, env gw.JobEnvelope) error {
		var payload []model.EntryHistoryInput
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return deps.EntryHistory.Update(ctx, env.SiteID, payload)
	}
}

func updateEmailsLogHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, env gw.JobEnvelope) error {
		var payload []model.FrmEmailLogInput
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		_, err := deps.EmailLog.UpdateAll(ctx, env.SiteID, payload)
		return err
	}
}

func updateEpShipmentHistoryHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, env gw.JobEnvelope) error {
		var payload []model.EpShipmentHistoryInput
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		_, err := deps.ShipmentHistory.Update(ctx, env.SiteID, payload)
		return err
	}
}
