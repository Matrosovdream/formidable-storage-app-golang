package usecase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	mrand "math/rand/v2"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"golang.org/x/sync/errgroup"
)

// SiteGenerateUseCase produces dummy email/field/entry-update rows for a site,
// inserts them into the DB, and reports the wall-clock cost of each phase.
//
// Both phases (in-memory generation and DB insertion) can run concurrently
// across goroutines when the caller passes concurrent=true. The toggle is
// intentionally exposed so calls can be benchmarked side-by-side.
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
	genDefaultAmount   = 10
	genMaxAmount       = 10000
	genDefaultEmailLen = 200
	genMaxEmailLen     = 100000
	genMaxWorkers      = 8
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

// randHexR is the hot-path variant of randHex that reuses a worker's PCG
// instead of paying the per-call crypto/rand cost. Output is fake-data quality
// — do not use for security-sensitive randomness.
func randHexR(r *mrand.Rand, n int) string {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(r.UintN(256))
	}
	return hex.EncodeToString(buf)
}

// numWorkers caps the worker pool at GOMAXPROCS (and at most genMaxWorkers),
// but never exceeds the number of items to process.
func numWorkers(amount int) int {
	w := runtime.GOMAXPROCS(0)
	if w > genMaxWorkers {
		w = genMaxWorkers
	}
	if w > amount {
		w = amount
	}
	if w < 1 {
		w = 1
	}
	return w
}

// partition splits the half-open range [0, total) into n contiguous sub-ranges.
func partition(total, n int) [][2]int {
	if n < 1 {
		n = 1
	}
	parts := make([][2]int, 0, n)
	base := total / n
	extra := total % n
	start := 0
	for i := 0; i < n; i++ {
		size := base
		if i < extra {
			size++
		}
		parts = append(parts, [2]int{start, start + size})
		start += size
	}
	return parts
}

// runFill dispatches `fill` across worker goroutines if concurrent, otherwise
// runs it once on the full range. Each invocation gets its own PCG.
func runFill(amount int, concurrent bool, fill func(r *mrand.Rand, start, end int)) int {
	if !concurrent {
		fill(newRNG(), 0, amount)
		return 1
	}
	n := numWorkers(amount)
	if n < 2 {
		fill(newRNG(), 0, amount)
		return 1
	}
	var wg sync.WaitGroup
	for _, p := range partition(amount, n) {
		if p[0] == p[1] {
			continue
		}
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			fill(newRNG(), p[0], p[1])
		}()
	}
	wg.Wait()
	return n
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
	const closeTag = "</body></html>"
	inner := targetLen - len(open) - len(closeTag)
	if inner < 0 {
		inner = targetLen
	}
	var b strings.Builder
	b.Grow(targetLen)
	b.WriteString(open)
	remaining := inner
	for remaining > 0 {
		chunk := 60 + r.IntN(120)
		if chunk > remaining {
			chunk = remaining
		}
		b.WriteString("<p>")
		b.WriteString(randWords(r, chunk-7))
		b.WriteString("</p>")
		remaining -= chunk
	}
	b.WriteString(closeTag)
	out := b.String()
	if len(out) > targetLen {
		out = out[:targetLen]
	}
	return out
}

// GenerateEmails creates `amount` dummy email-log rows with content of roughly
// `length` characters (plain + HTML) and bulk-inserts them via the repository.
// When `concurrent` is true, both phases fan out across a worker pool.
func (u *SiteGenerateUseCase) GenerateEmails(ctx context.Context, siteID int64, amount, length int, concurrent bool) (model.GenerateResponse, error) {
	if err := u.ensureSite(ctx, siteID); err != nil {
		return model.GenerateResponse{}, err
	}
	amount = clampAmount(amount)
	length = clampEmailLength(length)

	genStart := time.Now()
	rows := make([]repository.FrmEmailLogInput, amount)
	nowNano := time.Now().UnixNano()

	fill := func(r *mrand.Rand, start, end int) {
		for i := start; i < end; i++ {
			plain := randWords(r, length)
			html := randHTML(r, length)
			subject := fmt.Sprintf("%s #%d", emailSubjects[r.IntN(len(emailSubjects))], r.IntN(1_000_000))
			from := fmt.Sprintf("from_%s@%s", randHexR(r, 4), emailDomains[r.IntN(len(emailDomains))])
			to := fmt.Sprintf("to_%s@%s", randHexR(r, 4), emailDomains[r.IntN(len(emailDomains))])
			mailer := emailMailers[r.IntN(len(emailMailers))]
			msgID := fmt.Sprintf("gen-%d-%d-%s", nowNano, i, randHexR(r, 6))
			sent := time.Now().UTC().Add(-time.Duration(r.IntN(72)) * time.Hour)
			rows[i] = repository.FrmEmailLogInput{
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
			}
		}
	}
	workers := runFill(amount, concurrent, fill)
	genDur := time.Since(genStart)

	insStart := time.Now()
	if err := u.insertEmails(ctx, siteID, rows, concurrent); err != nil {
		return model.GenerateResponse{}, err
	}
	insDur := time.Since(insStart)

	return model.GenerateResponse{
		Success:    true,
		Kind:       "emails",
		SiteID:     siteID,
		Count:      len(rows),
		Length:     length,
		Concurrent: concurrent,
		Workers:    workers,
		Timings: model.GenerateTimings{
			GenerationMs: msFloat(genDur),
			InsertionMs:  msFloat(insDur),
			TotalMs:      msFloat(genDur + insDur),
		},
	}, nil
}

