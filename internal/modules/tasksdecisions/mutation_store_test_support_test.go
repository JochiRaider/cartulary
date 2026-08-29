package tasksdecisions_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

type taskState struct {
	RowVersion    int64
	Status        string
	BlockedReason sql.NullString
	CompletedAt   sql.NullTime
	OwnerUserID   sql.NullString
}

func taskSnapshot(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) taskState {
	t.Helper()
	var state taskState
	if err := db.QueryRow(context.Background(), `
SELECT r.row_version, t.status, t.blocked_reason, t.completed_at, t.owner_user_id::text
  FROM task_requests t
  JOIN records r
    ON r.record_id = t.record_id
 WHERE t.record_id = $1
`, recordID).Scan(&state.RowVersion, &state.Status, &state.BlockedReason, &state.CompletedAt, &state.OwnerUserID); err != nil {
		t.Fatalf("query task snapshot: %v", err)
	}
	return state
}

func requireTaskSnapshot(t testing.TB, got taskState, want taskState, context string) {
	t.Helper()
	if got.RowVersion != want.RowVersion ||
		got.Status != want.Status ||
		got.BlockedReason != want.BlockedReason ||
		got.CompletedAt.Valid != want.CompletedAt.Valid ||
		(got.CompletedAt.Valid && !got.CompletedAt.Time.Equal(want.CompletedAt.Time)) ||
		got.OwnerUserID != want.OwnerUserID {
		t.Fatalf("%s changed task snapshot: got %#v want %#v", context, got, want)
	}
}

type decisionState struct {
	RowVersion         int64
	Status             string
	Rationale          string
	IncomingSupersedes int
	OutgoingSupersedes int
}

func decisionSnapshot(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) decisionState {
	t.Helper()
	var state decisionState
	if err := db.QueryRow(context.Background(), `
SELECT r.row_version,
       d.status,
       d.rationale,
       COALESCE(incoming.count, 0)::integer,
       COALESCE(outgoing.count, 0)::integer
  FROM decisions d
  JOIN records r
    ON r.record_id = d.record_id
  LEFT JOIN (
        SELECT dst_record_id, COUNT(*) AS count
          FROM record_links
         WHERE link_type = 'supersedes'
           AND deleted_at IS NULL
         GROUP BY dst_record_id
  ) incoming
    ON incoming.dst_record_id = d.record_id
  LEFT JOIN (
        SELECT src_record_id, COUNT(*) AS count
          FROM record_links
         WHERE link_type = 'supersedes'
           AND deleted_at IS NULL
         GROUP BY src_record_id
  ) outgoing
    ON outgoing.src_record_id = d.record_id
 WHERE d.record_id = $1
`, recordID).Scan(&state.RowVersion, &state.Status, &state.Rationale, &state.IncomingSupersedes, &state.OutgoingSupersedes); err != nil {
		t.Fatalf("query decision snapshot: %v", err)
	}
	return state
}

func requireDecisionSnapshot(t testing.TB, got decisionState, want decisionState, context string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s changed decision snapshot: got %#v want %#v", context, got, want)
	}
}

func mustCreateDecision(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string) uuid.UUID {
	t.Helper()
	return mustCreateDecisionWithCollections(t, owner, actor, incidentID, clientTxnID, status, summary, nil).RecordID
}

