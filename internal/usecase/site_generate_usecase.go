package usecase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	mrand "math/rand/v2"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
)

// SiteGenerateUseCase produces dummy email/field/entry-update rows for a site,
// inserts them into the DB, and reports the wall-clock cost of each phase.
//
// It is intended for load testing and demo seeding from the Sanctum-protected
// /api/sites/generate/* endpoints — not for production traffic.
type SiteGenerateUseCase struct {
	sites       *repository.SiteRepository
	emails      *repository.FrmEmailLogRepository
	fields      *repository.FrmFieldRepository
	history     *repository.FrmEntryHistoryRepository
	updateTypes *repository.FrmEntryUpdateTypeRepository
}

func NewSiteGenerateUseCase(
	sites *repository.SiteRepository,
	emails *repository.FrmEmailLogRepository,
	fields *repository.FrmFieldRepository,
	history *repository.FrmEntryHistoryRepository,
	updateTypes *repository.FrmEntryUpdateTypeRepository,
) *SiteGenerateUseCase {
	return &SiteGenerateUseCase{
		sites:       sites,
		emails:      emails,
		fields:      fields,
		history:     history,
		updateTypes: updateTypes,
	}
}

const (
	genDefaultAmount     = 10
	genMaxAmount         = 10000
	genDefaultEmailLen   = 200
	genMaxEmailLen       = 100000
)

func clampAmount(n int) int {
	if n <= 0 {
		return genDefaultAmount
	}
	if n > genMaxAmount {
		return genMaxAmount
	}
	return n
}

func clampEmailLength(n int) int {
	if n <= 0 {
		return genDefaultEmailLen
	}
	if n > genMaxEmailLen {
		return genMaxEmailLen
	}
	return n
}

func msFloat(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}

func newRNG() *mrand.Rand {
	var seed [16]byte
	_, _ = rand.Read(seed[:])
	s1 := uint64(seed[0])<<56 | uint64(seed[1])<<48 | uint64(seed[2])<<40 | uint64(seed[3])<<32 |
		uint64(seed[4])<<24 | uint64(seed[5])<<16 | uint64(seed[6])<<8 | uint64(seed[7])
	s2 := uint64(seed[8])<<56 | uint64(seed[9])<<48 | uint64(seed[10])<<40 | uint64(seed[11])<<32 |
		uint64(seed[12])<<24 | uint64(seed[13])<<16 | uint64(seed[14])<<8 | uint64(seed[15])
	return mrand.New(mrand.NewPCG(s1, s2))
}

func randHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func (u *SiteGenerateUseCase) ensureSite(ctx context.Context, siteID int64) error {
	s, err := u.sites.FindByID(ctx, nil, siteID)
	if err != nil {
		return err
	}
	if s == nil {
		return ErrNotFound
	}
	return nil
}

// -------------------------------------------------------------------------
// Emails
// -------------------------------------------------------------------------

var (
	emailSubjects = []string{
		"Welcome aboard", "Order confirmation", "Password reset request",
		"Weekly digest", "Account update", "New message", "Invoice attached",
		"Payment received", "Action required", "Your shipment is on the way",
	}
	emailMailers = []string{"smtp", "sendgrid", "mailgun", "ses", "postmark"}
	emailDomains = []string{"example.com", "demo.local", "test.io", "mailtest.org", "inbox.dev"}
	loremWords   = []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing",
		"elit", "sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore",
		"et", "dolore", "magna", "aliqua", "enim", "ad", "minim", "veniam",
		"quis", "nostrud", "exercitation", "ullamco", "laboris", "nisi",
	}
)

func randWords(r *mrand.Rand, targetLen int) string {
	if targetLen <= 0 {
		return ""
	}
	b := make([]byte, 0, targetLen+16)
	for len(b) < targetLen {
		if len(b) > 0 {
			b = append(b, ' ')
		}
		b = append(b, loremWords[r.IntN(len(loremWords))]...)
	}
	if len(b) > targetLen {
		b = b[:targetLen]
	}
	return string(b)
}

