package entities_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func exerciseMentionTargetAdmission(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "mention-target-admission", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	data := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "admission-incident", "incident_key": "IR-MENTION-ADMIT", "title": "Mention target admission"})
	incident := uuid.MustParse(data["incident_id"].(string))
	for _, kind := range []string{"host", "identity"} {
		for _, change := range []string{"deleted", "merged", "reresolved_merged"} {
			t.Run(kind+"/"+change, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				target, survivor, source, mention := uuid.New(), uuid.New(), uuid.New(), uuid.New()
				if change == "reresolved_merged" && target.String() < survivor.String() {
					target, survivor = survivor, target
				}
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
				timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
				if change == "reresolved_merged" {
					entitytest.SeedResolvedMention(t, h.DB, actor, mention, source, survivor, field, kind, "Original Target")
					linktest.SeedRecordLink(t, h.DB, incident, actor, uuid.New(), source, survivor, linkType, "manual", nil)
				} else {
					entitytest.SeedMention(t, h.DB, actor, mention, source, field, kind, "Original Target", "unresolved", nil, nil)
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
				body := entitytest.MentionResolveRoutePayload(1, "admit-"+kind+"-"+change, "resolve_item", &target, nil)
				url := h.Server.HTTP.URL + "/api/v1/entity-mentions/" + mention.String() + "/resolve"

				payload, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
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
							t.Error("Mention resolution target worker did not drain")
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
						t.Fatalf("Mention resolution did not wait for current target: status=%d error=%v", status, r.err)
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
				if change == "reresolved_merged" {
					probe, err := h.Pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					_, lockErr := probe.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, survivor)
					rollbackErr := probe.Rollback(ctx)
					var pgErr *pgconn.PgError
					if !errors.As(lockErr, &pgErr) || pgErr.Code != "55P03" {
						t.Fatalf("previous target was not protected before requested-target wait: %v", lockErr)
					}
					if rollbackErr != nil {
						t.Fatal(rollbackErr)
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
				httptestx.RequireErrorEnvelope(t, attempted.response, http.StatusNotFound, "resolved_record_not_found")
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
					t.Fatal("rejected Mention resolution changed durable state or consumed a receipt")
				}

				do := func() *http.Response {
					return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, url, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
				}
				created := httptestx.RequireSuccessData(t, do(), http.StatusOK)
				row := created["entity_mention"].(map[string]any)
				for key, want := range map[string]any{"entity_mention_id": mention.String(), "source_record_id": source.String(), "source_field_key": field, "entity_type": kind, "raw_text": "Original Target", "normalized_text": "original target", "resolution_status": "resolved", "resolved_record_id": target.String(), "row_version": float64(2), "resolved_by_user_id": actor.String(), "resolution_method": "explicit_resolve_route"} {
					if row[key] != want {
						t.Fatalf("retry %s=%#v want %#v", key, row[key], want)
					}
				}
				if got := created["source_record"].(map[string]any); got["record_id"] != source.String() || got["row_version"] != float64(2) {
					t.Fatalf("retry source=%#v", got)
				}
				var exact bool
				if err := h.Pool.QueryRow(ctx, `SELECT m.source_record_id=$2 AND m.entity_type=$3 AND m.source_field_key=$4 AND m.raw_text='Original Target' AND m.normalized_text='original target' AND m.resolution_status='resolved' AND m.resolved_record_id=$5 AND m.resolved_by_user_id=$6 AND m.resolution_method='explicit_resolve_route' AND m.row_version=2 AND m.origin_kind='manual_entry' AND m.origin_locator='entities-test-support' AND m.ordinal=1 AND m.created_by_user_id=$6 AND r.row_version=2 AND l.dst_record_id=$5 AND l.link_type=$7 AND l.provenance='manual' AND l.confidence IS NULL AND l.created_by_user_id=$6 AND l.owner_user_id=$6 AND l.created_at=m.resolved_at AND l.decided_at=m.resolved_at AND l.deleted_at IS NULL AND l.deleted_by_user_id IS NULL FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id JOIN record_links l ON l.src_record_id=r.record_id WHERE m.entity_mention_id=$1 AND l.deleted_at IS NULL`, mention, source, kind, field, target, actor, linkType).Scan(&exact); err != nil || !exact {
					t.Fatalf("retained mention/source/Link=%t error=%v", exact, err)
				}
				for _, query := range []string{`SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=2`, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND deleted_at IS NULL`} {
					var count int
					if err := h.Pool.QueryRow(ctx, query, source).Scan(&count); err != nil || count != 1 {
						t.Fatalf("retry count=%d error=%v query=%s", count, err, query)
					}
				}
				if change == "reresolved_merged" {
					var count int
					if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND dst_record_id=$2 AND deleted_at IS NOT NULL AND deleted_by_user_id=$3`, source, survivor, actor).Scan(&count); err != nil || count != 1 {
						t.Fatalf("replaced prior Link count=%d error=%v", count, err)
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
					t.Fatal("committed mention replay changed after target invalidation")
				}
			})
		}
	}
}
