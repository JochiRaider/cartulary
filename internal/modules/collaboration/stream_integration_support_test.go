package collaboration_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func createDurableStreamIncident(
	t testing.TB,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	suffix string,
) string {
	t.Helper()
	incident := incidentscenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-collaboration-durable-" + suffix,
		"incident_key":  "IR-COLLAB-DURABLE-" + suffix,
		"title":         "Collaboration durable " + suffix,
	})
	return incident["incident_id"].(string)
}

func seedCurrentQuarantinedIncident(
	t testing.TB,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	pool *pgxpool.Pool,
	suffix string,
	invalidPayload bool,
	seededAt time.Time,
) (uuid.UUID, string, []byte) {
	t.Helper()
	incidentID := uuid.MustParse(createDurableStreamIncident(t, harness, admin, suffix))
	intent := requireJobIntent(t, incidentID, suffix, seededAt)
	if invalidPayload {
		intent.EventFamily = privatestream.EventFamilyRecordChanged
		intent.CanonicalPayload = json.RawMessage(`{"not":"a record change"}`)
		intent.IntentKey = "record_changed:" + suffix
		intent.SourceIdentity = "record:" + suffix
	}
	insertPersistedIntentFixture(t, pool, intent)
	if _, err := pool.Exec(context.Background(), `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, failure_count, quarantined_at,
    quarantine_reason, updated_at
) VALUES ($1, 0, 12, $2, 'invalid_event_payload', $2)
ON CONFLICT (incident_id) DO UPDATE
SET failure_count = 12,
    quarantined_at = EXCLUDED.quarantined_at,
    quarantine_reason = EXCLUDED.quarantine_reason,
	updated_at = EXCLUDED.updated_at
`, incidentID, seededAt); err != nil {
		t.Fatalf("seed quarantined collaboration cursor: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `
UPDATE collaboration_event_intents
   SET attempt_count = 12,
       last_error_code = 'invalid_event_payload',
       next_attempt_at = $2,
       updated_at = $2
 WHERE incident_id = $1
   AND intent_key = $3
`, incidentID, seededAt, intent.IntentKey); err != nil {
		t.Fatalf("seed quarantined collaboration intent: %v", err)
	}
	return incidentID, intent.IntentKey, append([]byte(nil), intent.CanonicalPayload...)
}

func jsonValuesEqual(left []byte, right []byte) bool {
	var leftValue any
	var rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func requeueCollaborationIncident(
	service collaboration.RecoveryCapability,
	ctx context.Context,
	incidentID uuid.UUID,
	mutatedAt time.Time,
) (collaboration.RequeueResult, error) {
	return service.RequeueIncident(ctx, collaboration.RequeueRequest{
		OperationID: uuid.New(),
		IncidentID:  incidentID,
		MutatedAt:   mutatedAt.UTC(),
	})
}

func requeueFailureKind(err error) collaboration.RequeueFailureKind {
	var failure *collaboration.RequeueFailure
	if errors.As(err, &failure) {
		return failure.Kind
	}
	return ""
}

type requeueCommitFailureDB struct {
	postgres.DB
}

func (db requeueCommitFailureDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return requeueCommitFailureTx{Tx: tx}, nil
}

type requeueCommitFailureTx struct {
	pgx.Tx
}

func (requeueCommitFailureTx) Commit(context.Context) error {
	return errors.New("injected collaboration requeue commit failure")
}

func requireJobIntent(t testing.TB, incidentID uuid.UUID, identity string, createdAt time.Time) privatestream.EventIntent {
	t.Helper()
	intent, err := privatestream.NewEventIntent(
		"job_progress:"+identity,
		incidentID,
		privatestream.EventFamilyJobProgress,
		collabprotocol.NewIncidentJobProgressPayload(
			identity,
			incidentID,
			collabprotocol.JobStatusQueued,
			collabprotocol.JobProgress{Completed: 0},
			createdAt,
		),
		"job:"+identity,
		0,
		createdAt,
	)
	if err != nil {
		t.Fatalf("create job intent: %v", err)
	}
	return intent
}

func insertPersistedIntentFixture(t testing.TB, pool *pgxpool.Pool, intent privatestream.EventIntent) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO collaboration_event_intents (
    intent_key, incident_id, event_family, canonical_payload, source_identity,
    mutation_ordinal, next_attempt_at, created_at, updated_at
) VALUES ($1, $2, $3, $4::jsonb, $5, 0, $6, $6, $6)
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, intent.CanonicalPayload, intent.SourceIdentity, intent.CreatedAt); err != nil {
		t.Fatalf("insert persisted intent fixture: %v", err)
	}
}

func appendCommittedIntent(t testing.TB, pool *pgxpool.Pool, writer privatestream.IntentWriter, intent privatestream.EventIntent) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin intent transaction: %v", err)
	}
	if err := writer.AppendTx(ctx, tx, intent); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("append durable intent: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit durable intent: %v", err)
	}
}

type recordingBroadcaster struct {
	mu            sync.Mutex
	failRemaining int
	messages      []collabprotocol.Message
}

func newDispatcherForTest(
	db postgres.DB,
	broadcaster interface {
		DeliverReplayable(collabprotocol.Message) error
	},
	now func() time.Time,
) *privatestream.Dispatcher {
	return privatestream.NewDispatcher(privatestream.NewPostgresStream(db, now), broadcaster, now)
}

func (b *recordingBroadcaster) DeliverReplayable(message collabprotocol.Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, message)
	if b.failRemaining > 0 {
		b.failRemaining--
		return errors.New("injected broadcast failure")
	}
	return nil
}

func (b *recordingBroadcaster) snapshot() []collabprotocol.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]collabprotocol.Message(nil), b.messages...)
}

func messagesForIncident(messages []collabprotocol.Message, incidentID string) []collabprotocol.Message {
	filtered := make([]collabprotocol.Message, 0, len(messages))
	for _, message := range messages {
		if message.IncidentID == incidentID {
			filtered = append(filtered, message)
		}
	}
	return filtered
}
