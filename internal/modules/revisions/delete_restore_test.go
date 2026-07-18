package revisions_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDeleteRestoreAdapterMatrix_Unit(t *testing.T) {
	t.Parallel()
	_, err := revisionassembly.NewCommandService(nil, nil)
	if !errors.Is(err, revisions.ErrInvalidCommandServiceDependency) {
		t.Fatalf("application composition did not complete every provider catalog before dependency validation: %v", err)
	}
}

func TestSoftDeleteRoutePreconditions_Unit(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase7-u-7-03-delete")
	login, actorID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U703")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC))
	body := map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-03-delete", "reason": " \r\n\t "}

	setMembershipRole(t, harness.DB, incidentID, actorID, "viewer")
	forbidden := deleteRecord(t, harness, login, recordID, body)
	httptestx.RequireErrorEnvelope(t, forbidden, http.StatusForbidden, "authorization_denied")

	missing := deleteRecord(t, harness, login, uuid.New(), body)
	httptestx.RequireErrorEnvelope(t, missing, http.StatusNotFound, "incident_not_found")

	setMembershipRole(t, harness.DB, incidentID, actorID, "editor")
	staleRecord := uuid.New()
	workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, actorID, staleRecord, "Stale Host", "stale-host", "", "")
	stale := deleteRecord(t, harness, login, staleRecord, map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-03-stale"})
	staleErr := httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "row_version_conflict")
	staleDetails := staleErr["error"].(map[string]any)["details"].(map[string]any)
	if staleDetails["base_row_version"] != float64(2) || staleDetails["current_row_version"] != float64(1) {
		t.Fatalf("unexpected stale details: %#v", staleDetails)
	}

	success := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, body), http.StatusOK)["data"].(map[string]any)
	if success["record_id"] != recordID.String() || success["incident_id"] != incidentID.String() || success["deleted"] != true || success["row_version"] != float64(2) {
		t.Fatalf("unexpected delete success payload: %#v", success)
	}
	if success["deleted_at"] == nil || success["deleted_by_user_id"] != actorID.String() || success["change_set_id"] == "" {
		t.Fatalf("delete response missing tombstone attribution: %#v", success)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id = $2`, recordID, actorID); got != 1 {
		t.Fatalf("record envelope was not soft-deleted, count=%d", got)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE record_id = $1`, recordID); got != 0 {
		t.Fatalf("soft-deleted host remained in ordinary projection, count=%d", got)
	}
	if got := nullableChangeSetReason(t, harness.DB, success["change_set_id"].(string)); got != nil {
		t.Fatalf("normalized-empty reason must persist as null, got %q", *got)
	}

	reasonedRecord := uuid.New()
	workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, actorID, reasonedRecord, "Reasoned Delete Host", "reasoned-delete-host", "", "")
	reasonedDelete := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, reasonedRecord, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-u-7-03-delete-reasoned",
		"reason":           "  Incident review reason\n",
	}), http.StatusOK)["data"].(map[string]any)
	if got := nullableChangeSetReason(t, harness.DB, reasonedDelete["change_set_id"].(string)); got == nil || *got != "Incident review reason" {
		t.Fatalf("non-empty delete reason must persist normalized text, got %#v", got)
	}

	tagID := seedRecordTag(t, harness.DB, incidentID, recordID, actorID)
	tagDelete := deleteRecord(t, harness, login, tagID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-03-delete-tag"})
	httptestx.RequireErrorEnvelope(t, tagDelete, http.StatusNotFound, "incident_not_found")

	replay := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-03-delete"}), http.StatusOK)["data"].(map[string]any)
	if replay["change_set_id"] != success["change_set_id"] || countRecordMutations(t, harness.DB, recordID, "soft_delete") != 1 {
		t.Fatalf("idempotent delete replay created a second mutation or changed payload: replay=%#v success=%#v", replay, success)
	}

	divergent := deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-03-delete", "reason": "different"})
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	already := deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-03-already"})
	httptestx.RequireErrorEnvelope(t, already, http.StatusConflict, "record_already_deleted")

	noteID := uuid.New()
	seedNoteRecord(t, harness.DB, incidentID, actorID, noteID)
	httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, noteID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-03-delete-note"}), http.StatusOK)
	patch := workbookscenariotest.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+noteID.String(), map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-u-7-03-patch-deleted",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "should not apply"}},
	}, workbookscenariotest.WithCookies(login.SessionCookie, login.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, patch, http.StatusConflict, "record_deleted_use_restore")
}

