package revisions_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type historySeed struct {
	IncidentID  uuid.UUID
	ActorID     uuid.UUID
	RecordID    uuid.UUID
	ChangeSetID uuid.UUID
	CreatedAt   time.Time
	Source      string
	SequenceNo  int
	TargetKind  string
	Operation   string
	RowVersion  int64
}

func seedRecord(t testing.TB, db *sql.DB, server *httptestx.Server, login workbookscenariotest.LoginResult, actorID uuid.UUID, incidentKey string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	incident := workbookscenariotest.CreateIncident(t, server, login, map[string]any{
		"client_txn_id": "txn-" + strings.ToLower(strings.ReplaceAll(incidentKey, "-", "")),
		"incident_key":  incidentKey,
		"title":         "History " + incidentKey,
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	workbookscenariotest.SeedHostRecord(t, db, incidentID, actorID, recordID, "History Host", "history_revision-host", "", "")
	return incidentID, recordID
}

func seedHistoryChangeSet(t testing.TB, db *sql.DB, seed historySeed) {
	t.Helper()
	seedChangeSet(t, db, seed)
	seedHistoryMutation(t, db, seed)
	if seed.RowVersion > 0 {
		beforePayload := map[string]any{"record_id": seed.RecordID.String(), "row_version": seed.RowVersion - 1}
		if seed.RowVersion == 1 {
			beforePayload = nil
		}
		afterPayload := map[string]any{"record_id": seed.RecordID.String(), "row_version": seed.RowVersion}
		beforeJSON := jsonOrNil(t, beforePayload)
		afterJSON := jsonOrNil(t, afterPayload)
		if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_revisions (change_set_id, record_id, row_version, before_json, after_json, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (record_id, row_version) DO NOTHING
`, seed.ChangeSetID, seed.RecordID, seed.RowVersion, beforeJSON, afterJSON, seed.CreatedAt); err != nil {
			t.Fatalf("seed record revision: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), `UPDATE records SET row_version = GREATEST(row_version, $1) WHERE record_id = $2`, seed.RowVersion, seed.RecordID); err != nil {
			t.Fatalf("advance record row version: %v", err)
		}
	}
}

func seedChangeSet(t testing.TB, db *sql.DB, seed historySeed) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (change_set_id) DO NOTHING
`, seed.ChangeSetID, seed.IncidentID, seed.ActorID, seed.Source, seed.CreatedAt); err != nil {
		t.Fatalf("seed change set: %v", err)
	}
}

func seedHistoryMutation(t testing.TB, db *sql.DB, seed historySeed) {
	t.Helper()
	seedChangeSet(t, db, seed)
	beforePayload := map[string]any{"record_id": seed.RecordID.String(), "operation": "before-" + seed.Operation}
	afterPayload := map[string]any{"record_id": seed.RecordID.String(), "operation": "after-" + seed.Operation}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_set_mutations (
    change_set_id,
    sequence_no,
    target_kind,
    target_id,
    operation_kind,
    before_version_id,
    after_version_id,
    before_value,
    after_value
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (change_set_id, sequence_no) DO NOTHING
`, seed.ChangeSetID, seed.SequenceNo, seed.TargetKind, seed.RecordID.String(), seed.Operation, versionID(seed, "before"), versionID(seed, "after"), jsonOrNil(t, beforePayload), jsonOrNil(t, afterPayload)); err != nil {
		t.Fatalf("seed change-set mutation: %v", err)
	}
}

func jsonOrNil(t testing.TB, value any) any {
	t.Helper()
	if value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal history fixture json: %v", err)
	}
	return string(payload)
}

func versionID(seed historySeed, suffix string) string {
	return seed.TargetKind + ":" + seed.RecordID.String() + ":" + suffix
}

func mustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}
