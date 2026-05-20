package entity

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type FrmEpShipment struct {
	ID                 int64          `db:"id"`
	EasypostShipmentID sql.NullString `db:"easypost_shipment_id"`
	EntryID            sql.NullInt64  `db:"entry_id"`
	SiteID             int64          `db:"site_id"`
	IsReturn           bool           `db:"is_return"`
	Status             sql.NullString `db:"status"`
	TrackingCode       sql.NullString `db:"tracking_code"`
	TrackingURL        sql.NullString `db:"tracking_url"`
	RefundStatus       sql.NullString `db:"refund_status"`
	Mode               sql.NullString `db:"mode"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}

type FrmEpShipmentHistory struct {
	ID                 int64          `db:"id"`
	EasypostShipmentID sql.NullString `db:"easypost_shipment_id"`
	UserID             sql.NullInt64  `db:"user_id"`
	SiteID             int64          `db:"site_id"`
	ChangeType         sql.NullString `db:"change_type"`
	Description        sql.NullString `db:"description"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}

type FrmEpShipmentAddress struct {
	ID                 int64          `db:"id"`
	EasypostID         string         `db:"easypost_id"`
	EasypostShipmentID string         `db:"easypost_shipment_id"`
	EntryID            sql.NullInt64  `db:"entry_id"`
	SiteID             int64          `db:"site_id"`
	AddressType        string         `db:"address_type"`
	Name               sql.NullString `db:"name"`
	Company            sql.NullString `db:"company"`
	Street1            sql.NullString `db:"street1"`
	Street2            sql.NullString `db:"street2"`
	City               sql.NullString `db:"city"`
	State              sql.NullString `db:"state"`
	Zip                sql.NullString `db:"zip"`
	Country            sql.NullString `db:"country"`
	Phone              sql.NullString `db:"phone"`
	Email              sql.NullString `db:"email"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}

type FrmEpShipmentLabel struct {
	ID                 int64          `db:"id"`
	EasypostID         string         `db:"easypost_id"`
	EasypostShipmentID string         `db:"easypost_shipment_id"`
	EntryID            sql.NullInt64  `db:"entry_id"`
	SiteID             int64          `db:"site_id"`
	DateAdvance        int            `db:"date_advance"`
	IntegratedForm     sql.NullString `db:"integrated_form"`
	LabelDate          sql.NullTime   `db:"label_date"`
	LabelResolution    sql.NullInt64  `db:"label_resolution"`
	LabelSize          sql.NullString `db:"label_size"`
	LabelType          sql.NullString `db:"label_type"`
	LabelFileType      sql.NullString `db:"label_file_type"`
	LabelURL           sql.NullString `db:"label_url"`
	LabelPDFURL        sql.NullString `db:"label_pdf_url"`
	LabelZPLURL        sql.NullString `db:"label_zpl_url"`
	LabelEPL2URL       sql.NullString `db:"label_epl2_url"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}

type FrmEpShipmentParcel struct {
	ID                 int64               `db:"id"`
	EasypostID         string              `db:"easypost_id"`
	EasypostShipmentID string              `db:"easypost_shipment_id"`
	EntryID            sql.NullInt64       `db:"entry_id"`
	SiteID             int64               `db:"site_id"`
	Length             decimal.NullDecimal `db:"length"`
	Width              decimal.NullDecimal `db:"width"`
	Height             decimal.NullDecimal `db:"height"`
	Weight             decimal.NullDecimal `db:"weight"`
	CreatedAt          time.Time           `db:"created_at"`
	UpdatedAt          time.Time           `db:"updated_at"`
}

type FrmEpShipmentRate struct {
	ID                     int64               `db:"id"`
	EasypostID             string              `db:"easypost_id"`
	EasypostShipmentID     string              `db:"easypost_shipment_id"`
	EntryID                sql.NullInt64       `db:"entry_id"`
	SiteID                 int64               `db:"site_id"`
	Mode                   string              `db:"mode"`
	Service                sql.NullString      `db:"service"`
	Carrier                sql.NullString      `db:"carrier"`
	Rate                   decimal.NullDecimal `db:"rate"`
	Currency               sql.NullString      `db:"currency"`
	RetailRate             decimal.NullDecimal `db:"retail_rate"`
	RetailCurrency         sql.NullString      `db:"retail_currency"`
	ListRate               decimal.NullDecimal `db:"list_rate"`
	ListCurrency           sql.NullString      `db:"list_currency"`
	BillingType            sql.NullString      `db:"billing_type"`
	DeliveryDays           sql.NullInt64       `db:"delivery_days"`
	DeliveryDate           sql.NullTime       `db:"delivery_date"`
	DeliveryDateGuaranteed sql.NullBool        `db:"delivery_date_guaranteed"`
	EstDeliveryDays        sql.NullInt64       `db:"est_delivery_days"`
	CreatedAt              time.Time           `db:"created_at"`
	UpdatedAt              time.Time           `db:"updated_at"`
}
