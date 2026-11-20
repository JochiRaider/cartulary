package revisions_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func requireTaskRestoreReferences(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	for _, variant := range []string{"party", "decision", "both", "neither"} {
		t.Run("Task restore references/"+variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			var party, decision *uuid.UUID
			type reference struct {
				id   uuid.UUID
				kind string
			}
			refs := []reference{}
			for _, kind := range []string{"party", "decision"} {
				if variant != kind && variant != "both" {
					continue
				}
				id := uuid.New()
				envelopetest.SeedRecordEnvelope(t, h.DB, incident, actor, id, kind)
				if kind == "party" {
					party = &id
					mustExec(t, h.DB, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Requester','organization')`, id, incident)
				} else {
					decision = &id
					mustExec(t, h.DB, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Decision','approved',$3,'containment',now(),'Rationale')`, id, incident, actor)
				}
				refs = append(refs, reference{id, kind})
			}
			task := uuid.New()
			envelopetest.SeedRecordEnvelope(t, h.DB, incident, actor, task, "task_request")
			mustExec(t, h.DB, `INSERT INTO task_requests(record_id,incident_id,title,status,owner_user_id,priority,task_kind,requester_party_id,decision_record_id,requester_party_text) VALUES($1,$2,'Task restore fixture','open',$3,'normal','collection',$4,$5,'Preserved requester')`, task, incident, actor, party, decision)
			deleted := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, task, map[string]any{"base_row_version": 1, "client_txn_id": "task-reference-delete-" + variant}), http.StatusOK)["data"].(map[string]any)
			if deleted["row_version"] != float64(2) || deleted["deleted"] != true {
				t.Fatalf("Task delete=%#v", deleted)
			}
			body := map[string]any{"base_row_version": 2, "client_txn_id": "task-reference-restore-" + variant}
			state := func() string {
				t.Helper()
				values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident)}
				for _, table := range []string{"records", "task_requests", "decisions", "parties", "record_links", "task_request_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
					var value string
					if err := h.DB.QueryRowContext(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
						t.Fatal(err)
					}
					values[table] = value
				}
				encoded, err := json.Marshal(values)
				if err != nil {
					t.Fatal(err)
				}
				return string(encoded)
			}
			for _, ref := range refs {
				before := state()
				sourceTable := "parties"
				if ref.kind == "decision" {
					sourceTable = "decisions"
				}
				for _, lockTable := range []string{"records", sourceTable} {
					lock, err := h.Pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer lock.Rollback(context.Background())
					if _, err := lock.Exec(ctx, "SELECT record_id FROM "+lockTable+" WHERE record_id=$1 FOR UPDATE", ref.id); err != nil {
						t.Fatal(err)
					}
					locked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, h, login, task, body), http.StatusConflict, "record_locked")["error"].(map[string]any)
					details := locked["details"].(map[string]any)
					if locked["retryable"] != true || len(details) != 1 || details["record_id"] != task.String() {
						t.Fatalf("Task dependency contention details=%#v", locked)
					}
					if err := lock.Rollback(ctx); err != nil {
						t.Fatal(err)
					}
					if state() != before {
						t.Fatal("dependency contention changed retained Task or publication state")
					}

				}
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, ref.id, map[string]any{"base_row_version": 1, "client_txn_id": "task-dependency-delete-" + variant + ref.kind}), http.StatusOK)
				before = state()
				blocked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, h, login, task, body), http.StatusConflict, "record_restore_blocked")["error"].(map[string]any)
				details := blocked["details"].(map[string]any)
				if blocked["retryable"] != false || len(details) != 2 || details["record_id"] != task.String() || details["reason_code"] != "inactive_task_reference" {
					t.Fatalf("Task reference blocker=%#v", blocked)
				}
				if state() != before {
					t.Fatal("inactive reference changed retained Task or publication state")
				}
				httptestx.RequireSuccessEnvelope(t, restoreRecord(t, h, login, ref.id, map[string]any{"base_row_version": 2, "client_txn_id": "task-dependency-repair-" + variant + ref.kind}), http.StatusOK)
			}
			restored := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, h, login, task, body), http.StatusOK)["data"].(map[string]any)
			if restored["row_version"] != float64(3) || restored["deleted"] != false || restored["deleted_at"] != nil || restored["deleted_by_user_id"] != nil {
				t.Fatalf("Task restore=%#v", restored)
			}
			var exact bool
			if err := h.DB.QueryRowContext(ctx, `SELECT t.title='Task restore fixture' AND p.title='Task restore fixture' AND t.status='open' AND p.status='open'
AND t.owner_user_id=$2 AND t.priority='normal' AND t.task_kind='collection'
AND t.requester_party_text='Preserved requester' AND p.requester_party_text='Preserved requester'
AND t.requester_party_id IS NOT DISTINCT FROM $3::uuid AND p.requester_party_id IS NOT DISTINCT FROM $3::uuid
AND t.decision_record_id IS NOT DISTINCT FROM $4::uuid AND p.decision_record_id IS NOT DISTINCT FROM $4::uuid
AND r.row_version=3 AND p.row_version=3 AND r.deleted_at IS NULL
FROM task_requests t JOIN records r USING(record_id) JOIN task_request_grid_projection p USING(record_id) WHERE t.record_id=$1`, task, actor, party, decision).Scan(&exact); err != nil || !exact {
				t.Fatalf("restored Task source/projection exact=%v error=%v", exact, err)
			}
			if countRecordMutations(t, h.DB, task, "restore") != 1 || countRecordRevisions(t, h.DB, task) != 2 {
				t.Fatal("Task restore history does not advance exactly once")
			}
			requireSingleRecordChangeIntent(t, h.DB, restored["change_set_id"].(string), task, "invalidate")
		})
	}
}