func randHTML(r *mrand.Rand, targetLen int) string {
	if targetLen <= 0 {
		return ""
	}
	const open = "<html><body>"
	const close = "</body></html>"
	inner := targetLen - len(open) - len(close)
	if inner < 0 {
		inner = targetLen
	}
	var body string
	remaining := inner
	for remaining > 0 {
		chunk := 60 + r.IntN(120)
		if chunk > remaining {
			chunk = remaining
		}
		body += "<p>" + randWords(r, chunk-7) + "</p>"
		remaining -= chunk
	}
	out := open + body + close
	if len(out) > targetLen {
		out = out[:targetLen]
	}
	return out
}

// GenerateEmails creates `amount` dummy email-log rows with content of roughly
// `length` characters (plain + HTML) and bulk-inserts them via the repository.
func (u *SiteGenerateUseCase) GenerateEmails(ctx context.Context, siteID int64, amount, length int) (model.GenerateResponse, error) {
	if err := u.ensureSite(ctx, siteID); err != nil {
		return model.GenerateResponse{}, err
	}
	amount = clampAmount(amount)
	length = clampEmailLength(length)

	genStart := time.Now()
	r := newRNG()
	nowNano := time.Now().UnixNano()
	rows := make([]repository.FrmEmailLogInput, 0, amount)
	for i := 0; i < amount; i++ {
		plain := randWords(r, length)
		html := randHTML(r, length)
		subject := fmt.Sprintf("%s #%d", emailSubjects[r.IntN(len(emailSubjects))], r.IntN(1_000_000))
		from := fmt.Sprintf("from_%s@%s", randHex(4), emailDomains[r.IntN(len(emailDomains))])
		to := fmt.Sprintf("to_%s@%s", randHex(4), emailDomains[r.IntN(len(emailDomains))])
		mailer := emailMailers[r.IntN(len(emailMailers))]
		msgID := fmt.Sprintf("gen-%d-%d-%s", nowNano, i, randHex(6))
		sent := time.Now().UTC().Add(-time.Duration(r.IntN(72)) * time.Hour)
		rows = append(rows, repository.FrmEmailLogInput{
			EntryID:       sql.NullInt64{Int64: int64(9000 + r.IntN(1000)), Valid: true},
			FormID:        sql.NullInt64{Int64: int64(1 + r.IntN(20)), Valid: true},
			Subject:       sql.NullString{String: subject, Valid: true},
			MessageID:     sql.NullString{String: msgID, Valid: true},
			EmailFrom:     sql.NullString{String: from, Valid: true},
			EmailTo:       sql.NullString{String: to, Valid: true},
			ContentPlain:  sql.NullString{String: plain, Valid: true},
			ContentHTML:   sql.NullString{String: html, Valid: true},
			Status:        int16(r.IntN(3)),
			DateSent:      sql.NullTime{Time: sent, Valid: true},
			Mailer:        sql.NullString{String: mailer, Valid: true},
			Attachments:   int16(r.IntN(3)),
			InitiatorName: sql.NullString{String: "generator", Valid: true},
		})
	}
	genDur := time.Since(genStart)

	insStart := time.Now()
	if err := u.emails.UpsertMany(ctx, nil, siteID, rows); err != nil {
		return model.GenerateResponse{}, err
	}
	insDur := time.Since(insStart)

	return model.GenerateResponse{
		Success: true,
		Kind:    "emails",
		SiteID:  siteID,
		Count:   len(rows),
		Length:  length,
		Timings: model.GenerateTimings{
			GenerationMs: msFloat(genDur),
			InsertionMs:  msFloat(insDur),
			TotalMs:      msFloat(genDur + insDur),
		},
	}, nil
}

// -------------------------------------------------------------------------
// Fields
// -------------------------------------------------------------------------

var fieldTypes = []string{"text", "email", "tel", "textarea", "number", "select", "checkbox", "url", "date"}

