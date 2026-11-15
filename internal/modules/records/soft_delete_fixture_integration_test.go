package records

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// The shared SQL tombstone fixture must leave envelopes that the Records
// snapshot validator accepts (Core 01 REQ-01-651) and that the Host/Identity
// mirror triggers accept at commit.
func TestSoftDeleteRecordFixtureKeepsEnvelopeLegal_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "records-soft-delete-fixture")
	actorID, incidentID := seedEnvelopeOwnerContext(t, db, "soft-delete-fixture")
	createdAt := time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC)

	t.Run("database clock tombstone advances updated_at and keeps the host mirror exact", func(t *testing.T) {
		recordID := seedTombstoneFixtureEnvelope(t, db, incidentID, actorID, "host", createdAt, createdAt, 1)
		if _, err := db.Exec(ctx, `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state, row_version,
    created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT r.record_id, r.incident_id, 'Tombstone fixture host', 'tombstone-fixture-host', 'canonical',
       r.row_version, r.created_at, r.updated_at, r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
`, recordID); err != nil {
			t.Fatalf("seed host mirror: %v", err)
		}

		envelopetest.SoftDeleteRecord(t, db, envelopetest.Tombstone{RecordID: recordID, DeletedBy: actorID})

		got := readTombstoneFixtureEnvelope(t, db, recordID)
		if got.deletedAt == nil || !got.deletedAt.After(createdAt) || !got.updatedAt.Equal(*got.deletedAt) ||
			got.deletedBy == nil || *got.deletedBy != actorID || got.rowVersion != 1 {
			t.Fatalf("database-clock tombstone = %+v; want deleted after %s with updated_at equal to deleted_at, version 1", got, createdAt)
		}
		var mirrorVersion int64
		var mirrorUpdatedAt time.Time
		if err := db.QueryRow(ctx, `SELECT row_version, updated_at FROM hosts WHERE record_id = $1`, recordID).Scan(&mirrorVersion, &mirrorUpdatedAt); err != nil {
			t.Fatalf("read host mirror: %v", err)
		}
		if mirrorVersion != got.rowVersion || !mirrorUpdatedAt.Equal(got.updatedAt) {
			t.Fatalf("host mirror version=%d updated_at=%s; want %d %s", mirrorVersion, mirrorUpdatedAt, got.rowVersion, got.updatedAt)
		}
		// Fire the deferred Host/Identity envelope triggers now instead of at a
		// commit that the rollback database never performs.
		if _, err := db.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
			t.Fatalf("deferred envelope mirror triggers rejected the tombstone: %v", err)
		}
		requireTombstoneFixtureSnapshotLegal(t, incidentID, recordID, "host", got)
	})

	t.Run("explicit instant keeps a later update and raises the requested version", func(t *testing.T) {
		updatedAt := createdAt.Add(10 * time.Minute)
		deletedAt := createdAt.Add(5 * time.Minute)
		recordID := seedTombstoneFixtureEnvelope(t, db, incidentID, actorID, "timeline_event", createdAt, updatedAt, 3)

		envelopetest.SoftDeleteRecord(t, db, envelopetest.Tombstone{RecordID: recordID, DeletedBy: actorID, At: deletedAt, RowVersion: 4})

		got := readTombstoneFixtureEnvelope(t, db, recordID)
		if got.deletedAt == nil || !got.deletedAt.Equal(deletedAt) || !got.updatedAt.Equal(updatedAt) || got.rowVersion != 4 {
			t.Fatalf("explicit tombstone = %+v; want deleted_at %s, updated_at %s, version 4", got, deletedAt, updatedAt)
		}
		requireTombstoneFixtureSnapshotLegal(t, incidentID, recordID, "timeline_event", got)
	})

	t.Run("barrier tombstone at the last update leaves the envelope and host mirror untouched", func(t *testing.T) {
		updatedAt := createdAt.Add(7 * time.Minute)
		recordID := seedTombstoneFixtureEnvelope(t, db, incidentID, actorID, "host", createdAt, updatedAt, 2)
		if _, err := db.Exec(ctx, `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state, row_version,
    created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT r.record_id, r.incident_id, 'Barrier fixture host', 'barrier-fixture-host', 'canonical',
       r.row_version, r.created_at, r.updated_at, r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
`, recordID); err != nil {
			t.Fatalf("seed barrier host mirror: %v", err)
		}
		var mirrorBefore string
		if err := db.QueryRow(ctx, `SELECT xmin::text FROM hosts WHERE record_id = $1`, recordID).Scan(&mirrorBefore); err != nil {
			t.Fatalf("read host mirror version: %v", err)
		}

		envelopetest.SoftDeleteRecord(t, db, envelopetest.Tombstone{RecordID: recordID, DeletedBy: actorID, AtLastUpdate: true})

		got := readTombstoneFixtureEnvelope(t, db, recordID)
		if got.deletedAt == nil || !got.deletedAt.Equal(updatedAt) || !got.updatedAt.Equal(updatedAt) || got.rowVersion != 2 {
			t.Fatalf("barrier tombstone = %+v; want deleted_at = updated_at = %s, version 2", got, updatedAt)
		}
		var mirrorAfter string
		if err := db.QueryRow(ctx, `SELECT xmin::text FROM hosts WHERE record_id = $1`, recordID).Scan(&mirrorAfter); err != nil {
			t.Fatalf("read host mirror version: %v", err)
		}
		if mirrorAfter != mirrorBefore {
			t.Fatalf("barrier tombstone rewrote the unchanged host mirror: xmin %s -> %s", mirrorBefore, mirrorAfter)
		}
		if _, err := db.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
			t.Fatalf("deferred envelope mirror triggers rejected the barrier tombstone: %v", err)
		}
		requireTombstoneFixtureSnapshotLegal(t, incidentID, recordID, "host", got)
	})

	t.Run("barrier tombstone writes no mirror-checked column on an envelope-only host", func(t *testing.T) {
		recordID := seedTombstoneFixtureEnvelope(t, db, incidentID, actorID, "host", createdAt, createdAt.Add(time.Minute), 1)

		envelopetest.SoftDeleteRecord(t, db, envelopetest.Tombstone{RecordID: recordID, DeletedBy: actorID, AtLastUpdate: true})

		// The Host has no source row; firing the deferred envelope mirror
		// trigger would reject it, so the barrier must not name its columns.
		if _, err := db.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
			t.Fatalf("barrier tombstone fired the envelope mirror trigger: %v", err)
		}
		got := readTombstoneFixtureEnvelope(t, db, recordID)
		if got.deletedAt == nil || !got.deletedAt.Equal(createdAt.Add(time.Minute)) || !got.updatedAt.Equal(createdAt.Add(time.Minute)) || got.rowVersion != 1 {
			t.Fatalf("envelope-only barrier tombstone = %+v", got)
		}
		requireTombstoneFixtureSnapshotLegal(t, incidentID, recordID, "host", got)
	})

	t.Run("an instant before creation or a missing record fails without writing", func(t *testing.T) {
		recordID := seedTombstoneFixtureEnvelope(t, db, incidentID, actorID, "timeline_event", createdAt, createdAt, 1)
		early := runTombstoneFixtureRecording(t, func(tb testing.TB) {
			envelopetest.SoftDeleteRecord(tb, db, envelopetest.Tombstone{RecordID: recordID, DeletedBy: actorID, At: createdAt.Add(-time.Second)})
		})
		if !early.failed || !strings.Contains(early.message, "created after the deletion instant") {
			t.Fatalf("early tombstone failure = %v %q; want a fail-closed creation-order failure", early.failed, early.message)
		}
		got := readTombstoneFixtureEnvelope(t, db, recordID)
		if got.deletedAt != nil || got.deletedBy != nil || !got.updatedAt.Equal(createdAt) || got.rowVersion != 1 {
			t.Fatalf("rejected tombstone changed the envelope: %+v", got)
		}
		missing := runTombstoneFixtureRecording(t, func(tb testing.TB) {
			envelopetest.SoftDeleteRecord(tb, db, envelopetest.Tombstone{RecordID: uuid.New(), DeletedBy: actorID})
		})
		if !missing.failed || !strings.Contains(missing.message, "is missing") {
			t.Fatalf("missing tombstone failure = %v %q; want a fail-closed missing-record failure", missing.failed, missing.message)
		}
	})
}

