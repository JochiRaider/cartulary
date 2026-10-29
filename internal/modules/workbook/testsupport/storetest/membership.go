package storetest

import (
	"context"
	"database/sql"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"testing"
)

func SeedUserPreferences(t testing.TB, db postgres.DB, incidentID, userID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), seedUserPreferences, incidentID, userID); err != nil {
		t.Fatalf("seed user workbook preferences: %v", err)
	}
}

func SeedUserPreferencesSQL(t testing.TB, db *sql.DB, incidentID, userID uuid.UUID) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), seedUserPreferences, incidentID, userID); err != nil {
		t.Fatalf("seed user workbook preferences: %v", err)
	}
}

const seedUserPreferences = `INSERT INTO user_workbook_preferences (incident_id, user_id, home_sheet_ref, created_at, updated_at) VALUES ($1, $2, NULL, now(), now()) ON CONFLICT (incident_id, user_id) DO NOTHING`
