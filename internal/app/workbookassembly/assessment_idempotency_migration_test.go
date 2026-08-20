package workbookassembly

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const (
	assessmentMigrationActorID     = "10000000-0000-4000-8000-000000000001"
	assessmentMigrationIncidentID  = "20000000-0000-4000-8000-000000000001"
	assessmentMigrationRecordID    = "30000000-0000-4000-8000-000000000001"
	assessmentMigrationChangeSetID = "40000000-0000-4000-8000-000000000001"
)

func TestAssessmentCreateIdempotencyMigration_Integration(t *testing.T) {
	t.Run("valid legacy conversion replay and deterministic round trip", func(t *testing.T) {
		harness := pgtest.Start(t)
		database := harness.MigrationDatabaseThroughT(t, 34)
		ctx := context.Background()
		seedAssessmentMigrationActor(t, ctx, database.SQL())

		legacy := assessmentLegacyMigrationPayload(assessmentMigrationRecordID, 7)
		canonical := assessmentCanonicalMigrationPayload(assessmentMigrationRecordID, assessmentMigrationIncidentID, 8)
		unrelated := `{"opaque":"unchanged"}`
		insertAssessmentMigrationRow(t, ctx, database.SQL(), "assessments.rows.create", "legacy", []byte("legacy-hash"), 201, legacy)
		insertAssessmentMigrationRow(t, ctx, database.SQL(), "assessments.rows.create", "canonical", []byte("canonical-hash"), 201, canonical)
		insertAssessmentMigrationRow(t, ctx, database.SQL(), "other.rows.create", "unrelated", []byte("unrelated-hash"), 418, unrelated)

		if err := database.ApplyThrough(ctx, 35); err != nil {
			t.Fatalf("apply assessment idempotency cutover: %v", err)
		}
		legacyUp := assessmentMigrationResponse(t, ctx, database.SQL(), "legacy")
		canonicalUp := assessmentMigrationResponse(t, ctx, database.SQL(), "canonical")
		unrelatedUp := assessmentMigrationResponse(t, ctx, database.SQL(), "unrelated")
		assertAssessmentCanonicalMigrationPayload(t, legacyUp, 7)
		assertJSONSemanticEqual(t, canonicalUp, canonical)
		assertJSONSemanticEqual(t, unrelatedUp, unrelated)

		scope := assessmentMigrationIncidentID + ":" + assessments.AssessmentsViewSchemaID
		replayed, err := decodeAssessmentCreateResult(legacyUp, scope, 201, []byte("legacy-hash"))
		if err != nil {
			t.Fatalf("decode migrated replay: %v", err)
		}
		if replayed.RecordID.String() != assessmentMigrationRecordID || replayed.ChangeSetID.String() != assessmentMigrationChangeSetID || replayed.RowVersion != 7 {
			t.Fatalf("migrated replay identity/version = %#v", replayed)
		}

		if err := database.RollbackThrough(ctx, 34); err != nil {
			t.Fatalf("roll back assessment idempotency cutover: %v", err)
		}
		assertAssessmentLegacyMigrationPayload(t, assessmentMigrationResponse(t, ctx, database.SQL(), "legacy"), 7)
		assertAssessmentLegacyMigrationPayload(t, assessmentMigrationResponse(t, ctx, database.SQL(), "canonical"), 8)
		assertJSONSemanticEqual(t, assessmentMigrationResponse(t, ctx, database.SQL(), "unrelated"), unrelated)

		if err := database.ApplyThrough(ctx, 35); err != nil {
			t.Fatalf("reapply assessment idempotency cutover: %v", err)
		}
		assertAssessmentCanonicalMigrationPayload(t, assessmentMigrationResponse(t, ctx, database.SQL(), "legacy"), 7)
		assertAssessmentCanonicalMigrationPayload(t, assessmentMigrationResponse(t, ctx, database.SQL(), "canonical"), 8)
	})

	t.Run("up preflight rejects unsupported state without mutation or leakage", func(t *testing.T) {
		harness := pgtest.Start(t)
		database := harness.MigrationDatabaseThroughT(t, 34)
		ctx := context.Background()
		seedAssessmentMigrationActor(t, ctx, database.SQL())
		invalid := `{"view_schema_id":"cartulary.view.assessments.v1","change_set_id":"` + assessmentMigrationChangeSetID + `","row":{"record_id":"` + assessmentMigrationRecordID + `","incident_id":"` + assessmentMigrationIncidentID + `","row_version":1},"unexpected":"secret-marker"}`
		insertAssessmentMigrationRow(t, ctx, database.SQL(), "assessments.rows.create", "invalid", []byte("invalid-hash"), 201, invalid)

		err := database.ApplyThrough(ctx, 35)
		assertAssessmentMigrationPreflightError(t, err, "assessment_create_idempotency_v1_preflight_failed")
		if strings.Contains(err.Error(), "secret-marker") || strings.Contains(err.Error(), assessmentMigrationRecordID) || strings.Contains(err.Error(), assessmentMigrationIncidentID) {
			t.Fatalf("migration error leaked stored data: %v", err)
		}
		assertJSONSemanticEqual(t, assessmentMigrationResponse(t, ctx, database.SQL(), "invalid"), invalid)
	})

	t.Run("down preflight rejects unsupported state without mutation", func(t *testing.T) {
		harness := pgtest.Start(t)
		database := harness.MigrationDatabaseThroughT(t, 34)
		ctx := context.Background()
		seedAssessmentMigrationActor(t, ctx, database.SQL())
		canonical := assessmentCanonicalMigrationPayload(assessmentMigrationRecordID, assessmentMigrationIncidentID, 3)
		insertAssessmentMigrationRow(t, ctx, database.SQL(), "assessments.rows.create", "rollback-invalid", []byte("rollback-hash"), 201, canonical)
		if err := database.ApplyThrough(ctx, 35); err != nil {
			t.Fatalf("apply assessment idempotency cutover: %v", err)
		}
		invalid := `{"schema_id":"cartulary.assessments.create_result.v2"}`
		if _, err := database.SQL().ExecContext(ctx, `UPDATE route_idempotency SET response_json = $1::jsonb WHERE client_txn_id = 'rollback-invalid'`, invalid); err != nil {
			t.Fatalf("corrupt rollback fixture: %v", err)
		}
		err := database.RollbackThrough(ctx, 34)
		assertAssessmentMigrationPreflightError(t, err, "assessment_create_idempotency_v1_rollback_preflight_failed")
		assertJSONSemanticEqual(t, assessmentMigrationResponse(t, ctx, database.SQL(), "rollback-invalid"), invalid)
	})
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

func assessmentLegacyMigrationPayload(recordID string, rowVersion int64) string {
	return `{"view_schema_id":"cartulary.view.assessments.v1","change_set_id":"` + assessmentMigrationChangeSetID + `","row":{"record_id":"` + recordID + `","row_version":` + assessmentMigrationInt(rowVersion) + `,"future_field":true}}`
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

func assertAssessmentCanonicalMigrationPayload(t *testing.T, payload []byte, rowVersion int64) {
	t.Helper()
	assertJSONSemanticEqual(t, payload, assessmentCanonicalMigrationPayload(assessmentMigrationRecordID, assessmentMigrationIncidentID, rowVersion))
}

func assertAssessmentLegacyMigrationPayload(t *testing.T, payload []byte, rowVersion int64) {
	t.Helper()
	assertJSONSemanticEqual(t, payload, assessmentLegacyMigrationPayload(assessmentMigrationRecordID, rowVersion))
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

func assertAssessmentMigrationPreflightError(t *testing.T, err error, reason string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), reason) {
		t.Fatalf("migration error = %v; want %q", err, reason)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || !strings.Contains(postgresError.Detail, "invalid=1") {
		t.Fatalf("migration error detail = %q; want invalid=1", postgresErrorDetail(postgresError))
	}
}

func postgresErrorDetail(err *pgconn.PgError) string {
	if err == nil {
		return ""
	}
	return err.Detail
}
