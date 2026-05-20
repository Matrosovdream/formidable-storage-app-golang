package http_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

// registerAndToken: helper that registers a fresh user via the API and returns the token.
func registerAndToken(t *testing.T, s *helpers.TestServer) string {
	email := fmt.Sprintf("sites-%d@example.com", time.Now().UnixNano())
	pw := "passw0rd!"
	resp := s.Do(t, "POST", "/api/register", "", map[string]any{
		"name": "U", "email": email, "password": pw, "password_confirmation": pw,
	})
	require.Equal(t, 201, resp.StatusCode)
	var out struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
		Token string `json:"token"`
	}
	helpers.DecodeJSON(t, resp, &out)
	t.Cleanup(func() {
		_, _ = s.DB.Exec("DELETE FROM personal_access_tokens WHERE tokenable_id = $1", out.User.ID)
		_, _ = s.DB.Exec("DELETE FROM users WHERE id = $1", out.User.ID)
	})
	return out.Token
}

func TestRoute_SitesList_RequiresAuth(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "GET", "/api/sites/list", "", nil)
	require.Equal(t, 401, resp.StatusCode)
}

func TestRoute_SitesCRUD(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)

	// create form
	resp := s.Do(t, "GET", "/api/sites/create", token, nil)
	require.Equal(t, 200, resp.StatusCode)

	// store
	url := fmt.Sprintf("https://route-test-%d.example/", time.Now().UnixNano())
	resp = s.Do(t, "POST", "/api/sites/store", token, map[string]any{"name": "Routed", "url": url})
	require.Equal(t, 201, resp.StatusCode)
	var stored struct {
		ID    int64  `json:"id"`
		Token string `json:"token"`
	}
	helpers.DecodeJSON(t, resp, &stored)
	require.NotZero(t, stored.ID)
	require.NotEmpty(t, stored.Token)
	t.Cleanup(func() { _, _ = s.DB.Exec("DELETE FROM sites WHERE id = $1", stored.ID) })

	// list (must include our site)
	resp = s.Do(t, "GET", "/api/sites/list", token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var listBody struct {
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &listBody)
	require.True(t, containsSiteID(listBody.Data, stored.ID))

	// view
	resp = s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d", stored.ID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var view map[string]any
	helpers.DecodeJSON(t, resp, &view)
	require.Equal(t, url, view["url"])
	require.NotNil(t, view["stats"])

	// delete
	resp = s.Do(t, "DELETE", fmt.Sprintf("/api/sites/delete/%d", stored.ID), token, nil)
	require.Equal(t, 200, resp.StatusCode)

	// view 404
	resp = s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d", stored.ID), token, nil)
	require.Equal(t, 404, resp.StatusCode)
}

func TestRoute_SitesStore_ValidationError(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)

	resp := s.Do(t, "POST", "/api/sites/store", token, map[string]any{"name": "", "url": "not-a-url"})
	require.Equal(t, 422, resp.StatusCode)
	var body map[string]any
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, false, body["success"])
	require.Contains(t, body["errors"], "Name")
	require.Contains(t, body["errors"], "URL")
}

func containsSiteID(items []struct{ ID int64 `json:"id"` }, id int64) bool {
	for _, x := range items {
		if x.ID == id {
			return true
		}
	}
	return false
}
