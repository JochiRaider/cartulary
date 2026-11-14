package tasksdecisions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseTaskDecisionWriterOverlap(t *testing.T, h *appsupport.ServerHarness, incident, actor uuid.UUID, createURL, mode string, options []func(*http.Request)) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	decision := uuid.New()
	envelopetest.SeedRecordEnvelope(t, h.DB, incident, actor, decision, "decision")
	if _, err := h.DB.ExecContext(ctx, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Shared Decision','approved',$3,'containment',now(),'Retained rationale')`, decision, incident, actor); err != nil {
		t.Fatal(err)
	}
	method, wantStatus, wantVersion := http.MethodPost, http.StatusCreated, float64(1)
	urls := []string{createURL, createURL}
	bodies := make([]map[string]any, 2)
	for i := range bodies {
		title := fmt.Sprintf("Concurrent %s Task %d", mode, i)
		bodies[i] = map[string]any{"client_txn_id": uuid.NewString(), "task.title": title, "task.task_kind": "collection", "task.requester_party_text": "Preserved requester", "task.decision_record_id": decision.String()}
		if mode == "patch" {
			created := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, createURL, map[string]any{"client_txn_id": uuid.NewString(), "task.title": title, "task.task_kind": "collection", "task.requester_party_text": "Preserved requester"}, options...), http.StatusCreated)
			method, wantStatus, wantVersion = http.MethodPatch, http.StatusOK, 2
			urls[i] = h.Server.HTTP.URL + "/api/v1/records/" + created["row"].(map[string]any)["record_id"].(string)
			bodies[i] = map[string]any{"client_txn_id": uuid.NewString(), "view_schema_id": "cartulary.view.task_requests.v1", "base_row_version": 1, "changes": []map[string]any{{"field_key": "task.decision_record_id", "value": decision.String()}}}
		}
	}
	holder, err := h.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Rollback(context.Background())
	var holderPID int32
	if err := holder.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
		t.Fatal(err)
	}
	if _, err := holder.Exec(ctx, `SELECT record_id FROM decisions WHERE record_id=$1 FOR NO KEY UPDATE`, decision); err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		index    int
		response *http.Response
		err      error
	}
	done := make(chan outcome, 2)
	started, consumed := 0, 0
	defer func() {
		cancel()
		for consumed < started {
			select {
			case r := <-done:
				consumed++
				if r.response != nil {
					_ = r.response.Body.Close()
				}
			case <-time.After(5 * time.Second):
				t.Error("Task Decision writers did not drain")
				return
			}
		}
	}()
	launch := func(i int) {
		payload, err := json.Marshal(bodies[i])
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequestWithContext(ctx, method, urls[i], strings.NewReader(string(payload)))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		for _, option := range options {
			option(request)
		}
		started++
		go func() { r, e := http.DefaultClient.Do(request); done <- outcome{i, r, e} }()
	}
	waitFor := func(query string, args ...any) int32 {
		for {
			var pid int32
			if err := h.Pool.QueryRow(ctx, query, args...).Scan(&pid); err != nil {
				t.Fatal(err)
			}
			if pid != 0 {
				return pid
			}
			select {
			case r := <-done:
				consumed++
				status := 0
				if r.response != nil {
					status = r.response.StatusCode
					_ = r.response.Body.Close()
				}
				t.Fatalf("writer %d escaped typed-source contention: status=%d error=%v", r.index, status, r.err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	launch(0)
	firstPID := waitFor(`SELECT coalesce(min(pid),0) FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid))`, holderPID)
	launch(1)
	// The second writer either shares the first's envelope lock and waits at the
	// typed source, or is correctly serialized behind the first at its envelope.
	waitFor(`SELECT coalesce(min(pid),0) FROM pg_stat_activity WHERE datname=current_database() AND pid<>$2 AND ($1::integer=ANY(pg_blocking_pids(pid)) OR $2::integer=ANY(pg_blocking_pids(pid)))`, holderPID, firstPID)
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	responses := make([]*http.Response, 2)
	for consumed < started {
		select {
		case r := <-done:
			consumed++
			if r.err != nil {
				t.Errorf("writer %d HTTP error=%v", r.index, r.err)
			}
			responses[r.index] = r.response
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	for _, response := range responses {
		if response != nil {
			defer response.Body.Close()
		}
	}
	if t.Failed() {
		return
	}
	ids := make([]uuid.UUID, 2)
	for i, response := range responses {
		data := httptestx.RequireSuccessData(t, response, wantStatus)
		row := data["row"].(map[string]any)
		ids[i] = uuid.MustParse(row["record_id"].(string))
		cells := row["cells"].(map[string]any)
		title := fmt.Sprintf("Concurrent %s Task %d", mode, i)
		if row["row_version"] != wantVersion || cells["task.title"].(map[string]any)["value"] != title || cells["task.decision_record_id"].(map[string]any)["value"] != decision.String() || cells["task.requester_party_text"].(map[string]any)["value"] != "Preserved requester" {
			t.Fatalf("writer %d row=%#v", i, row)
		}
		var exact bool
		if err := h.Pool.QueryRow(ctx, `SELECT t.title=$2 AND t.task_kind='collection' AND t.requester_party_text='Preserved requester' AND t.decision_record_id=$3 AND r.row_version=$4 AND r.deleted_at IS NULL AND l.dst_record_id=$3 AND l.field_key='task.decision_record_id' AND l.link_type='references_record' AND l.provenance='manual' AND l.confidence IS NULL AND l.owner_user_id=$5 AND l.created_by_user_id=$5 AND l.created_at=r.updated_at AND l.decided_at=r.updated_at AND l.deleted_at IS NULL AND l.deleted_by_user_id IS NULL FROM task_requests t JOIN records r USING(record_id) JOIN record_links l ON l.src_record_id=t.record_id WHERE t.record_id=$1`, ids[i], title, decision, int64(wantVersion), actor).Scan(&exact); err != nil || !exact {
			t.Fatalf("writer %d retained Task/Link=%t error=%v", i, exact, err)
		}
		var revisions int
		if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=$2`, ids[i], int64(wantVersion)).Scan(&revisions); err != nil || revisions != 1 {
			t.Fatalf("writer %d revisions=%d error=%v", i, revisions, err)
		}
	}
	if ids[0] == ids[1] {
		t.Fatal("distinct Task writers shared an identity")
	}
	var count int
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=ANY($1::uuid[]) AND dst_record_id=$2 AND deleted_at IS NULL`, ids, decision).Scan(&count); err != nil || count != 2 {
		t.Fatalf("committed Task Decision links=%d error=%v", count, err)
	}
}
