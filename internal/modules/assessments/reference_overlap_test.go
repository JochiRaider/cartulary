package assessments_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseAssessmentReferenceOverlap(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	subjects := []uuid.UUID{uuid.New(), uuid.New()}
	bodies := make([]map[string]any, 2)
	url := h.Server.HTTP.URL + "/api/v1/incidents/" + incident.String() + "/views/" + assessments.AssessmentsViewSchemaID + "/rows"
	for i, subject := range subjects {
		entitytest.SeedHostRecord(t, h.DB, incident, actor, subject, "Overlap subject", subject.String(), "", "")
		bodies[i] = validAssessmentBody(uuid.NewString(), subject, "host", "confirmed")
		bodies[i]["assessment.support_refs"] = map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_record_ref", "linked_record_id": subjects[1-i].String()}}}
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
	if _, err := holder.Exec(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, actor); err != nil {
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
				t.Error("Assessment writers did not drain")
				return
			}
		}
	}()
	launch := func(i int) {
		payload, err := json.Marshal(bodies[i])
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
				t.Fatalf("writer %d escaped assessor contention: status=%d error=%v", r.index, status, r.err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	launch(0)
	firstPID := waitFor(`SELECT coalesce(min(pid),0) FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid))`, holderPID)
	launch(1)
	// A partial reference set lets both requests reach the assessor while holding
	// opposite subjects. Complete admission serializes the second before that point.
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
		data := httptestx.RequireSuccessData(t, response, http.StatusCreated)
		row := data["row"].(map[string]any)
		ids[i] = uuid.MustParse(row["record_id"].(string))
		cells := row["cells"].(map[string]any)
		if row["row_version"] != float64(1) {
			t.Fatalf("writer %d row=%#v", i, row)
		}
		requireCellValue(t, cells, "assessment.subject_ref", subjects[i].String())
		requireCellValue(t, cells, "assessment.subject_type", "host")
		requireCellValue(t, cells, "assessment.assessment_state", "confirmed")
		requireCellValue(t, cells, "assessment.rationale", "Valid rationale.")
		requireCellValue(t, cells, "assessment.assessor", actor.String())
		requireCellValue(t, cells, "assessment.supporting_link_count", float64(1))
		var exact bool
		if err := h.Pool.QueryRow(ctx, `SELECT a.subject_record_id=$2 AND a.subject_type='host' AND a.assessment_state='confirmed' AND a.rationale='Valid rationale.' AND a.assessor_user_id=$4 AND r.row_version=1 AND r.deleted_at IS NULL AND p.subject_ref=$2 AND l.dst_record_id=$3 AND l.link_type='supported_by' AND l.field_key IS NULL AND l.provenance='manual' AND l.confidence IS NULL AND l.owner_user_id=$4 AND l.created_by_user_id=$4 AND l.created_at=r.created_at AND l.decided_at=r.created_at AND l.deleted_at IS NULL AND l.deleted_by_user_id IS NULL FROM assessments a JOIN records r USING(record_id) JOIN assessment_grid_projection p USING(record_id) JOIN record_links l ON l.src_record_id=a.record_id WHERE a.record_id=$1`, ids[i], subjects[i], subjects[1-i], actor).Scan(&exact); err != nil || !exact {
			t.Fatalf("writer %d retained Assessment/Link=%t error=%v", i, exact, err)
		}
		var revisions int
		if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=1`, ids[i]).Scan(&revisions); err != nil || revisions != 1 {
			t.Fatalf("writer %d revisions=%d error=%v", i, revisions, err)
		}
	}
	if ids[0] == ids[1] {
		t.Fatal("distinct Assessment writers shared an identity")
	}
	var count int
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=ANY($1::uuid[]) AND deleted_at IS NULL`, ids).Scan(&count); err != nil || count != 2 {
		t.Fatalf("committed support links=%d error=%v", count, err)
	}
}
