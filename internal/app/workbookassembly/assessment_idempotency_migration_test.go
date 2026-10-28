package workbookassembly

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"

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
	payload := assessmentCanonicalMigrationPayload(assessmentMigrationRecordID, assessmentMigrationIncidentID, 7)
	insertAssessmentMigrationRow(t, ctx, db, "assessments.rows.create", "canonical", []byte("hash"), 201, payload)
	stored := assessmentMigrationResponse(t, ctx, db, "canonical")
	assertJSONSemanticEqual(t, stored, payload)
	replayed, err := decodeAssessmentCreateResult(stored, assessmentMigrationIncidentID+":"+assessments.AssessmentsViewSchemaID, 201, []byte("hash"))
	if err != nil {
		t.Fatal(err)
	}
	if replayed.RecordID.String() != assessmentMigrationRecordID || replayed.ChangeSetID.String() != assessmentMigrationChangeSetID || replayed.RowVersion != 7 {
		t.Fatalf("replay identity/version=%#v", replayed)
	}

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
