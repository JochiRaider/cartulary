package workbook_test

import (
	"crypto/sha256"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestTimelineClipboardProvenanceRejection_Integration(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-provenance-rejection", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "provenance-rejection-incident", "incident_key": "PROVENANCE-REJECTION", "title": "Provenance rejection"})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	const view = timeline.TimelineViewSchemaID
	sourceState := func(t testing.TB) [32]byte {
		t.Helper()
		var state string
		if err := h.DB.QueryRowContext(t.Context(), `SELECT jsonb_build_object(
 'envelopes',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY record_id),'[]') FROM records r WHERE incident_id=$1),
 'events',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY record_id),'[]') FROM timeline_events r WHERE incident_id=$1),
 'provenance',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY p.record_id,p.source_row_ordinal,p.source_column_ordinal,p.source_identity_hash),'[]') FROM timeline_source_provenance p JOIN timeline_events r USING(record_id) WHERE r.incident_id=$1),
 'projection',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY record_id),'[]') FROM timeline_grid_projection p WHERE incident_id=$1)
 )::text`, incidentID).Scan(&state); err != nil {
			t.Fatal(err)
		}
		return sha256.Sum256([]byte(state))
	}
	request := func(txn, text string, targets []map[string]any) map[string]any {
		return map[string]any{"view_schema_id": view, "client_txn_id": txn, "clipboard_text": text, "format": "tsv", "header_mode": "none", "start_field_key": "timeline.activity_synopsis_text", "columns": []string{"timeline.activity_synopsis_text"}, "targets": targets}
	}
	reject := func(t *testing.T, body map[string]any) {
		t.Helper()
		before := snapshotWorkbookState(t, h, login, incidentID, view)
		sourceBefore := sourceState(t)
		failure := httptestx.RequireErrorEnvelope(t, doClipboardPaste(t, h, login, incidentID, view, body), http.StatusBadRequest, "invalid_mutation_payload")["error"].(map[string]any)
		if !reflect.DeepEqual(failure["details"], map[string]any{"field": "clipboard_text", "reason_code": "invalid_value"}) || failure["message"] != "invalid mutation payload" {
			t.Fatalf("unsafe or incorrect provenance rejection: %#v", failure)
		}
		requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, view))
		if sourceState(t) != sourceBefore {
			t.Fatal("rejected provenance changed source, envelopes, provenance or projection")
		}
	}
	for _, tc := range []struct{ name, raw, txn string }{
		{"NUL", "hidden\x00raw", "provenance-nul"},
		{"raw byte bound", strings.Repeat("x", 65537), "provenance-raw-bound"},
		{"metadata byte bound", "raw", strings.Repeat("m", 65536)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reject(t, request(tc.txn, "First valid\tretained\nSecond invalid\t"+tc.raw, []map[string]any{{"kind": "create"}, {"kind": "create"}}))
		})
	}
	created := requireWorkbookCreate(t, h, login, incidentID, view, map[string]any{"client_txn_id": "provenance-capacity-target", "timeline.activity_synopsis_text": "Original"})
	record := uuid.MustParse(created["row"].(map[string]any)["record_id"].(string))
	hash := sha256.Sum256([]byte(`{"source_kind":"capture"}`))
	if _, err := h.DB.ExecContext(t.Context(), `INSERT INTO timeline_source_provenance(record_id,source_identity_hash,source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,created_at) SELECT $1,$2,n,1,'capture','{"source_kind":"capture"}','null','retained','2026-01-02T03:04:05.123456Z'::timestamptz FROM generate_series(1,4096) n`, record, hash[:]); err != nil {
		t.Fatal(err)
	}
	t.Run("capacity rolls back earlier row", func(t *testing.T) {
		reject(t, request("provenance-capacity", "Earlier create\tretained\nChanged patch\tnew provenance", []map[string]any{{"kind": "create"}, {"kind": "record", "record_id": record.String(), "base_row_version": 1}}))
	})
	t.Run("exact raw bound and receipt replay", func(t *testing.T) {
		raw := strings.Repeat("x", 65536)
		body := request("provenance-exact-bound", "Accepted\t"+raw, []map[string]any{{"kind": "create"}})
		accepted := requireClipboardPaste(t, h, login, incidentID, view, body, http.StatusOK)
		rows := accepted["rows"].([]any)
		if len(rows) != 1 {
			t.Fatal("accepted raw boundary did not create exactly one row")
		}
		id := rows[0].(map[string]any)["record_id"].(string)
		var count int
		var actual string
		if err := h.DB.QueryRowContext(t.Context(), `SELECT count(*),min(raw_value) FROM timeline_source_provenance WHERE record_id=$1`, id).Scan(&count, &actual); err != nil || count != 1 || actual != raw {
			t.Fatalf("accepted provenance changed: count=%d err=%v", count, err)
		}
		before := snapshotWorkbookState(t, h, login, incidentID, view)
		sourceBefore := sourceState(t)
		replayed := requireClipboardPaste(t, h, login, incidentID, view, body, http.StatusOK)
		if !reflect.DeepEqual(accepted, replayed) {
			t.Fatal("receipt replay changed the original result")
		}
		requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, view))
		if sourceState(t) != sourceBefore {
			t.Fatal("receipt replay changed retained provenance")
		}
	})
}