func TestRestoreTombstonePreconditions_Unit(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase7-u-7-04-restore")
	login, actorID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U704")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 12, 10, 0, 0, time.UTC))
	deletePayload := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-04-delete"}), http.StatusOK)["data"].(map[string]any)
	tombstoneVersion := int64(deletePayload["row_version"].(float64))

	setMembershipRole(t, harness.DB, incidentID, actorID, "editor")
	forbidden := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion, "client_txn_id": "txn-u-7-04-forbidden"})
	httptestx.RequireErrorEnvelope(t, forbidden, http.StatusForbidden, "authorization_denied")

	setMembershipRole(t, harness.DB, incidentID, actorID, "reviewer")
	lockTx, err := harness.DB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin restore lock holder: %v", err)
	}
	if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, recordID); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("hold restore lock: %v", err)
	}
	locked := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion - 1, "client_txn_id": "txn-u-7-04-locked"})
	lockedErr := httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
	if lockedErr["error"].(map[string]any)["retryable"] != true {
		t.Fatalf("record_locked must be retryable: %#v", lockedErr)
	}
	if err := lockTx.Rollback(); err != nil {
		t.Fatalf("release restore lock: %v", err)
	}

	stale := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion - 1, "client_txn_id": "txn-u-7-04-stale"})
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "row_version_conflict")

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 12, 11, 0, 0, time.UTC))
	success := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion, "client_txn_id": "txn-u-7-04-restore", "reason": nil}), http.StatusOK)["data"].(map[string]any)
	if success["deleted"] != false || success["deleted_at"] != nil || success["deleted_by_user_id"] != nil || success["row_version"] != float64(tombstoneVersion+1) {
		t.Fatalf("unexpected restore success payload: %#v", success)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND deleted_at IS NULL AND deleted_by_user_id IS NULL`, recordID); got != 1 {
		t.Fatalf("record envelope was not restored, count=%d", got)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE record_id = $1`, recordID); got != 1 {
		t.Fatalf("restored host was not eligible for ordinary projection, count=%d", got)
	}
	if countRecordMutations(t, harness.DB, recordID, "soft_delete") != 1 || countRecordMutations(t, harness.DB, recordID, "restore") != 1 {
		t.Fatalf("delete/restore history mutations were not append-only")
	}
	if countRecordRevisions(t, harness.DB, recordID) != 2 {
		t.Fatalf("expected one delete and one restore revision")
	}

	replay := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion, "client_txn_id": "txn-u-7-04-restore"}), http.StatusOK)["data"].(map[string]any)
	if replay["change_set_id"] != success["change_set_id"] || countRecordMutations(t, harness.DB, recordID, "restore") != 1 {
		t.Fatalf("idempotent restore replay created a second mutation or changed payload: replay=%#v success=%#v", replay, success)
	}

	reasonedRecord := uuid.New()
	workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, actorID, reasonedRecord, "Reasoned Restore Host", "reasoned-restore-host", "", "")
	reasonedDelete := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, reasonedRecord, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-u-7-04-delete-reasoned",
	}), http.StatusOK)["data"].(map[string]any)
	reasonedTombstoneVersion := int64(reasonedDelete["row_version"].(float64))
	reasonedRestore := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, reasonedRecord, map[string]any{
		"base_row_version": reasonedTombstoneVersion,
		"client_txn_id":    "txn-u-7-04-restore-reasoned",
		"reason":           " Restore approval reason ",
	}), http.StatusOK)["data"].(map[string]any)
	if got := nullableChangeSetReason(t, harness.DB, reasonedRestore["change_set_id"].(string)); got == nil || *got != "Restore approval reason" {
		t.Fatalf("non-empty restore reason must persist normalized text, got %#v", got)
	}

	notDeleted := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion + 1, "client_txn_id": "txn-u-7-04-not-deleted"})
	httptestx.RequireErrorEnvelope(t, notDeleted, http.StatusConflict, "record_not_deleted")
}

