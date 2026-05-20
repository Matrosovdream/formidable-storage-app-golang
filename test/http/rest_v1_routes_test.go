package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	gw "github.com/Matrosovdream/formidable-storage-app-golang/internal/gateway/messaging"
	"github.com/Matrosovdream/formidable-storage-app-golang/test/helpers"
	"github.com/stretchr/testify/require"
)

func TestRoute_RestV1_Status_RequiresBearer(t *testing.T) {
	s := helpers.NewTestServer(t)
	resp := s.Do(t, "GET", "/rest/v1/status", "", nil)
	require.Equal(t, 401, resp.StatusCode)
}

func TestRoute_RestV1_Status_WithBearer(t *testing.T) {
	s := helpers.NewTestServer(t)
	_, token := helpers.SeedSiteWithToken(t, s.DB)
	resp := s.Do(t, "GET", "/rest/v1/status", token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var body struct{ Status, Version, Timestamp string }
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, "ok", body.Status)
	require.Equal(t, "1.0", body.Version)
	require.NotEmpty(t, body.Timestamp)
}

func TestRoute_RestV1_FieldsUpdateAll_EnqueuesJob(t *testing.T) {
	s := helpers.NewTestServer(t)
	siteID, token := helpers.SeedSiteWithToken(t, s.DB)

	resp := s.Do(t, "POST", "/rest/v1/fields/update-all", token, map[string]any{
		"data": []map[string]any{
			{"field_id": 1, "key": "name", "type": "text", "label": "Name"},
			{"field_id": 2, "key": "email", "type": "email", "label": "Email"},
		},
	})
	require.Equal(t, 200, resp.StatusCode)
	var body struct {
		Queued bool `json:"queued"`
		Count  int  `json:"count"`
	}
	helpers.DecodeJSON(t, resp, &body)
	require.True(t, body.Queued)
	require.Equal(t, 2, body.Count)

	// Inspect the queue: one envelope must exist for our site.
	raw, err := s.Redis.RPop(context.Background(), s.Deps.Cfg.Queue.Stream).Bytes()
	require.NoError(t, err)
	var env gw.JobEnvelope
	require.NoError(t, json.Unmarshal(raw, &env))
	require.Equal(t, gw.JobUpdateFrmFields, env.Name)
	require.Equal(t, siteID, env.SiteID)
}

func TestRoute_RestV1_EntryHistoryUpdate_EnqueuesJob(t *testing.T) {
	s := helpers.NewTestServer(t)
	siteID, token := helpers.SeedSiteWithToken(t, s.DB)

	resp := s.Do(t, "POST", "/rest/v1/entry/history/update", token, map[string]any{
		"data": []map[string]any{
			{"entry_id": 9001, "field_id": 1, "update_type_code": "created", "new_value": "Alice"},
		},
	})
	require.Equal(t, 200, resp.StatusCode)
	raw, err := s.Redis.RPop(context.Background(), s.Deps.Cfg.Queue.Stream).Bytes()
	require.NoError(t, err)
	var env gw.JobEnvelope
	require.NoError(t, json.Unmarshal(raw, &env))
	require.Equal(t, gw.JobUpdateFrmEntryHistory, env.Name)
	require.Equal(t, siteID, env.SiteID)
}

