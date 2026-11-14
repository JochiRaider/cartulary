package timeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseTimelineEnvelopeSourceOrder(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-lock-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := authflowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	data := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "lock-order-incident", "incident_key": "IR-LOCK-ORDER", "title": "Timeline lock ordering"})
	incident, source := uuid.MustParse(data["incident_id"].(string)), uuid.New()
	timelinetest.SeedTimelineRecord(t, h.DB, incident, actor, source)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	holder, err := h.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Rollback(context.Background())
	var pid int32
	if err := holder.QueryRow(ctx, `SELECT pg_backend_pid() FROM records WHERE record_id=$1 FOR UPDATE`, source).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	body := map[string]any{"client_txn_id": "lock-order-patch", "view_schema_id": "cartulary.view.timeline.v2", "base_row_version": 2, "changes": []any{map[string]any{"field_key": "timeline.activity_synopsis_text", "value": "Patch after coordinated write"}}}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+source.String(), strings.NewReader(string(encoded)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	type result struct {
		response *http.Response
		err      error
	}
	done := make(chan result, 1)
	go func() { r, e := http.DefaultClient.Do(request); done <- result{r, e} }()
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
				t.Error("patch worker did not drain")
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
			if r.response != nil {
				defer r.response.Body.Close()
				t.Fatalf("patch bypassed envelope wait: status=%d error=%v", r.response.StatusCode, r.err)
			}
			t.Fatal(r.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	// A coordinating writer already holding the envelope must be able to finish
	// its typed-row work while the ordinary patch waits for that envelope.
	var locked uuid.UUID
	if err := holder.QueryRow(ctx, `SELECT record_id FROM timeline_events WHERE record_id=$1 FOR UPDATE NOWAIT`, source).Scan(&locked); err != nil {
		t.Fatalf("envelope waiter retained typed source lock: %v", err)
	}
	if locked != source {
		t.Fatalf("locked source=%s", locked)
	}
	if _, err := holder.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, source); err != nil {
		t.Fatal(err)
	}
	if _, err := holder.Exec(ctx, `UPDATE timeline_events SET row_version=2,raw_activity_text='Coordinated source text' WHERE record_id=$1`, source); err != nil {
		t.Fatal(err)
	}
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var attempted result
	select {
	case attempted = <-done:
		consumed = true
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if attempted.err != nil {
		t.Fatal(attempted.err)
	}
	httptestx.RequireSuccessEnvelope(t, attempted.response, http.StatusOK)
	var coherent bool
	if err := h.DB.QueryRowContext(ctx, `SELECT r.row_version=3 AND e.row_version=3 AND e.raw_activity_text='Coordinated source text' AND e.activity_synopsis_text='Patch after coordinated write' AND r.updated_by_user_id=$2 AND e.updated_by_user_id=$2 AND (SELECT count(*) FROM record_revisions WHERE record_id=$1)=1 FROM records r JOIN timeline_events e USING(record_id) WHERE r.record_id=$1`, source, actor).Scan(&coherent); err != nil {
		t.Fatal(err)
	}
	if !coherent {
		t.Fatal("patch did not use current protected source and envelope")
	}
}
