package revisions_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	projectiontest "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type sourceSnapshotScanRow struct {
	raw []byte
	err error
}

func (row sourceSnapshotScanRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	*(dest[0].(*[]byte)) = append([]byte(nil), row.raw...)
	return nil
}

func TestDeleteRestoreAdapterMatrix_Unit(t *testing.T) {
	t.Parallel()
	t.Run("source snapshot JSON boundary", func(t *testing.T) {
		payload := []byte(`{"record":{"row_version":9007199254740993},"source":{"row_version":9223372036854775807}}`)
		snapshot, err := deleterestorecontract.ScanSnapshot(sourceSnapshotScanRow{raw: payload})
		if err != nil {
			t.Fatal(err)
		}
		if snapshot["record"].(map[string]any)["row_version"] != json.Number("9007199254740993") || snapshot["source"].(map[string]any)["row_version"] != json.Number("9223372036854775807") {
			t.Fatalf("snapshot lost exact integers: %#v", snapshot)
		}
		for _, suffix := range []string{" {}", " null", " broken"} {
			if _, err := deleterestorecontract.ScanSnapshot(sourceSnapshotScanRow{raw: append(append([]byte(nil), payload...), suffix...)}); err == nil {
				t.Fatalf("trailing snapshot content %q accepted", suffix)
			}
		}
		if _, err := deleterestorecontract.ScanSnapshot(sourceSnapshotScanRow{raw: []byte(`{"record":`)}); err == nil {
			t.Fatal("malformed snapshot accepted")
		}
		scanErr := errors.New("source row unavailable")
		if _, err := deleterestorecontract.ScanSnapshot(sourceSnapshotScanRow{err: scanErr}); !errors.Is(err, scanErr) {
			t.Fatalf("scan failure = %v", err)
		}
	})
	t.Run("source scalar kinds", func(t *testing.T) {
		record := map[string]any{"record_id": "10000000-0000-4000-8000-000000000001", "incident_id": "20000000-0000-4000-8000-000000000002"}
		for _, test := range []struct {
			name           string
			kind           deleterestorecontract.SourceScalarKind
			valid, invalid []any
		}{
			{"string", deleterestorecontract.SourceString, []any{"", "2026-01-01T00:00:00+03:00"}, []any{true, json.Number("1"), nil}},
			{"UUID", deleterestorecontract.SourceUUID, []any{"abcdef00-1234-4000-8000-000000000001"}, []any{"ABCDEF00-1234-4000-8000-000000000001", uuid.Nil.String(), "bad", 1}},
			{"timestamp", deleterestorecontract.SourceTimestamp,
				[]any{"0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999+00:00", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00.123456+00:00", "2026-01-01T00:00:00.000001Z"},
				[]any{"0000-01-01T00:00:00Z", "0000-12-31T23:59:59+00:00", "10000-01-01T00:00:00Z", "2026-01-01", "2026-01-01T00:00:00+03:00", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.123456789Z", "2026-01-01T00:00:00.1234560Z", "2026-01-01T00:00:00.100000Z", 1}},
			{"integer", deleterestorecontract.SourceInteger, []any{json.Number("0"), json.Number("9007199254740993"), json.Number("9223372036854775807")}, []any{json.Number("9223372036854775808"), json.Number("1.5"), json.Number("1e2"), "1", float64(1)}},
			{"boolean", deleterestorecontract.SourceBoolean, []any{true, false}, []any{"true", json.Number("1"), nil}},
		} {
			t.Run(test.name, func(t *testing.T) {
				check := func(value any, nullable bool) error {
					source := maps.Clone(record)
					source["value"] = value
					return deleterestorecontract.ValidateSourceObject(record, source,
						deleterestorecontract.SourceFieldGroup{Names: "record_id incident_id", Kind: deleterestorecontract.SourceUUID},
						deleterestorecontract.SourceFieldGroup{Names: "value", Kind: test.kind, Nullable: nullable})
				}
				for _, value := range test.valid {
					if err := check(value, false); err != nil {
						t.Errorf("valid %#v rejected: %v", value, err)
					}
				}
				for _, value := range test.invalid {
					if err := check(value, false); err == nil {
						t.Errorf("invalid %#v accepted", value)
					}
				}
				if err := check(nil, true); err != nil {
					t.Errorf("nullable null rejected: %v", err)
				}
			})
		}
	})
	wantRecordTypes := map[string]struct{}{
		"artifact": {}, "assessment": {}, "decision": {}, "evidence": {}, "host": {},
		"identity": {}, "indicator": {}, "party": {}, "task_request": {}, "timeline_event": {},
	}
	contributions := mustRevisionProviderContributions(t)
	gotRecordTypes := make(map[string]struct{}, len(wantRecordTypes))
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			if record.DeleteRestoreSource == nil || record.ValidateSnapshotSource == nil {
				t.Fatalf("record type %q has nil delete/restore source", record.RecordType)
			}
			if _, exists := gotRecordTypes[record.RecordType]; exists {
				t.Fatalf("record type %q has duplicate delete/restore sources", record.RecordType)
			}
			gotRecordTypes[record.RecordType] = struct{}{}
		}
	}
	if !reflect.DeepEqual(gotRecordTypes, wantRecordTypes) {
		t.Fatalf("delete/restore record-type matrix = %#v, want %#v", gotRecordTypes, wantRecordTypes)
	}
	runtime, err := revisionassembly.Build(contributions...)
	if err != nil {
		t.Fatalf("build Revisions runtime: %v", err)
	}
	_, err = runtime.NewCommandService(nil, nil, nil, nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("application composition did not complete every provider catalog before dependency validation: %v", err)
	}
}

func TestDeleteRestoreConcreteSourceAdapterMatrix_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        "history_revision-delete-restore-source-matrix",
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
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
	for _, contribution := range mustRevisionProviderContributions(t) {
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
	projectionRebuilder := projectionRuntime.ImportRebuilder()
	liveRecords := projectionRuntime.RevisionLiveRecords()
	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin source adapter matrix transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := projectionRebuilder.RebuildImportedIncidentTx(ctx, tx, incidentID); err != nil {
		t.Fatalf("rebuild characterization projections: %v", err)
	}

	var now time.Time
	if err := tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		t.Fatalf("read source transition clock: %v", err)
	}
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
		t.Run(recordType+" source schema", func(t *testing.T) {
			if contribution.ValidateSnapshotSource == nil {
				t.Fatal("source validator absent")
			}
			original, err := json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			if err := contribution.ValidateSnapshotSource(record, source); err != nil {
				t.Fatalf("complete source rejected: %v", err)
			}
			if recordType == "evidence" {
				requireEvidenceHistoricalLifecycle(t, contribution.ValidateSnapshotSource, record, source)
				requireEvidenceHistoricalStoredText(t, contribution.ValidateSnapshotSource, record, source)
				requireEvidenceHistoricalStorageReference(t, contribution.ValidateSnapshotSource, record, source)
			}
			if recordType == "indicator" {
				requireIndicatorHistoricalRepresentation(t, contribution.ValidateSnapshotSource, record, source)
			}
			if recordType == "timeline_event" {
				if source["capture_state"] != "reviewed" || source["reviewed_by_user_id"] != actorID.String() || source["reviewed_at"] != source["recorded_at"] {
					t.Fatalf("Timeline reviewed fixture attribution is incomplete: %#v", source)
				}
				requireTimelineHistoricalIntrinsicState(t, contribution.ValidateSnapshotSource, record, source)
			}
			if recordType == "decision" {
				requireDecisionHistoricalIntrinsicState(t, contribution.ValidateSnapshotSource, record, source)
				requireTaskDecisionHistoricalStoredText(t, contribution.ValidateSnapshotSource, record, source, recordType)
			}
			if recordType == "task_request" {
				requireTaskHistoricalIntrinsicState(t, contribution.ValidateSnapshotSource, record, source)
				requireTaskDecisionHistoricalStoredText(t, contribution.ValidateSnapshotSource, record, source, recordType)
			}
			for _, key := range slices.Sorted(maps.Keys(source)) {
				t.Run(key, func(t *testing.T) {
					missing := maps.Clone(source)
					delete(missing, key)
					if err := contribution.ValidateSnapshotSource(record, missing); err == nil {
						t.Error("missing member accepted")
					}
					mistyped := maps.Clone(source)
					mistyped[key] = map[string]any{"wrong": "scalar"}
					if err := contribution.ValidateSnapshotSource(record, mistyped); err == nil {
						t.Error("object in scalar member accepted")
					}
				})
			}
			for _, key := range []string{"record_id", "incident_id", "projection_label"} {
				invalid := maps.Clone(source)
				invalid[key] = "ffffffff-ffff-4fff-8fff-ffffffffffff"
				if err := contribution.ValidateSnapshotSource(record, invalid); err == nil {
					t.Errorf("foreign or extra member %s accepted", key)
				}
			}
			after, err := json.Marshal(source)
			if err != nil || string(after) != string(original) {
				t.Fatalf("validation changed retained source: %s / %s (%v)", original, after, err)
			}
		})
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
		if recordType == "host" || recordType == "identity" {
			// The source adapter mirrors independently advanced envelope facts,
			// not the delete/restore command's deliberately different actor/clock.
			mirrorClock := now.Add(time.Hour).UTC().Truncate(time.Microsecond)
			if _, err := tx.Exec(ctx, `UPDATE records SET row_version=7,updated_at=$2,updated_by_user_id=$3 WHERE record_id=$1`, recordID, mirrorClock, actorID); err != nil {
				t.Fatal(err)
			}
			var before string
			if err := tx.QueryRow(ctx, `SELECT to_jsonb(r)::text FROM records r WHERE record_id=$1`, recordID).Scan(&before); err != nil {
				t.Fatal(err)
			}
			table := "hosts"
			if recordType == "identity" {
				table = "identities"
			}
			for _, deleted := range []bool{true, false} {
				if err := contribution.DeleteRestoreSource.UpdateSourceDeleteStateTx(ctx, tx, recordID, uuid.Nil, now.Add(-time.Hour), deleted); err != nil {
					t.Fatalf("%s envelope mirror: %v", recordType, err)
				}
				var exact bool
				if err := tx.QueryRow(ctx, "SELECT row_version=7 AND updated_at=$2 AND updated_by_user_id=$3 FROM "+table+" WHERE record_id=$1", recordID, mirrorClock, actorID).Scan(&exact); err != nil || !exact {
					t.Fatalf("%s mirror did not use envelope metadata: %v %v", recordType, exact, err)
				}
				var after string
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(r)::text FROM records r WHERE record_id=$1`, recordID).Scan(&after); err != nil || after != before {
					t.Fatalf("%s source adapter changed envelope: %v", recordType, err)
				}
			}
		}
		preparation, err := contribution.DeleteRestoreSource.PrepareStateTransitionTx(
			ctx,
			tx,
			deleterestorecontract.StateTransitionRequest{
				Kind:       deleterestorecontract.StateTransitionDelete,
				IncidentID: incidentID,
				RecordID:   recordID,
			},
		)
		if err != nil || preparation.Blocked != nil {
			t.Fatalf("%s delete preparation = %#v err %v", recordType, preparation, err)
		}
		validation, err := contribution.DeleteRestoreSource.ValidateStateTransitionTx(
			ctx, tx, deleterestorecontract.StateTransitionRequest{
				Kind: deleterestorecontract.StateTransitionDelete, IncidentID: incidentID, RecordID: recordID,
			},
		)
		if err != nil || validation.Blocked != nil {
			t.Fatalf("%s delete validation = %#v err %v", recordType, validation, err)
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
		snapshot, err := sources[recordType].DeleteRestoreSource.SnapshotTx(ctx, tx, recordID)
		if err != nil {
			t.Fatal(err)
		}
		if err := sources[recordType].ValidateSnapshotSource(snapshot["record"].(map[string]any), snapshot["source"].(map[string]any)); err != nil {
			t.Fatalf("deleted %s source rejected: %v", recordType, err)
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
		snapshot, err := sources[recordType].DeleteRestoreSource.SnapshotTx(ctx, tx, recordID)
		if err != nil {
			t.Fatal(err)
		}
		if err := sources[recordType].ValidateSnapshotSource(snapshot["record"].(map[string]any), snapshot["source"].(map[string]any)); err != nil {
			t.Fatalf("restored %s source rejected: %v", recordType, err)
		}
	}
}

func requireDecisionHistoricalIntrinsicState(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	base := maps.Clone(captured)
	base["created_at"], base["updated_at"] = "2026-01-01T00:00:00Z", "2026-01-03T00:00:00+00:00"
	base["summary"], base["rationale"] = "Retained Decision", "Retained\n\trationale"
	check := func(t testing.TB, value map[string]any, valid bool) {
		t.Helper()
		before, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		err = validate(record, value)
		if (err == nil) != valid {
			t.Fatalf("historical Decision validity=%t, want %t: %v; value=%#v", err == nil, valid, err, value)
		}
		after, err := json.Marshal(value)
		if err != nil || string(before) != string(after) {
			t.Fatal("historical source was repaired or changed")
		}
	}
	for _, status := range []string{"proposed", "approved", "rejected", "superseded", "executed"} {
		for _, kind := range []string{"scope", "containment", "communication", "evidence", "reporting"} {
			for _, decided := range []string{"2000-01-01T00:00:00.000001Z", "2030-01-01T00:00:00+00:00"} {
				value := maps.Clone(base)
				value["status"], value["decision_type"], value["decided_at"] = status, kind, decided
				check(t, value, true)
			}
		}
	}
	for _, field := range []string{"record_id", "incident_id", "summary", "status", "owner_user_id", "decision_type", "decided_at", "rationale", "created_at", "updated_at"} {
		t.Run("required "+field, func(t *testing.T) {
			value := maps.Clone(base)
			value[field] = nil
			check(t, value, false)
		})
	}
	for _, field := range []string{"status", "decision_type"} {
		for _, invalid := range []string{"", "unknown", "Approved", "approved "} {
			value := maps.Clone(base)
			value[field] = invalid
			check(t, value, false)
		}
	}
	for _, field := range []string{"summary", "rationale"} {
		for _, invalid := range []string{"", " \n\t\u00a0", "bad\x00source"} {
			value := maps.Clone(base)
			value[field] = invalid
			check(t, value, false)
		}
	}
}

func requireTaskHistoricalIntrinsicState(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	base := maps.Clone(captured)
	base["created_at"], base["updated_at"] = "2026-01-01T00:00:00Z", "2026-01-03T00:00:00Z"
	base["status"], base["title"], base["priority"], base["task_kind"] = "open", "Retained task", "normal", "request"
	base["blocked_reason"], base["completed_at"] = nil, nil
	check := func(t testing.TB, value map[string]any, valid bool) {
		t.Helper()
		before, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		err = validate(record, value)
		if (err == nil) != valid {
			t.Fatalf("Task Request validity=%t want=%t err=%v value=%#v", err == nil, valid, err, value)
		}
		after, err := json.Marshal(value)
		if err != nil || string(after) != string(before) {
			t.Fatal("historical Task Request changed")
		}
	}
	for _, status := range []string{"open", "in_progress", "blocked", "done", "canceled"} {
		for _, priority := range []string{"low", "normal", "high", "urgent"} {
			for _, kind := range []string{"question", "request", "collection", "containment", "follow_up"} {
				value := maps.Clone(base)
				value["status"], value["priority"], value["task_kind"] = status, priority, kind
				if status == "blocked" {
					value["blocked_reason"] = "Waiting for evidence"
				}
				if status == "done" {
					value["completed_at"] = "2026-01-01T00:00:00Z"
				}
				check(t, value, true)
				value["owner_user_id"] = nil
				check(t, value, status == "done" || status == "canceled")
			}
		}
	}
	for _, field := range []string{"record_id", "incident_id", "title", "status", "priority", "task_kind", "created_at", "updated_at"} {
		value := maps.Clone(base)
		value[field] = nil
		check(t, value, false)
	}
	for _, field := range []string{"status", "priority", "task_kind"} {
		for _, bad := range []string{"", "Unknown", "unknown", "open "} {
			value := maps.Clone(base)
			value[field] = bad
			check(t, value, false)
		}
	}
	for _, bad := range []string{"", " \t\n\u00a0", "bad\x00title"} {
		value := maps.Clone(base)
		value["title"] = bad
		check(t, value, false)
	}
	for _, field := range []string{"workstream", "requester_party_text", "blocked_reason", "external_ticket_ref", "closure_summary"} {
		value := maps.Clone(base)
		value[field] = "bad\x00source"
		check(t, value, false)
	}
	for _, status := range []string{"open", "in_progress", "blocked", "done", "canceled"} {
		for _, blocker := range []any{nil, "", " \t\n", "A blocker"} {
			value := maps.Clone(base)
			value["status"], value["blocked_reason"] = status, blocker
			if status == "done" {
				value["completed_at"] = "2026-01-01T00:00:00Z"
			}
			valid := blocker == nil
			if status == "blocked" {
				valid = blocker == "A blocker"
			}
			check(t, value, valid)
		}
		for _, completed := range []any{nil, "2000-01-01T00:00:00Z", "2026-01-01T00:00:00Z", "2030-01-01T00:00:00Z"} {
			value := maps.Clone(base)
			value["status"], value["completed_at"] = status, completed
			if status == "blocked" {
				value["blocked_reason"] = "A blocker"
			}
			valid := completed == nil
			if status == "done" {
				valid = completed == "2026-01-01T00:00:00Z" || completed == "2030-01-01T00:00:00Z"
			}
			check(t, value, valid)
		}
	}
	// Due dates are explicit authored instants, independent of persistence clocks.
	for _, due := range []any{nil, "2000-01-01T00:00:00Z", "2030-01-01T00:00:00Z"} {
		value := maps.Clone(base)
		value["due_at"] = due
		check(t, value, true)
	}
}

func mustRevisionProviderContributions(t testing.TB) []revisions.ProviderContribution {
	t.Helper()
	contributions, err := revisionassembly.CurrentProviderContributions(referencefixture.IndicatorRegistry{})
	if err != nil {
		t.Fatalf("compose current Revisions provider contributions: %v", err)
	}
	return contributions
}

func TestSoftDeleteRoutePreconditions_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "history_revision-u-7-03-delete", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U703")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	var clockBase time.Time
	if err := harness.DB.QueryRow(`SELECT clock_timestamp() + interval '5 minutes'`).Scan(&clockBase); err != nil {
		t.Fatal(err)
	}
	httptestx.SetClockFixed(t, harness.Server, clockBase)
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
	patch := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+noteID.String(), map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-u-7-03-patch-deleted",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "should not apply"}},
	}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, patch, http.StatusConflict, "record_deleted_use_restore")
}

func TestSourceOwnerTransitionValidation_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "source-owner-transition-validation", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident, host := seedRecord(t, harness.DB, harness.Server, login, actor, "IR-OWNER-TRANSITION")
	seedHostProjection(t, harness.DB, incident, host)
	requirePartyDeleteReferencesAfterLockWait(t, harness, login, incident, actor)
	for _, variant := range []string{"collector", "source", "both_same", "both_distinct", "neither"} {
		t.Run("Evidence restore Party references/"+variant, func(t *testing.T) {
			requireEvidenceRestorePartyReferences(t, harness, login, incident, actor, variant)
		})
	}
	requireEvidenceRestoreBlobBridge(t, harness, login, incident, actor)
	requireTaskRestoreReferences(t, harness, login, incident, actor)
	requireDecisionDeleteTaskReferences(t, harness, login, incident, actor)
	requireDecisionDeleteSupersession(t, harness, login, incident, actor)
	requireDecisionRestoreSupersession(t, harness, login, incident, actor)
	deleted := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, host, map[string]any{"base_row_version": 1, "client_txn_id": "owner-transition-delete"}), http.StatusOK)["data"].(map[string]any)
	if deleted["deleted"] != true || deleted["row_version"] != float64(2) {
		t.Fatalf("Host delete through both owner phases=%#v", deleted)
	}
	restored := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, host, map[string]any{"base_row_version": 2, "client_txn_id": "owner-transition-restore"}), http.StatusOK)["data"].(map[string]any)
	if restored["deleted"] != false || restored["row_version"] != float64(3) || restored["deleted_at"] != nil || restored["deleted_by_user_id"] != nil {
		t.Fatalf("Host restore through both owner phases=%#v", restored)
	}
	if countRecordMutations(t, harness.DB, host, "soft_delete") != 1 || countRecordMutations(t, harness.DB, host, "restore") != 1 || countRows(t, harness.DB, `SELECT count(*) FROM host_grid_projection WHERE record_id=$1`, host) != 1 {
		t.Fatal("owner phases changed Host history or restored projection")
	}
}

func requirePartyDeleteReferencesAfterLockWait(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	party, evidence := uuid.New(), uuid.New()
	envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, party, "party")
	envelopetest.SeedRecordEnvelope(t, harness.DB, incident, actor, evidence, "evidence")
	mustExec(t, harness.DB, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Waited Party','organization')`, party, incident)
	mustExec(t, harness.DB, `INSERT INTO evidence(record_id,incident_id,title) VALUES($1,$2,'Incoming Evidence')`, evidence, incident)
	barrier, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer barrier.Rollback(context.Background())
	var holderPID int32
	if err := barrier.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
		t.Fatal(err)
	}
	var locked uuid.UUID
	if err := barrier.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR UPDATE`, party).Scan(&locked); err != nil || locked != party {
		t.Fatalf("lock Party: id=%s err=%v", locked, err)
	}
	body := map[string]any{"base_row_version": 1, "client_txn_id": "party-delete-after-lock-wait"}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+party.String(), strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	type result struct {
		response *http.Response
		err      error
	}
	finished := make(chan result, 1)
	go func() {
		response, err := http.DefaultClient.Do(request)
		finished <- result{response, err}
	}()
	consumed := false
	t.Cleanup(func() {
		cancel()
		if consumed {
			return
		}
		select {
		case outcome := <-finished:
			if outcome.response != nil {
				_ = outcome.response.Body.Close()
			}
		case <-time.After(5 * time.Second):
			t.Error("Party delete HTTP worker did not drain")
		}
	})
	for {
		var waiting bool
		if err := harness.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, holderPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case outcome := <-finished:
			consumed = true
			if outcome.response != nil {
				_ = outcome.response.Body.Close()
			}
			t.Fatalf("delete returned before waiting on the held Party: %v", outcome.err)
		case <-ctx.Done():
			t.Fatal("delete did not reach the held Party envelope", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if _, err := barrier.Exec(ctx, `UPDATE evidence SET collector_party_id=$2 WHERE record_id=$1`, evidence, party); err != nil {
		t.Fatal(err)
	}
	if err := barrier.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var outcome result
	select {
	case outcome = <-finished:
		consumed = true
	case <-ctx.Done():
		t.Fatal("delete did not finish after releasing Party", ctx.Err())
	}
	if outcome.err != nil {
		t.Fatal(outcome.err)
	}
	envelope := httptestx.RequireErrorEnvelope(t, outcome.response, http.StatusConflict, "record_delete_blocked")
	details := envelope["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "active_incoming_party_reference" {
		t.Fatalf("owner blocker=%#v", details)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=1 AND deleted_at IS NULL AND deleted_by_user_id IS NULL`, party) != 1 || countRows(t, harness.DB, `SELECT count(*) FROM record_revisions WHERE record_id=$1`, party) != 0 || countRecordMutations(t, harness.DB, party, "soft_delete") != 0 {
		t.Fatal("blocked delete changed the envelope or retained history")
	}
	mustExec(t, harness.DB, `UPDATE evidence SET collector_party_id=NULL WHERE record_id=$1`, evidence)
	success := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, party, body), http.StatusOK)["data"].(map[string]any)
	if success["row_version"] != float64(2) || success["deleted"] != true || countRecordMutations(t, harness.DB, party, "soft_delete") != 1 {
		t.Fatalf("same-request retry after clearing reference failed: %#v", success)
	}
}

