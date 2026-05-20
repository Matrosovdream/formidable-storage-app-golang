package messaging

import (
	"encoding/json"
	"time"
)

// Job names — keep stable; they are the wire identifiers between producer and consumer.
const (
	JobUpdateFrmFields         = "UpdateFrmFields"
	JobUpdateFrmEntryHistory   = "UpdateFrmEntryHistory"
	JobUpdateEmailsLog         = "UpdateEmailsLog"
	JobUpdateEpShipmentHistory = "UpdateEpShipmentHistory"
)

type JobEnvelope struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	SiteID     int64           `json:"site_id"`
	Payload    json.RawMessage `json:"payload"`
	Attempts   int             `json:"attempts"`
	EnqueuedAt time.Time       `json:"enqueued_at"`
}

// StatTypeFor maps a job name to its queue-stats counter key.
func StatTypeFor(jobName string) string {
	switch jobName {
	case JobUpdateFrmFields:
		return "fields"
	case JobUpdateFrmEntryHistory:
		return "entry_history"
	case JobUpdateEmailsLog:
		return "emails"
	case JobUpdateEpShipmentHistory:
		return "shipment_history"
	}
	return "unknown"
}
