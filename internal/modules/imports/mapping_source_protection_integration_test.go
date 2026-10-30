package imports_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestMappingApprovalProtectsPreparedSource_Integration(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	for _, family := range []string{"generic", "analytical"} {
		for _, scenario := range []string{"changed", "unchanged", "canceled", "held_through_persistence"} {
			t.Run(family+"/"+scenario, func(t *testing.T) {
				content := "summary\noriginal\n"
				token := family + "-" + scenario
				mapping := map[string]any{"client_txn_id": token, "target_view_schema_id": "cartulary.view.timeline.v2", "header_row_ref": 1, "data_start_row_ref": 2, "unknown_column_policy": "preserve_raw_capture", "source_columns": []map[string]any{{"source_column_ordinal": 1, "source_header_text": "summary", "field_key": "timeline.activity_synopsis_text", "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"}}}
				if family == "analytical" {
					content = sourceAdmissionHeader + sourceAdmissionRow
					mapping = networkFlowMappingPayload(token)
				}
				session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "upload-"+token, content, token+".csv")
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				snapshot := func() string {
					t.Helper()
					var got string
					if err := h.DB.QueryRowContext(ctx, `SELECT jsonb_build_object('unit',to_jsonb(u),'status',s.session_status,'selected',s.selected_unit_ids)::text FROM import_units u JOIN import_sessions s USING(import_session_id) WHERE import_unit_id::text=$1`, unit).Scan(&got); err != nil {
						t.Fatal(err)
					}
					return got
				}
				before := snapshot()
				blocker, err := h.DB.Conn(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer blocker.Close()
				var pid int
				if err = blocker.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
					t.Fatal(err)
				}
				tx, err := blocker.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback()
				var ref string
				if scenario == "held_through_persistence" {
					if _, err = blocker.ExecContext(ctx, `SELECT pg_advisory_lock(179,301)`); err != nil {
						t.Fatal(err)
					}
					defer blocker.ExecContext(context.Background(), `SELECT pg_advisory_unlock(179,301)`)
					sql := fmt.Sprintf(`CREATE FUNCTION mapping_source_protection_gate() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.import_unit_id='%s'::uuid AND NEW.approved_mapping_json IS NOT NULL THEN PERFORM pg_advisory_xact_lock(179,301); END IF; RETURN NEW; END $$; CREATE TRIGGER mapping_source_protection_gate BEFORE UPDATE ON import_units FOR EACH ROW EXECUTE FUNCTION mapping_source_protection_gate()`, unit)
					if _, err = h.DB.ExecContext(ctx, sql); err != nil {
						t.Fatal(err)
					}
					defer h.DB.ExecContext(context.Background(), `DROP TRIGGER mapping_source_protection_gate ON import_units; DROP FUNCTION mapping_source_protection_gate()`)
				} else {
					if err = tx.QueryRowContext(ctx, `SELECT source_stream_ref FROM import_source_streams WHERE import_unit_id::text=$1 FOR UPDATE`, unit).Scan(&ref); err != nil {
						t.Fatal(err)
					}
				}
				raw, err := json.Marshal(mapping)
				if err != nil {
					t.Fatal(err)
				}
				requestCtx, requestCancel := context.WithCancel(ctx)
				defer requestCancel()
				request, err := http.NewRequestWithContext(requestCtx, http.MethodPut, h.Server.HTTP.URL+"/api/v1/import-sessions/"+session+"/units/"+unit+"/mapping", bytes.NewReader(raw))
				if err != nil {
					t.Fatal(err)
				}
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
				request.AddCookie(login.Session)
				request.AddCookie(login.CSRF)
				type result struct {
					response *http.Response
					err      error
				}
				done := make(chan result, 1)
				go func() { r, e := http.DefaultClient.Do(request); done <- result{r, e} }()
				blocking := `SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid))`
				if got := waitForSQLCount(t, h.DB, 10*time.Second, 1, blocking, pid); got != 1 {
					t.Fatal("approval did not reach controlled lock boundary")
				}
				switch scenario {
				case "changed":
					changed := []byte(content + "\n")
					if _, err = tx.ExecContext(ctx, `UPDATE import_source_streams SET source_bytes=$2,source_byte_size=$3,source_content_sha256=$4 WHERE source_stream_ref=$1`, ref, changed, len(changed), fmt.Sprintf("%x", sha256.Sum256(changed))); err != nil {
						t.Fatal(err)
					}
				case "canceled":
					requestCancel()
					if got := waitForSQLCount(t, h.DB, 10*time.Second, 0, blocking, pid); got != 0 {
						t.Fatal("canceled approval retained its blocked transaction")
					}
				case "held_through_persistence":
					probe, probeErr := h.DB.BeginTx(ctx, nil)
					if probeErr != nil {
						t.Fatal(probeErr)
					}
					_, probeErr = probe.ExecContext(ctx, `SELECT source_stream_ref FROM import_source_streams WHERE import_unit_id::text=$1 FOR UPDATE NOWAIT`, unit)
					_ = probe.Rollback()
					if probeErr == nil {
						t.Fatal("source protection was released before mapping publication")
					}
					if _, err = blocker.ExecContext(ctx, `SELECT pg_advisory_unlock(179,301)`); err != nil {
						t.Fatal(err)
					}
				}
				if err = tx.Commit(); err != nil {
					t.Fatal(err)
				}
				var outcome result
				select {
				case outcome = <-done:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				switch scenario {
				case "canceled":
					if outcome.response != nil {
						outcome.response.Body.Close()
					}
					if outcome.err == nil {
						t.Fatal("canceled request unexpectedly succeeded")
					}
				case "changed":
					if outcome.err != nil {
						t.Fatal(outcome.err)
					}
					envelope := httptestx.RequireErrorEnvelope(t, outcome.response, http.StatusConflict, "import_apply_blocked")
					if envelope["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "source_changed" {
						t.Fatal(envelope)
					}
				default:
					if outcome.err != nil {
						t.Fatal(outcome.err)
					}
					httptestx.RequireSuccessEnvelope(t, outcome.response, http.StatusOK)
					replay := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, "/api/v1/import-sessions/"+session+"/units/"+unit+"/mapping", mapping)
					httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)
				}
				if scenario == "changed" || scenario == "canceled" {
					if got := snapshot(); got != before {
						t.Fatalf("failed approval changed durable unit/session: before=%s after=%s", before, got)
					}
					var count int
					if err = h.DB.QueryRowContext(ctx, `SELECT count(*) FROM route_idempotency WHERE route_key='imports.units.mapping' AND client_txn_id=$1`, token).Scan(&count); err != nil || count != 0 {
						t.Fatalf("failed approval idempotency: %d %v", count, err)
					}
				}
			})
		}
	}
}