func TestRestoreTombstonePreconditions_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "history_revision-u-7-04-restore", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U704")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	var clockBase time.Time
	if err := harness.DB.QueryRow(`SELECT clock_timestamp() + interval '5 minutes'`).Scan(&clockBase); err != nil {
		t.Fatal(err)
	}
	httptestx.SetClockFixed(t, harness.Server, clockBase)
	deletePayload := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-u-7-04-delete"}), http.StatusOK)["data"].(map[string]any)
	tombstoneVersion := int64(deletePayload["row_version"].(float64))
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM entity_active_identifier_claims WHERE record_id = $1`, recordID); got != 0 {
		t.Fatalf("deleted Host retained %d active identifier claims", got)
	}

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

	httptestx.SetClockFixed(t, harness.Server, clockBase.Add(time.Minute))
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
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM entity_active_identifier_claims WHERE record_id = $1`, recordID); got != 1 {
		t.Fatalf("restored Host reacquired %d active identifier claims, want 1", got)
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

	blockedRecordID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, blockedRecordID, "Blocked Restore Host", "blocked-restore-host", "", "")
	blockedDelete := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, blockedRecordID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-u-7-04-delete-blocked",
	}), http.StatusOK)["data"].(map[string]any)
	blockingRecordID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, blockingRecordID, "Blocking Active Host", "blocked-restore-host", "", "")
	blockedVersion := int64(blockedDelete["row_version"].(float64))
	blocked := restoreRecord(t, harness, login, blockedRecordID, map[string]any{
		"base_row_version": blockedVersion,
		"client_txn_id":    "txn-u-7-04-restore-blocked",
	})
	blockedEnvelope := httptestx.RequireErrorEnvelope(t, blocked, http.StatusConflict, "record_restore_blocked")
	details := blockedEnvelope["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "active_entity_identifier_conflict" ||
		details["entity_type"] != "host" ||
		details["identifier_class"] != "hostname" ||
		details["normalized_value"] != "blocked-restore-host" ||
		details["blocking_record_id"] != blockingRecordID.String() {
		t.Fatalf("unsafe or incomplete restore conflict details: %#v", details)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND row_version = $2 AND deleted_at IS NOT NULL`, blockedRecordID, blockedVersion); got != 1 {
		t.Fatalf("blocked restore changed the tombstone envelope, count=%d", got)
	}
	if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE target_id = $1 AND operation_kind = 'restore'`, blockedRecordID.String()); got != 0 {
		t.Fatalf("blocked restore appended %d history mutations", got)
	}

	notDeleted := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": tombstoneVersion + 1, "client_txn_id": "txn-u-7-04-not-deleted"})
	httptestx.RequireErrorEnvelope(t, notDeleted, http.StatusConflict, "record_not_deleted")
}