// requireTaskRestoreLifecycleBlock covers REQ-02-109a and AC-606: restoring a
// Task whose row violates a REQ-02-103 guard is blocked with
// record_restore_blocked / task_lifecycle_inconsistent and commits nothing.
// Each case breaks exactly one guard; once repaired, the restore succeeds.
func requireTaskRestoreLifecycleBlock(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	for _, guard := range []struct {
		name   string
		breaks string
	}{
		{"blocked_requires_reason", `UPDATE task_requests SET status='blocked', blocked_reason=NULL WHERE record_id=$1`},
		{"blocked_requires_reason_blank", `UPDATE task_requests SET status='blocked', blocked_reason='   ' WHERE record_id=$1`},
		{"non_blocked_clears_reason", `UPDATE task_requests SET blocked_reason='Waiting on vendor' WHERE record_id=$1`},
		{"done_requires_completed_at", `UPDATE task_requests SET status='done', completed_at=NULL WHERE record_id=$1`},
		{"completed_at_not_before_created_at", `UPDATE task_requests SET status='done', completed_at=created_at - interval '1 day' WHERE record_id=$1`},
		{"non_done_clears_completed_at", `UPDATE task_requests SET completed_at=created_at WHERE record_id=$1`},
		{"active_requires_owner", `UPDATE task_requests SET owner_user_id=NULL WHERE record_id=$1`},
	} {
		t.Run("Task restore lifecycle block/"+guard.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			task := uuid.New()
			envelopetest.SeedRecordEnvelope(t, h.DB, incident, actor, task, "task_request")
			mustExec(t, h.DB, `INSERT INTO task_requests(record_id,incident_id,title,status,owner_user_id,priority,task_kind) VALUES($1,$2,'Lifecycle restore fixture','open',$3,'normal','collection')`, task, incident, actor)
			httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, task, map[string]any{"base_row_version": 1, "client_txn_id": "task-lifecycle-delete-" + guard.name}), http.StatusOK)
			mustExec(t, h.DB, guard.breaks, task)
			state := func() string {
				t.Helper()
				values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident)}
				for _, table := range []string{"records", "task_requests", "task_request_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "route_idempotency"} {
					var value string
					if err := h.DB.QueryRowContext(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
						t.Fatal(err)
					}
					values[table] = value
				}
				encoded, err := json.Marshal(values)
				if err != nil {
					t.Fatal(err)
				}
				return string(encoded)
			}
			body := map[string]any{"base_row_version": 2, "client_txn_id": "task-lifecycle-restore-" + guard.name}
			before := state()
			blocked := httptestx.RequireErrorEnvelope(t, restoreRecord(t, h, login, task, body), http.StatusConflict, "record_restore_blocked")["error"].(map[string]any)
			details := blocked["details"].(map[string]any)
			if blocked["retryable"] != false || len(details) != 2 || details["record_id"] != task.String() || details["reason_code"] != "task_lifecycle_inconsistent" {
				t.Fatalf("Task lifecycle blocker=%#v", blocked)
			}
			if state() != before {
				t.Fatal("blocked Task restore changed source, history, projection or publication state")
			}
			mustExec(t, h.DB, `UPDATE task_requests SET status='open', blocked_reason=NULL, completed_at=NULL, owner_user_id=$2 WHERE record_id=$1`, task, actor)
			restored := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, h, login, task, body), http.StatusOK)["data"].(map[string]any)
			if restored["row_version"] != float64(3) || restored["deleted"] != false {
				t.Fatalf("repaired Task restore=%#v", restored)
			}
		})
	}
}
