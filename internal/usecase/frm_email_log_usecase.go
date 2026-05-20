package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
)

type FrmEmailLogUseCase struct {
	emails *repository.FrmEmailLogRepository
}

func NewFrmEmailLogUseCase(emails *repository.FrmEmailLogRepository) *FrmEmailLogUseCase {
	return &FrmEmailLogUseCase{emails: emails}
}

// UpdateAll bulk-upserts the email log rows. Behaviour mirrors Laravel's `updateDataMultiple`.
func (u *FrmEmailLogUseCase) UpdateAll(ctx context.Context, siteID int64, in []model.FrmEmailLogInput) (int, error) {
	if len(in) == 0 {
		return 0, nil
	}
	rows := make([]repository.FrmEmailLogInput, len(in))
	for i, r := range in {
		rows[i] = repository.FrmEmailLogInput{
			EntryID:       nullInt64Ptr(r.EntryID),
			FormID:        nullInt64Ptr(r.FormID),
			Subject:       nullStringPtr(r.Subject),
			MessageID:     nullStringPtr(r.MessageID),
			EmailFrom:     nullStringPtr(r.EmailFrom),
			EmailTo:       nullStringPtr(r.EmailTo),
			People:        nullStringPtr(r.People),
			Headers:       nullStringPtr(r.Headers),
			ErrorText:     nullStringPtr(r.ErrorText),
			ContentPlain:  nullStringPtr(r.ContentPlain),
			ContentHTML:   nullStringPtr(r.ContentHTML),
			Status:        r.Status,
			DateSent:      nullTimeFromString(r.DateSent),
			Mailer:        nullStringPtr(r.Mailer),
			Attachments:   r.Attachments,
			InitiatorName: nullStringPtr(r.InitiatorName),
			InitiatorFile: nullStringPtr(r.InitiatorFile),
			OriginalLogID: nullInt64Ptr(r.OriginalLogID),
		}
	}
	if err := u.emails.UpsertMany(ctx, nil, siteID, rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}

// UpdateAllRaw is identical to UpdateAll in this Go port — both paths hit the same SQL.
// In Laravel the difference was queue vs sync; the REST controller picks which one to call.
func (u *FrmEmailLogUseCase) UpdateAllRaw(ctx context.Context, siteID int64, in []model.FrmEmailLogInput) (int, error) {
	return u.UpdateAll(ctx, siteID, in)
}

// List returns a paginated list of email-log rows scoped to siteID.
func (u *FrmEmailLogUseCase) List(ctx context.Context, siteID int64, req model.ListEmailsLogRequest) (repository.ListResult[model.EmailLogItem], error) {
	page := req.PageNum
	if page < 1 {
		page = 1
	}
	perPage := req.Paginate
	if perPage <= 0 {
		perPage = 25
	}
	res, err := u.emails.List(ctx, nil, siteID, repository.ListParams{
		Filters: req.Filters,
		SortBy:  req.SortBy,
		SortDir: req.SortDir,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return repository.ListResult[model.EmailLogItem]{}, err
	}
	return repository.ListResult[model.EmailLogItem]{
		Data:     converter.ToEmailLogItems(res.Data),
		Total:    res.Total,
		Page:     res.Page,
		PerPage:  res.PerPage,
		LastPage: res.LastPage,
	}, nil
}

func (u *FrmEmailLogUseCase) FindByEntry(ctx context.Context, siteID, entryID int64) ([]model.EmailLogItem, error) {
	items, err := u.emails.FindByEntry(ctx, nil, siteID, entryID)
	if err != nil {
		return nil, err
	}
	return converter.ToEmailLogItems(items), nil
}

func nullStringPtr(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func nullInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullTimeFromString(p *string) sql.NullTime {
	if p == nil || *p == "" {
		return sql.NullTime{}
	}
	if t, err := time.Parse(time.RFC3339, *p); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	if t, err := time.Parse("2006-01-02 15:04:05", *p); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	return sql.NullTime{}
}