func deleteRecord(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
}

func restoreRecord(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/restore", body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
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
	collaborationsupport.RequireSingleRecordChangedIntent(t, db, changeSetID, recordID.String(), changeKind)
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
INSERT INTO decisions (record_id, incident_id, summary, status, decision_type, owner_user_id, decided_at, rationale)
VALUES ($1, $2, 'Matrix decision', 'proposed', 'containment', $3, '2026-01-01T00:00:00Z', 'Matrix decision rationale')
`, recordIDs["decision"], incidentID, actorID)
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
INSERT INTO task_requests (record_id, incident_id, title, status, priority, task_kind, owner_user_id)
VALUES ($1, $2, 'Matrix task', 'open', 'normal', 'request', $3)
`, recordIDs["task_request"], incidentID, actorID)
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
	indicatortest.SeedRecord(t, db, incidentID, actorUserID, recordID, "domain_name", "atomic", "history-revision.example.test")
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
	DeletedRecords       int
	ChangeSets           int
	Mutations            int
	Revisions            int
	Idempotency          int
	ProjectionRows       int
	CollaborationIntents int
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
		CollaborationIntents: collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{
			SourceRecordID: recordID.String(),
		}),
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

func requireTaskDecisionHistoricalStoredText(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any, recordType string) {
	t.Helper()
	fields := []struct {
		kind, field         string
		limit               int
		optional, multiline bool
	}{
		{"task_request", "title", 512, false, false},
		{"task_request", "workstream", 512, true, false},
		{"task_request", "requester_party_text", 256, true, false},
		{"task_request", "blocked_reason", 4096, true, true},
		{"task_request", "external_ticket_ref", 1024, true, false},
		{"task_request", "closure_summary", 16384, true, true},
		{"decision", "summary", 512, false, false},
		{"decision", "rationale", 16384, false, true},
	}
	for _, field := range fields {
		if field.kind != recordType {
			continue
		}
		t.Run("stored text "+field.field, func(t *testing.T) {
			check := func(value any, valid bool) {
				t.Helper()
				source := maps.Clone(captured)
				source[field.field] = value
				if field.field == "blocked_reason" && value != nil {
					source["status"] = "blocked"
				}
				before, err := json.Marshal(source)
				if err != nil {
					t.Fatal(err)
				}
				err = validate(record, source)
				if (err == nil) != valid {
					t.Fatalf("stored %s validity=%t want=%t err=%v", field.field, err == nil, valid, err)
				}
				after, err := json.Marshal(source)
				if err != nil || string(before) != string(after) {
					t.Fatal("stored text was rewritten")
				}
			}
			check(nil, field.optional)
			for _, valid := range []string{"Café  MiXeD\u200dText", strings.Repeat("😀", field.limit), strings.Repeat("é", field.limit)} {
				check(valid, true)
			}
			for _, invalid := range []string{"", " ", " leading", "trailing\u00a0", "Cafe\u0301", strings.Repeat("😀", field.limit+1), "A\r\nB", "A\rB"} {
				check(invalid, false)
			}
			check("A\n\t B", field.multiline)
			for control := rune(0); control <= 159; control++ {
				if control > 31 && control < 127 {
					continue
				}
				if field.multiline && (control == '\n' || control == '\t') {
					continue
				}
				for _, value := range []string{string(control) + "Text", "Te" + string(control) + "xt", "Text" + string(control)} {
					check(value, false)
				}
			}
		})
	}
}