func mustCreateEvidence(t testing.TB, pool postgres.DB, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	request, admissionFailure := evidence.AdmitCreateJSON(strings.NewReader(fmt.Sprintf(
		`{"client_txn_id":%q,"evidence.title":%q}`,
		clientTxnID, title,
	)))
	if admissionFailure != nil {
		t.Fatalf("admit evidence %s: %#v", clientTxnID, admissionFailure)
	}
	result, err := appsupport.NewEvidenceMutationOwner(pool, conflicttest.NewCodec("workbook")).Create(
		context.Background(),
		evidence.CreateCommand{
			ActorUserID: actor.ID, IncidentID: incidentID, Admission: request,
			RequestID: "req-" + clientTxnID, Now: testTime(0),
		},
	)
	if err != nil {
		t.Fatalf("create evidence %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func createTaskDecision(
	owner *tasksdecisions.MutationFacade,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request tasksdecisions.CreateRequest,
	requestID string,
	now time.Time,
) (tasksdecisions.MutationResult, error) {
	return owner.Create(context.Background(), tasksdecisions.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incidentID, Request: request,
		RequestHash: tasksdecisions.CreateRequestHash(request), RequestID: requestID,
		RouteKey: "workbook.rows.create", Now: now,
	})
}

func mustCreateDecisionWithCollections(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string, collections map[string]tasksdecisions.CollectionActionPayload) tasksdecisions.MutationResult {
	t.Helper()
	values := map[string]tasksdecisions.FieldValue{
		"decision.summary":       {Text: &summary},
		"decision.decision_type": {Text: stringPtr("containment")},
		"decision.rationale":     {Text: stringPtr("The decision is needed for coordinated response.")},
	}
	if status != "" {
		values["decision.status"] = tasksdecisions.FieldValue{Text: &status}
	}
	request := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}
	result, err := createTaskDecision(owner, actor, incidentID, request, "req-"+clientTxnID, testTime(0))
	if err != nil {
		t.Fatalf("create decision %s: %v", clientTxnID, err)
	}
	return result
}

func mustCreateTask(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, values map[string]tasksdecisions.FieldValue, collections map[string]tasksdecisions.CollectionActionPayload) tasksdecisions.MutationResult {
	t.Helper()
	request := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}
	result, err := createTaskDecision(owner, actor, incidentID, request, "req-"+clientTxnID, testTime(0))
	if err != nil {
		t.Fatalf("create task %s: %v", clientTxnID, err)
	}
	return result
}

func patchRecord(owner *tasksdecisions.MutationFacade, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...tasksdecisions.PatchChange) (tasksdecisions.MutationResult, error) {
	request := tasksdecisions.PatchRequest{
		ViewSchemaID:   viewSchemaID,
		BaseRowVersion: baseRowVersion,
		ClientTxnID:    clientTxnID,
		Changes:        changes,
	}
	return owner.Patch(context.Background(), tasksdecisions.PatchCommand{
		ActorUserID: actor.ID, RecordID: recordID, Request: request,
		RequestHash: tasksdecisions.PatchRequestHash(request), RequestID: "req-" + clientTxnID,
		RouteKey: "workbook.records.patch", ConflictRouteKey: "workbook.records.conflicts.resolve",
		Now: testTime(30 * time.Minute),
	})
}

func mustPatch(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...tasksdecisions.PatchChange) tasksdecisions.MutationResult {
	t.Helper()
	result, err := patchRecord(owner, actor, recordID, viewSchemaID, baseRowVersion, clientTxnID, changes...)
	if err != nil {
		t.Fatalf("patch %s: %v", clientTxnID, err)
	}
	return result
}

func valueChange(fieldKey string, value tasksdecisions.FieldValue) tasksdecisions.PatchChange {
	return tasksdecisions.PatchChange{FieldKey: fieldKey, Value: &value}
}

func collectionChange(fieldKey string, value tasksdecisions.CollectionActionPayload) tasksdecisions.PatchChange {
	return tasksdecisions.PatchChange{FieldKey: fieldKey, Collection: &value}
}

func collectionActions(actions ...tasksdecisions.CollectionAction) tasksdecisions.CollectionActionPayload {
	return tasksdecisions.CollectionActionPayload{Actions: actions}
}

func addOptionalSurfaceRecordRef(recordID uuid.UUID) tasksdecisions.CollectionAction {
	return tasksdecisions.CollectionAction{Op: "add_record_ref", LinkedRecordID: &recordID}
}

func removeRecordRef(recordID uuid.UUID) tasksdecisions.CollectionAction {
	return tasksdecisions.CollectionAction{Op: "remove_record_ref", ItemRef: "record_ref:" + recordID.String()}
}

func countRecords(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM records WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count records: %v", err)
	}
	return count
}

func recordVersion(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) int64 {
	t.Helper()
	var version int64
	if err := db.QueryRow(context.Background(), `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&version); err != nil {
		t.Fatalf("query record version: %v", err)
	}
	return version
}

func countTaskDecisionLinks(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, taskID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND link_type = 'references_record'
   AND field_key = 'task.decision_record_id'
   AND deleted_at IS NULL
`, taskID).Scan(&count); err != nil {
		t.Fatalf("count task decision links: %v", err)
	}
	return count
}