func deleteRecord(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return workbookscenariotest.DoJSON(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, workbookscenariotest.WithCookies(login.SessionCookie, login.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
}

func restoreRecord(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return workbookscenariotest.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/restore", body, workbookscenariotest.WithCookies(login.SessionCookie, login.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
}

func setMembershipRole(t testing.TB, db *sql.DB, incidentID uuid.UUID, userID uuid.UUID, role string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `UPDATE incident_memberships SET role = $3, updated_at = now(), updated_by_user_id = $2 WHERE incident_id = $1 AND user_id = $2`, incidentID, userID, role); err != nil {
		t.Fatalf("set membership role: %v", err)
	}
}

func seedHostProjection(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO host_grid_projection (record_id, incident_id, row_version, display_name, hostname, host_state, edited_at)
SELECT h.record_id, h.incident_id, r.row_version, h.display_name, h.hostname, h.host_state, r.updated_at
  FROM hosts h
  JOIN records r ON r.record_id = h.record_id
 WHERE h.incident_id = $1 AND h.record_id = $2
ON CONFLICT (record_id) DO UPDATE
SET row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    hostname = EXCLUDED.hostname,
    host_state = EXCLUDED.host_state,
    edited_at = EXCLUDED.edited_at
`, incidentID, recordID); err != nil {
		t.Fatalf("seed host projection: %v", err)
	}
}

func seedNoteRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	workbookscenariotest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "artifact")
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
VALUES ($1, $2, 'note', 'History Note', 'Patch-after-delete note body', $3)
`, recordID, incidentID, actorUserID); err != nil {
		t.Fatalf("seed note record: %v", err)
	}
}

func seedIndicatorRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	workbookscenariotest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "indicator")
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO indicators (
    record_id,
    incident_id,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, 'domain', 'atomic', 'phase7.example.test', 'phase7.example.test', 'domain:phase7.example.test', $3, $3)
`, recordID, incidentID, actorUserID); err != nil {
		t.Fatalf("seed indicator record: %v", err)
	}
	return recordID
}

func seedRecordTag(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorUserID uuid.UUID) uuid.UUID {
	t.Helper()
	tagID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, $3, 'History tag', 'phase 7 tag', $4)
`, tagID, incidentID, recordID, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}
	return tagID
}

type MutationState struct {
	DeletedRecords int
	ChangeSets     int
	Mutations      int
	Revisions      int
	Idempotency    int
	ProjectionRows int
}

func StateCounts(t testing.TB, db *sql.DB, recordID uuid.UUID) MutationState {
	t.Helper()
	return MutationState{
		DeletedRecords: countRows(t, db, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND deleted_at IS NOT NULL`, recordID),
		ChangeSets:     countRows(t, db, `SELECT COUNT(*) FROM change_sets WHERE change_set_id IN (SELECT change_set_id FROM change_set_mutations WHERE target_id = $1)`, recordID.String()),
		Mutations:      countRows(t, db, `SELECT COUNT(*) FROM change_set_mutations WHERE target_id = $1`, recordID.String()),
		Revisions:      countRows(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID),
		Idempotency:    countRows(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE scope_key = $1`, recordID.String()),
		ProjectionRows: countRows(t, db, `SELECT COUNT(*) FROM host_grid_projection WHERE record_id = $1`, recordID),
	}
}

func countRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return count
}

func countRecordMutations(t testing.TB, db *sql.DB, recordID uuid.UUID, operation string) int {
	t.Helper()
	return countRows(t, db, `SELECT COUNT(*) FROM change_set_mutations WHERE target_kind = 'record' AND target_id = $1 AND operation_kind = $2`, recordID.String(), operation)
}

func countRecordRevisions(t testing.TB, db *sql.DB, recordID uuid.UUID) int {
	t.Helper()
	return countRows(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID)
}

func nullableChangeSetReason(t testing.TB, db *sql.DB, changeSetID string) *string {
	t.Helper()
	var reason sql.NullString
	if err := db.QueryRowContext(context.Background(), `SELECT reason FROM change_sets WHERE change_set_id::text = $1`, changeSetID).Scan(&reason); err != nil {
		t.Fatalf("load change set reason: %v", err)
	}
	if !reason.Valid {
		return nil
	}
	return &reason.String
}
