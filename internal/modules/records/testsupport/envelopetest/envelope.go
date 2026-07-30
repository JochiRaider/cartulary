package envelopetest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func RequireRowVersionStable(t testing.TB, before int64, after int64) {
	t.Helper()
	if after != before {
		t.Fatalf("expected row_version to remain stable, before=%d after=%d", before, after)
	}
}

func SeedRecordEnvelope(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordID uuid.UUID,
	recordType string,
) {
	t.Helper()

	if err := execDB(db, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
ON CONFLICT (record_id) DO NOTHING
`, recordID, incidentID, recordType, actorUserID); err != nil {
		t.Fatalf("seed record envelope: %v", err)
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
		return fmt.Errorf("unsupported Records test database %T", db)
	}
}
