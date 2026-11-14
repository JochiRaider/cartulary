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
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseTimelineCreateReferenceOrder(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-create-reference-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incidentText := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "create-reference-order", "incident_key": "IR-CREATE-REFERENCE-ORDER", "title": "Create reference ordering"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	send := func(path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	payload := func(actions ...map[string]any) map[string]any {
		return map[string]any{"kind": "collection_actions_v1", "actions": actions}
	}
	resolved := func(id uuid.UUID, text string) map[string]any {
		return map[string]any{"op": "add_resolved_ref", "resolved_record_id": id.String(), "raw_text": text}
	}
	for _, mode := range []string{"explicit same field", "cross field", "evidence", "automatic", "initially absent", "initially ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			low, high := uuid.New(), uuid.New()
			low[0] = 0x10
			high[0] = 0xf0
			if mode == "evidence" {
				row := httptestx.RequireSuccessData(t, send("/api/v1/incidents/"+incidentText+"/views/cartulary.view.evidence.v1/rows", map[string]any{"client_txn_id": "order-evidence", "evidence.title": "Attached target"}), http.StatusCreated)["row"].(map[string]any)
				low = uuid.MustParse(row["record_id"].(string))
				high = low
				for i := 15; i >= 0; i-- {
					high[i]++
					if high[i] != 0 {
						break
					}
				}
			}
			if mode != "evidence" {
				if mode == "cross field" {
					entitytest.SeedIdentityRecord(t, h.DB, incident, actor, low, "Low", low.String()+"@example.test", "", "")
				} else {
					entitytest.SeedHostRecord(t, h.DB, incident, actor, low, "Low", low.String(), "", "")
				}
			}
			entitytest.SeedHostRecord(t, h.DB, incident, actor, high, "High", high.String(), "", "")
			token := "Frozen-" + high.String()
			body := map[string]any{"client_txn_id": "reference-order-" + mode, "timeline.host_refs": payload(resolved(high, "High"), resolved(low, "Low"))}
			frozen := mode == "initially absent" || mode == "initially ambiguous"
			competitor := uuid.New()
			switch mode {
			case "cross field":
				body["timeline.host_refs"] = payload(resolved(high, "High"))
				body["timeline.identity_refs"] = payload(resolved(low, "Low"))
			case "evidence":
				body["timeline.host_refs"] = payload(resolved(high, "High"))
				body["timeline.attached_evidence_ids"] = payload(map[string]any{"op": "add_record_ref", "linked_record_id": low.String()})
			case "automatic":
				entitytest.SeedEntityAlias(t, h.DB, incident, actor, high, "host", "High-"+high.String())
				entitytest.SeedEntityAlias(t, h.DB, incident, actor, low, "host", "Low-"+low.String())
				body["timeline.host_refs"] = payload(map[string]any{"op": "add_token", "raw_text": "High-" + high.String()}, map[string]any{"op": "add_token", "raw_text": "Low-" + low.String()})
			case "initially absent", "initially ambiguous":
				body["timeline.host_refs"] = payload(resolved(low, "Low"), map[string]any{"op": "add_token", "raw_text": token})
				if mode == "initially ambiguous" {
					entitytest.SeedHostRecord(t, h.DB, incident, actor, competitor, "Competitor", competitor.String(), "", "")
					entitytest.SeedEntityAlias(t, h.DB, incident, actor, high, "host", token)
					entitytest.SeedEntityAlias(t, h.DB, incident, actor, competitor, "host", token)
				}
			}
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
			path := "/api/v1/incidents/" + incidentText + "/views/cartulary.view.timeline.v2/rows"
			request, err := http.NewRequestWithContext(ctx, http.MethodPost, h.Server.HTTP.URL+path, strings.NewReader(string(encoded)))
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
						t.Error("create reference worker did not drain")
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
					t.Fatalf("create did not wait: status=%d error=%v", status, r.err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-time.After(10 * time.Millisecond):
				}
			}
			var locked uuid.UUID
			if err := holder.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, high).Scan(&locked); err != nil {
				t.Fatalf("create retained a higher reference before the lower target: %v", err)
			}
			if locked != high {
				t.Fatal("reference probe locked wrong identity")
			}
			if mode == "initially absent" {
				if _, err := holder.Exec(ctx, `INSERT INTO entity_aliases(entity_alias_id,incident_id,record_id,entity_type,raw_text,normalized_text,classification,created_by_user_id) VALUES($1,$2,$3,'host',$4::text,$4::citext,'suggestion_only',$5)`, uuid.New(), incident, high, token, actor); err != nil {
					t.Fatal(err)
				}
			} else if mode == "initially ambiguous" {
				if _, err := holder.Exec(ctx, `UPDATE entity_aliases SET deleted_at=now() WHERE record_id=$1`, competitor); err != nil {
					t.Fatal(err)
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
			data := httptestx.RequireSuccessData(t, attempted.response, http.StatusCreated)
			row := data["row"].(map[string]any)
			source := row["record_id"].(string)
			if row["row_version"] != float64(1) {
				t.Fatal("create version differs")
			}
			items := func(field string) []any {
				return row["cells"].(map[string]any)[field].(map[string]any)["value"].(map[string]any)["items"].([]any)
			}
			hosts := items("timeline.host_refs")
			wantHosts := 2
			if mode == "cross field" || mode == "evidence" {
				wantHosts = 1
			}
			if len(hosts) != wantHosts {
				t.Fatal("create Mention membership count differs")
			}
			for i, raw := range hosts {
				item := raw.(map[string]any)
				wantID := high
				wantText := "High"
				wantKind := "resolved_ref"
				if i == 1 {
					wantID = low
					wantText = "Low"
				}
				if mode == "automatic" {
					wantText += "-" + wantID.String()
					if item["auto_resolved"] != true || item["resolution_method"] != "auto_match" {
						t.Fatal("prepared automatic target lost provenance")
					}
				}
				if frozen {
					wantID = low
					wantText = "Low"
					if i == 1 {
						wantKind = "unresolved_mention"
						wantText = token
					}
				}
				if item["item_kind"] != wantKind || item["entity_type"] != "host" || item["raw_text"] != wantText || item["mention_row_version"] != float64(1) {
					t.Fatalf("prepared Mention=%#v", item)
				}
				if wantKind == "resolved_ref" {
					if item["resolved_record_id"] != wantID.String() {
						t.Fatal("prepared Mention target/order differs")
					}
				} else {
					for _, key := range []string{"resolved_record_id", "resolution_method", "auto_resolved", "provenance", "confidence", "matched_alias_text"} {
						if _, ok := item[key]; ok {
							t.Fatalf("initially nonunique token leaked %s", key)
						}
					}
				}
			}
			wantMentions, wantLinks := 2, 2
			if mode == "cross field" {
				refs := items("timeline.identity_refs")
				if len(refs) != 1 || refs[0].(map[string]any)["resolved_record_id"] != low.String() || refs[0].(map[string]any)["raw_text"] != "Low" {
					t.Fatal("cross-field target differs")
				}
			}
			if mode == "evidence" {
				refs := items("timeline.attached_evidence_ids")
				if len(refs) != 1 || refs[0].(map[string]any)["linked_record_id"] != low.String() {
					t.Fatal("Evidence target differs")
				}
				wantMentions = 1
			}
			if frozen {
				wantLinks = 1
			}
			var exact bool
			if err := h.DB.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM entity_mentions WHERE source_record_id=$1)=$2 AND (SELECT count(*) FROM record_links WHERE src_record_id=$1 AND deleted_at IS NULL)=$3 AND (SELECT count(*) FROM record_revisions WHERE record_id=$1)=1`, source, wantMentions, wantLinks).Scan(&exact); err != nil || !exact {
				t.Fatalf("create retained effects=%t error=%v", exact, err)
			}
			before := snapshotTimelineMutationState(t, h.DB, incidentText, source)
			replay := httptestx.RequireSuccessData(t, send(path, body), http.StatusOK)
			if !reflect.DeepEqual(replay, data) {
				t.Fatal("create replay payload differs")
			}
			if after := snapshotTimelineMutationState(t, h.DB, incidentText, source); after != before {
				t.Fatal("create replay wrote duplicate effects")
			}
		})
	}
}