// GenerateFields creates `amount` dummy field rows with timestamp-derived field_ids
// (so they don't collide with seeded fields) and upserts them via the repository.
func (u *SiteGenerateUseCase) GenerateFields(ctx context.Context, siteID int64, amount int) (model.GenerateResponse, error) {
	if err := u.ensureSite(ctx, siteID); err != nil {
		return model.GenerateResponse{}, err
	}
	amount = clampAmount(amount)

	genStart := time.Now()
	r := newRNG()
	base := time.Now().UnixNano() % 1_000_000_000
	rows := make([]repository.FrmFieldInput, 0, amount)
	for i := 0; i < amount; i++ {
		fid := base + int64(i)
		typ := fieldTypes[r.IntN(len(fieldTypes))]
		key := fmt.Sprintf("gen_field_%d_%s", fid, randHex(3))
		label := fmt.Sprintf("Generated %s field #%d", typ, fid)
		rows = append(rows, repository.FrmFieldInput{
			FieldID: fid,
			Key:     key,
			Type:    typ,
			Label:   label,
		})
	}
	genDur := time.Since(genStart)

	insStart := time.Now()
	if err := u.fields.UpsertMany(ctx, nil, siteID, rows); err != nil {
		return model.GenerateResponse{}, err
	}
	insDur := time.Since(insStart)

	return model.GenerateResponse{
		Success: true,
		Kind:    "fields",
		SiteID:  siteID,
		Count:   len(rows),
		Timings: model.GenerateTimings{
			GenerationMs: msFloat(genDur),
			InsertionMs:  msFloat(insDur),
			TotalMs:      msFloat(genDur + insDur),
		},
	}, nil
}

// -------------------------------------------------------------------------
// Entry updates (frm_entry_history)
// -------------------------------------------------------------------------

// GenerateEntryUpdates creates `amount` dummy entry-history rows attached to
// random existing fields of the site. If the site has no fields yet, one
// placeholder field is created first.
//
// Note: the prep reads (existing fields + update types) are counted under
// generation time, since they happen before the bulk INSERT.
func (u *SiteGenerateUseCase) GenerateEntryUpdates(ctx context.Context, siteID int64, amount int) (model.GenerateResponse, error) {
	if err := u.ensureSite(ctx, siteID); err != nil {
		return model.GenerateResponse{}, err
	}
	amount = clampAmount(amount)

	genStart := time.Now()

	siteFields, err := u.fields.FindBySite(ctx, nil, siteID)
	if err != nil {
		return model.GenerateResponse{}, err
	}
	if len(siteFields) == 0 {
		fid := time.Now().UnixNano() % 1_000_000_000
		seed := []repository.FrmFieldInput{{
			FieldID: fid,
			Key:     fmt.Sprintf("gen_placeholder_%s", randHex(3)),
			Type:    "text",
			Label:   "Generated placeholder field",
		}}
		if err := u.fields.UpsertMany(ctx, nil, siteID, seed); err != nil {
			return model.GenerateResponse{}, err
		}
		siteFields, err = u.fields.FindBySite(ctx, nil, siteID)
		if err != nil {
			return model.GenerateResponse{}, err
		}
	}

	types, err := u.updateTypes.FindAll(ctx, nil)
	if err != nil {
		return model.GenerateResponse{}, err
	}

	r := newRNG()
	rows := make([]repository.FrmEntryHistoryInput, 0, amount)
	for i := 0; i < amount; i++ {
		f := siteFields[r.IntN(len(siteFields))]
		oldV := sql.NullString{String: fmt.Sprintf("old-%s", randHex(4)), Valid: true}
		newV := sql.NullString{String: fmt.Sprintf("new-%s", randHex(4)), Valid: true}
		changeAt := time.Now().UTC().Add(-time.Duration(r.IntN(168)) * time.Hour)
		var typeID sql.NullInt64
		if len(types) > 0 {
			typeID = sql.NullInt64{Int64: types[r.IntN(len(types))].ID, Valid: true}
		}
		rows = append(rows, repository.FrmEntryHistoryInput{
			EntryID:      sql.NullInt64{Int64: int64(9000 + r.IntN(1000)), Valid: true},
			FieldID:      f.ID,
			UpdateTypeID: typeID,
			OldValue:     oldV,
			NewValue:     newV,
			ChangeDate:   sql.NullTime{Time: changeAt, Valid: true},
		})
	}
	genDur := time.Since(genStart)

	insStart := time.Now()
	if err := u.history.InsertMany(ctx, nil, siteID, rows); err != nil {
		return model.GenerateResponse{}, err
	}
	insDur := time.Since(insStart)

	return model.GenerateResponse{
		Success: true,
		Kind:    "entry_updates",
		SiteID:  siteID,
		Count:   len(rows),
		Timings: model.GenerateTimings{
			GenerationMs: msFloat(genDur),
			InsertionMs:  msFloat(insDur),
			TotalMs:      msFloat(genDur + insDur),
		},
	}, nil
}
