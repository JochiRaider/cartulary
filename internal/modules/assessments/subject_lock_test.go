package assessments_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseAssessmentSubjectAdmission(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	for _, kind := range []string{"host", "identity"} {
		for _, change := range []string{"deleted", "merged", "support_deleted"} {
			t.Run(kind+"/"+change, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				subject, survivor := uuid.New(), uuid.New()
				for _, id := range []uuid.UUID{subject, survivor} {
					if kind == "host" {
						entitytest.SeedHostRecord(t, h.DB, incident, actor, id, "Subject host", id.String(), "", "")
					} else {
						entitytest.SeedIdentityRecord(t, h.DB, incident, actor, id, "Subject identity", id.String()+"@example.test", "", "")
					}
				}
				table, stateField := "hosts", "host_state"
				if kind == "identity" {
					table, stateField = "identities", "identity_state"
				}
				state := func() string {
					t.Helper()
					values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident)}
					for _, name := range []string{"records", "hosts", "identities", "assessments", "record_links", "assessment_grid_projection", "host_grid_projection", "identity_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
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
					if _, err := tx.Exec(ctx, `UPDATE records SET row_version=2 WHERE record_id=$1`, subject); err != nil {
						t.Fatal(err)
					}
					if _, err := tx.Exec(ctx, "UPDATE "+table+" SET row_version=2 WHERE record_id=$1", subject); err != nil {
						t.Fatal(err)
					}
					if change != "merged" {
						if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at='2026-11-14T10:00:00Z',deleted_by_user_id=$2 WHERE record_id=$1`, subject, actor); err != nil {
							t.Fatal(err)
						}
					} else if _, err := tx.Exec(ctx, "UPDATE "+table+" SET "+stateField+"='merged',merged_into_record_id=$2 WHERE record_id=$1", subject, survivor); err != nil {
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
				subjectRef, invalidField, supportCount := subject, "assessment.subject_ref", float64(0)
				if change == "support_deleted" {
					subjectRef, invalidField, supportCount = survivor, "assessment.support_refs", 1
				}
				body := validAssessmentBody("subject-admission-"+kind+"-"+change, subjectRef, kind, "confirmed")
				if change == "support_deleted" {
					body["assessment.support_refs"] = map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{
						{"op": "add_record_ref", "linked_record_id": subject.String()},
						{"op": "add_record_ref", "linked_record_id": subject.String()},
					}}
				}
				payload, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				request, err := http.NewRequestWithContext(ctx, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incident.String()+"/views/"+assessments.AssessmentsViewSchemaID+"/rows", strings.NewReader(string(payload)))
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
							t.Error("Assessment subject worker did not drain")
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
						t.Fatalf("Assessment did not wait for current subject: status=%d error=%v", status, r.err)
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
				if details["field"] != invalidField || details["reason_code"] != "invalid_value" {
					t.Fatalf("subject failure=%#v", failure)
				}
				// Reverse only this admission probe's synthetic lifecycle transition.
				repair, err := h.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer repair.Rollback(context.Background())
				if _, err := repair.Exec(ctx, `UPDATE records SET row_version=1,deleted_at=NULL,deleted_by_user_id=NULL WHERE record_id=$1`, subject); err != nil {
					t.Fatal(err)
				}
				if _, err := repair.Exec(ctx, "UPDATE "+table+" SET row_version=1,"+stateField+"='canonical',merged_into_record_id=NULL WHERE record_id=$1", subject); err != nil {
					t.Fatal(err)
				}
				if err := repair.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				if after := state(); after != before {
					t.Fatal("rejected Assessment changed durable state or consumed a receipt")
				}
				created := httptestx.RequireSuccessData(t, postAssessment(t, h, login, incident, body), http.StatusCreated)
				row := created["row"].(map[string]any)
				cells := row["cells"].(map[string]any)
				if row["row_version"] != float64(1) {
					t.Fatalf("retry row=%#v", row)
				}
				requireCellValue(t, cells, "assessment.subject_ref", subjectRef.String())
				requireCellValue(t, cells, "assessment.subject_type", kind)
				requireCellValue(t, cells, "assessment.assessment_state", "confirmed")
				requireCellValue(t, cells, "assessment.rationale", "Valid rationale.")
				requireCellValue(t, cells, "assessment.assessor", actor.String())
				requireCellValue(t, cells, "assessment.supporting_link_count", supportCount)
				id := uuid.MustParse(row["record_id"].(string))
				var exact bool
				if err := h.Pool.QueryRow(ctx, `SELECT a.subject_record_id=$2 AND a.subject_type=$3 AND a.assessment_state='confirmed' AND a.rationale='Valid rationale.' AND a.assessor_user_id=$4 AND a.confidence_score IS NULL AND r.row_version=1 AND r.deleted_at IS NULL AND p.subject_ref=$2 FROM assessments a JOIN records r USING(record_id) JOIN assessment_grid_projection p USING(record_id) WHERE a.record_id=$1`, id, subjectRef, kind, actor).Scan(&exact); err != nil || !exact {
					t.Fatalf("retry source/projection=%t error=%v", exact, err)
				}
				var linkCount int
				if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND dst_record_id=$2 AND link_type='supported_by' AND deleted_at IS NULL`, id, subject).Scan(&linkCount); err != nil || float64(linkCount) != supportCount {
					t.Fatalf("retry support links=%d error=%v", linkCount, err)
				}
				for _, query := range []string{`SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=1`, `SELECT count(*) FROM change_set_mutations WHERE target_kind='assessment' AND target_id=$1::uuid::text AND operation_kind='create' AND coalesce(before_value,'null'::jsonb)='null'::jsonb`} {
					var count int
					if err := h.Pool.QueryRow(ctx, query, id).Scan(&count); err != nil || count != 1 {
						t.Fatalf("retry history count=%d error=%v query=%s", count, err, query)
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
				replayed := httptestx.RequireSuccessData(t, postAssessment(t, h, login, incident, body), http.StatusOK)
				if !reflect.DeepEqual(replayed, created) || state() != committed {
					t.Fatalf("committed replay changed after %s subject", change)
				}
			})
		}
	}
}
