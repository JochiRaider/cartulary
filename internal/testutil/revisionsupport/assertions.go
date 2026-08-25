package revisionsupport

import (
	"database/sql"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

func RequireExactlyOneChangeSet(t testing.TB, before int, after int) {
	t.Helper()
	if after-before != 1 {
		t.Fatalf("expected exactly one change_set delta, before=%d after=%d", before, after)
	}
}

func RequireActorAttribution(
	t testing.TB,
	gotActorUserID string,
	wantActorUserID string,
	gotSource string,
	wantSource string,
) {
	t.Helper()
	if gotActorUserID != wantActorUserID {
		t.Fatalf("unexpected actor attribution: got %q want %q", gotActorUserID, wantActorUserID)
	}
	if gotSource != wantSource {
		t.Fatalf("unexpected mutation source: got %q want %q", gotSource, wantSource)
	}
}

func RequireOneRecordChangeIntentPerRevisionSQL(t testing.TB, db *sql.DB, changeSetID string) {
	t.Helper()
	collaborationsupport.RequireOneRecordChangeIntentPerRevision(t, db, changeSetID)
}

func RequireOneRecordChangeIntentPerRevisionPostgres(t testing.TB, db postgres.DB, changeSetID string) {
	t.Helper()
	collaborationsupport.RequireOneRecordChangeIntentPerRevision(t, db, changeSetID)
}
