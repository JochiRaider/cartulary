package workbookassembly

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const (
	assessmentMigrationActorID     = "10000000-0000-4000-8000-000000000001"
	assessmentMigrationIncidentID  = "20000000-0000-4000-8000-000000000001"
	assessmentMigrationRecordID    = "30000000-0000-4000-8000-000000000001"
	assessmentMigrationChangeSetID = "40000000-0000-4000-8000-000000000001"
)

func TestAssessmentCreateIdempotencyBaseline_Integration(t *testing.T) {

	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	ctx := t.Context()
	seedAssessmentMigrationActor(t, ctx, db)
	sum := sha256.Sum256([]byte("canonical-request"))
	requestHash := sum[:]
	payload := assessmentCanonicalMigrationPayload(assessmentMigrationRecordID, assessmentMigrationIncidentID, 7)
	insertAssessmentMigrationRow(t, ctx, db, "assessments.rows.create", "canonical", requestHash, 201, payload)
	stored := assessmentMigrationResponse(t, ctx, db, "canonical")
	assertJSONSemanticEqual(t, stored, payload)
	replayed, err := decodeAssessmentCreateResult(stored, assessmentMigrationIncidentID+":"+assessments.AssessmentsViewSchemaID, 201, requestHash)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Result.RecordID.String() != assessmentMigrationRecordID || replayed.Result.ChangeSetID.String() != assessmentMigrationChangeSetID || replayed.Result.RowVersion != 7 || replayed.DefaultAssessedAt != nil {
		t.Fatalf("replay identity/version=%#v", replayed)
	}
	defaultTime := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	replayed.DefaultAssessedAt = &defaultTime
	encoded, err := encodeAssessmentCreateResult(assessmentMigrationIncidentID+":"+assessments.AssessmentsViewSchemaID, replayed)
	if err != nil {
		t.Fatal(err)
	}
	insertAssessmentMigrationRow(t, ctx, db, "assessments.rows.create", "normalized", requestHash, 201, string(encoded))
	normalized := assessmentMigrationResponse(t, ctx, db, "normalized")
	assertJSONSemanticEqual(t, normalized, string(encoded))
	decoded, err := decodeAssessmentCreateResult(normalized, assessmentMigrationIncidentID+":"+assessments.AssessmentsViewSchemaID, 201, requestHash)
	if err != nil || decoded.DefaultAssessedAt == nil || !decoded.DefaultAssessedAt.Equal(defaultTime) || decoded.Result.RecordID != replayed.Result.RecordID {
		t.Fatalf("normalized receipt round trip: %#v, %v", decoded, err)
	}
	assertJSONSemanticEqual(t, assessmentMigrationResponse(t, ctx, db, "canonical"), payload)

}

func seedAssessmentMigrationActor(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required)
VALUES ($1, 'assessment-migration@example.test', 'Assessment migration', 'hash', false)
`, assessmentMigrationActorID); err != nil {
		t.Fatalf("seed assessment migration actor: %v", err)
	}
}

func insertAssessmentMigrationRow(t *testing.T, ctx context.Context, db *sql.DB, routeKey, clientTxnID string, requestHash []byte, status int, payload string) {
	t.Helper()
	scope := assessmentMigrationIncidentID + ":" + assessments.AssessmentsViewSchemaID
	if routeKey != "assessments.rows.create" {
		scope = "unrelated"
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
`, routeKey, scope, clientTxnID, assessmentMigrationActorID, requestHash, status, payload); err != nil {
		t.Fatalf("seed route idempotency row %q: %v", clientTxnID, err)
	}
}

func assessmentCanonicalMigrationPayload(recordID, incidentID string, rowVersion int64) string {
	return `{"schema_id":"cartulary.assessments.create_result.v1","record_id":"` + recordID + `","change_set_id":"` + assessmentMigrationChangeSetID + `","row_version":` + assessmentMigrationInt(rowVersion) + `,"row":{"record_id":"` + recordID + `","incident_id":"` + incidentID + `","row_version":` + assessmentMigrationInt(rowVersion) + `,"future_field":true}}`
}

func assessmentMigrationInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func assessmentMigrationResponse(t *testing.T, ctx context.Context, db *sql.DB, clientTxnID string) []byte {
	t.Helper()
	var payload []byte
	if err := db.QueryRowContext(ctx, `SELECT response_json FROM route_idempotency WHERE client_txn_id = $1`, clientTxnID).Scan(&payload); err != nil {
		t.Fatalf("load route idempotency row %q: %v", clientTxnID, err)
	}
	return payload
}

func assertJSONSemanticEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotValue any
	var wantValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("decode got JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatalf("decode want JSON: %v", err)
	}
	gotJSON, _ := json.Marshal(gotValue)
	wantJSON, _ := json.Marshal(wantValue)
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Fatalf("JSON mismatch\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}