func requireEvidenceHistoricalLifecycle(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	type tuple struct {
		lifecycle, upload string
		linked            bool
	}
	legal := []tuple{
		{"requested", "pending", false}, {"pending_receipt", "pending", false}, {"received", "pending", false}, {"quarantined", "pending", false},
		{"requested", "available", true}, {"pending_receipt", "available", true}, {"received", "available", true}, {"available", "available", true}, {"released", "available", true},
		{"requested", "quarantined", true}, {"pending_receipt", "quarantined", true}, {"received", "quarantined", true}, {"quarantined", "quarantined", true},
	}
	check := func(value tuple, valid bool) {
		t.Helper()
		source := maps.Clone(captured)
		source["lifecycle_state"], source["upload_state"] = value.lifecycle, value.upload
		source["object_blob_id"], source["blob_hash"], source["storage_ref"] = nil, nil, nil
		if value.linked {
			source["object_blob_id"] = "abcdef00-1234-4000-8000-000000000001"
			source["storage_ref"] = "object://abcdef00-1234-4000-8000-000000000001"
			source["blob_hash"] = strings.Repeat("a", 64)
		}
		before, err := json.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		err = validate(record, source)
		if (err == nil) != valid {
			t.Fatalf("historical Evidence tuple=%+v valid=%t want=%t err=%v", value, err == nil, valid, err)
		}
		after, err := json.Marshal(source)
		if err != nil || string(before) != string(after) {
			t.Fatal("historical Evidence was rewritten")
		}
	}
	for _, value := range legal {
		check(value, true)
	}
	for _, bad := range []string{"", "unknown", "Available", "available ", "available\x00"} {
		check(tuple{bad, "pending", false}, false)
		check(tuple{"requested", bad, false}, false)
	}
	for _, lifecycle := range []string{"available", "released"} {
		check(tuple{lifecycle, "pending", false}, false)
		check(tuple{lifecycle, "available", false}, false)
		for _, upload := range []string{"pending", "failed", "quarantined"} {
			check(tuple{lifecycle, upload, true}, false)
		}
	}
	for _, upload := range []string{"pending", "available", "failed"} {
		check(tuple{"quarantined", upload, true}, false)
	}
}

