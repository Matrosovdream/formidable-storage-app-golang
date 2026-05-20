package http_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestRoute_Health(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "GET", "/health", "", nil)
	require.Equal(t, 200, resp.StatusCode)
	var body map[string]any
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, "ok", body["status"])
}

func TestRoute_Login_BadCredentialsReturns422EnvelopeWithEmailError(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "POST", "/api/login", "", map[string]any{"email": "nobody@example.com", "password": "x"})
	require.Equal(t, 422, resp.StatusCode)
	var body map[string]any
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, false, body["success"])
	errs := body["errors"].(map[string]any)
	require.Contains(t, errs, "email")
}

func TestRoute_RegisterThenLoginThenMeThenLogout(t *testing.T) {
	s := helpers.NewTestServer(t)
	email := fmt.Sprintf("api-test-%d@example.com", time.Now().UnixNano())
	pw := "passw0rd!"

	// Register
	resp := s.Do(t, "POST", "/api/register", "", map[string]any{
		"name": "Alice", "email": email, "password": pw, "password_confirmation": pw,
	})
	require.Equal(t, 201, resp.StatusCode)
	var reg struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
		Token string `json:"token"`
	}
	helpers.DecodeJSON(t, resp, &reg)
	require.NotZero(t, reg.User.ID)
	require.NotEmpty(t, reg.Token)
	t.Cleanup(func() {
		_, _ = s.DB.Exec("DELETE FROM personal_access_tokens WHERE tokenable_id = $1", reg.User.ID)
		_, _ = s.DB.Exec("DELETE FROM users WHERE id = $1", reg.User.ID)
	})

	// Login
	resp = s.Do(t, "POST", "/api/login", "", map[string]any{"email": email, "password": pw})
	require.Equal(t, 200, resp.StatusCode)
	var login struct{ Token string `json:"token"` }
	helpers.DecodeJSON(t, resp, &login)
	require.NotEmpty(t, login.Token)

	// Me with token
	resp = s.Do(t, "GET", "/api/user", login.Token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var me map[string]any
	helpers.DecodeJSON(t, resp, &me)
	require.Equal(t, email, me["email"])

	// Me WITHOUT token → null
	resp = s.Do(t, "GET", "/api/user", "", nil)
	require.Equal(t, 200, resp.StatusCode)
	var nullBody any
	helpers.DecodeJSON(t, resp, &nullBody)
	require.Nil(t, nullBody)

	// Logout
	resp = s.Do(t, "POST", "/api/logout", login.Token, nil)
	require.Equal(t, 200, resp.StatusCode)

	// After logout the same token is invalid → 401
	resp = s.Do(t, "POST", "/api/logout", login.Token, nil)
	require.Equal(t, 401, resp.StatusCode)
}

func TestRoute_LogoutWithoutToken_Returns401(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "POST", "/api/logout", "", nil)
	require.Equal(t, 401, resp.StatusCode)
}
