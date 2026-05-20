package usecase

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	db         *sqlx.DB
	validate   *validator.Validate
	users      *repository.UserRepository
	tokens     *repository.PersonalAccessTokenRepository
	bcryptCost int
	tokenTTL   time.Duration // 0 = no expiry
}

func NewAuthUseCase(
	db *sqlx.DB,
	v *validator.Validate,
	users *repository.UserRepository,
	tokens *repository.PersonalAccessTokenRepository,
	bcryptCost int,
	tokenLifetimeMinutes int,
) *AuthUseCase {
	var ttl time.Duration
	if tokenLifetimeMinutes > 0 {
		ttl = time.Duration(tokenLifetimeMinutes) * time.Minute
	}
	return &AuthUseCase{db: db, validate: v, users: users, tokens: tokens, bcryptCost: bcryptCost, tokenTTL: ttl}
}

func (a *AuthUseCase) issueToken(ctx context.Context, q repository.Querier, userID int64) (string, error) {
	plain, err := generatePlaintextToken()
	if err != nil {
		return "", err
	}
	t := &entity.PersonalAccessToken{
		TokenableType: "App\\Models\\User",
		TokenableID:   userID,
		Name:          "auth",
		Token:         sha256Hex(plain),
		Abilities:     sql.NullString{String: `["*"]`, Valid: true},
	}
	if a.tokenTTL > 0 {
		t.ExpiresAt = sql.NullTime{Time: time.Now().Add(a.tokenTTL).UTC(), Valid: true}
	}
	if err := a.tokens.Create(ctx, q, t); err != nil {
		return "", err
	}
	return FormatToken(t.ID, plain), nil
}

func (a *AuthUseCase) Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error) {
	if err := a.validate.Struct(req); err != nil {
		return model.LoginResponse{}, &ValidationError{Fields: fieldsFromValidator(err)}
	}
	u, err := a.users.FindByEmail(ctx, nil, req.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}
	if u == nil || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return model.LoginResponse{}, ErrInvalidCredentials
	}

	tok, err := a.issueToken(ctx, nil, u.ID)
	if err != nil {
		return model.LoginResponse{}, err
	}
	return model.LoginResponse{User: converter.ToUserResponse(u), Message: "Logged in", Token: tok}, nil
}

func (a *AuthUseCase) Register(ctx context.Context, req model.RegisterRequest) (model.LoginResponse, error) {
	if err := a.validate.Struct(req); err != nil {
		return model.LoginResponse{}, &ValidationError{Fields: fieldsFromValidator(err)}
	}

	existing, err := a.users.FindByEmail(ctx, nil, req.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}
	if existing != nil {
		return model.LoginResponse{}, &ValidationError{Fields: map[string][]string{"email": {"The email has already been taken."}}}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), a.bcryptCost)
	if err != nil {
		return model.LoginResponse{}, err
	}

	var resp model.LoginResponse
	err = repository.WithTx(ctx, a.db, func(tx *sqlx.Tx) error {
		u := &entity.User{Name: req.Name, Email: req.Email, Password: string(hash)}
		if err := a.users.Create(ctx, tx, u); err != nil {
			return err
		}
		tok, err := a.issueToken(ctx, tx, u.ID)
		if err != nil {
			return err
		}
		resp = model.LoginResponse{User: converter.ToUserResponse(u), Message: "Registered", Token: tok}
		return nil
	})
	return resp, err
}

func (a *AuthUseCase) Logout(ctx context.Context, tokenID int64) error {
	return a.tokens.Delete(ctx, nil, tokenID)
}

// VerifyToken parses a "<id>|<plain>" string and returns the matching token row + user, or an error.
func (a *AuthUseCase) VerifyToken(ctx context.Context, raw string) (*entity.PersonalAccessToken, *entity.User, error) {
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return nil, nil, ErrUnauthorized
	}
	tok, err := a.tokens.FindByHash(ctx, nil, sha256Hex(parts[1]))
	if err != nil {
		return nil, nil, err
	}
	if tok == nil {
		return nil, nil, ErrUnauthorized
	}
	if tok.ExpiresAt.Valid && tok.ExpiresAt.Time.Before(time.Now()) {
		return nil, nil, ErrUnauthorized
	}
	u, err := a.users.FindByID(ctx, nil, tok.TokenableID)
	if err != nil {
		return nil, nil, err
	}
	if u == nil {
		return nil, nil, ErrUnauthorized
	}
	_ = a.tokens.Touch(ctx, nil, tok.ID)
	return tok, u, nil
}

func fieldsFromValidator(err error) map[string][]string {
	out := map[string][]string{}
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range vErrs {
			field := fe.Field()
			out[field] = append(out[field], fe.Error())
		}
	} else {
		out["_"] = []string{err.Error()}
	}
	return out
}
