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
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseTimelineSupersedeReferenceOrder(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-supersede-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incidentText := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "supersede-order", "incident_key": "IR-SUPERSEDE-ORDER", "title": "Supersede endpoint ordering"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	for _, mode := range []string{"lower replacement", "lower source", "replacement superseded during wait"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			low, high := uuid.New(), uuid.New()
			low[0], high[0] = 0x10, 0xf0
			source, replacement := high, low
			if mode == "lower source" {
				source, replacement = low, high
			}
			for _, id := range []uuid.UUID{source, replacement} {
				timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, id)
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
			body := map[string]any{"client_txn_id": "supersede-order-" + mode, "base_row_version": 1, "replacement_record_id": replacement.String(), "reason": "Ordered replacement"}
			encoded, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			path := h.Server.HTTP.URL + "/api/v1/records/" + source.String() + "/supersede"
			request, err := http.NewRequestWithContext(ctx, http.MethodPost, path, strings.NewReader(string(encoded)))
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
						t.Error("supersede worker did not drain")
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
					t.Fatalf("supersede bypassed lower endpoint wait: status=%d error=%v", status, r.err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-time.After(10 * time.Millisecond):
				}
			}
			var locked uuid.UUID
			if err := holder.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, high).Scan(&locked); err != nil {
				t.Fatalf("lower endpoint waiter retained higher envelope: %v", err)
			}
			if locked != high {
				t.Fatal("wrong higher envelope")
			}
			for _, id := range []uuid.UUID{low, high} {
				if err := holder.QueryRow(ctx, `SELECT record_id FROM timeline_events WHERE record_id=$1 FOR UPDATE NOWAIT`, id).Scan(&locked); err != nil {
					t.Fatalf("endpoint waiter retained typed source: %v", err)
				}
				if locked != id {
					t.Fatal("wrong typed source")
				}
			}
			reject := mode == "replacement superseded during wait"
			if reject {
				// This is a coherent source fixture change, not a production history write.
				if _, err := holder.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, replacement); err != nil {
					t.Fatal(err)
				}
				if _, err := holder.Exec(ctx, `UPDATE timeline_events SET row_version=2,capture_state='superseded',superseded_at=now(),superseded_by_user_id=$2 WHERE record_id=$1`, replacement, actor); err != nil {
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
			send := func() *http.Response {
				return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
			}
			if reject {
				check := func(resp *http.Response) {
					result := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "illegal_transition")
					httptestx.RequireErrorDetail(t, result, "from_status", "reviewed")
					httptestx.RequireErrorDetail(t, result, "to_status", "superseded")
					httptestx.RequireErrorDetailStrings(t, result, "violated_guards", []string{"replacement_must_not_be_superseded"})
				}
				check(attempted.response)
				var unchanged bool
				if err := h.DB.QueryRowContext(ctx, `SELECT r.row_version=1 AND e.row_version=1 AND e.capture_state='reviewed'
      AND NOT EXISTS(SELECT 1 FROM record_revisions WHERE record_id=$1)
      AND NOT EXISTS(SELECT 1 FROM record_links WHERE src_record_id=$1 OR dst_record_id=$1)
      AND NOT EXISTS(SELECT 1 FROM timeline_grid_projection WHERE record_id=$1)
      AND NOT EXISTS(SELECT 1 FROM change_sets WHERE client_txn_id=$2)
      AND NOT EXISTS(SELECT 1 FROM route_idempotency WHERE client_txn_id=$2)
      FROM records r JOIN timeline_events e USING(record_id) WHERE r.record_id=$1`, source, body["client_txn_id"]).Scan(&unchanged); err != nil {
					t.Fatal(err)
				}
				if !unchanged {
					t.Fatal("rejected supersession changed source or retained effects")
				}
				before := snapshotTimelineMutationState(t, h.DB, incidentText, source.String())
				check(send())
				if after := snapshotTimelineMutationState(t, h.DB, incidentText, source.String()); after != before {
					t.Fatal("repeated rejected supersession changed retained state")
				}
				return
			}
			data := httptestx.RequireSuccessData(t, attempted.response, http.StatusOK)
			if data["row_version"] != float64(2) || data["capture_state"] != "superseded" || data["replacement_record_id"] != replacement.String() || data["reason"] != "Ordered replacement" {
				t.Fatalf("unexpected supersession payload: %#v", data)
			}
			var coherent bool
			if err := h.DB.QueryRowContext(ctx, `SELECT r.row_version=2 AND e.row_version=2 AND e.capture_state='superseded' AND e.raw_activity_text='record-support-source-row'
    AND p.row_version=2 AND p.capture_state='superseded' AND p.replacement_record_id=$2
    AND (SELECT row_version=1 AND capture_state='reviewed' FROM timeline_events WHERE record_id=$2)
    AND (SELECT row_version=1 FROM records WHERE record_id=$2)
    AND (SELECT count(*) FROM record_links WHERE src_record_id=$2 AND dst_record_id=$1 AND link_type='supersedes' AND field_key IS NULL AND deleted_at IS NULL)=1
    FROM records r JOIN timeline_events e USING(record_id) JOIN timeline_grid_projection p USING(record_id) WHERE r.record_id=$1`, source, replacement).Scan(&coherent); err != nil {
				t.Fatal(err)
			}
			if !coherent {
				t.Fatal("accepted supersession did not preserve exact source/replacement/projection/link state")
			}
			requireMutationRecorded(t, h.DB, data["change_set_id"].(string), source.String(), actor.String(), "timeline.records.supersede", body["client_txn_id"].(string), 2, 1)
			asserttest.RequireSupersedeCoupledChangeSet(t, asserttest.SQLDatabase(h.DB), data["change_set_id"].(string), source.String(), replacement.String(), 2)
			before := snapshotTimelineMutationState(t, h.DB, incidentText, source.String())
			replay := httptestx.RequireSuccessData(t, send(), http.StatusOK)
			if !reflect.DeepEqual(data, replay) {
				t.Fatal("supersession replay changed committed payload")
			}
			if after := snapshotTimelineMutationState(t, h.DB, incidentText, source.String()); after != before {
				t.Fatal("supersession replay changed retained state")
			}
		})
	}
}
