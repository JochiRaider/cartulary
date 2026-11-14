package revisions_test

import (
	"database/sql"
	"encoding/json"
	"maps"
	"testing"
	"time"

	"github.com/google/uuid"
)

func requirePreservedIdentifierHistory(t testing.TB, db *sql.DB, changeSetID string, id, incidentID, recordID, actorID uuid.UUID, createdAt, deletedAt time.Time, rollback bool) {
	t.Helper()
	value := map[string]any{
		"entity_preserved_identifier_id": id.String(), "incident_id": incidentID.String(), "record_id": recordID.String(),
		"entity_type": "host", "identifier_type": "fqdn", "raw_value": "loser.example.test", "normalized_value": "loser.example.test",
		"classification": "exact_match_reuse", "created_by_user_id": actorID.String(), "created_at": createdAt.UTC().Format(time.RFC3339Nano), "deleted_at": nil,
	}
	var before map[string]any
	after := maps.Clone(value)
	operation := "create"
	if rollback {
		before = maps.Clone(value)
		after["deleted_at"] = deletedAt.UTC().Format(time.RFC3339Nano)
		operation = "rollback"
	}
	encode := func(value map[string]any) any {
		if value == nil {
			return nil
		}
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return payload
	}
	var count int
	var exact bool
	if err := db.QueryRow(`SELECT count(*),coalesce(bool_and(operation_kind=$3 AND before_value IS NOT DISTINCT FROM $4::jsonb AND after_value=$5::jsonb),false)
FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='entity_preserved_identifier' AND target_id=$2`, changeSetID, "entity_preserved_identifier:"+id.String(), operation, encode(before), encode(after)).Scan(&count, &exact); err != nil {
		t.Fatal(err)
	}
	if count != 1 || !exact {
		t.Fatalf("complete physical preserved history: count=%d exact=%v rollback=%v", count, exact, rollback)
	}
	var deleted any
	if rollback {
		deleted = deletedAt
	}
	if got := countRows(t, db, `SELECT count(*) FROM entity_preserved_identifiers WHERE entity_preserved_identifier_id=$1 AND record_id=$2 AND created_by_user_id=$3 AND created_at=$4 AND deleted_at IS NOT DISTINCT FROM $5::timestamptz`, id, recordID, actorID, createdAt, deleted); got != 1 {
		t.Fatal("preserved source identity or creation/deletion facts changed")
	}
}
