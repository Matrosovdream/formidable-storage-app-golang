package usecase

import (
	"context"
	"encoding/hex"
	"crypto/rand"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type SiteUseCase struct {
	db         *sqlx.DB
	validate   *validator.Validate
	sites      *repository.SiteRepository
	siteTokens *repository.SiteTokenRepository
	queueStats *QueueStatsUseCase
}

func NewSiteUseCase(
	db *sqlx.DB,
	v *validator.Validate,
	sites *repository.SiteRepository,
	siteTokens *repository.SiteTokenRepository,
	queueStats *QueueStatsUseCase,
) *SiteUseCase {
	return &SiteUseCase{db: db, validate: v, sites: sites, siteTokens: siteTokens, queueStats: queueStats}
}

func (u *SiteUseCase) List(ctx context.Context) ([]model.SiteResponse, error) {
	items, err := u.sites.FindAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	return converter.ToSiteResponses(items), nil
}

func (u *SiteUseCase) View(ctx context.Context, siteID int64) (model.SiteWithStatsResponse, error) {
	s, err := u.sites.FindByID(ctx, nil, siteID)
	if err != nil {
		return model.SiteWithStatsResponse{}, err
	}
	if s == nil {
		return model.SiteWithStatsResponse{}, ErrNotFound
	}

	counts, err := u.sites.Counts(ctx, nil, siteID)
	if err != nil {
		return model.SiteWithStatsResponse{}, err
	}
	queue, _ := u.queueStats.CountsForSite(ctx, siteID)

	resp := model.SiteWithStatsResponse{
		SiteResponse: converter.ToSiteResponse(s),
		Stats: model.SiteStatsDTO{
			Fields:       counts.Fields,
			Emails:       counts.Emails,
			EntryHistory: counts.EntryHistory,
			Queue:        queue,
		},
	}

	tok, err := u.siteTokens.FindBySiteID(ctx, nil, siteID)
	if err == nil && tok != nil {
		resp.Token = tok.Token
	}
	return resp, nil
}

func (u *SiteUseCase) Create(ctx context.Context, req model.CreateSiteRequest) (model.SiteWithStatsResponse, error) {
	if err := u.validate.Struct(req); err != nil {
		return model.SiteWithStatsResponse{}, &ValidationError{Fields: fieldsFromValidator(err)}
	}
	var out model.SiteWithStatsResponse
	err := repository.WithTx(ctx, u.db, func(tx *sqlx.Tx) error {
		s := &entity.Site{Name: req.Name, URL: req.URL}
		if err := u.sites.Create(ctx, tx, s); err != nil {
			return err
		}
		token, err := generateSiteToken()
		if err != nil {
			return err
		}
		st := &entity.SiteToken{SiteID: s.ID, Token: token}
		if err := u.siteTokens.Create(ctx, tx, st); err != nil {
			return err
		}
		out = model.SiteWithStatsResponse{
			SiteResponse: converter.ToSiteResponse(s),
			Token:        token,
			Stats:        model.SiteStatsDTO{Queue: map[string]int64{}},
		}
		return nil
	})
	return out, err
}

func (u *SiteUseCase) Delete(ctx context.Context, siteID int64) error {
	s, err := u.sites.FindByID(ctx, nil, siteID)
	if err != nil {
		return err
	}
	if s == nil {
		return ErrNotFound
	}
	return u.sites.Delete(ctx, nil, siteID)
}

// GetByToken locates the site for a REST bearer token, used by middleware.
func (u *SiteUseCase) GetByToken(ctx context.Context, token string) (*entity.Site, error) {
	st, err := u.siteTokens.FindByToken(ctx, nil, token)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, ErrUnauthorized
	}
	return u.sites.FindByID(ctx, nil, st.SiteID)
}

func generateSiteToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
