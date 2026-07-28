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

func InsertLegacyJobProgressV1(
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
		t.Fatalf("insert legacy job-progress intent: %v", err)
	}
}

func CountLegacyAndV2JobProgress(
	t testing.TB,
	pool *pgxpool.Pool,
	legacyKey string,
	v2Pattern string,
	sourceIdentity string,
) (int, int) {
	t.Helper()
	var legacyCount, v2Count int
	if err := pool.QueryRow(context.Background(), `
SELECT count(*) FILTER (WHERE intent_key = $1),
       count(*) FILTER (WHERE intent_key LIKE $2)
  FROM collaboration_event_intents
 WHERE source_identity = $3
`, legacyKey, v2Pattern, sourceIdentity).Scan(&legacyCount, &v2Count); err != nil {
		t.Fatalf("count legacy/v2 job-progress intents: %v", err)
	}
	return legacyCount, v2Count
}