func TestRoute_RestV1_EntryHistoryView(t *testing.T) {
	s := helpers.NewTestServer(t)
	ctx := context.Background()
	siteID, token := helpers.SeedSiteWithToken(t, s.DB)
	fieldPK := helpers.SeedField(t, s.DB, siteID, 1, "name", "Name")
	helpers.SeedUpdateType(t, s.DB, "created", "Created")
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO frm_entry_history (entry_id, site_id, field_id, old_value, new_value, change_date)
		 VALUES (50, $1, $2, NULL, 'Carol', NOW())`,
		siteID, fieldPK)
	require.NoError(t, err)

	resp := s.Do(t, "POST", "/rest/v1/entry/history/view/50", token, nil)
	require.Equal(t, 200, resp.StatusCode)
	var body struct {
		EntryID int64 `json:"entry_id"`
		Updates []struct {
			NewValue *string `json:"new_value"`
		} `json:"updates"`
	}
	helpers.DecodeJSON(t, resp, &body)
	require.Equal(t, int64(50), body.EntryID)
	require.Len(t, body.Updates, 1)
	require.NotNil(t, body.Updates[0].NewValue)
	require.Equal(t, "Carol", *body.Updates[0].NewValue)
}

func TestRoute_RestV1_EmailsLog_UpdateAllRaw_PersistsAndListWorks(t *testing.T) {
	s := helpers.NewTestServer(t)
	siteID, token := helpers.SeedSiteWithToken(t, s.DB)

	// Sync write.
	mk := func(i int) map[string]any {
		return map[string]any{
			"message_id":    fmt.Sprintf("rest-msg-%d", i),
			"subject":       fmt.Sprintf("Subj %d", i),
			"email_from":    "noreply@example.com",
			"email_to":      fmt.Sprintf("u%d@example.com", i),
			"content_plain": "plain",
			"content_html":  "<p>html</p>",
			"status":        i % 2,
		}
	}
	resp := s.Do(t, "POST", "/rest/v1/emailslog/update-all/raw", token, map[string]any{
		"data": []map[string]any{mk(1), mk(2), mk(3)},
	})
	require.Equal(t, 200, resp.StatusCode)
	var raw struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
	}
	helpers.DecodeJSON(t, resp, &raw)
	require.True(t, raw.Success)
	require.Equal(t, 3, raw.Count)

	// List back.
	resp = s.Do(t, "POST", "/rest/v1/emailslog/list", token, map[string]any{"paginate": 10})
	require.Equal(t, 200, resp.StatusCode)
	var page struct {
		Total int64 `json:"total"`
		Data  []struct {
			MessageID *string `json:"message_id"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &page)
	require.Equal(t, int64(3), page.Total)
	require.Len(t, page.Data, 3)
	_ = siteID
}

func TestRoute_RestV1_EmailsLog_UpdateAll_EnqueuesJob(t *testing.T) {
	s := helpers.NewTestServer(t)
	_, token := helpers.SeedSiteWithToken(t, s.DB)

	resp := s.Do(t, "POST", "/rest/v1/emailslog/update-all", token, map[string]any{
		"data": []map[string]any{
			{"message_id": "q-msg-1", "subject": "Q1", "email_to": "x@example.com", "content_plain": "p", "content_html": "<h/>", "status": 0},
		},
	})
	require.Equal(t, 200, resp.StatusCode)
	raw, err := s.Redis.RPop(context.Background(), s.Deps.Cfg.Queue.Stream).Bytes()
	require.NoError(t, err)
	var env gw.JobEnvelope
	require.NoError(t, json.Unmarshal(raw, &env))
	require.Equal(t, gw.JobUpdateEmailsLog, env.Name)
}

func TestRoute_RestV1_EpShipmentHistory_UpdateAndList(t *testing.T) {
	s := helpers.NewTestServer(t)
	ctx := context.Background()
	siteID, token := helpers.SeedSiteWithToken(t, s.DB)

	// Direct DB insert so List has rows to read back without consuming the job from the queue.
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO frm_easypost_shipment_history (easypost_shipment_id, user_id, site_id, change_type, description)
		VALUES ('shp_1', 7, $1, 'created', 'first event')`, siteID)
	require.NoError(t, err)

	// List
	resp := s.Do(t, "POST", "/rest/v1/ep-shipment-history/list", token, map[string]any{})
	require.Equal(t, 200, resp.StatusCode)
	var page struct {
		Total int64 `json:"total"`
		Data  []struct {
			ChangeType *string `json:"change_type"`
		} `json:"data"`
	}
	helpers.DecodeJSON(t, resp, &page)
	require.GreaterOrEqual(t, page.Total, int64(1))
	require.NotEmpty(t, page.Data)

	// UpdateAll → must enqueue
	resp = s.Do(t, "POST", "/rest/v1/ep-shipment-history/update-all", token, map[string]any{
		"data": []map[string]any{
			{"easypost_shipment_id": "shp_2", "change_type": "updated", "description": "label printed"},
		},
	})
	require.Equal(t, 200, resp.StatusCode)
	raw, err := s.Redis.RPop(context.Background(), s.Deps.Cfg.Queue.Stream).Bytes()
	require.NoError(t, err)
	var env gw.JobEnvelope
	require.NoError(t, json.Unmarshal(raw, &env))
	require.Equal(t, gw.JobUpdateEpShipmentHistory, env.Name)
}
