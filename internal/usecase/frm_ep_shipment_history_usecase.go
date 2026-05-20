package usecase

import (
	"context"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
)

type FrmEpShipmentHistoryUseCase struct {
	history *repository.FrmEpShipmentHistoryRepository
}

func NewFrmEpShipmentHistoryUseCase(h *repository.FrmEpShipmentHistoryRepository) *FrmEpShipmentHistoryUseCase {
	return &FrmEpShipmentHistoryUseCase{history: h}
}

func (u *FrmEpShipmentHistoryUseCase) Update(ctx context.Context, siteID int64, in []model.EpShipmentHistoryInput) (int, error) {
	if len(in) == 0 {
		return 0, nil
	}
	rows := make([]repository.FrmEpShipmentHistoryInput, len(in))
	for i, r := range in {
		rows[i] = repository.FrmEpShipmentHistoryInput{
			EasypostShipmentID: nullStringPtr(r.EasypostShipmentID),
			UserID:             nullInt64Ptr(r.UserID),
			ChangeType:         nullStringPtr(r.ChangeType),
			Description:        nullStringPtr(r.Description),
		}
	}
	if err := u.history.InsertMany(ctx, nil, siteID, rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (u *FrmEpShipmentHistoryUseCase) List(ctx context.Context, siteID int64, req model.ListEpShipmentHistoryRequest) (repository.ListResult[model.EpShipmentHistoryItem], error) {
	page := req.PageNum
	if page < 1 {
		page = 1
	}
	perPage := req.Paginate
	if perPage <= 0 {
		perPage = 25
	}
	res, err := u.history.List(ctx, nil, siteID, repository.ListParams{
		Filters: req.Filters,
		SortBy:  req.SortBy,
		SortDir: req.SortDir,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return repository.ListResult[model.EpShipmentHistoryItem]{}, err
	}
	return repository.ListResult[model.EpShipmentHistoryItem]{
		Data:     converter.ToEpShipmentHistoryItems(res.Data),
		Total:    res.Total,
		Page:     res.Page,
		PerPage:  res.PerPage,
		LastPage: res.LastPage,
	}, nil
}