func requireEvidenceHistoricalStoredText(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	for _, test := range []struct {
		name  string
		value any
		valid bool
	}{
		{"null", nil, true}, {"digest", strings.Repeat("a", 64), true},
		{"leading NUL", "\x00digest", false}, {"internal NUL", "dig\x00est", false}, {"trailing NUL", "digest\x00", false}, {"invalid UTF8", string([]byte{0xff}), false},
	} {
		t.Run("Evidence stored hash "+test.name, func(t *testing.T) {
			source := maps.Clone(captured)
			source["blob_hash"] = test.value
			before := maps.Clone(source)
			err := validate(record, source)
			if (err == nil) != test.valid {
				t.Fatalf("stored hash valid=%v want=%v err=%v", err == nil, test.valid, err)
			}
			if !reflect.DeepEqual(before, source) {
				t.Fatal("stored hash was rewritten")
			}
		})
	}

	for _, field := range []struct {
		name  string
		limit int
	}{{"title", 512}, {"storage_ref", 1024}, {"collector_party_text", 256}, {"source_party_text", 256}} {
		t.Run("Evidence stored text "+field.name, func(t *testing.T) {
			check := func(value any, valid bool) {
				t.Helper()
				source := maps.Clone(captured)
				source[field.name] = value
				if value == "object://abcdef00-1234-4000-8000-000000000001" {
					source["object_blob_id"] = "abcdef00-1234-4000-8000-000000000001"
					source["upload_state"] = "available"
					source["blob_hash"] = strings.Repeat("a", 64)
				}
				before, err := json.Marshal(source)
				if err != nil {
					t.Fatal(err)
				}
				err = validate(record, source)
				if (err == nil) != valid {
					t.Fatalf("stored Evidence %s valid=%t want=%t err=%v", field.name, err == nil, valid, err)
				}
				after, err := json.Marshal(source)
				if err != nil || string(before) != string(after) {
					t.Fatal("historical Evidence text changed")
				}
			}
			check(nil, true)
			for _, value := range []string{"Café  MiXeD\u200dText", strings.Repeat("😀", field.limit), strings.Repeat("é", field.limit)} {
				check(value, true)
			}
			if field.name == "storage_ref" {
				check("object://abcdef00-1234-4000-8000-000000000001", true)
			}
			for _, value := range []string{"", " ", " leading", "trailing\u00a0", "Cafe\u0301", strings.Repeat("😀", field.limit+1), "A\r\nB", "A\rB"} {
				check(value, false)
			}
			for control := rune(0); control <= 159; control++ {
				if control > 31 && control < 127 {
					continue
				}
				for _, value := range []string{string(control) + "Text", "Te" + string(control) + "xt", "Text" + string(control)} {
					check(value, false)
				}
			}
		})
	}
}

