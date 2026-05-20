package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/jmoiron/sqlx"
)

type FrmEntryHistoryUseCase struct {
	db         *sqlx.DB
	history    *repository.FrmEntryHistoryRepository
	fields     *repository.FrmFieldRepository
	fieldUC    *FrmFieldUseCase
	updateTypes *repository.FrmEntryUpdateTypeRepository
	cache      *CacheUseCase
}

func NewFrmEntryHistoryUseCase(
	db *sqlx.DB,
	history *repository.FrmEntryHistoryRepository,
	fields *repository.FrmFieldRepository,
	fieldUC *FrmFieldUseCase,
	updateTypes *repository.FrmEntryUpdateTypeRepository,
	c *CacheUseCase,
) *FrmEntryHistoryUseCase {
	return &FrmEntryHistoryUseCase{db: db, history: history, fields: fields, fieldUC: fieldUC, updateTypes: updateTypes, cache: c}
}

// Update inserts history rows for siteID. Resolves field_id → frm_fields.id and update_type_code → frm_entry_update_types.id.
func (u *FrmEntryHistoryUseCase) Update(ctx context.Context, siteID int64, in []model.EntryHistoryInput) error {
	if len(in) == 0 {
		return nil
	}

	// Build a fields map (FrmField.FieldID → entity.FrmField.ID) so payload field_id can be translated to the pk.
	allFields, err := u.fields.FindBySite(ctx, nil, siteID)
	if err != nil {
		return err
	}
	fieldByExternal := make(map[int64]int64, len(allFields))
	for _, f := range allFields {
		if f.FieldID.Valid {
			fieldByExternal[f.FieldID.Int64] = f.ID
		}
	}

	// Resolve update type codes to ids (cached).
	typesByCode, err := u.UpdateTypesByCode(ctx)
	if err != nil {
		return err
	}

	rows := make([]repository.FrmEntryHistoryInput, 0, len(in))
	affectedEntries := make(map[int64]struct{}, len(in))

	for _, r := range in {
		pkFieldID, ok := fieldByExternal[r.FieldID]
		if !ok {
			// Auto-create a placeholder field row so history can reference a real FK.
			f := entity.FrmField{SiteID: siteID, FieldID: sql.NullInt64{Int64: r.FieldID, Valid: true}}
			if err := repository.WithTx(ctx, u.db, func(tx *sqlx.Tx) error {
				_, err := tx.ExecContext(ctx, `INSERT INTO frm_fields (field_id, site_id) VALUES ($1, $2)`, r.FieldID, siteID)
				return err
			}); err != nil {
				return err
			}
			// re-fetch
			allFields, err = u.fields.FindBySite(ctx, nil, siteID)
			if err != nil {
				return err
			}
			for _, f2 := range allFields {
				if f2.FieldID.Valid && f2.FieldID.Int64 == r.FieldID {
					fieldByExternal[r.FieldID] = f2.ID
					pkFieldID = f2.ID
					break
				}
			}
			_ = f
		}

		var typeID sql.NullInt64
		if r.UpdateTypeCode != "" {
			if t, ok := typesByCode[r.UpdateTypeCode]; ok {
				typeID = sql.NullInt64{Int64: t, Valid: true}
			}
		}

		var changeDate sql.NullTime
		if r.ChangeDate != nil {
			if t, err := time.Parse(time.RFC3339, *r.ChangeDate); err == nil {
				changeDate = sql.NullTime{Time: t, Valid: true}
			}
		}

		var userID sql.NullInt64
		if r.UserID != nil {
			userID = sql.NullInt64{Int64: *r.UserID, Valid: true}
		}

		var oldV, newV sql.NullString
		if r.OldValue != nil {
			oldV = sql.NullString{String: *r.OldValue, Valid: true}
		}
		if r.NewValue != nil {
			newV = sql.NullString{String: *r.NewValue, Valid: true}
		}

		rows = append(rows, repository.FrmEntryHistoryInput{
			EntryID:      sql.NullInt64{Int64: r.EntryID, Valid: true},
			FieldID:      pkFieldID,
			UserID:       userID,
			UpdateTypeID: typeID,
			OldValue:     oldV,
			NewValue:     newV,
			ChangeDate:   changeDate,
		})
		affectedEntries[r.EntryID] = struct{}{}
	}

	if err := repository.WithTx(ctx, u.db, func(tx *sqlx.Tx) error {
		return u.history.InsertMany(ctx, tx, siteID, rows)
	}); err != nil {
		return err
	}

	if u.cache != nil {
		for entryID := range affectedEntries {
			_ = u.cache.ForgetEntryMeta(ctx, siteID, entryID)
		}
	}
	return nil
}

func (u *FrmEntryHistoryUseCase) UpdateTypesByCode(ctx context.Context) (map[string]int64, error) {
	out := map[string]int64{}
	build := func() (any, error) {
		all, err := u.updateTypes.FindAll(ctx, nil)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{}
		for _, t := range all {
			if t.Code.Valid {
				m[t.Code.String] = t.ID
			}
		}
		return m, nil
	}
	if u.cache == nil {
		v, err := build()
		if err != nil {
			return out, err
		}
		return v.(map[string]int64), nil
	}
	if err := u.cache.Remember(ctx, u.cache.Keys.UpdateTypes(), 0, &out, build); err != nil {
		return out, err
	}
	return out, nil
}

// GetByEntry returns the formatted history for one entry, including resolved field + update-type metadata.
func (u *FrmEntryHistoryUseCase) GetByEntry(ctx context.Context, siteID, entryID int64) (model.EntryUpdatesResponse, error) {
	items, err := u.history.FindByEntry(ctx, nil, siteID, entryID)
	if err != nil {
		return model.EntryUpdatesResponse{}, err
	}

	fmap, _ := u.fieldUC.FieldsMap(ctx, siteID)
	allTypes, _ := u.updateTypes.FindAll(ctx, nil)
	typesByID := make(map[int64]*entity.FrmEntryUpdateType, len(allTypes))
	for i := range allTypes {
		typesByID[allTypes[i].ID] = &allTypes[i]
	}

	out := make([]model.EntryUpdateItem, 0, len(items))
	for i := range items {
		h := &items[i]
		var key, label *string
		// fields map is keyed by external FieldID, but history.FieldID is the pk.
		// Look up via fmap iteration since the cardinality is small.
		for _, f := range fmap {
			if f.ID == h.FieldID {
				key = f.Key
				label = f.Label
				break
			}
		}
		var updateType *string
		if h.UpdateTypeID.Valid {
			if t, ok := typesByID[h.UpdateTypeID.Int64]; ok && t.Code.Valid {
				v := t.Code.String
				updateType = &v
			}
		}
		out = append(out, model.EntryUpdateItem{
			ID:         h.ID,
			FieldID:    h.FieldID,
			FieldKey:   key,
			FieldLabel: label,
			UpdateType: updateType,
			OldValue:   converter.NullString(h.OldValue),
			NewValue:   converter.NullString(h.NewValue),
			ChangeDate: converter.NullTime(h.ChangeDate),
			CreatedAt:  h.CreatedAt,
		})
	}

	return model.EntryUpdatesResponse{EntryID: entryID, Updates: out}, nil
}