func countReferenceLinks(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sourceID uuid.UUID, fieldKey string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND field_key = $2
   AND deleted_at IS NULL
`, sourceID, fieldKey).Scan(&count); err != nil {
		t.Fatalf("count reference links: %v", err)
	}
	return count
}

func requireManualReferenceLink(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sourceID uuid.UUID, targetID uuid.UUID, fieldKey string, linkType string) {
	t.Helper()
	var provenance string
	var confidence sql.NullInt64
	if err := db.QueryRow(context.Background(), `
SELECT provenance, confidence
  FROM record_links
 WHERE src_record_id = $1
   AND dst_record_id = $2
   AND field_key = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, sourceID, targetID, fieldKey, linkType).Scan(&provenance, &confidence); err != nil {
		t.Fatalf("query manual reference link %s %s -> %s: %v", fieldKey, sourceID, targetID, err)
	}
	if provenance != "manual" || confidence.Valid {
		t.Fatalf("manual link %s must preserve provenance=manual confidence=NULL, got provenance=%q confidence=%#v", fieldKey, provenance, confidence)
	}
}

func countSupersedesLinks(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sourceID uuid.UUID, targetID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
   AND deleted_at IS NULL
`, sourceID, targetID).Scan(&count); err != nil {
		t.Fatalf("count supersedes links: %v", err)
	}
	return count
}

func insertSupersedesLink(t testing.TB, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, actorID uuid.UUID, now time.Time) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'supersedes', NULL, 'manual', NULL, $4, $4, $5, $5)
`, incidentID, sourceID, targetID, actorID, now); err != nil {
		t.Fatalf("insert supersedes link %s -> %s: %v", sourceID, targetID, err)
	}
}

func queryOne(t testing.TB, store *workbook.WorkbookContributionCatalog, incidentID uuid.UUID, viewSchemaID string, fieldKey string, value any, recordID uuid.UUID) map[string]any {
	t.Helper()
	rows, err := workbookroutetest.QueryRows(store, context.Background(), incidentID, viewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": value}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query one %s: %v", fieldKey, err)
	}
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("record %s missing from rows %#v", recordID, rows)
	return nil
}

func rowsContain(rows []map[string]any, recordID uuid.UUID) bool {
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return true
		}
	}
	return false
}

func requireMutationValidation(t testing.TB, err error, field string, reason string) {
	t.Helper()
	var ownerValidation *tasksdecisions.ValidationError
	if errors.As(err, &ownerValidation) {
		if ownerValidation.Field != field || ownerValidation.ReasonCode != reason {
			t.Fatalf("unexpected owner validation error: %#v", ownerValidation)
		}
		return
	}
	t.Fatalf("expected owner validation error, got %v", err)
}

func requireLifecycle(t testing.TB, err error) {
	t.Helper()
	var ownerLifecycleErr *tasksdecisions.LifecycleValidationError
	if errors.As(err, &ownerLifecycleErr) {
		return
	}
	t.Fatalf("expected lifecycle validation error, got %v", err)
}

func requireCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requireCellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got == nil || got == "" {
		t.Fatalf("expected non-empty %s value, got %#v", fieldKey, got)
	}
}

func requireCellNumericValue(t testing.TB, row map[string]any, fieldKey string, want int64) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	switch value := got.(type) {
	case int:
		if int64(value) == want {
			return
		}
	case int32:
		if int64(value) == want {
			return
		}
	case int64:
		if value == want {
			return
		}
	case float64:
		if int64(value) == want {
			return
		}
	}
	t.Fatalf("unexpected %s value: got %#v want %d", fieldKey, got, want)
}

func requireCollectionItemCount(t testing.TB, row map[string]any, fieldKey string, want int) {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" {
		t.Fatalf("expected %s to be collection_value_v1, got %#v", fieldKey, value)
	}
	var got int
	switch items := value["items"].(type) {
	case []any:
		got = len(items)
	case []map[string]any:
		got = len(items)
	default:
		t.Fatalf("unexpected %s items shape: %#v", fieldKey, value["items"])
	}
	if got != want {
		t.Fatalf("unexpected %s item count: got %d want %d items=%#v", fieldKey, got, want, value["items"])
	}
}

func testTime(offset time.Duration) time.Time {
	return time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC).Add(offset)
}

func stringPtr(value string) *string {
	return &value
}
