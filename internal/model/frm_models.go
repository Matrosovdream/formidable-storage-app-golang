package model

import "time"

// ---------- FrmField ----------

type FrmFieldInput struct {
	FieldID int64  `json:"field_id" validate:"required"`
	Key     string `json:"key"`
	Type    string `json:"type"`
	Label   string `json:"label"`
}

type UpdateFieldsRequest struct {
	Data []FrmFieldInput `json:"data" validate:"required,dive"`
}

type FrmFieldResponse struct {
	ID        int64     `json:"id"`
	FieldID   *int64    `json:"field_id"`
	SiteID    int64     `json:"site_id"`
	Key       *string   `json:"key"`
	Type      *string   `json:"type"`
	Label     *string   `json:"label"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------- FrmEntryHistory ----------

type EntryHistoryInput struct {
	EntryID          int64   `json:"entry_id"           validate:"required"`
	FieldID          int64   `json:"field_id"           validate:"required"`
	UserID           *int64  `json:"user_id"`
	UpdateTypeCode   string  `json:"update_type_code"`
	OldValue         *string `json:"old_value"`
	NewValue         *string `json:"new_value"`
	ChangeDate       *string `json:"change_date"`
}

type UpdateEntryHistoryRequest struct {
	Data []EntryHistoryInput `json:"data" validate:"required,dive"`
}

type EntryUpdateItem struct {
	ID         int64      `json:"id"`
	FieldID    int64      `json:"field_id"`
	FieldKey   *string    `json:"field_key"`
	FieldLabel *string    `json:"field_label"`
	UpdateType *string    `json:"update_type"`
	OldValue   *string    `json:"old_value"`
	NewValue   *string    `json:"new_value"`
	ChangeDate *time.Time `json:"change_date"`
	CreatedAt  time.Time  `json:"created_at"`
}

type EntryUpdatesResponse struct {
	EntryID int64             `json:"entry_id"`
	Updates []EntryUpdateItem `json:"updates"`
}

// ---------- FrmEmailLog ----------

type FrmEmailLogInput struct {
	EntryID       *int64  `json:"entry_id"`
	FormID        *int64  `json:"form_id"`
	Subject       *string `json:"subject"`
	MessageID     *string `json:"message_id"`
	EmailFrom     *string `json:"email_from"`
	EmailTo       *string `json:"email_to"`
	People        *string `json:"people"`
	Headers       *string `json:"headers"`
	ErrorText     *string `json:"error_text"`
	ContentPlain  *string `json:"content_plain"`
	ContentHTML   *string `json:"content_html"`
	Status        int16   `json:"status"`
	DateSent      *string `json:"date_sent"`
	Mailer        *string `json:"mailer"`
	Attachments   int16   `json:"attachments"`
	InitiatorName *string `json:"initiator_name"`
	InitiatorFile *string `json:"initiator_file"`
	OriginalLogID *int64  `json:"original_log_id"`
}

type UpdateEmailsLogRequest struct {
	Data []FrmEmailLogInput `json:"data" validate:"required,dive"`
}

type ListEmailsLogRequest struct {
	Filters  map[string]any `json:"filters"`
	SortBy   string         `json:"sort_by"`
	SortDir  string         `json:"sort_dir"`
	Paginate int            `json:"paginate"`
	PageNum  int            `json:"page_num"`
}

type EmailLogItem struct {
	ID            int64      `json:"id"`
	EntryID       *int64     `json:"entry_id"`
	SiteID        int64      `json:"site_id"`
	FormID        *int64     `json:"form_id"`
	Subject       *string    `json:"subject"`
	MessageID     *string    `json:"message_id"`
	EmailFrom     *string    `json:"email_from"`
	EmailTo       *string    `json:"email_to"`
	Status        int16      `json:"status"`
	DateSent      *time.Time `json:"date_sent"`
	Mailer        *string    `json:"mailer"`
	Attachments   int16      `json:"attachments"`
	InitiatorName *string    `json:"initiator_name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ---------- FrmEpShipmentHistory ----------

type EpShipmentHistoryInput struct {
	EasypostShipmentID *string `json:"easypost_shipment_id"`
	UserID             *int64  `json:"user_id"`
	ChangeType         *string `json:"change_type"`
	Description        *string `json:"description"`
}

type UpdateEpShipmentHistoryRequest struct {
	Data []EpShipmentHistoryInput `json:"data" validate:"required,dive"`
}

type ListEpShipmentHistoryRequest struct {
	Filters  map[string]any `json:"filters"`
	SortBy   string         `json:"sort_by"`
	SortDir  string         `json:"sort_dir"`
	Paginate int            `json:"paginate"`
	PageNum  int            `json:"page_num"`
}

type EpShipmentHistoryItem struct {
	ID                 int64     `json:"id"`
	EasypostShipmentID *string   `json:"easypost_shipment_id"`
	UserID             *int64    `json:"user_id"`
	SiteID             int64     `json:"site_id"`
	ChangeType         *string   `json:"change_type"`
	Description        *string   `json:"description"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ---------- Data (per-entry summaries) ----------

type DataEntryRow struct {
	EntryID     int64      `json:"entry_id"`
	SiteID      int64      `json:"site_id"`
	LastUpdate  *time.Time `json:"last_update"`
	EmailCount  int64      `json:"email_count"`
	UpdateCount int64      `json:"update_count"`
}

type DataEntriesRequest struct {
	EntryID *int64 `query:"entry_id" json:"entry_id"`
	SortBy  string `query:"sort_by"  json:"sort_by"`
	SortDir string `query:"sort_dir" json:"sort_dir"`
	Page    int    `query:"page"     json:"page"`
	PerPage int    `query:"per_page" json:"per_page"`
}
