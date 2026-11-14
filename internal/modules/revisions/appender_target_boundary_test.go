package revisions_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseAppenderTargetBoundaries(t *testing.T) {
	h := appsupport.StartStore(t, "append-target-boundary")
	actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "append-target@example.test", DisplayName: "Append target"})
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "append-target-incident", "IR-TARGET", "Append target")
	record := uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, record, "Boundary Host", "boundary-host", "", "")
	appender := revisionsupport.MustComposition(t).Runtime.Appender()
	for _, tc := range []struct {
		name, kind string
		typed      bool
		wantError  error
	}{
		{"raw host", "host", false, revisions.ErrInvalidTargetSemantics},
		{"raw generic record", "record", false, revisions.ErrInvalidTargetSemantics},
		{"typed wrong record family", "indicator", true, revisions.ErrInvalidTargetSemantics},
		{"typed non-row target", "entity_alias", true, revisions.ErrInvalidTargetSemantics},
		{"unknown raw target", "unregistered_target", false, revisions.ErrMissingTargetSemantics},
		{"unknown typed target", "unregistered_target", true, revisions.ErrMissingTargetSemantics},
		{"typed generic record", "record", true, nil},
		{"owned non-row alias", "entity_alias", false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := h.DB.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			now := time.Now().UTC().Truncate(time.Microsecond)
			changed, err := appender.AppendChangeSetTx(t.Context(), tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "append.target.test", CreatedAt: now})
			if err != nil {
				t.Fatal(err)
			}
			targetID := record.String()
			var expectedAlias map[string]any
			if tc.typed {
				snapshot, captureErr := appender.CaptureRecordSnapshotTx(t.Context(), tx, record)
				if captureErr != nil {
					t.Fatal(captureErr)
				}
				err = appender.AppendRecordMutationTx(t.Context(), tx, revisions.AppendRecordMutationParams{ChangeSetID: changed, SequenceNo: 1, TargetKind: tc.kind, RecordID: record, OperationKind: "create", AfterSnapshot: &snapshot})
			} else {
				value := map[string]any{"display_name": "uncaptured raw row"}
				if tc.kind == "entity_alias" {
					alias := uuid.New()
					targetID = "entity_alias:" + alias.String()
					expectedAlias = expectedAliasHistory(alias, incident.ID, record, actor.ID, "host", "Boundary Alias", now)
					value = expectedAlias
					if _, err := tx.Exec(t.Context(), `INSERT INTO entity_aliases
(entity_alias_id,incident_id,record_id,entity_type,raw_text,normalized_text,classification,created_by_user_id,created_at)
VALUES($1,$2,$3,'host','Boundary Alias','Boundary Alias','suggestion_only',$4,$5)`, alias, incident.ID, record, actor.ID, now); err != nil {
						t.Fatal(err)
					}
				}
				err = appender.AppendNonRowMutationTx(t.Context(), tx, revisions.AppendNonRowMutationParams{ChangeSetID: changed, SequenceNo: 1, TargetKind: tc.kind, TargetID: targetID, OperationKind: "create", AfterValue: value})
			}
			if !errors.Is(err, tc.wantError) {
				t.Errorf("append error=%v; want %v", err, tc.wantError)
			}
			var count int
			if err := tx.QueryRow(t.Context(), "SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1", changed).Scan(&count); err != nil {
				t.Fatal(err)
			}
			wantCount := 0
			if tc.wantError == nil {
				wantCount = 1
			}
			if count != wantCount {
				t.Fatalf("persisted mutations=%d; want %d", count, wantCount)
			}
			if tc.wantError != nil {
				return
			}
			var metadataExact bool
			entryIDs := []uuid.UUID{}
			if tc.typed {
				entryIDs = []uuid.UUID{record}
			}
			if err := tx.QueryRow(t.Context(), `SELECT sequence_no=1 AND target_kind=$2 AND target_id=$3 AND operation_kind='create'
AND coalesce(before_value,'null'::jsonb)='null'::jsonb AND before_version_id IS NULL AND after_version_id IS NULL
AND history_record_ids=ARRAY[$4::uuid] AND history_entry_record_ids=$5::uuid[]
FROM change_set_mutations WHERE change_set_id=$1`, changed, tc.kind, targetID, record, entryIDs).Scan(&metadataExact); err != nil || !metadataExact {
				t.Fatalf("persisted metadata exact=%v error=%v", metadataExact, err)
			}
			var valueExact bool
			if tc.typed {
				err = tx.QueryRow(t.Context(), `SELECT
(SELECT count(*) FROM jsonb_object_keys(m.after_value))=3
AND m.after_value->>'snapshot_schema_id'='cartulary.revisions.snapshot.host.v1'
AND (m.after_value->'record')-'created_at'-'updated_at'-'deleted_at'=to_jsonb(r)-'created_at'-'updated_at'-'deleted_at'
AND (m.after_value->'record'->>'created_at')::timestamptz=r.created_at
AND (m.after_value->'record'->>'updated_at')::timestamptz=r.updated_at
AND (m.after_value->'record'->>'deleted_at')::timestamptz IS NOT DISTINCT FROM r.deleted_at
AND (m.after_value->'source')-'created_at'-'updated_at'=to_jsonb(h)-'created_at'-'updated_at'
AND (m.after_value->'source'->>'created_at')::timestamptz=h.created_at
AND (m.after_value->'source'->>'updated_at')::timestamptz=h.updated_at
FROM change_set_mutations m JOIN records r ON r.record_id=$2 JOIN hosts h ON h.record_id=r.record_id
WHERE m.change_set_id=$1`, changed, record).Scan(&valueExact)
			} else {
				encoded, encodeErr := json.Marshal(expectedAlias)
				if encodeErr != nil {
					t.Fatal(encodeErr)
				}
				err = tx.QueryRow(t.Context(), "SELECT after_value=$2::jsonb FROM change_set_mutations WHERE change_set_id=$1", changed, encoded).Scan(&valueExact)
			}
			if err != nil || !valueExact {
				t.Fatalf("persisted complete value exact=%v error=%v", valueExact, err)
			}
		})
	}
}
