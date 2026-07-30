package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	FieldHostRefs     = "timeline.host_refs"
	FieldIdentityRefs = "timeline.identity_refs"
	FieldSummary      = "timeline.activity_synopsis_text"
	FieldSourceText   = "timeline.raw_activity_text"
)

var (
	RecordID        = uuid.MustParse("40000000-0000-0000-0000-000000000201")
	SiblingRecordID = uuid.MustParse("40000000-0000-0000-0000-000000000202")
	MixedRecordID   = uuid.MustParse("40000000-0000-0000-0000-000000000203")
)

func SeedTimelineRecord(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordID uuid.UUID,
) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "timeline_event")
	if err := execDB(db, `
INSERT INTO timeline_events (
    record_id, incident_id, activity_synopsis_text, capture_state,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, 'record-support-source-row', 'reviewed', $3, $3)
`, recordID, incidentID, actorUserID); err != nil {
		t.Fatalf("seed timeline record: %v", err)
	}
}

func execDB(db any, query string, args ...any) error {
	switch typed := db.(type) {
	case postgres.DB:
		_, err := typed.Exec(context.Background(), query, args...)
		return err
	case *sql.DB:
		_, err := typed.ExecContext(context.Background(), query, args...)
		return err
	default:
		return fmt.Errorf("unsupported Timeline test database %T", db)
	}
}

func CollectionActions(actions ...map[string]any) map[string]any {
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": append([]map[string]any(nil), actions...),
	}
}

func AddTokenAction(rawText string) map[string]any {
	return map[string]any{"op": "add_token", "raw_text": rawText}
}

func AddResolvedRefAction(rawText string, resolvedRecordID uuid.UUID) map[string]any {
	return map[string]any{
		"op":                 "add_resolved_ref",
		"raw_text":           rawText,
		"resolved_record_id": resolvedRecordID.String(),
	}
}

func TimelineCollectionPatchPayload(
	fieldKey string,
	baseRowVersion int64,
	clientTxnID string,
	actionPayload map[string]any,
) map[string]any {
	return map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": baseRowVersion,
		"client_txn_id":    clientTxnID,
		"changes": []map[string]any{{
			"field_key":      fieldKey,
			"action_payload": actionPayload,
		}},
	}
}
