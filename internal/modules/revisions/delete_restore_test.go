package revisions_test

import (
	"context"
	"database/sql"
	"errors"
	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	projectiontest "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDeleteRestoreAdapterMatrix_Unit(t *testing.T) {
	t.Parallel()
	wantAdapters := map[string]string{
		"artifact":       "github.com/JochiRaider/cartulary/internal/modules/artifacts/deleterestore.Source",
		"assessment":     "github.com/JochiRaider/cartulary/internal/modules/assessments/deleterestore.Source",
		"decision":       "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/deleterestore.DecisionSource",
		"evidence":       "github.com/JochiRaider/cartulary/internal/modules/evidence/deleterestore.Source",
		"host":           "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/deleterestore.HostSource",
		"identity":       "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/deleterestore.IdentitySource",
		"indicator":      "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/deleterestore.Source",
		"party":          "github.com/JochiRaider/cartulary/internal/modules/parties/deleterestore.Source",
		"task_request":   "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/deleterestore.TaskRequestSource",
		"timeline_event": "github.com/JochiRaider/cartulary/internal/modules/timeline/deleterestore.Source",
	}
	contributions := revisionassembly.CurrentProviderContributions()
	gotAdapters := make(map[string]string, len(wantAdapters))
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			if record.DeleteRestoreSource == nil {
				t.Fatalf("record type %q has nil delete/restore source", record.RecordType)
			}
			sourceType := reflect.TypeOf(record.DeleteRestoreSource)
			if sourceType.Kind() == reflect.Pointer {
				sourceType = sourceType.Elem()
			}
			if _, exists := gotAdapters[record.RecordType]; exists {
				t.Fatalf("record type %q has duplicate delete/restore sources", record.RecordType)
			}
			gotAdapters[record.RecordType] = sourceType.PkgPath() + "." + sourceType.Name()
		}
	}
	if !reflect.DeepEqual(gotAdapters, wantAdapters) {
		t.Fatalf("delete/restore adapter matrix = %#v, want %#v", gotAdapters, wantAdapters)
	}
	runtime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         collaboration.NewIntentAppender(),
		},
		contributions...,
	)
	if err != nil {
		t.Fatalf("build Revisions runtime: %v", err)
	}
	_, err = runtime.NewCommandService(nil, nil, nil, nil, nil)
	if !errors.Is(err, revisions.ErrInvalidCommandServiceDependency) {
		t.Fatalf("application composition did not complete every provider catalog before dependency validation: %v", err)
	}
}

