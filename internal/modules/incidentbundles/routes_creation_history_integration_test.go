package incidentbundles_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

// The fixture explicitly names newly seeded version-one records. Production
// source capture and append APIs author their history before any test export;
// this does not discover, backfill or rewrite an existing retained chain.
func seedPortableCreationHistory(t testing.TB, h *appsupport.ServerHarness, ids ...uuid.UUID) {
	t.Helper()
	tx, err := h.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	appender := h.Revisions.Appender()
	for _, id := range ids {
		var incidentID, actorID uuid.UUID
		var kind string
		var createdAt time.Time
		var version int64
		var retained bool
		if err := tx.QueryRow(context.Background(), `SELECT incident_id,created_by_user_id,record_type,created_at,row_version,EXISTS(SELECT 1 FROM record_revisions WHERE record_id=$1) FROM records WHERE record_id=$1`, id).Scan(&incidentID, &actorID, &kind, &createdAt, &version, &retained); err != nil {
			t.Fatal(err)
		}
		if version != 1 || retained {
			t.Fatalf("creation fixture requires new version-one record %s", id)
		}
		snapshot, err := appender.CaptureRecordSnapshotTx(context.Background(), tx, id)
		if err != nil {
			t.Fatal(err)
		}
		changeSetID, err := appender.AppendChangeSetTx(context.Background(), tx, revisions.AppendChangeSetParams{IncidentID: incidentID, ActorUserID: actorID, Source: "fixture.record.create", CreatedAt: createdAt})
		if err != nil {
			t.Fatal(err)
		}
		afterVersion := fmt.Sprintf("%s:%s:1", kind, id)
		if err := appender.AppendRecordMutationTx(context.Background(), tx, revisions.AppendRecordMutationParams{ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "record", RecordID: id, OperationKind: "create", AfterVersionID: &afterVersion, AfterSnapshot: &snapshot}); err != nil {
			t.Fatal(err)
		}
		if err := appender.AppendHistoricalRevisionTx(context.Background(), tx, revisions.HistoricalRevisionInput{ChangeSetID: changeSetID, RecordID: id, RowVersion: 1, AfterSnapshot: &snapshot}); err != nil {
			t.Fatal(err)
		}
		// Complete the authored historical fixture inside its uncommitted setup.
		if _, err := tx.Exec(context.Background(), `UPDATE record_revisions SET created_at=$2 WHERE change_set_id=$1`, changeSetID, createdAt); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// This variant changes the fixture's single creation, not an existing deployment
// history. Requiring exactly one creation per member prevents broad rewriting.
func replaceEvidenceCreationFixtureHash(t testing.TB, files map[string][]byte, recordID, hash string) {
	t.Helper()
	for path, afterKey := range map[string]string{"data/change_set_mutations.ndjson": "after_value", "data/record_revisions.ndjson": "after_json"} {
		rows := decodeNDJSONRows(t, files[path])
		matched := 0
		for _, row := range rows {
			after, ok := row[afterKey].(map[string]any)
			if !ok {
				continue
			}
			record, ok := after["record"].(map[string]any)
			if !ok || record["record_id"] != recordID {
				continue
			}
			if record["record_type"] != "evidence" {
				t.Fatal("Evidence fixture matched another record type")
			}
			beforeKey := "before_json"
			if afterKey == "after_value" {
				beforeKey = "before_value"
				if row["operation_kind"] != "create" {
					t.Fatal("Evidence fixture is not creation-only")
				}
			}
			if row[beforeKey] != nil {
				t.Fatal("Evidence creation fixture has retained before state")
			}
			source := after["source"].(map[string]any)
			source["blob_hash"] = hash
			matched++
		}
		if matched != 1 {
			t.Fatalf("%s matched %d Evidence creations", path, matched)
		}
		files[path] = encodeNDJSONRows(t, rows)
	}
}

func seedPortableEvidenceQuarantine(t testing.TB, h *appsupport.ServerHarness, incidentID, recordID, actorID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tx, err := h.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	appender := h.Revisions.Appender()
	before, err := appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		t.Fatal(err)
	}
	var prior int64
	if err := tx.QueryRow(ctx, `SELECT row_version FROM records WHERE record_id=$1`, recordID).Scan(&prior); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := tx.Exec(ctx, `UPDATE object_blobs SET upload_state='quarantined' WHERE incident_id=$1 AND object_blob_id=(SELECT object_blob_id FROM evidence WHERE record_id=$2)`, incidentID, recordID); err != nil {
		t.Fatal(err)
	}
	if tag, err := tx.Exec(ctx, `UPDATE evidence SET lifecycle_state='quarantined',upload_state='quarantined',updated_at=$2 WHERE record_id=$1`, recordID, now); err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("quarantine fixture affected=%d error=%v", tag.RowsAffected(), err)
	}
	version, err := records.NewStore().AdvanceVersionTx(ctx, tx, recordID, actorID, now)
	if err != nil {
		t.Fatal(err)
	}
	after, err := appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{IncidentID: incidentID, ActorUserID: actorID, Source: "fixture.evidence.quarantine", CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	beforeVersion, afterVersion := fmt.Sprintf("evidence:%s:%d", recordID, prior), fmt.Sprintf("evidence:%s:%d", recordID, version)
	if err := appender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{ChangeSetID: changed, SequenceNo: 1, TargetKind: "evidence", RecordID: recordID, OperationKind: "patch", BeforeVersionID: &beforeVersion, AfterVersionID: &afterVersion, BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
		t.Fatal(err)
	}
	if err := appender.AppendHistoricalRevisionTx(ctx, tx, revisions.HistoricalRevisionInput{ChangeSetID: changed, RecordID: recordID, RowVersion: version, BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