func requireEvidenceHistoricalStorageReference(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	const blobID = "abcdef00-1234-4000-8000-000000000001"
	for _, test := range []struct {
		name   string
		ref    any
		linked bool
		valid  bool
	}{
		{"null unlinked", nil, false, true},
		{"null linked", nil, true, true},
		{"external unlinked", "ticket://MiXeD/retained", false, true},
		{"external linked", "ticket://MiXeD/retained", true, true},
		{"matching", "object://abcdef00-1234-4000-8000-000000000001", true, true},
		{"orphan", "object://abcdef00-1234-4000-8000-000000000001", false, false},
		{"mismatch", "object://abcdef00-1234-4000-8000-000000000002", true, false},
		{"empty identity", "object://", true, false},
		{"invalid identity", "object://not-a-uuid", true, false},
		{"nil identity", "object://00000000-0000-0000-0000-000000000000", true, false},
		{"uppercase identity", "object://ABCDEF00-1234-4000-8000-000000000001", true, false},
		{"compact identity", "object://abcdef00123440008000000000000001", true, false},
		{"extra path", "object://abcdef00-1234-4000-8000-000000000001/object", true, false},
	} {
		t.Run("Evidence historical reference "+test.name, func(t *testing.T) {
			source := maps.Clone(captured)
			source["lifecycle_state"], source["upload_state"] = "requested", "pending"
			source["object_blob_id"], source["blob_hash"], source["storage_ref"] = nil, nil, test.ref
			if test.linked {
				source["object_blob_id"], source["blob_hash"], source["upload_state"] = blobID, strings.Repeat("a", 64), "available"
			}
			before, err := json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			err = validate(record, source)
			if (err == nil) != test.valid {
				t.Fatalf("historical reference valid=%t want=%t err=%v", err == nil, test.valid, err)
			}
			after, err := json.Marshal(source)
			if err != nil || string(before) != string(after) {
				t.Fatal("historical reference source changed")
			}
		})
	}
}

