package entities_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

const mentionHistoryCreatedAt = "2026-08-01T12:00:00.123456Z"

func setMentionHistoryCreationFixture(t testing.TB, db *sql.DB, mentionID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(`UPDATE entity_mentions SET ordinal=7,created_at=$2 WHERE entity_mention_id=$1`, mentionID, mentionHistoryCreatedAt); err != nil {
		t.Fatal(err)
	}
}

func mentionHistoryResolutionClock(t testing.TB, db *sql.DB, mentionID uuid.UUID) time.Time {
	t.Helper()
	var at time.Time
	if err := db.QueryRow(`SELECT resolved_at FROM entity_mentions WHERE entity_mention_id=$1`, mentionID).Scan(&at); err != nil {
		t.Fatal(err)
	}
	return at
}

// These literal expectations are independent of the production history codec.
func expectedMentionHistory(mentionID, sourceID, actorID uuid.UUID, locator, normalized, status string, version int64, resolved *uuid.UUID, at *time.Time) map[string]any {
	value := map[string]any{
		"entity_mention_id": mentionID.String(), "source_record_id": sourceID.String(),
		"entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual_entry",
		"origin_locator": locator, "raw_text": "WS-023", "normalized_text": normalized,
		"resolution_status": status, "row_version": version, "ordinal": 7,
		"created_by_user_id": actorID.String(), "created_at": mentionHistoryCreatedAt,
		"resolved_record_id": nil, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
	}
	if resolved != nil {
		value["resolved_record_id"], value["resolved_by_user_id"] = resolved.String(), actorID.String()
		value["resolved_at"], value["resolution_method"] = at.UTC().Format(time.RFC3339Nano), "explicit_resolve_route"
	}
	return value
}

func requireRetainedMentionPair(t testing.TB, db *sql.DB, changeSetID string, mentionID uuid.UUID, before, after map[string]any) {
	t.Helper()
	rows, err := db.Query(`SELECT before_value,after_value FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='entity_mention' AND target_id=$2`, changeSetID, mentionID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatalf("missing retained mention: %v", rows.Err())
	}
	var gotBefore, gotAfter []byte
	if err := rows.Scan(&gotBefore, &gotAfter); err != nil {
		t.Fatal(err)
	}
	for index, pair := range []struct {
		got  []byte
		want map[string]any
	}{{gotBefore, before}, {gotAfter, after}} {
		want, err := json.Marshal(pair.want)
		if err != nil {
			t.Fatal(err)
		}
		decode := func(payload []byte) any {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(payload))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			return value
		}
		if !reflect.DeepEqual(decode(pair.got), decode(want)) {
			t.Fatalf("retained mention side %d = %s want %s", index, pair.got, want)
		}
	}
	if rows.Next() {
		t.Fatal("duplicate retained mention mutation")
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}
