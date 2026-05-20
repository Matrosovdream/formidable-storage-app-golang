package http_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestRoute_DataEntries_WithSeededHistoryAndEmails(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	ctx := context.Background()

	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)
	fieldPK := helpers.SeedField(t, s.DB, siteID, 555, "name", "Name")
	helpers.SeedUpdateType(t, s.DB, "created", "Created")

	// One email-log row and one history row, both scoped to entry_id=42.
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO frm_emails_log
			(entry_id, site_id, subject, message_id, email_from, email_to, content_plain, content_html, status, date_sent)
		VALUES (42, $1, 'subj', 'mid-42', 'a@x', 'b@x', 'p', '<h/>', 1, NOW())
	`, siteID)
	require.NoError(t, err)
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO frm_entry_history (entry_id, site_id, field_id, old_value, new_value, change_date)
		VALUES (42, $1, $2, NULL, 'value', NOW())
	`, siteID, fieldPK)
	require.NoError(t, err)

	// list entries
	resp := s.Do(t, "GET", fmt.Sprintf("/api/data/entries/%d", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var page struct {
		Total int64 `json:"total"`
		Data  []struct {
			EntryID     int64 `json:"entry_id"`
			EmailCount  int64 `json:"email_count"`
			UpdateCount int64 `json:"update_count"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &page)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(42), page.Data[0].EntryID)
	require.Equal(t, int64(1), page.Data[0].EmailCount)
	require.Equal(t, int64(1), page.Data[0].UpdateCount)

	// entry updates
	resp = s.Do(t, "GET", fmt.Sprintf("/api/data/entries/%d/42/updates", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var upd struct {
		EntryID int64 `json:"entry_id"`
		Updates []struct {
			NewValue *string `json:"new_value"`
		} `json:"updates"`
	}
	helpers.DecodeJSON(t, resp, &upd)
	require.Equal(t, int64(42), upd.EntryID)
	require.Len(t, upd.Updates, 1)
	require.NotNil(t, upd.Updates[0].NewValue)
	require.Equal(t, "value", *upd.Updates[0].NewValue)

	// entry emails
	resp = s.Do(t, "GET", fmt.Sprintf("/api/data/entries/%d/42/emails", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var emails struct {
		Data []struct {
			MessageID *string `json:"message_id"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &emails)
	require.Len(t, emails.Data, 1)
	require.NotNil(t, emails.Data[0].MessageID)
	require.Equal(t, "mid-42", *emails.Data[0].MessageID)
}

func TestRoute_DataEntries_RequiresAuth(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "GET", "/api/data/entries/1", "", nil)
	require.Equal(t, 401, resp.StatusCode)
}