func requireIndicatorHistoricalRepresentation(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	base := maps.Clone(captured)
	base["indicator_type"], base["value_kind"] = "text", "atomic"
	base["display_value"], base["normalized_value"], base["dedupe_key"] = "MiXeD\n\tCafé\u200dtext", "MiXeD\n\tCafé\u200dtext", "text:retained-key"
	base["hash_algorithm"], base["hash_value"], base["defanged_value"], base["stix_pattern"] = nil, nil, nil, nil
	check := func(name string, changes map[string]any, valid bool) {
		t.Helper()
		t.Run("Indicator historical representation "+name, func(t *testing.T) {
			source := maps.Clone(base)
			maps.Copy(source, changes)
			before, err := json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			err = validate(record, source)
			if (err == nil) != valid {
				t.Fatalf("historical Indicator valid=%t want=%t err=%v", err == nil, valid, err)
			}
			after, err := json.Marshal(source)
			if err != nil || string(before) != string(after) {
				t.Fatal("historical Indicator source changed")
			}
		})
	}
	for _, kind := range []string{"atomic", "pattern", "reference"} {
		check("text "+kind, map[string]any{"value_kind": kind}, true)
	}
	for _, example := range []struct{ typ, value string }{
		{"ipv4_addr", "203.0.113.7"}, {"ipv6_addr", "2001:db8::1"}, {"domain_name", "example.test"},
		{"url", "https://example.test/Path"}, {"sha256", strings.Repeat("a", 64)}, {"email_addr", "User@example.test"},
		{"registry_key", `HKEY_LOCAL_MACHINE\SOFTWARE\EXAMPLE`}, {"process_name", "PowerShell.EXE"},
	} {
		check(example.typ, map[string]any{"indicator_type": example.typ, "display_value": example.value, "normalized_value": example.value, "dedupe_key": example.typ + ":" + example.value}, true)
	}
	for _, field := range []string{"indicator_type", "value_kind"} {
		for _, bad := range []string{"", "unknown", "Atomic", "text ", "literal", "ip_addr"} {
			check(field+" "+bad, map[string]any{field: bad}, false)
		}
	}
	for _, field := range []string{"display_value", "normalized_value"} {
		check(field+" scalar boundary", map[string]any{field: strings.Repeat("😀", 8192)}, true)
		check(field+" oversized", map[string]any{field: strings.Repeat("😀", 8193)}, false)
	}
	for _, field := range []string{"display_value", "normalized_value", "dedupe_key"} {
		for _, bad := range []any{nil, "", " \t\n"} {
			check(field+" empty", map[string]any{field: bad}, false)
		}
	}
	for _, field := range []string{"display_value", "normalized_value", "dedupe_key", "defanged_value", "hash_algorithm", "hash_value", "stix_pattern"} {
		check(field+" NUL", map[string]any{field: "A\x00B"}, false)
	}
	for _, field := range []string{"defanged_value", "stix_pattern"} {
		for _, value := range []any{nil, "", "  Presentation\n\tMiXeD\u200d  "} {
			check(field+" preserved", map[string]any{field: value}, true)
		}
	}
	for _, pair := range []struct {
		algorithm, hash any
		valid           bool
	}{
		{nil, nil, true}, {"sha256", "abc012", true}, {"future-algorithm", "09af", true},
		{"sha256", nil, false}, {nil, "abc012", false}, {"", "abc012", false}, {"sha256", "", false},
		{"SHA256", "abc012", false}, {" sha256", "abc012", false}, {"sha256 ", "abc012", false},
		{"sha256", "ABC012", false}, {"sha256", "abc012 ", false}, {"sha256", "xyz", false},
	} {
		check("hash pair", map[string]any{"hash_algorithm": pair.algorithm, "hash_value": pair.hash}, pair.valid)
	}
	for _, ip := range []struct{ typ, value string }{{"ipv4_addr", "203.0.113.7"}, {"ipv6_addr", "2001:db8::1"}} {
		for _, kind := range []string{"pattern", "reference"} {
			check(ip.typ+" "+kind, map[string]any{"indicator_type": ip.typ, "value_kind": kind, "display_value": ip.value, "normalized_value": ip.value}, false)
		}
		check(ip.typ+" mismatched normalized", map[string]any{"indicator_type": ip.typ, "display_value": ip.value, "normalized_value": "other"}, false)
		check(ip.typ+" hash pair", map[string]any{"indicator_type": ip.typ, "display_value": ip.value, "normalized_value": ip.value, "hash_algorithm": "sha256", "hash_value": "abc012"}, false)
	}
}