func (u *SiteGenerateUseCase) insertEmails(ctx context.Context, siteID int64, rows []repository.FrmEmailLogInput, concurrent bool) error {
	if len(rows) == 0 {
		return nil
	}
	if !concurrent {
		return u.emails.UpsertMany(ctx, nil, siteID, rows)
	}
	n := numWorkers(len(rows))
	if n < 2 {
		return u.emails.UpsertMany(ctx, nil, siteID, rows)
	}
	g, gctx := errgroup.WithContext(ctx)
	for _, p := range partition(len(rows), n) {
		if p[0] == p[1] {
			continue
		}
		p := p
		g.Go(func() error {
			return u.emails.UpsertMany(gctx, nil, siteID, rows[p[0]:p[1]])
		})
	}
	return g.Wait()
}

// -------------------------------------------------------------------------
// Fields
// -------------------------------------------------------------------------

var fieldTypes = []string{"text", "email", "tel", "textarea", "number", "select", "checkbox", "url", "date"}

// GenerateFields creates `amount` dummy field rows with timestamp-derived field_ids
// (so they don't collide with seeded fields) and upserts them via the repository.
// When `concurrent` is true, both phases fan out across a worker pool.
func (u *SiteGenerateUseCase) GenerateFields(ctx context.Context, siteID int64, amount int, concurrent bool) (model.GenerateResponse, error) {
	if err := u.ensureSite(ctx, siteID); err != nil {
		return model.GenerateResponse{}, err
	}
	amount = clampAmount(amount)

	genStart := time.Now()
	base := time.Now().UnixNano() % 1_000_000_000
	rows := make([]repository.FrmFieldInput, amount)

	fill := func(r *mrand.Rand, start, end int) {
		for i := start; i < end; i++ {
			fid := base + int64(i)
			typ := fieldTypes[r.IntN(len(fieldTypes))]
			key := fmt.Sprintf("gen_field_%d_%s", fid, randHexR(r, 3))
			label := fmt.Sprintf("Generated %s field #%d", typ, fid)
			rows[i] = repository.FrmFieldInput{
				FieldID: fid,
				Key:     key,
				Type:    typ,
				Label:   label,
			}
		}
	}
	workers := runFill(amount, concurrent, fill)
	genDur := time.Since(genStart)

	insStart := time.Now()
	if err := u.insertFields(ctx, siteID, rows, concurrent); err != nil {
		return model.GenerateResponse{}, err
	}
	insDur := time.Since(insStart)

	return model.GenerateResponse{
		Success:    true,
		Kind:       "fields",
		SiteID:     siteID,
		Count:      len(rows),
		Concurrent: concurrent,
		Workers:    workers,
		Timings: model.GenerateTimings{
			GenerationMs: msFloat(genDur),
			InsertionMs:  msFloat(insDur),
			TotalMs:      msFloat(genDur + insDur),
		},
	}, nil
}

func (u *SiteGenerateUseCase) insertFields(ctx context.Context, siteID int64, rows []repository.FrmFieldInput, concurrent bool) error {
	if len(rows) == 0 {
		return nil
	}
	if !concurrent {
		return u.fields.UpsertMany(ctx, nil, siteID, rows)
	}
	n := numWorkers(len(rows))
	if n < 2 {
		return u.fields.UpsertMany(ctx, nil, siteID, rows)
	}
	g, gctx := errgroup.WithContext(ctx)
	for _, p := range partition(len(rows), n) {
		if p[0] == p[1] {
			continue
		}
		p := p
		g.Go(func() error {
			return u.fields.UpsertMany(gctx, nil, siteID, rows[p[0]:p[1]])
		})
	}
	return g.Wait()
}

