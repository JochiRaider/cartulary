package timeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseTimelineExistingReferenceOrder(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "existing-reference-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incidentText := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "existing-reference-order", "incident_key": "IR-EXISTING-REFERENCE-ORDER", "title": "Existing reference order"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	send := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	create := func(txn string, values map[string]any) map[string]any {
		values["client_txn_id"] = txn
		return httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incidentText+"/views/cartulary.view.timeline.v2/rows", values), http.StatusCreated)["row"].(map[string]any)
	}
	serial := byte(0)
	targets := func(kind string) (uuid.UUID, uuid.UUID) {
		serial++
		low, high := uuid.UUID{14: serial, 15: 1}, uuid.UUID{14: serial, 15: 2}
		for _, id := range []uuid.UUID{low, high} {
			if kind == "host" {
				entitytest.SeedHostRecord(t, h.DB, incident, actor, id, "Target", id.String(), "", "")
			} else {
				entitytest.SeedIdentityRecord(t, h.DB, incident, actor, id, "Target", id.String()+"@example.test", "", "")
			}
		}
		return low, high
	}
	// The lower endpoint holder must be free to complete work on every later
	// endpoint while a public request waits. No timing assumption proves ordering.
	blocked := func(t *testing.T, low uuid.UUID, free []uuid.UUID, method, path string, body map[string]any, change func(context.Context, pgx.Tx), afterRelease func(context.Context)) map[string]any {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()
		holder, err := h.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer holder.Rollback(context.Background())
		var pid int32
		if err := holder.QueryRow(ctx, `SELECT pg_backend_pid() FROM records WHERE record_id=$1 FOR NO KEY UPDATE`, low).Scan(&pid); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequestWithContext(ctx, method, h.Server.HTTP.URL+path, strings.NewReader(string(encoded)))
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
			_ = holder.Rollback(context.Background())
			if !consumed {
				select {
				case r := <-done:
					if r.response != nil {
						_ = r.response.Body.Close()
					}
				case <-time.After(5 * time.Second):
					t.Error("existing-reference worker did not drain")
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
				t.Fatalf("write bypassed target wait: status=%d error=%v", status, r.err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-time.After(10 * time.Millisecond):
			}
		}
		for _, id := range free {
			var got uuid.UUID
			if err := holder.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, id).Scan(&got); err != nil {
				t.Fatalf("writer retained later endpoint %s: %v", id, err)
			}
			if got != id {
				t.Fatal("endpoint probe identity differs")
			}
		}
		if change != nil {
			change(ctx, holder)
		}
		if err := holder.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if afterRelease != nil {
			afterRelease(ctx)
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
		return httptestx.RequireSuccessData(t, attempted.response, http.StatusOK)
	}
	for _, kind := range []string{"host", "identity"} {
		for _, family := range []string{"patch", "direct"} {
			for _, op := range []string{"resolve_item", "dismiss_item", "revert_to_unresolved"} {
				t.Run(kind+"/"+family+"/"+op, func(t *testing.T) {
					low, high := targets(kind)
					field := "timeline." + kind + "_refs"
					txn := kind + "-" + family + "-" + op
					row := create(txn+"-source", map[string]any{field: payload(map[string]any{"op": "add_resolved_ref", "raw_text": "Original", "resolved_record_id": low.String()})})
					source := uuid.MustParse(row["record_id"].(string))
					item := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)[0].(map[string]any)
					mention := uuid.MustParse(item["entity_mention_id"].(string))
					action := map[string]any{"op": op, "item_ref": item["item_ref"]}
					if op == "resolve_item" {
						action["resolved_record_id"] = high.String()
					}
					method, path := http.MethodPatch, "/api/v1/records/"+source.String()
					body := map[string]any{"client_txn_id": txn, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(action)}}}
					if family == "direct" {
						method, path = http.MethodPost, "/api/v1/entity-mentions/"+mention.String()+"/resolve"
						body = map[string]any{"client_txn_id": txn, "base_mention_row_version": 1, "action": op}
						if op == "resolve_item" {
							body["resolved_record_id"] = high.String()
						}
					}
					data := blocked(t, low, []uuid.UUID{source, high}, method, path, body, nil, nil)
					state, links := "unresolved", 0
					if op == "resolve_item" {
						state, links = "resolved", 1
					}
					if op == "dismiss_item" {
						state = "dismissed"
					}
					var exact bool
					if err := h.DB.QueryRowContext(t.Context(), `SELECT m.resolution_status=$3 AND m.row_version=2 AND m.raw_text='Original' AND m.entity_type=$4 AND r.row_version=2 AND e.row_version=2 AND p.row_version=2 AND CASE WHEN $3='resolved' THEN m.resolved_record_id=$5 ELSE m.resolved_record_id IS NULL END AND (SELECT count(*) FROM record_links WHERE src_record_id=$1 AND deleted_at IS NULL)=$6 AND NOT EXISTS(SELECT 1 FROM record_links WHERE src_record_id=$1 AND dst_record_id=$7 AND deleted_at IS NULL) FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id JOIN timeline_events e ON e.record_id=r.record_id JOIN timeline_grid_projection p ON p.record_id=r.record_id WHERE m.source_record_id=$1 AND m.entity_mention_id=$2`, source, mention, state, kind, high, links, low).Scan(&exact); err != nil || !exact {
						t.Fatalf("protected lifecycle effects=%t error=%v", exact, err)
					}
					before := snapshotTimelineMutationState(t, h.DB, incidentText, source.String())
					replay := httptestx.RequireSuccessData(t, send(method, path, body), http.StatusOK)
					if !reflect.DeepEqual(replay, data) {
						t.Fatal("protected lifecycle replay changed payload")
					}
					if after := snapshotTimelineMutationState(t, h.DB, incidentText, source.String()); after != before {
						t.Fatal("protected lifecycle replay wrote effects")
					}
				})
			}
		}
	}
	for _, mode := range []string{"batch opposite targets", "batch creates", "batch mixed", "batch initially absent", "batch initially ambiguous", "patch initially absent", "patch initially ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			low, high := targets("host")
			token := "Frozen-" + high.String()
			lowToken := "Low-" + low.String()
			highToken := "High-" + high.String()
			entitytest.SeedEntityAlias(t, h.DB, incident, actor, low, "host", lowToken)
			entitytest.SeedEntityAlias(t, h.DB, incident, actor, high, "host", highToken)
			ambiguous := strings.HasSuffix(mode, "ambiguous")
			frozen := strings.Contains(mode, "initially")
			competitor := uuid.New()
			if ambiguous {
				entitytest.SeedHostRecord(t, h.DB, incident, actor, competitor, "Competitor", competitor.String(), "", "")
				entitytest.SeedEntityAlias(t, h.DB, incident, actor, high, "host", token)
				entitytest.SeedEntityAlias(t, h.DB, incident, actor, competitor, "host", token)
			}
			first := create(mode+"-first", map[string]any{})
			second := create(mode+"-second", map[string]any{})
			source1, source2 := uuid.MustParse(first["record_id"].(string)), uuid.MustParse(second["record_id"].(string))
			field := "timeline.host_refs"
			text := highToken + "\n" + lowToken
			if frozen {
				text = lowToken + "\n" + token
			}
			method, path := http.MethodPost, "/api/v1/incidents/"+incidentText+"/views/cartulary.view.timeline.v2/clipboard-paste"
			body := map[string]any{"client_txn_id": mode, "view_schema_id": "cartulary.view.timeline.v2", "clipboard_text": text, "format": "tsv", "start_field_key": field, "columns": []string{field}, "targets": []map[string]any{{"kind": "record", "record_id": source1.String(), "base_row_version": 1}, {"kind": "record", "record_id": source2.String(), "base_row_version": 1}}}
			createdFirst := mode == "batch creates" || mode == "batch mixed"
			createdSecond := mode == "batch creates"
			if createdFirst {
				body["targets"].([]map[string]any)[0] = map[string]any{"kind": "create"}
			}
			if createdSecond {
				body["targets"].([]map[string]any)[1] = map[string]any{"kind": "create"}
			}
			patch := strings.HasPrefix(mode, "patch")
			if patch {
				method, path = http.MethodPatch, "/api/v1/records/"+source1.String()
				body = map[string]any{"client_txn_id": mode, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "action_payload": payload(map[string]any{"op": "add_token", "raw_text": lowToken}, map[string]any{"op": "add_token", "raw_text": token})}}}
			}
			var change func(context.Context, pgx.Tx)
			if frozen {
				change = func(ctx context.Context, tx pgx.Tx) {
					if ambiguous {
						if _, err := tx.Exec(ctx, `UPDATE entity_aliases SET deleted_at=now() WHERE record_id=$1`, competitor); err != nil {
							t.Fatal(err)
						}
					} else {
						if _, err := tx.Exec(ctx, `INSERT INTO entity_aliases(entity_alias_id,incident_id,record_id,entity_type,raw_text,normalized_text,classification,created_by_user_id) VALUES($1,$2,$3,'host',$4::text,$4::citext,'suggestion_only',$5)`, uuid.New(), incident, high, token, actor); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			data := blocked(t, low, []uuid.UUID{source1, source2, high}, method, path, body, change, nil)
			var items []any
			if patch {
				row := data["row"].(map[string]any)
				if row["record_id"] != source1.String() || row["row_version"] != float64(2) {
					t.Fatal("patch row identity/version differs")
				}
				items = row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)
			} else {
				rows := data["rows"].([]any)
				if len(rows) != 2 || len(data["conflicts"].([]any)) != 0 {
					t.Fatal("batch result geometry differs")
				}
				for i, raw := range rows {
					row := raw.(map[string]any)
					want := []uuid.UUID{source1, source2}[i]
					wantVersion := float64(2)
					if (i == 0 && createdFirst) || (i == 1 && createdSecond) {
						parsed, err := uuid.Parse(row["record_id"].(string))
						if err != nil || parsed == uuid.Nil {
							t.Fatal("created batch row identity missing")
						}
						want, wantVersion = parsed, 1
						if i == 0 {
							source1 = parsed
						} else {
							source2 = parsed
						}
					}
					if row["record_id"] != want.String() || row["row_version"] != wantVersion {
						t.Fatal("batch row order/identity/version differs")
					}
					values := row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)
					if len(values) != 1 {
						t.Fatal("batch Mention membership differs")
					}
					items = append(items, values[0])
				}
			}
			if len(items) != 2 {
				t.Fatal("prepared token count differs")
			}
			for i, raw := range items {
				item := raw.(map[string]any)
				wantID, wantText := high, highToken
				if i == 1 {
					wantID, wantText = low, lowToken
				}
				if frozen {
					wantID, wantText = low, lowToken
					if i == 1 {
						wantText = token
					}
				}
				if item["raw_text"] != wantText || item["entity_type"] != "host" || item["mention_row_version"] != float64(1) {
					t.Fatal("prepared token text/type/version differs")
				}
				if frozen && i == 1 {
					if item["item_kind"] != "unresolved_mention" {
						t.Fatal("initially nonunique token was promoted")
					}
					for _, key := range []string{"resolved_record_id", "resolution_method", "auto_resolved", "provenance", "confidence", "matched_alias_text"} {
						if _, ok := item[key]; ok {
							t.Fatalf("unresolved token leaked %s", key)
						}
					}
				} else if item["item_kind"] != "resolved_ref" || item["resolved_record_id"] != wantID.String() || item["resolution_method"] != "auto_match" || item["auto_resolved"] != true {
					t.Fatal("prepared resolved target/order differs")
				}
			}
			before1, before2 := snapshotTimelineMutationState(t, h.DB, incidentText, source1.String()), snapshotTimelineMutationState(t, h.DB, incidentText, source2.String())
			replay := httptestx.RequireSuccessData(t, send(method, path, body), http.StatusOK)
			if !reflect.DeepEqual(replay, data) {
				t.Fatal("prepared request replay differs")
			}
			if snapshotTimelineMutationState(t, h.DB, incidentText, source1.String()) != before1 || snapshotTimelineMutationState(t, h.DB, incidentText, source2.String()) != before2 {
				t.Fatal("prepared request replay wrote effects")
			}
		})
	}
	for _, family := range []string{"patch", "direct"} {
		t.Run("reprepare/"+family, func(t *testing.T) {
			low, high := targets("host")
			tiny := low
			tiny[15] = 0
			entitytest.SeedHostRecord(t, h.DB, incident, actor, tiny, "Next previous", tiny.String(), "", "")
			source, mention, link := uuid.New(), uuid.New(), uuid.New()
			timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
			resolution := "explicit_resolve_route"
			entitytest.SeedMention(t, h.DB, actor, mention, source, "timeline.host_refs", "host", "Original", "resolved", &low, &resolution)
			linktest.SeedRecordLink(t, h.DB, incident, actor, link, source, low, "observed_on_host", "manual", nil)
			later, err := h.Pool.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer later.Rollback(context.Background())
			var laterPID int32
			if err := later.QueryRow(t.Context(), `SELECT pg_backend_pid() FROM records WHERE record_id=$1 FOR NO KEY UPDATE`, tiny).Scan(&laterPID); err != nil {
				t.Fatal(err)
			}
			method, path := http.MethodPatch, "/api/v1/records/"+source.String()
			body := map[string]any{"client_txn_id": "reprepare-" + family, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 2, "changes": []map[string]any{{"field_key": "timeline.host_refs", "action_payload": payload(map[string]any{"op": "resolve_item", "item_ref": "entity_mention:" + mention.String(), "resolved_record_id": high.String()})}}}
			if family == "direct" {
				method, path = http.MethodPost, "/api/v1/entity-mentions/"+mention.String()+"/resolve"
				body = map[string]any{"client_txn_id": "reprepare-" + family, "base_mention_row_version": 2, "action": "resolve_item", "resolved_record_id": high.String()}
			}
			change := func(ctx context.Context, tx pgx.Tx) {
				// Advance the synthetic source fixture coherently while discovery waits.
				// Its first retained production revision is the subsequent accepted write.
				if _, err := tx.Exec(ctx, `SELECT record_id FROM timeline_events WHERE record_id=$1 FOR UPDATE NOWAIT`, source); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, source); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE timeline_events SET row_version=2 WHERE record_id=$1`, source); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE entity_mentions SET row_version=2,resolved_record_id=$2 WHERE entity_mention_id=$1`, mention, tiny); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE record_links SET dst_record_id=$2 WHERE record_link_id=$1`, link, tiny); err != nil {
					t.Fatal(err)
				}
			}
			afterRelease := func(ctx context.Context) {
				for {
					var waiting bool
					if err := h.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, laterPID).Scan(&waiting); err != nil {
						t.Fatal(err)
					}
					if waiting {
						break
					}
					select {
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
				for _, id := range []uuid.UUID{source, low, high} {
					var got uuid.UUID
					if err := later.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, id).Scan(&got); err != nil {
						t.Fatalf("repreparation retained prior endpoint %s: %v", id, err)
					}
					if got != id {
						t.Fatal("repreparation endpoint identity differs")
					}
				}
				if err := later.Commit(ctx); err != nil {
					t.Fatal(err)
				}
			}
			data := blocked(t, low, []uuid.UUID{source, high}, method, path, body, change, afterRelease)
			var exact bool
			if err := h.DB.QueryRowContext(t.Context(), `SELECT m.row_version=3 AND m.resolution_status='resolved' AND m.resolved_record_id=$3 AND m.raw_text='Original' AND r.row_version=3 AND e.row_version=3 AND (SELECT count(*) FROM record_links WHERE src_record_id=$1 AND deleted_at IS NULL)=1 AND EXISTS(SELECT 1 FROM record_links WHERE src_record_id=$1 AND dst_record_id=$3 AND deleted_at IS NULL) AND (SELECT count(*) FROM record_revisions WHERE record_id=$1)=1 FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id JOIN timeline_events e ON e.record_id=r.record_id WHERE m.source_record_id=$1 AND m.entity_mention_id=$2`, source, mention, high).Scan(&exact); err != nil || !exact {
				t.Fatalf("reprepared current effect=%t error=%v", exact, err)
			}
			before := snapshotTimelineMutationState(t, h.DB, incidentText, source.String())
			replay := httptestx.RequireSuccessData(t, send(method, path, body), http.StatusOK)
			if !reflect.DeepEqual(replay, data) || snapshotTimelineMutationState(t, h.DB, incidentText, source.String()) != before {
				t.Fatal("reprepared write replay changed result or effects")
			}
		})
	}
	t.Run("Evidence removal", func(t *testing.T) {
		evidence := httptestx.RequireSuccessData(t, send(http.MethodPost, "/api/v1/incidents/"+incidentText+"/views/cartulary.view.evidence.v1/rows", map[string]any{"client_txn_id": "reference-remove-evidence", "evidence.title": "Remove target"}), http.StatusCreated)["row"].(map[string]any)
		target := uuid.MustParse(evidence["record_id"].(string))
		source := target
		for i := 15; i >= 0; i-- {
			source[i]++
			if source[i] != 0 {
				break
			}
		}
		timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
		path := "/api/v1/records/" + source.String()
		field := "timeline.attached_evidence_ids"
		patch := func(txn string, base int, action map[string]any) map[string]any {
			return map[string]any{"client_txn_id": txn, "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": base, "changes": []map[string]any{{"field_key": field, "action_payload": payload(action)}}}
		}
		attached := httptestx.RequireSuccessData(t, send(http.MethodPatch, path, patch("reference-attach", 1, map[string]any{"op": "add_record_ref", "linked_record_id": target.String()})), http.StatusOK)["row"].(map[string]any)
		items := attached["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)
		if len(items) != 1 || attached["row_version"] != float64(2) {
			t.Fatal("Evidence removal fixture differs")
		}
		body := patch("reference-remove", 2, map[string]any{"op": "remove_record_ref", "item_ref": items[0].(map[string]any)["item_ref"]})
		data := blocked(t, target, []uuid.UUID{source}, http.MethodPatch, path, body, nil, nil)
		row := data["row"].(map[string]any)
		value := row["cells"].(map[string]any)[field].(map[string]any)["value"]
		if row["record_id"] != source.String() || row["row_version"] != float64(3) || !reflect.DeepEqual(value, map[string]any{"kind": "collection_value_v1", "ordered": false, "items": []any{}}) {
			t.Fatal("protected Evidence removal differs")
		}
		var active int
		if err := h.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND deleted_at IS NULL`, source).Scan(&active); err != nil || active != 0 {
			t.Fatalf("removed Evidence links=%d error=%v", active, err)
		}
		before := snapshotTimelineMutationState(t, h.DB, incidentText, source.String())
		replay := httptestx.RequireSuccessData(t, send(http.MethodPatch, path, body), http.StatusOK)
		if !reflect.DeepEqual(replay, data) || snapshotTimelineMutationState(t, h.DB, incidentText, source.String()) != before {
			t.Fatal("Evidence removal replay changed result or effects")
		}
	})
}
