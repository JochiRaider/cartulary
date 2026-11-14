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
)

func exerciseTimelineAutoTargetAdmission(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-auto-admission", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	data := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "admission-incident", "incident_key": "IR-AUTO-ADMIT", "title": "Auto target admission"})
	incident := uuid.MustParse(data["incident_id"].(string))
	for _, mode := range []string{"create", "patch"} {
		for _, kind := range []string{"host", "identity"} {
			for _, change := range []string{"deleted", "merged", "alias_removed", "alias_renamed", "ambiguous", "replacement", "stable"} {
				t.Run(mode+"/"+kind+"/"+change, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
					defer cancel()
					target, survivor, source, competingAlias := uuid.New(), uuid.New(), uuid.New(), uuid.New()
					token := "Candidate-" + target.String()
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
					entitytest.SeedEntityAlias(t, h.DB, incident, actor, target, kind, token)
					if mode == "patch" {
						timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
					}
					state := func() string {
						t.Helper()
						values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident)}
						for _, name := range []string{"entity_aliases", "records", "hosts", "identities", "record_links", "timeline_events", "entity_mentions", "timeline_grid_projection", "host_grid_projection", "identity_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
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

					holder, err := h.Pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer holder.Rollback(context.Background())
					var pid int32
					if err := holder.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
						t.Fatal(err)
					}
					if _, err := holder.Exec(ctx, `SELECT record_id FROM records WHERE record_id=ANY($1::uuid[]) ORDER BY record_id FOR NO KEY UPDATE`, []uuid.UUID{target, survivor}); err != nil {
						t.Fatal(err)
					}
					if _, err := holder.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, target); err != nil {
						t.Fatal(err)
					}
					if _, err := holder.Exec(ctx, "UPDATE "+table+" SET row_version=2 WHERE record_id=$1", target); err != nil {
						t.Fatal(err)
					}
					switch change {
					case "deleted":
						if _, err := holder.Exec(ctx, `UPDATE records SET deleted_at='2026-11-14T10:00:00Z',deleted_by_user_id=$2 WHERE record_id=$1`, target, actor); err != nil {
							t.Fatal(err)
						}
					case "merged":
						if _, err := holder.Exec(ctx, "UPDATE "+table+" SET "+stateField+"='merged',merged_into_record_id=$2 WHERE record_id=$1", target, survivor); err != nil {
							t.Fatal(err)
						}
					case "alias_removed", "replacement":
						if _, err := holder.Exec(ctx, `UPDATE entity_aliases SET deleted_at='2026-11-14T10:00:00Z' WHERE record_id=$1`, target); err != nil {
							t.Fatal(err)
						}
					case "alias_renamed":
						if _, err := holder.Exec(ctx, `UPDATE entity_aliases SET raw_text=$2::text,normalized_text=$2::citext WHERE record_id=$1`, target, "Different-"+target.String()); err != nil {
							t.Fatal(err)
						}
					}
					if change == "ambiguous" || change == "replacement" {
						if _, err := holder.Exec(ctx, `INSERT INTO entity_aliases(entity_alias_id,incident_id,record_id,entity_type,raw_text,normalized_text,classification,created_by_user_id) VALUES($1,$2,$3,$4,$5::text,$5::citext,'suggestion_only',$6)`, competingAlias, incident, survivor, kind, token, actor); err != nil {
							t.Fatal(err)
						}
					}
					action := timelinetest.CollectionActions(timelinetest.AddTokenAction(token))
					body := map[string]any{"client_txn_id": "auto-" + mode + "-" + kind + "-" + change, "timeline.activity_synopsis_text": "Current alias admission", field: action}
					method, wantStatus, wantVersion := http.MethodPost, http.StatusCreated, int64(1)
					url := h.Server.HTTP.URL + "/api/v1/incidents/" + incident.String() + "/views/" + timeline.TimelineViewSchemaID + "/rows"
					if mode == "patch" {
						method, wantStatus, wantVersion = http.MethodPatch, http.StatusOK, 2
						url = h.Server.HTTP.URL + "/api/v1/records/" + source.String()
						body = timelinetest.TimelineCollectionPatchPayload(field, 1, "auto-"+mode+"-"+kind+"-"+change, action)
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
								t.Error("Automatic Timeline reference target worker did not drain")
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
							t.Fatalf("Automatic Timeline reference did not wait for current target: status=%d error=%v", status, r.err)
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

					created := httptestx.RequireSuccessData(t, attempted.response, wantStatus)
					row := created["row"].(map[string]any)
					gotSource := uuid.MustParse(row["record_id"].(string))
					if mode == "patch" && gotSource != source {
						t.Fatalf("patch changed source identity: %s", gotSource)
					}
					source = gotSource
					if row["row_version"] != float64(wantVersion) {
						t.Fatalf("row version=%#v want %d", row["row_version"], wantVersion)
					}
					resolved := change == "stable"
					wantKind, wantState, wantLinks := "unresolved_mention", "unresolved", 0
					if resolved {
						wantKind, wantState, wantLinks = "resolved_ref", "resolved", 1
					}
					item := requireSingleCollectionItem(t, row, field)
					if item["item_kind"] != wantKind || item["raw_text"] != token || item["entity_type"] != kind || item["mention_row_version"] != float64(1) {
						t.Fatalf("current candidate result=%#v", item)
					}
					if resolved {
						for key, want := range map[string]any{"resolved_record_id": target.String(), "resolution_method": "auto_match", "auto_resolved": true, "provenance": "auto_match", "confidence": float64(100)} {
							if got, present := item[key]; !present || got != want {
								t.Fatalf("resolved %s=%#v want %#v", key, got, want)
							}
						}
					} else {
						for _, key := range []string{"resolved_record_id", "resolution_method", "auto_resolved", "provenance", "confidence", "matched_alias_text"} {
							if _, present := item[key]; present {
								t.Fatalf("unresolved token leaked %s: %#v", key, item)
							}
						}
					}
					var exact bool
					if err := h.Pool.QueryRow(ctx, `SELECT m.entity_type=$2 AND m.source_field_key=$3 AND m.raw_text=$4 AND m.normalized_text=$4 AND m.resolution_status=$5 AND m.row_version=1 AND m.origin_kind='manual_entry' AND m.ordinal=1 AND m.created_by_user_id=$6 AND r.row_version=$7 AND p.has_unresolved_mentions=(NOT $8::boolean) AND CASE WHEN $8 THEN m.resolved_record_id=$9 AND m.resolved_by_user_id=$6 AND m.resolved_at=m.created_at AND m.resolution_method='auto_match' ELSE m.resolved_record_id IS NULL AND m.resolved_by_user_id IS NULL AND m.resolved_at IS NULL AND m.resolution_method IS NULL END FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id JOIN timeline_grid_projection p ON p.record_id=r.record_id WHERE m.source_record_id=$1`, source, kind, field, token, wantState, actor, wantVersion, resolved, target).Scan(&exact); err != nil || !exact {
						t.Fatalf("retained mention/source/projection=%t error=%v", exact, err)
					}
					var count int
					if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM entity_mentions WHERE source_record_id=$1`, source).Scan(&count); err != nil || count != 1 {
						t.Fatalf("mention count=%d error=%v", count, err)
					}
					if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=$1`, source).Scan(&count); err != nil || count != wantLinks {
						t.Fatalf("Link count=%d error=%v", count, err)
					}
					if resolved {
						if err := h.Pool.QueryRow(ctx, `SELECT l.dst_record_id=$2 AND l.link_type=$3 AND l.provenance='auto_match' AND l.confidence=100 AND l.owner_user_id=$4 AND l.created_by_user_id=$4 AND l.created_at=m.resolved_at AND l.decided_at=m.resolved_at AND l.deleted_at IS NULL AND l.deleted_by_user_id IS NULL FROM record_links l JOIN entity_mentions m ON m.source_record_id=l.src_record_id WHERE l.src_record_id=$1`, source, target, linkType, actor).Scan(&exact); err != nil || !exact {
							t.Fatalf("auto-match Link=%t error=%v", exact, err)
						}
					}
					if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=$2`, source, wantVersion).Scan(&count); err != nil || count != 1 {
						t.Fatalf("source revision count=%d error=%v", count, err)
					}
					// Restore only controlled candidate fixtures. An already committed unresolved
					// result must replay unchanged even though the alias is now eligible again.
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
					if _, err := repair.Exec(ctx, `UPDATE entity_aliases SET raw_text=$2::text,normalized_text=$2::citext,deleted_at=NULL WHERE record_id=$1`, target, token); err != nil {
						t.Fatal(err)
					}
					if _, err := repair.Exec(ctx, `DELETE FROM entity_aliases WHERE entity_alias_id=$1`, competingAlias); err != nil {
						t.Fatal(err)
					}
					if err := repair.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					committed := state()
					replay := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, method, url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusOK)
					if !reflect.DeepEqual(replay, created) || state() != committed {
						t.Fatal("committed auto-resolution replay changed after candidate restoration")
					}
				})
			}
		}
	}
}