// -------------------------------------------------------------------------
// Entry updates (frm_entry_history)
// -------------------------------------------------------------------------

// GenerateEntryUpdates creates `amount` dummy entry-history rows attached to
// random existing fields of the site. If the site has no fields yet, one
// placeholder field is created first.
//
// Note: the prep reads (existing fields + update types) are counted under
// generation time, since they happen before the bulk INSERT. When concurrent
// is true, those two reads also run in parallel.
func (u *SiteGenerateUseCase) GenerateEntryUpdates(ctx context.Context, siteID int64, amount int, concurrent bool) (model.GenerateResponse, error) {
	if err := u.ensureSite(ctx, siteID); err != nil {
		return model.GenerateResponse{}, err
	}
	amount = clampAmount(amount)

	genStart := time.Now()

	siteFields, types, err := u.loadFieldsAndTypes(ctx, siteID, concurrent)
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

	rows := make([]repository.FrmEntryHistoryInput, amount)
	fill := func(r *mrand.Rand, start, end int) {
		for i := start; i < end; i++ {
			f := siteFields[r.IntN(len(siteFields))]
			oldV := sql.NullString{String: fmt.Sprintf("old-%s", randHexR(r, 4)), Valid: true}
			newV := sql.NullString{String: fmt.Sprintf("new-%s", randHexR(r, 4)), Valid: true}
			changeAt := time.Now().UTC().Add(-time.Duration(r.IntN(168)) * time.Hour)
			var typeID sql.NullInt64
			if len(types) > 0 {
				typeID = sql.NullInt64{Int64: types[r.IntN(len(types))].ID, Valid: true}
			}
			rows[i] = repository.FrmEntryHistoryInput{
				EntryID:      sql.NullInt64{Int64: int64(9000 + r.IntN(1000)), Valid: true},
				FieldID:      f.ID,
				UpdateTypeID: typeID,
				OldValue:     oldV,
				NewValue:     newV,
				ChangeDate:   sql.NullTime{Time: changeAt, Valid: true},
			}
		}
	}
	workers := runFill(amount, concurrent, fill)
	genDur := time.Since(genStart)

	insStart := time.Now()
	if err := u.insertHistory(ctx, siteID, rows, concurrent); err != nil {
		return model.GenerateResponse{}, err
	}
	insDur := time.Since(insStart)

	return model.GenerateResponse{
		Success:    true,
		Kind:       "entry_updates",
		SiteID:     siteID,
		Count:      len(rows),
		Concurrent: concurrent,
		Workers:    workers,
		Timings: model.GenerateTimings{
			GenerationMs: msFloat(genDur),
			InsertionMs:  msFloat(insDur),
			TotalMs:      msFloat(genDur + insDur),
		},
	}, nil
}

// loadFieldsAndTypes runs the two pre-read queries serially or in parallel
// depending on the toggle.
func (u *SiteGenerateUseCase) loadFieldsAndTypes(ctx context.Context, siteID int64, concurrent bool) (siteFields []entity.FrmField, types []entity.FrmEntryUpdateType, err error) {
	if !concurrent {
		sf, ferr := u.fields.FindBySite(ctx, nil, siteID)
		if ferr != nil {
			return nil, nil, ferr
		}
		ts, terr := u.updateTypes.FindAll(ctx, nil)
		if terr != nil {
			return nil, nil, terr
		}
		return sf, ts, nil
	}
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		sf, ferr := u.fields.FindBySite(gctx, nil, siteID)
		if ferr != nil {
			return ferr
		}
		siteFields = sf
		return nil
	})
	g.Go(func() error {
		ts, terr := u.updateTypes.FindAll(gctx, nil)
		if terr != nil {
			return terr
		}
		types = ts
		return nil
	})
	if werr := g.Wait(); werr != nil {
		return nil, nil, werr
	}
	return siteFields, types, nil
}

func (u *SiteGenerateUseCase) insertHistory(ctx context.Context, siteID int64, rows []repository.FrmEntryHistoryInput, concurrent bool) error {
	if len(rows) == 0 {
		return nil
	}
	if !concurrent {
		return u.history.InsertMany(ctx, nil, siteID, rows)
	}
	n := numWorkers(len(rows))
	if n < 2 {
		return u.history.InsertMany(ctx, nil, siteID, rows)
	}
	g, gctx := errgroup.WithContext(ctx)
	for _, p := range partition(len(rows), n) {
		if p[0] == p[1] {
			continue
		}
		p := p
		g.Go(func() error {
			return u.history.InsertMany(gctx, nil, siteID, rows[p[0]:p[1]])
		})
	}
	return g.Wait()
}
