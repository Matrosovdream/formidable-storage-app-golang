package http_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestRoute_SiteViewEmails_FieldsAndUpdates(t *testing.T) {
	s := helpers.NewTestServer(t)
	token := registerAndToken(t, s)
	ctx := context.Background()

	siteID, _ := helpers.SeedSiteWithToken(t, s.DB)
	fieldPK := helpers.SeedField(t, s.DB, siteID, 777, "name", "Name")
	helpers.SeedUpdateType(t, s.DB, "updated", "Updated")

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO frm_emails_log
			(entry_id, site_id, subject, message_id, email_from, email_to, content_plain, content_html, status, date_sent)
		VALUES (101, $1, 'subj', 'mid-101', 'a@x', 'b@x', 'plain-body', '<b>html</b>', 1, NOW())
	`, siteID)
	require.NoError(t, err)
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO frm_entry_history (entry_id, site_id, field_id, old_value, new_value, change_date)
		VALUES (101, $1, $2, 'old', 'new', NOW())
	`, siteID, fieldPK)
	require.NoError(t, err)

	// emails — content fields must be present
	resp := s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d/emails", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var emails struct {
		Total int64 `json:"total"`
		Data  []struct {
			MessageID    *string `json:"message_id"`
			ContentPlain *string `json:"content_plain"`
			ContentHTML  *string `json:"content_html"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &emails)
	require.Equal(t, int64(1), emails.Total)
	require.Len(t, emails.Data, 1)
	require.NotNil(t, emails.Data[0].MessageID)
	require.Equal(t, "mid-101", *emails.Data[0].MessageID)
	require.NotNil(t, emails.Data[0].ContentPlain)
	require.Equal(t, "plain-body", *emails.Data[0].ContentPlain)
	require.NotNil(t, emails.Data[0].ContentHTML)
	require.Equal(t, "<b>html</b>", *emails.Data[0].ContentHTML)

	// fields
	resp = s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d/fields", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var fields struct {
		Total int64 `json:"total"`
		Data  []struct {
			FieldID *int64  `json:"field_id"`
			Key     *string `json:"key"`
			Label   *string `json:"label"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &fields)
	require.Equal(t, int64(1), fields.Total)
	require.Len(t, fields.Data, 1)
	require.NotNil(t, fields.Data[0].FieldID)
	require.Equal(t, int64(777), *fields.Data[0].FieldID)
	require.NotNil(t, fields.Data[0].Label)
	require.Equal(t, "Name", *fields.Data[0].Label)

	// entry-updates
	resp = s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d/entry-updates", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var updates struct {
		Total int64 `json:"total"`
		Data  []struct {
			EntryID  *int64  `json:"entry_id"`
			FieldKey *string `json:"field_key"`
			NewValue *string `json:"new_value"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &updates)
	require.Equal(t, int64(1), updates.Total)
	require.Len(t, updates.Data, 1)
	require.NotNil(t, updates.Data[0].EntryID)
	require.Equal(t, int64(101), *updates.Data[0].EntryID)
	require.NotNil(t, updates.Data[0].FieldKey)
	require.Equal(t, "name", *updates.Data[0].FieldKey)
	require.NotNil(t, updates.Data[0].NewValue)
	require.Equal(t, "new", *updates.Data[0].NewValue)

	// filter by external field_id should map → pk and return the row
	resp = s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d/entry-updates?field_id=777", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	helpers.DecodeJSON(t, resp, &updates)
	require.Equal(t, int64(1), updates.Total)

	// filter by missing external field_id → empty page
	resp = s.Do(t, "GET", fmt.Sprintf("/api/sites/view/%d/entry-updates?field_id=99999", siteID), token, nil)
	require.Equal(t, 200, resp.StatusCode)
	helpers.DecodeJSON(t, resp, &updates)
	require.Equal(t, int64(0), updates.Total)
	require.Len(t, updates.Data, 0)
}

func TestRoute_SiteViewEmails_RequiresAuth(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "GET", "/api/sites/view/1/emails", "", nil)
	require.Equal(t, 401, resp.StatusCode)
}
