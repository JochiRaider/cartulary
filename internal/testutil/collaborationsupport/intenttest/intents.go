package intenttest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IntentRecord struct {
	IntentKey        string
	EventFamily      string
	SourceIdentity   string
	CanonicalPayload []byte
}

func CountBySourceIdentity(t testing.TB, pool *pgxpool.Pool, sourceIdentity string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE source_identity = $1
`, sourceIdentity).Scan(&count); err != nil {
		t.Fatalf("count Collaboration intents: %v", err)
	}
	return count
}

func LoadBySourceIdentity(t testing.TB, pool *pgxpool.Pool, sourceIdentity string) IntentRecord {
	t.Helper()
	var record IntentRecord
	if err := pool.QueryRow(context.Background(), `
SELECT intent_key, event_family, source_identity, canonical_payload
  FROM collaboration_event_intents
 WHERE source_identity = $1
`, sourceIdentity).Scan(&record.IntentKey, &record.EventFamily, &record.SourceIdentity, &record.CanonicalPayload); err != nil {
		t.Fatalf("load collaboration intent: %v", err)
	}
	return record
}

func InsertPersistedJobProgressIntentFixture(
	t testing.TB,
	pool *pgxpool.Pool,
	intentKey string,
	incidentID uuid.UUID,
	payload []byte,
	sourceIdentity string,
	createdAt time.Time,
) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO collaboration_event_intents (
    intent_key, incident_id, event_family, canonical_payload, source_identity,
    mutation_ordinal, next_attempt_at, created_at, updated_at
) VALUES ($1, $2, 'job_progress', $3::jsonb, $4, 0, $5, $5, $5)
`, intentKey, incidentID, payload, sourceIdentity, createdAt.UTC()); err != nil {
		t.Fatalf("insert persisted job-progress intent fixture: %v", err)
	}
}
