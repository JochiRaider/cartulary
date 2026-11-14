package revisions_test

import (
	"database/sql"
	"github.com/google/uuid"
	"net/http"
	"testing"
)

func requireMentionCurrentness(t testing.TB, db *sql.DB, incidentID, sourceID, mentionID uuid.UUID, rollback func() *http.Response) {
	t.Helper()
	cases := map[string]map[string]any{}
	for field, value := range map[string]any{
		"raw_text": "different raw text", "normalized_text": "different normalized text",
		"origin_kind": "clipboard_paste", "origin_locator": "different locator",
		"ordinal": 7, "created_at": "2026-05-10T12:20:59Z", "row_version": 7,
		"resolved_at": "2026-05-10T12:21:01Z", "resolution_method": "different method",
	} {
		cases[field] = map[string]any{field: value}
	}
	requireNonRowCurrentness(t, db, incidentID, sourceID, mentionID, "entity_mentions", "entity_mention_id", cases, rollback)
}