func requireTimelineHistoricalIntrinsicState(t *testing.T, validate revisions.SnapshotSourceValidator, record, captured map[string]any) {
	t.Helper()
	base := maps.Clone(captured)
	base["capture_state"], base["row_version"], base["activity_time_pair_state"] = "rough", json.Number("1"), "disabled"
	for _, field := range []string{"reviewed_by_user_id", "reviewed_at", "superseded_by_user_id", "superseded_at", "activity_utc_text", "activity_local_text"} {
		base[field] = nil
	}
	base["activity_utc_generated"], base["activity_local_generated"] = false, false
	check := func(changes map[string]any, valid bool) {
		t.Helper()
		source := maps.Clone(base)
		maps.Copy(source, changes)
		before, err := json.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		err = validate(record, source)
		if (err == nil) != valid {
			t.Fatalf("historical Timeline changes=%#v valid=%t want=%t err=%v", changes, err == nil, valid, err)
		}
		after, err := json.Marshal(source)
		if err != nil || string(before) != string(after) {
			t.Fatal("historical Timeline source changed")
		}
	}
	const actor = "abcdef00-1234-4000-8000-000000000001"
	const at = "2026-01-01T12:00:00Z"
	for _, state := range []string{"rough", "enriched", "reviewed", "superseded"} {
		for reviewMask := 0; reviewMask < 4; reviewMask++ {
			for supersedeMask := 0; supersedeMask < 4; supersedeMask++ {
				changes := map[string]any{"capture_state": state}
				if reviewMask&1 != 0 {
					changes["reviewed_by_user_id"] = actor
				}
				if reviewMask&2 != 0 {
					changes["reviewed_at"] = at
				}
				if supersedeMask&1 != 0 {
					changes["superseded_by_user_id"] = actor
				}
				if supersedeMask&2 != 0 {
					changes["superseded_at"] = at
				}
				valid := ((state == "rough" || state == "enriched") && reviewMask == 0 && supersedeMask == 0) ||
					(state == "reviewed" && reviewMask == 3 && supersedeMask == 0) ||
					(state == "superseded" && (reviewMask == 0 || reviewMask == 3) && supersedeMask == 3)
				check(changes, valid)
			}
		}
	}
	for _, field := range []string{"capture_state", "activity_time_pair_state"} {
		for _, bad := range []string{"", "unknown", "Rough", "rough ", "linked", "rolled back"} {
			check(map[string]any{field: bad}, false)
		}
	}
	for _, version := range []string{"1", "9007199254740993", "9223372036854775807"} {
		check(map[string]any{"row_version": json.Number(version)}, true)
	}
	for _, version := range []string{"0", "-1"} {
		check(map[string]any{"row_version": json.Number(version)}, false)
	}
	for _, field := range []string{"date_entered_text", "analyst_text", "mitre_stage_text", "device_object_text", "ip_address_text", "activity_utc_text", "activity_local_text", "raw_activity_text", "activity_synopsis_text", "data_source_text"} {
		t.Run("Timeline historical text "+field, func(t *testing.T) {
			for _, value := range []any{nil, "", "  Cafe\u0301\r\n\tMiXeD\u200d  ", "<script>inert</script> =SUM(A1) https://example.test", strings.Repeat("😀", 32768)} {
				check(map[string]any{field: value}, true)
			}
			check(map[string]any{field: strings.Repeat("😀", 32769)}, false)
			check(map[string]any{field: string([]byte{0xff})}, false)
			for control := rune(0); control <= 159; control++ {
				if (control > 31 && control < 127) || control == '\t' || control == '\n' || control == '\r' {
					continue
				}
				for _, value := range []string{string(control) + "Text", "Te" + string(control) + "xt", "Text" + string(control)} {
					check(map[string]any{field: value}, false)
				}
			}
		})
	}
	const utc, local = "2026-01-01T12:00:00Z", "2026-01-01T07:00:00-05:00"
	for _, pair := range []struct {
		state              string
		utc, local         any
		utcFlag, localFlag bool
		valid              bool
	}{
		{"disabled", nil, nil, false, false, true}, {"disabled", utc, local, false, true, true},
		{"conversion_unavailable", nil, local, false, true, true}, {"conversion_unavailable", utc, nil, true, false, true},
		{"empty", nil, "", false, false, true}, {"empty", utc, nil, false, false, false},
		{"paired_generated", utc, local, false, true, true}, {"paired_generated", utc, local, true, false, true},
		{"paired_generated", utc, local, false, false, false}, {"paired_generated", utc, nil, true, false, false},
		{"paired_user_preserved", utc, local, false, false, true}, {"paired_user_preserved", utc, local, false, true, false},
		{"paired_mismatch", utc, "2026-01-01T08:00:00-05:00", false, false, true},
		{"paired_mismatch", utc, "2026-01-01T08:00:00-05:00", true, false, true},
		{"paired_mismatch", nil, local, false, false, false},
		{"disabled", utc, local, true, true, false}, {"disabled", nil, local, true, false, false},
		{"disabled", utc, "", false, true, false}, {"disabled", "2026-01-01 12:00:00Z", local, true, false, false},
		{"disabled", "2026-01-01T12:00Z", local, true, false, false}, {"disabled", utc, "2026-01-01T07:00:00", false, true, false},
		{"disabled", utc, "2026-01-01T07:00-05:00", false, true, false}, {"disabled", "bad", local, true, false, false},
	} {
		check(map[string]any{"activity_time_pair_state": pair.state, "activity_utc_text": pair.utc, "activity_local_text": pair.local, "activity_utc_generated": pair.utcFlag, "activity_local_generated": pair.localFlag}, pair.valid)
	}
}
