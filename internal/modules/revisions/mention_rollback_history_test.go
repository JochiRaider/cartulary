package revisions_test

import (
	"database/sql"
	"encoding/json"
	"maps"
	"testing"

	"github.com/google/uuid"
)

func requireCompleteMentionRollbackHistory(t testing.TB, db *sql.DB, changeSetID string, mentionID, sourceID, actorID, linkID uuid.UUID) {
	t.Helper()
	var targetID uuid.UUID
	if err := db.QueryRow(`SELECT dst_record_id FROM record_links WHERE record_link_id=$1`, linkID).Scan(&targetID); err != nil {
		t.Fatal(err)
	}
	before := map[string]any{
		"entity_mention_id": mentionID.String(), "source_record_id": sourceID.String(),
		"entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual_entry",
		"origin_locator": "history_revision", "raw_text": "Mention Host", "normalized_text": "mention host",
		"resolution_status": "resolved", "row_version": 2, "ordinal": 1,
		"created_by_user_id": actorID.String(), "created_at": "2026-05-10T12:21:00Z",
		"resolved_record_id": targetID.String(), "resolved_by_user_id": actorID.String(),
		"resolved_at": "2026-05-10T12:21:00Z", "resolution_method": "explicit_resolve_route",
	}
	after := maps.Clone(before)
	after["resolution_status"], after["row_version"] = "unresolved", 3
	for _, field := range []string{"resolved_record_id", "resolved_by_user_id", "resolved_at", "resolution_method"} {
		after[field] = nil
	}
	encode := func(value any) []byte {
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return payload
	}
	var count int
	var exact bool
	if err := db.QueryRow(`SELECT count(*),coalesce(bool_and(operation_kind='rollback' AND before_value=$3::jsonb AND after_value=$4::jsonb),false)
FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='entity_mention' AND target_id=$2`, changeSetID, mentionID.String(), encode(before), encode(after)).Scan(&count, &exact); err != nil {
		t.Fatal(err)
	}
	if count != 1 || !exact {
		t.Fatalf("complete rollback Mention history: count=%d exact=%v", count, exact)
	}
	if got := countRows(t, db, `SELECT count(*) FROM entity_mentions WHERE entity_mention_id=$1 AND ordinal=1 AND created_by_user_id=$2 AND created_at='2026-05-10T12:21:00Z' AND row_version=3 AND resolution_status='unresolved' AND resolved_record_id IS NULL AND resolved_by_user_id IS NULL AND resolved_at IS NULL AND resolution_method IS NULL`, mentionID, actorID); got != 1 {
		t.Fatal("rollback changed creation facts or failed to clear complete resolution tuple")
	}
}
