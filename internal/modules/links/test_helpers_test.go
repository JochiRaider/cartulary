package links_test

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func collectionItems(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	items := make([]map[string]any, 0)
	switch rawItems := value["items"].(type) {
	case []any:
		for _, rawItem := range rawItems {
			items = append(items, rawItem.(map[string]any))
		}
	case []map[string]any:
		items = append(items, rawItems...)
	default:
		t.Fatalf("unexpected %s items type %T", fieldKey, value["items"])
	}
	return items
}

func singleCollectionItem(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()
	items := collectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, items)
	}
	return items[0]
}

func historyItems(t testing.TB, envelope map[string]any) []map[string]any {
	t.Helper()
	raw := envelope["data"].(map[string]any)["items"].([]any)
	items := make([]map[string]any, 0, len(raw))
	for _, rawItem := range raw {
		items = append(items, rawItem.(map[string]any))
	}
	return items
}

func historyItemForTarget(t testing.TB, items []map[string]any, targetKind string, targetID string) map[string]any {
	t.Helper()
	for _, item := range items {
		summary := item["diff_summary"].(map[string]any)
		units := summary["units"].([]any)
		for _, rawUnit := range units {
			unit := rawUnit.(map[string]any)
			if unit["target_kind"] == targetKind && unit["target_id"] == targetID {
				return item
			}
		}
	}
	t.Fatalf("missing history item target_kind=%s target_id=%s in %#v", targetKind, targetID, items)
	return nil
}

func requireRollbackActionContains(t testing.TB, item map[string]any, expected string) {
	t.Helper()
	raw := item["available_rollback_actions"].([]any)
	for _, value := range raw {
		if value.(string) == expected {
			if _, ok := item["history_entry_ref"].(string); expected == "history_entry" && !ok {
				t.Fatalf("history_entry rollback item missing history_entry_ref: %#v", item)
			}
			return
		}
	}
	t.Fatalf("rollback actions missing %s in %#v", expected, item)
}

func countRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return count
}

func stringScalar(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func stringScalarPG(t testing.TB, db postgres.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func columnNamesPG(t testing.TB, db postgres.DB, tableName string) []string {
	t.Helper()
	rows, err := db.Query(context.Background(), `
SELECT column_name
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
 ORDER BY ordinal_position
`, tableName)
	if err != nil {
		t.Fatalf("query %s columns: %v", tableName, err)
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			t.Fatalf("scan %s column: %v", tableName, err)
		}
		columns = append(columns, column)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s columns: %v", tableName, err)
	}
	return columns
}

func mustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}

func requireCanonicalRecordLinkMutationValue(t testing.TB, value map[string]any) {
	t.Helper()
	wantKeys := []string{
		"confidence", "created_at", "created_by_user_id", "decided_at",
		"deleted_at", "deleted_by_user_id", "dst_record_id", "field_key",
		"incident_id", "link_type", "owner_user_id", "provenance",
		"record_link_id", "src_record_id",
	}
	gotKeys := make([]string, 0, len(value))
	for key := range value {
		gotKeys = append(gotKeys, key)
	}
	slices.Sort(gotKeys)
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("record-link mutation keys got %v want %v", gotKeys, wantKeys)
	}
	for _, key := range []string{"record_link_id", "incident_id", "src_record_id", "dst_record_id", "owner_user_id", "created_by_user_id"} {
		requireCanonicalUUIDScalar(t, value, key, false)
	}
	requireCanonicalUUIDScalar(t, value, "deleted_by_user_id", true)
	for _, key := range []string{"decided_at", "created_at"} {
		requireCanonicalTimestampScalar(t, value, key, false)
	}
	requireCanonicalTimestampScalar(t, value, "deleted_at", true)
	if _, ok := value["link_type"].(string); !ok {
		t.Fatalf("record-link link_type is %T, want string", value["link_type"])
	}
	if _, ok := value["provenance"].(string); !ok {
		t.Fatalf("record-link provenance is %T, want string", value["provenance"])
	}
	if value["field_key"] != nil {
		if _, ok := value["field_key"].(string); !ok {
			t.Fatalf("record-link field_key is %T, want string or nil", value["field_key"])
		}
	}
	if value["confidence"] != nil {
		if _, ok := value["confidence"].(int); !ok {
			t.Fatalf("record-link confidence is %T, want JSON-compatible integer or nil", value["confidence"])
		}
	}
}

func requireCanonicalRecordTagMutationValue(t testing.TB, value map[string]any) {
	t.Helper()
	wantKeys := []string{
		"created_at", "created_by_user_id", "deleted_at", "deleted_by_user_id",
		"incident_id", "normalized_tag_name", "record_id", "record_tag_id",
		"tag_name", "updated_at",
	}
	gotKeys := make([]string, 0, len(value))
	for key := range value {
		gotKeys = append(gotKeys, key)
	}
	slices.Sort(gotKeys)
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("record-tag mutation keys got %v want %v", gotKeys, wantKeys)
	}
	for _, key := range []string{"record_tag_id", "incident_id", "record_id", "created_by_user_id"} {
		requireCanonicalUUIDScalar(t, value, key, false)
	}
	requireCanonicalUUIDScalar(t, value, "deleted_by_user_id", true)
	for _, key := range []string{"created_at", "updated_at"} {
		requireCanonicalTimestampScalar(t, value, key, false)
	}
	requireCanonicalTimestampScalar(t, value, "deleted_at", true)
	for _, key := range []string{"tag_name", "normalized_tag_name"} {
		if text, ok := value[key].(string); !ok || text == "" {
			t.Fatalf("record-tag %s got %#v, want non-empty string", key, value[key])
		}
	}
}

func requireCanonicalUUIDScalar(t testing.TB, value map[string]any, key string, nullable bool) {
	t.Helper()
	if nullable && value[key] == nil {
		return
	}
	text, ok := value[key].(string)
	if !ok {
		t.Fatalf("%s got %T, want UUID string", key, value[key])
	}
	parsed, err := uuid.Parse(text)
	if err != nil || parsed.String() != text {
		t.Fatalf("%s got noncanonical UUID %q: %v", key, text, err)
	}
}

func requireCanonicalTimestampScalar(t testing.TB, value map[string]any, key string, nullable bool) {
	t.Helper()
	if nullable && value[key] == nil {
		return
	}
	text, ok := value[key].(string)
	if !ok {
		t.Fatalf("%s got %T, want timestamp string", key, value[key])
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || parsed.UTC().Format(time.RFC3339Nano) != text || !strings.HasSuffix(text, "Z") {
		t.Fatalf("%s got noncanonical UTC RFC3339Nano timestamp %q: %v", key, text, err)
	}
}