type tombstoneFixtureEnvelope struct {
	createdAt  time.Time
	updatedAt  time.Time
	deletedAt  *time.Time
	deletedBy  *uuid.UUID
	createdBy  uuid.UUID
	updatedBy  uuid.UUID
	rowVersion int64
}

func seedTombstoneFixtureEnvelope(t *testing.T, db dbport.DB, incidentID, actorID uuid.UUID, recordType string, createdAt, updatedAt time.Time, rowVersion int64) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	if _, err := db.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id,
    created_at, updated_at, row_version
) VALUES ($1, $2, $3, $4, $4, $5, $6, $7)
`, recordID, incidentID, recordType, actorID, createdAt, updatedAt, rowVersion); err != nil {
		t.Fatalf("seed %s envelope: %v", recordType, err)
	}
	return recordID
}

func readTombstoneFixtureEnvelope(t *testing.T, db dbport.DB, recordID uuid.UUID) tombstoneFixtureEnvelope {
	t.Helper()
	var got tombstoneFixtureEnvelope
	if err := db.QueryRow(context.Background(), `
SELECT created_at, updated_at, deleted_at, deleted_by_user_id, created_by_user_id, updated_by_user_id, row_version
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&got.createdAt, &got.updatedAt, &got.deletedAt, &got.deletedBy, &got.createdBy, &got.updatedBy, &got.rowVersion); err != nil {
		t.Fatalf("read tombstoned envelope: %v", err)
	}
	return got
}

func requireTombstoneFixtureSnapshotLegal(t *testing.T, incidentID, recordID uuid.UUID, recordType string, got tombstoneFixtureEnvelope) {
	t.Helper()
	clock := func(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
	value := map[string]any{
		"record_id":          recordID.String(),
		"incident_id":        incidentID.String(),
		"record_type":        recordType,
		"created_at":         clock(got.createdAt),
		"created_by_user_id": got.createdBy.String(),
		"updated_at":         clock(got.updatedAt),
		"updated_by_user_id": got.updatedBy.String(),
		"row_version":        got.rowVersion,
		"deleted_at":         nil,
		"deleted_by_user_id": nil,
	}
	if got.deletedAt != nil {
		value["deleted_at"] = clock(*got.deletedAt)
	}
	if got.deletedBy != nil {
		value["deleted_by_user_id"] = got.deletedBy.String()
	}
	if err := ValidateSnapshotEnvelope(incidentID, recordID, recordType, value); err != nil {
		t.Fatalf("tombstoned envelope rejected by the Records snapshot validator: %v", err)
	}
}

type tombstoneFixtureRecorder struct {
	testing.TB
	failed  bool
	message string
}

func (r *tombstoneFixtureRecorder) Helper() {}

func (r *tombstoneFixtureRecorder) Fatalf(format string, args ...any) {
	r.failed = true
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func runTombstoneFixtureRecording(t *testing.T, body func(testing.TB)) *tombstoneFixtureRecorder {
	t.Helper()
	recorder := &tombstoneFixtureRecorder{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		body(recorder)
	}()
	<-done
	return recorder
}