func TestDeleteRestoreConcreteSourceAdapterMatrix_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        "history_revision-delete-restore-source-matrix",
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, hostID := seedRecord(
		t,
		harness.DB,
		harness.Server,
		login,
		actorID,
		"IR-DELETE-RESTORE-SOURCE-MATRIX",
	)
	recordIDs := seedDeleteRestoreAdapterRecords(t, harness.DB, incidentID, actorID, hostID)

	sources := map[string]revisions.RecordProviderContribution{}
	for _, contribution := range revisionassembly.CurrentProviderContributions() {
		for _, record := range contribution.Records {
			sources[record.RecordType] = record
		}
	}
	wantViews := map[string]string{
		"artifact":       "cartulary.view.notes.v1",
		"assessment":     "cartulary.view.assessments.v1",
		"decision":       "cartulary.view.decisions.v1",
		"evidence":       "cartulary.view.evidence.v1",
		"host":           "cartulary.view.hosts.v1",
		"identity":       "cartulary.view.identities.v1",
		"indicator":      "cartulary.view.indicators.v1",
		"party":          "cartulary.view.parties.v1",
		"task_request":   "cartulary.view.task_requests.v1",
		"timeline_event": "cartulary.view.timeline.v2",
	}
	ctx := context.Background()
	projectionRuntime := projectiontest.MustBuild(t, harness.Pool)
	projectionServices := projectionRuntime.RevisionServices()
	liveRecords := projectionRuntime.RevisionLiveRecords()
	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin source adapter matrix transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := projectionServices.RebuildIncidentTx(ctx, tx, incidentID); err != nil {
		t.Fatalf("rebuild characterization projections: %v", err)
	}

	now := time.Date(2026, 7, 29, 18, 0, 0, 0, time.UTC)
	liveRows := 0
	for recordType, wantView := range wantViews {
		contribution, ok := sources[recordType]
		if !ok {
			t.Fatalf("record type %q has no source contribution", recordType)
		}
		recordID := recordIDs[recordType]
		snapshot, err := contribution.DeleteRestoreSource.SnapshotTx(ctx, tx, recordID)
		if err != nil {
			t.Fatalf("%s snapshot: %v", recordType, err)
		}
		record, recordOK := snapshot["record"].(map[string]any)
		source, sourceOK := snapshot["source"].(map[string]any)
		if !recordOK || !sourceOK ||
			record["record_id"] != recordID.String() ||
			record["record_type"] != recordType ||
			source["record_id"] != recordID.String() {
			t.Fatalf("%s snapshot shape = %#v", recordType, snapshot)
		}
		viewSchemaID, err := contribution.DeleteRestoreSource.ViewSchemaID(ctx, tx, recordID)
		if err != nil || viewSchemaID != wantView {
			t.Fatalf("%s view consequence = %q, %v; want %q", recordType, viewSchemaID, err, wantView)
		}
		if liveRecords.Supports(viewSchemaID) {
			projectionRow, err := liveRecords.LoadRowTx(ctx, tx, viewSchemaID, recordID)
			if err != nil {
				t.Fatalf("%s live projection row: %v", recordType, err)
			}
			liveRows++
			if projectionRow["record_id"] != recordID.String() {
				t.Fatalf("%s projection record identity = %#v", recordType, projectionRow["record_id"])
			}
			if reflect.DeepEqual(projectionRow, snapshot) || projectionRow["record"] != nil || projectionRow["source"] != nil {
				t.Fatalf("%s live projection row unexpectedly matches the authoritative {record, source} snapshot: %#v", recordType, projectionRow)
			}
		}
		reasonCode, blocked, err := contribution.DeleteRestoreSource.ValidateDeletePreconditionsTx(
			ctx,
			tx,
			incidentID,
			recordID,
		)
		if err != nil || blocked || reasonCode != "" {
			t.Fatalf("%s delete precondition = reason %q blocked %v err %v", recordType, reasonCode, blocked, err)
		}
		if err := contribution.DeleteRestoreSource.UpdateSourceDeleteStateTx(
			ctx,
			tx,
			recordID,
			actorID,
			now,
			true,
		); err != nil {
			t.Fatalf("%s source delete: %v", recordType, err)
		}
	}
	if liveRows == 0 {
		t.Fatal("characterization did not exercise a live projection row")
	}
	for _, recordType := range []string{"assessment"} {
		recordID := recordIDs[recordType]
		table := recordType + "s"
		var matched bool
		if err := tx.QueryRow(
			ctx,
			"SELECT deleted_at = $2 AND deleted_by_user_id = $3 FROM "+table+" WHERE record_id = $1",
			recordID,
			now,
			actorID,
		).Scan(&matched); err != nil || !matched {
			t.Fatalf("%s source tombstone = %v, %v", recordType, matched, err)
		}
	}
	for recordType, contribution := range sources {
		if err := contribution.DeleteRestoreSource.UpdateSourceDeleteStateTx(
			ctx,
			tx,
			recordIDs[recordType],
			actorID,
			now.Add(time.Minute),
			false,
		); err != nil {
			t.Fatalf("%s source restore: %v", recordType, err)
		}
	}
	for _, recordType := range []string{"assessment"} {
		recordID := recordIDs[recordType]
		table := recordType + "s"
		var cleared bool
		if err := tx.QueryRow(
			ctx,
			"SELECT deleted_at IS NULL AND deleted_by_user_id IS NULL FROM "+table+" WHERE record_id = $1",
			recordID,
		).Scan(&cleared); err != nil || !cleared {
			t.Fatalf("%s source tombstone clear = %v, %v", recordType, cleared, err)
		}
	}
}

func TestSoftDeleteRoutePreconditions_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-u-7-03-delete")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
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
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, staleRecord, "Stale Host", "stale-host", "", "")
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
	requireSingleRecordChangeIntent(t, harness.DB, success["change_set_id"].(string), recordID, "remove")
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
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, reasonedRecord, "Reasoned Delete Host", "reasoned-delete-host", "", "")
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
	noteDelete := deleteRecord(t, harness, login, noteID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-03-delete-note"})
	if noteDelete.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(noteDelete.Body)
		t.Fatalf("delete note status = %d body=%s", noteDelete.StatusCode, body)
	}
	httptestx.RequireSuccessEnvelope(t, noteDelete, http.StatusOK)
	patch := appsupport.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+noteID.String(), map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-u-7-03-patch-deleted",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "should not apply"}},
	}, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, patch, http.StatusConflict, "record_deleted_use_restore")
}

