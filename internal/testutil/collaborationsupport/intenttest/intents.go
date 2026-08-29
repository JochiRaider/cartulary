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

type PersistedIntentFixture struct {
	IntentKey        string
	IncidentID       uuid.UUID
	EventFamily      string
	CanonicalPayload []byte
	SourceIdentity   string
	MutationOrdinal  int
	CreatedAt        time.Time
}

type PersistedReplayEventFixture struct {
	EventID          uuid.UUID
	IncidentID       uuid.UUID
	StreamSeq        int64
	IntentKey        string
	EventFamily      string
	CanonicalPayload []byte
	EmittedAt        time.Time
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

func InsertPersistedIntentFixture(
	t testing.TB,
	pool *pgxpool.Pool,
	fixture PersistedIntentFixture,
) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO collaboration_event_intents (
    intent_key, incident_id, event_family, canonical_payload, source_identity,
    mutation_ordinal, next_attempt_at, created_at, updated_at
) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $7, $7)
`, fixture.IntentKey, fixture.IncidentID, fixture.EventFamily, fixture.CanonicalPayload,
		fixture.SourceIdentity, fixture.MutationOrdinal, fixture.CreatedAt.UTC()); err != nil {
		t.Fatalf("insert persisted Collaboration intent fixture: %v", err)
	}
}

func InsertPersistedReplayEventFixtures(
	t testing.TB,
	pool *pgxpool.Pool,
	fixtures ...PersistedReplayEventFixture,
) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin persisted Collaboration replay fixture: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	highWaterByIncident := map[uuid.UUID]int64{}
	for _, fixture := range fixtures {
		if fixture.StreamSeq > highWaterByIncident[fixture.IncidentID] {
			highWaterByIncident[fixture.IncidentID] = fixture.StreamSeq
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO collaboration_replay_events (
    event_id, incident_id, stream_seq, intent_key, event_family, canonical_payload, emitted_at
) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
`, fixture.EventID, fixture.IncidentID, fixture.StreamSeq, fixture.IntentKey,
			fixture.EventFamily, fixture.CanonicalPayload, fixture.EmittedAt.UTC()); err != nil {
			t.Fatalf("insert persisted Collaboration replay fixture: %v", err)
		}
	}
	for incidentID, highWater := range highWaterByIncident {
		if _, err := tx.Exec(ctx, `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, updated_at
) VALUES ($1, $2, now())
ON CONFLICT (incident_id) DO UPDATE
SET high_water_stream_seq = EXCLUDED.high_water_stream_seq,
    updated_at = EXCLUDED.updated_at
`, incidentID, highWater); err != nil {
			t.Fatalf("upsert persisted Collaboration replay cursor fixture: %v", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit persisted Collaboration replay fixture: %v", err)
	}
}

func ReplacePersistedReplayPayload(
	t testing.TB,
	pool *pgxpool.Pool,
	eventID uuid.UUID,
	payload []byte,
) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
UPDATE collaboration_replay_events
   SET canonical_payload = $2::jsonb
 WHERE event_id = $1
`, eventID, payload); err != nil {
		t.Fatalf("replace persisted Collaboration replay payload fixture: %v", err)
	}
}
