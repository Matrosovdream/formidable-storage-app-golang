package http_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestRoute_SiteGenerate_RequiresAuth(t *testing.T) {
	s := helpers.NewTestServer(t)

	for _, path := range []string{
		"/api/sites/generate/1/emails",
		"/api/sites/generate/1/fields",
		"/api/sites/generate/1/entry-updates",
	} {
		resp := s.Do(t, "POST", path, "", nil)
		require.Equalf(t, 401, resp.StatusCode, "POST %s without auth", path)
	}
}

func TestRoute_SiteGenerate_InvalidSiteID_Returns404(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)

	resp := s.Do(t, "POST", "/api/sites/generate/not-a-number/emails", token, nil)
	require.Equal(t, 404, resp.StatusCode)
}

func TestRoute_SiteGenerate_UnknownSite_Returns404(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)

	resp := s.Do(t, "POST", "/api/sites/generate/9999999/emails", token, nil)
	require.Equal(t, 404, resp.StatusCode)
}

func TestRoute_SiteGenerate_Emails_DefaultAmount(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)

	resp := s.Do(t, "POST", fmt.Sprintf("/api/sites/generate/%d/emails", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)

	var body struct {
		Success bool   `json:"success"`
		Kind    string `json:"kind"`
		SiteID  int64  `json:"site_id"`
		Count   int    `json:"count"`
		Length  int    `json:"length"`
		Timings struct {
			GenerationMs float64 `json:"generation_ms"`
			InsertionMs  float64 `json:"insertion_ms"`
			TotalMs      float64 `json:"total_ms"`
		} `json:"timings"`
	}
	helpers.DecodeJSON(t, resp, &body)

	require.True(t, body.Success)
	require.Equal(t, "emails", body.Kind)
	require.Equal(t, siteID, body.SiteID)
	require.Equal(t, 10, body.Count, "no amount → default 10")
	require.Equal(t, 200, body.Length, "no length → default 200")

	var n int
	require.NoError(t, s.DB.GetContext(context.Background(), &n,
		"SELECT COUNT(*) FROM frm_emails_log WHERE site_id = $1", siteID))
	require.Equal(t, 10, n)
}

func TestRoute_SiteGenerate_Emails_CustomAmountAndLength(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)

	resp := s.Do(t, "POST",
		fmt.Sprintf("/api/sites/generate/%d/emails?amount=3&length=150", siteID),
		token, nil)
	require.Equal(t, 200, resp.StatusCode)

	var body struct {
		Count  int `json:"count"`
		Length int `json:"length"`
	}
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, 3, body.Count)
	require.Equal(t, 150, body.Length)
}

func TestRoute_SiteGenerate_Fields(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)

	resp := s.Do(t, "POST",
		fmt.Sprintf("/api/sites/generate/%d/fields?amount=4", siteID),
		token, nil)
	require.Equal(t, 200, resp.StatusCode)

	var body struct {
		Success bool   `json:"success"`
		Kind    string `json:"kind"`
		SiteID  int64  `json:"site_id"`
		Count   int    `json:"count"`
	}
	helpers.DecodeJSON(t, resp, &body)
	require.True(t, body.Success)
	require.Equal(t, "fields", body.Kind)
	require.Equal(t, siteID, body.SiteID)
	require.Equal(t, 4, body.Count)

	var n int
	require.NoError(t, s.DB.GetContext(context.Background(), &n,
		"SELECT COUNT(*) FROM frm_fields WHERE site_id = $1", siteID))
	require.Equal(t, 4, n)
}

func TestRoute_SiteGenerate_EntryUpdates_CreatesPlaceholderWhenNoFields(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)

	// Site has zero fields → the use case must seed a placeholder field, then
	// insert history rows.
	resp := s.Do(t, "POST",
		fmt.Sprintf("/api/sites/generate/%d/entry-updates?amount=2", siteID),
		token, nil)
	require.Equal(t, 200, resp.StatusCode)

	var body struct {
		Success bool   `json:"success"`
		Kind    string `json:"kind"`
		SiteID  int64  `json:"site_id"`
		Count   int    `json:"count"`
	}
	helpers.DecodeJSON(t, resp, &body)
	require.True(t, body.Success)
	require.Equal(t, "entry_updates", body.Kind)
	require.Equal(t, siteID, body.SiteID)
	require.Equal(t, 2, body.Count)

	ctx := context.Background()
	var histCount, fieldCount int
	require.NoError(t, s.DB.GetContext(ctx, &histCount,
		"SELECT COUNT(*) FROM frm_entry_history WHERE site_id = $1", siteID))
	require.Equal(t, 2, histCount)
	require.NoError(t, s.DB.GetContext(ctx, &fieldCount,
		"SELECT COUNT(*) FROM frm_fields WHERE site_id = $1", siteID))
	require.GreaterOrEqual(t, fieldCount, 1, "placeholder field should be created")
}

func TestRoute_SiteGenerate_EntryUpdates_UsesExistingFields(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)
	helpers.SeedField(t, s.DB, siteID, 1234, "name", "Name")

	resp := s.Do(t, "POST",
		fmt.Sprintf("/api/sites/generate/%d/entry-updates?amount=5", siteID),
		token, nil)
	require.Equal(t, 200, resp.StatusCode)

	var body struct {
		Count int `json:"count"`
	}
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, 5, body.Count)

	var n int
	require.NoError(t, s.DB.GetContext(context.Background(), &n,
		"SELECT COUNT(*) FROM frm_entry_history WHERE site_id = $1", siteID))
	require.Equal(t, 5, n)
}
