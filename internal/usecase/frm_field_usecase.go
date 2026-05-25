package usecase

import (
	"context"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/cache"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
)

type FrmFieldUseCase struct {
	fields *repository.FrmFieldRepository
	cache  *CacheUseCase
}

func NewFrmFieldUseCase(fields *repository.FrmFieldRepository, c *CacheUseCase) *FrmFieldUseCase {
	return &FrmFieldUseCase{fields: fields, cache: c}
}

func (u *FrmFieldUseCase) UpdateAll(ctx context.Context, siteID int64, in []model.FrmFieldInput) error {
	inputs := make([]repository.FrmFieldInput, len(in))
	for i, f := range in {
		inputs[i] = repository.FrmFieldInput{FieldID: f.FieldID, Key: f.Key, Type: f.Type, Label: f.Label}
	}
	if err := u.fields.UpsertMany(ctx, nil, siteID, inputs); err != nil {
		return err
	}
	if u.cache != nil {
		_ = u.cache.Forget(ctx, u.cache.Keys.FieldsMap(siteID))
	}
	return nil
}

// ListBySite returns a paginated list of fields for siteID, applying the optional filters/sort.
func (u *FrmFieldUseCase) ListBySite(ctx context.Context, siteID int64, req model.SiteFieldsRequest) (repository.ListResult[model.FrmFieldResponse], error) {
	filters := map[string]any{}
	if req.FieldID != nil {
		filters["field_id"] = *req.FieldID
	}
	if req.Type != nil && *req.Type != "" {
		filters["type"] = *req.Type
	}
	if req.Key != nil && *req.Key != "" {
		filters["key"] = *req.Key
	}
	if req.Label != nil && *req.Label != "" {
		filters["label"] = *req.Label
	}

	res, err := u.fields.List(ctx, nil, siteID, repository.ListParams{
		Filters: filters,
		SortBy:  req.SortBy,
		SortDir: req.SortDir,
		Page:    req.Page,
		PerPage: req.PerPage,
	})
	if err != nil {
		return repository.ListResult[model.FrmFieldResponse]{}, err
	}

	data := make([]model.FrmFieldResponse, len(res.Data))
	for i := range res.Data {
		data[i] = convertField(&res.Data[i])
	}
	return repository.ListResult[model.FrmFieldResponse]{
		Data:     data,
		Total:    res.Total,
		Page:     res.Page,
		PerPage:  res.PerPage,
		LastPage: res.LastPage,
	}, nil
}

// FieldsMap returns a site's fields keyed by FieldID (cached).
func (u *FrmFieldUseCase) FieldsMap(ctx context.Context, siteID int64) (map[int64]model.FrmFieldResponse, error) {
	out := map[int64]model.FrmFieldResponse{}
	build := func() (any, error) {
		fields, err := u.fields.FindBySite(ctx, nil, siteID)
		if err != nil {
			return nil, err
		}
		m := map[int64]model.FrmFieldResponse{}
		for i := range fields {
			f := &fields[i]
			fr := convertField(f)
			if fr.FieldID != nil {
				m[*fr.FieldID] = fr
			}
		}
		return m, nil
	}

	if u.cache == nil {
		v, err := build()
		if err != nil {
			return out, err
		}
		return v.(map[int64]model.FrmFieldResponse), nil
	}
	key := u.cache.Keys.FieldsMap(siteID)
	if err := u.cache.Remember(ctx, key, 0, &out, build); err != nil {
		return out, err
	}
	return out, nil
}

// Ensure cache package import is used.
var _ cache.Driver = (*cache.RedisDriver)(nil)
