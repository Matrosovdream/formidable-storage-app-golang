package converter

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
)

// ---------- User / Auth ----------

func ToUserResponse(u *entity.User) model.UserResponse {
	return model.UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		EmailVerifiedAt: NullTime(u.EmailVerifiedAt),
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// ---------- Site ----------

func ToSiteResponse(s *entity.Site) model.SiteResponse {
	return model.SiteResponse{
		ID:        s.ID,
		Name:      s.Name,
		URL:       s.URL,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func ToSiteResponses(in []entity.Site) []model.SiteResponse {
	out := make([]model.SiteResponse, len(in))
	for i := range in {
		out[i] = ToSiteResponse(&in[i])
	}
	return out
}

// ---------- FrmField ----------

func ToFrmFieldResponse(f *entity.FrmField) model.FrmFieldResponse {
	return model.FrmFieldResponse{
		ID:        f.ID,
		FieldID:   NullInt64(f.FieldID),
		SiteID:    f.SiteID,
		Key:       NullString(f.Key),
		Type:      NullString(f.Type),
		Label:     NullString(f.Label),
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

// ---------- FrmEmailLog ----------

func ToEmailLogItem(l *entity.FrmEmailLog) model.EmailLogItem {
	return model.EmailLogItem{
		ID:            l.ID,
		EntryID:       NullInt64(l.EntryID),
		SiteID:        l.SiteID,
		FormID:        NullInt64(l.FormID),
		Subject:       NullString(l.Subject),
		MessageID:     NullString(l.MessageID),
		EmailFrom:     NullString(l.EmailFrom),
		EmailTo:       NullString(l.EmailTo),
		ContentPlain:  NullString(l.ContentPlain),
		ContentHTML:   NullString(l.ContentHTML),
		Status:        l.Status,
		DateSent:      NullTime(l.DateSent),
		Mailer:        NullString(l.Mailer),
		Attachments:   l.Attachments,
		InitiatorName: NullString(l.InitiatorName),
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
}

func ToEmailLogItems(in []entity.FrmEmailLog) []model.EmailLogItem {
	out := make([]model.EmailLogItem, len(in))
	for i := range in {
		out[i] = ToEmailLogItem(&in[i])
	}
	return out
}

// ---------- FrmEpShipmentHistory ----------

func ToEpShipmentHistoryItem(h *entity.FrmEpShipmentHistory) model.EpShipmentHistoryItem {
	return model.EpShipmentHistoryItem{
		ID:                 h.ID,
		EasypostShipmentID: NullString(h.EasypostShipmentID),
		UserID:             NullInt64(h.UserID),
		SiteID:             h.SiteID,
		ChangeType:         NullString(h.ChangeType),
		Description:        NullString(h.Description),
		CreatedAt:          h.CreatedAt,
		UpdatedAt:          h.UpdatedAt,
	}
}

func ToEpShipmentHistoryItems(in []entity.FrmEpShipmentHistory) []model.EpShipmentHistoryItem {
	out := make([]model.EpShipmentHistoryItem, len(in))
	for i := range in {
		out[i] = ToEpShipmentHistoryItem(&in[i])
	}
	return out
}
