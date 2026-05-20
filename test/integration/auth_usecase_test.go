package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/config"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_RegisterThenLoginThenVerify(t *testing.T) {
	db := helpers.DB(t)
	v := config.NewValidator()

	users := repository.NewUserRepository(db)
	pats := repository.NewPersonalAccessTokenRepository(db)
	uc := usecase.NewAuthUseCase(db, v, users, pats, 4, 0) // bcrypt cost 4 for test speed

	email := fmt.Sprintf("auth-test-%d@example.com", time.Now().UnixNano())
	password := "passw0rd!"

	resp, err := uc.Register(context.Background(), model.RegisterRequest{
		Name:                 "Alice",
		Email:                email,
		Password:             password,
		PasswordConfirmation: password,
	})
	require.NoError(t, err)
	require.Greater(t, resp.User.ID, int64(0))
	require.NotEmpty(t, resp.Token)
	require.True(t, strings.Contains(resp.Token, "|"))
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM personal_access_tokens WHERE tokenable_id = $1", resp.User.ID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", resp.User.ID)
	})

	// Verifies the issued token resolves to the right user.
	_, u, err := uc.VerifyToken(context.Background(), resp.Token)
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, resp.User.ID, u.ID)

	// Login with the same credentials issues a new token.
	loginResp, err := uc.Login(context.Background(), model.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.Equal(t, resp.User.ID, loginResp.User.ID)
	require.NotEmpty(t, loginResp.Token)
	require.NotEqual(t, resp.Token, loginResp.Token)

	// Wrong password.
	_, err = uc.Login(context.Background(), model.LoginRequest{
		Email:    email,
		Password: "wrong-password",
	})
	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
}

func TestAuthUseCase_RegisterDuplicateEmail(t *testing.T) {
	db := helpers.DB(t)
	v := config.NewValidator()

	users := repository.NewUserRepository(db)
	pats := repository.NewPersonalAccessTokenRepository(db)
	uc := usecase.NewAuthUseCase(db, v, users, pats, 4, 0)

	email := fmt.Sprintf("dup-%d@example.com", time.Now().UnixNano())
	pw := "passw0rd!"
	req := model.RegisterRequest{
		Name:                 "First",
		Email:                email,
		Password:             pw,
		PasswordConfirmation: pw,
	}
	resp, err := uc.Register(context.Background(), req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM personal_access_tokens WHERE tokenable_id = $1", resp.User.ID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", resp.User.ID)
	})

	_, err = uc.Register(context.Background(), req)
	require.Error(t, err)
	ve, ok := err.(*usecase.ValidationError)
	require.True(t, ok)
	require.Contains(t, ve.Fields, "email")
}
