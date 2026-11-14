package timeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseTimelineExplicitTargetAdmission(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-target-admission", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	data := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "admission-incident", "incident_key": "IR-TIMELINE-ADMIT", "title": "Timeline target admission"})
	incident := uuid.MustParse(data["incident_id"].(string))
	for _, mode := range []string{"create", "patch"} {
		for _, kind := range []string{"host", "identity"} {
			for _, change := range []string{"deleted", "merged"} {
				t.Run(mode+"/"+kind+"/"+change, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
					defer cancel()
					target, survivor, source := uuid.New(), uuid.New(), uuid.New()
					table, stateField, field, linkType := "hosts", "host_state", timelinetest.FieldHostRefs, "observed_on_host"
					for _, id := range []uuid.UUID{target, survivor} {
						if kind == "host" {
							entitytest.SeedHostRecord(t, h.DB, incident, actor, id, "Target", id.String(), "", "")
						} else {
							entitytest.SeedIdentityRecord(t, h.DB, incident, actor, id, "Target", id.String()+"@example.test", "", "")
						}
					}
					if kind == "identity" {
						table, stateField, field, linkType = "identities", "identity_state", timelinetest.FieldIdentityRefs, "observed_as_identity"
					}
					if mode == "patch" {
						timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
					}
					state := func() string {
						t.Helper()
						values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident)}
						for _, name := range []string{"records", "hosts", "identities", "record_links", "timeline_events", "entity_mentions", "timeline_grid_projection", "host_grid_projection", "identity_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
							var value string
							if err := h.DB.QueryRowContext(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+name+" t").Scan(&value); err != nil {
								t.Fatal(err)
							}
							values[name] = value
						}
						encoded, err := json.Marshal(values)
						if err != nil {
							t.Fatal(err)
						}
						return string(encoded)
					}
					invalidate := func(tx pgx.Tx) {
						t.Helper()
						if _, err := tx.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, target); err != nil {
							t.Fatal(err)
						}
						if _, err := tx.Exec(ctx, "UPDATE "+table+" SET row_version=2 WHERE record_id=$1", target); err != nil {
							t.Fatal(err)
						}
						if change == "deleted" {
							if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at='2026-11-14T10:00:00Z',deleted_by_user_id=$2 WHERE record_id=$1`, target, actor); err != nil {
								t.Fatal(err)
							}
						} else if _, err := tx.Exec(ctx, "UPDATE "+table+" SET "+stateField+"='merged',merged_into_record_id=$2 WHERE record_id=$1", target, survivor); err != nil {
							t.Fatal(err)
						}
					}
					before := state()
					holder, err := h.Pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer holder.Rollback(context.Background())
					var pid int32
					if err := holder.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
						t.Fatal(err)
					}
					invalidate(holder)
					action := timelinetest.CollectionActions(timelinetest.AddResolvedRefAction("Original Target", target))
					body := map[string]any{"client_txn_id": "admit-" + kind + "-" + change + "-" + mode, "timeline.activity_synopsis_text": "Explicit target admission", field: action}
					method, wantStatus, wantVersion := http.MethodPost, http.StatusCreated, int64(1)
					url := h.Server.HTTP.URL + "/api/v1/incidents/" + incident.String() + "/views/" + timeline.TimelineViewSchemaID + "/rows"
					if mode == "patch" {
						method, wantStatus, wantVersion = http.MethodPatch, http.StatusOK, 2
						url = h.Server.HTTP.URL + "/api/v1/records/" + source.String()
						body = timelinetest.TimelineCollectionPatchPayload(field, 1, "admit-"+kind+"-"+change+"-"+mode, action)
					}

					payload, err := json.Marshal(body)
					if err != nil {
						t.Fatal(err)
					}
					request, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(payload)))
					if err != nil {
						t.Fatal(err)
					}
					request.Header.Set("Content-Type", "application/json")
					request.AddCookie(login.Session)
					request.AddCookie(login.CSRF)
					request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
					type outcome struct {
						response *http.Response
						err      error
					}
					done := make(chan outcome, 1)
					go func() { r, e := http.DefaultClient.Do(request); done <- outcome{r, e} }()
					consumed := false
					defer func() {
						cancel()
						if !consumed {
							select {
							case r := <-done:
								if r.response != nil {
									_ = r.response.Body.Close()
								}
							case <-time.After(5 * time.Second):
								t.Error("Explicit Timeline reference target worker did not drain")
							}
						}
					}()
					for {
						var waiting bool
						if err := h.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
							t.Fatal(err)
						}
						if waiting {
							break
						}
						select {
						case r := <-done:
							consumed = true
							status := 0
							if r.response != nil {
								status = r.response.StatusCode
								_ = r.response.Body.Close()
							}
							t.Fatalf("Explicit Timeline reference did not wait for current target: status=%d error=%v", status, r.err)
						case <-ctx.Done():
							t.Fatal(ctx.Err())
						case <-time.After(10 * time.Millisecond):
						}
					}
					if err := holder.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					var attempted outcome
					select {
					case attempted = <-done:
						consumed = true
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					}
					if attempted.err != nil {
						t.Fatal(attempted.err)
					}
					failure := httptestx.RequireErrorEnvelope(t, attempted.response, http.StatusBadRequest, "invalid_mutation_payload")["error"].(map[string]any)
					details := failure["details"].(map[string]any)
					if details["reason_code"] != "invalid_value" || (mode == "patch" && details["field"] != "changes") {
						t.Fatalf("target rejection=%#v", failure)
					}
					// Reverse only this admission probe's synthetic lifecycle transition.
					repair, err := h.Pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer repair.Rollback(context.Background())
					if _, err := repair.Exec(ctx, `UPDATE records SET row_version=1,deleted_at=NULL,deleted_by_user_id=NULL WHERE record_id=$1`, target); err != nil {
						t.Fatal(err)
					}
					if _, err := repair.Exec(ctx, "UPDATE "+table+" SET row_version=1,"+stateField+"='canonical',merged_into_record_id=NULL WHERE record_id=$1", target); err != nil {
						t.Fatal(err)
					}
					if err := repair.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					if after := state(); after != before {
						t.Fatal("rejected Explicit Timeline reference changed durable state or consumed a receipt")
					}

					do := func() *http.Response {
						return httptestx.DoJSON(t, http.DefaultClient, method, url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
					}
					created := httptestx.RequireSuccessData(t, do(), wantStatus)
					row := created["row"].(map[string]any)
					gotSource := uuid.MustParse(row["record_id"].(string))
					if mode == "patch" && gotSource != source {
						t.Fatalf("patch changed source identity: %s", gotSource)
					}
					source = gotSource
					if row["row_version"] != float64(wantVersion) {
						t.Fatalf("retry row=%#v", row)
					}
					item := requireSingleCollectionItem(t, row, field)
					for key, want := range map[string]any{"item_kind": "resolved_ref", "raw_text": "Original Target", "resolved_record_id": target.String(), "resolution_method": "add_resolved_ref", "provenance": "manual", "confidence": nil} {
						if got, present := item[key]; !present || got != want {
							t.Fatalf("retry %s=%#v want %#v", key, item[key], want)
						}
					}
					var exact bool
					if err := h.Pool.QueryRow(ctx, `SELECT m.entity_type=$2 AND m.source_field_key=$3 AND m.raw_text='Original Target' AND m.normalized_text='Original Target' AND m.resolution_status='resolved' AND m.resolved_record_id=$4 AND m.resolved_by_user_id=$5 AND m.resolution_method='add_resolved_ref' AND m.row_version=1 AND m.origin_kind='manual_entry' AND m.ordinal=1 AND m.created_by_user_id=$5 AND m.created_at=m.resolved_at AND r.row_version=$7 AND l.dst_record_id=$4 AND l.link_type=$6 AND l.provenance='manual' AND l.confidence IS NULL AND l.created_by_user_id=$5 AND l.owner_user_id=$5 AND l.created_at=m.resolved_at AND l.decided_at=m.resolved_at AND l.deleted_at IS NULL AND l.deleted_by_user_id IS NULL FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id JOIN record_links l ON l.src_record_id=r.record_id WHERE m.source_record_id=$1`, source, kind, field, target, actor, linkType, wantVersion).Scan(&exact); err != nil || !exact {
						t.Fatalf("retained explicit mention/source/Link=%t error=%v", exact, err)
					}
					var revisionCount int
					if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=$2`, source, wantVersion).Scan(&revisionCount); err != nil || revisionCount != 1 {
						t.Fatalf("retry revisions=%d error=%v", revisionCount, err)
					}
					for _, table := range []string{"entity_mentions", "record_links"} {
						column := "source_record_id"
						if table == "record_links" {
							column = "src_record_id"
						}
						var count int
						if err := h.Pool.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE "+column+"=$1", source).Scan(&count); err != nil || count != 1 {
							t.Fatalf("retry %s count=%d error=%v", table, count, err)
						}
					}
					again, err := h.Pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer again.Rollback(context.Background())
					invalidate(again)
					if err := again.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					committed := state()
					replay := httptestx.RequireSuccessData(t, do(), http.StatusOK)
					if !reflect.DeepEqual(replay, created) || state() != committed {
						t.Fatal("committed Timeline replay changed after target invalidation")
					}
				})
			}
		}
	}
}