func TestRestoreTombstonePreconditions_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-u-7-04-restore")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
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
	requireSingleRecordChangeIntent(t, harness.DB, success["change_set_id"].(string), recordID, "invalidate")

	replay := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion, "client_txn_id": "txn-u-7-04-restore"}), http.StatusOK)["data"].(map[string]any)
	if replay["change_set_id"] != success["change_set_id"] || countRecordMutations(t, harness.DB, recordID, "restore") != 1 {
		t.Fatalf("idempotent restore replay created a second mutation or changed payload: replay=%#v success=%#v", replay, success)
	}

	reasonedRecord := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, reasonedRecord, "Reasoned Restore Host", "reasoned-restore-host", "", "")
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

func deleteRecord(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return appsupport.DoJSON(t, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
}

func restoreRecord(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/restore", body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
}

func setMembershipRole(t testing.TB, db *sql.DB, incidentID uuid.UUID, userID uuid.UUID, role string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `UPDATE incident_memberships SET role = $3, updated_at = now(), updated_by_user_id = $2 WHERE incident_id = $1 AND user_id = $2`, incidentID, userID, role); err != nil {
		t.Fatalf("set membership role: %v", err)
	}
}

func requireSingleRecordChangeIntent(
	t testing.TB,
	db *sql.DB,
	changeSetID string,
	recordID uuid.UUID,
	changeKind string,
) {
	t.Helper()
	if got := countRows(t, db, `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE source_change_set_id = $1
   AND source_record_id = $2
   AND event_family = 'record_changed'
   AND canonical_payload -> 'affected_views' -> 0 ->> 'change_kind' = $3
`, changeSetID, recordID, changeKind); got != 1 {
		t.Fatalf(
			"record change intent count for change set %s record %s kind %s = %d, want 1",
			changeSetID,
			recordID,
			changeKind,
			got,
		)
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

func seedDeleteRestoreAdapterRecords(
	t testing.TB,
	db *sql.DB,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	hostID uuid.UUID,
) map[string]uuid.UUID {
	t.Helper()
	recordIDs := map[string]uuid.UUID{
		"artifact":       uuid.New(),
		"assessment":     uuid.New(),
		"decision":       uuid.New(),
		"evidence":       uuid.New(),
		"host":           hostID,
		"identity":       uuid.New(),
		"indicator":      uuid.New(),
		"party":          uuid.New(),
		"task_request":   uuid.New(),
		"timeline_event": uuid.New(),
	}
	entitytest.SeedIdentityRecord(
		t,
		db,
		incidentID,
		actorID,
		recordIDs["identity"],
		"Matrix Identity",
		"matrix@example.test",
		"matrix@example.test",
		"matrix",
	)
	timelinetest.SeedTimelineRecord(t, db, incidentID, actorID, recordIDs["timeline_event"])
	assessmenttest.SeedAssessment(
		t,
		db,
		incidentID,
		actorID,
		recordIDs["assessment"],
		hostID,
		"host",
		"suspected",
	)
	for _, recordType := range []string{
		"artifact",
		"decision",
		"evidence",
		"party",
		"task_request",
	} {
		envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, recordIDs[recordType], recordType)
	}
	mustExec(t, db, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, title, created_by_user_id)
VALUES ($1, $2, 'note', 'Matrix note', $3)
`, recordIDs["artifact"], incidentID, actorID)
	mustExec(t, db, `
INSERT INTO decisions (record_id, incident_id, summary)
VALUES ($1, $2, 'Matrix decision')
`, recordIDs["decision"], incidentID)
	mustExec(t, db, `
INSERT INTO evidence (record_id, incident_id, title)
VALUES ($1, $2, 'Matrix evidence')
`, recordIDs["evidence"], incidentID)
	indicatortest.SeedRecord(t, db, incidentID, actorID, recordIDs["indicator"], "ipv4_addr", "atomic", "192.0.2.77")
	mustExec(t, db, `
INSERT INTO parties (record_id, incident_id, display_name, party_kind)
VALUES ($1, $2, 'Matrix Party', 'person')
`, recordIDs["party"], incidentID)
	mustExec(t, db, `
INSERT INTO task_requests (record_id, incident_id, title)
VALUES ($1, $2, 'Matrix task')
`, recordIDs["task_request"], incidentID)
	return recordIDs
}

func seedNoteRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "artifact")
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
	indicatortest.SeedRecord(t, db, incidentID, actorUserID, recordID, "domain_name", "atomic", "history_revision.example.test")
	return recordID
}

func seedRecordTag(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorUserID uuid.UUID) uuid.UUID {
	t.Helper()
	tagID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, $3, 'History tag', 'history_revision tag', $4)
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
