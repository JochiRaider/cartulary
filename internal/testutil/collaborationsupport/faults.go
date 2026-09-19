package collaborationsupport

import (
	"context"
	"database/sql"
	"testing"
)

// FailIntentInserts installs a real failing transaction participant in a
// dedicated test database. The handle must have migration privileges; product
// requests still use their restricted runtime handle. Cleanup restores writes.
func FailIntentInserts(t testing.TB, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
CREATE FUNCTION fail_test_collaboration_intent() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'injected Collaboration intent failure'; END;
$$;
CREATE TRIGGER fail_test_collaboration_intent BEFORE INSERT ON collaboration_event_intents
FOR EACH ROW EXECUTE FUNCTION fail_test_collaboration_intent();`); err != nil {
		t.Fatalf("install Collaboration intent failure: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, `DROP TRIGGER fail_test_collaboration_intent ON collaboration_event_intents; DROP FUNCTION fail_test_collaboration_intent()`); err != nil {
			t.Errorf("remove Collaboration intent failure: %v", err)
		}
	})
}
