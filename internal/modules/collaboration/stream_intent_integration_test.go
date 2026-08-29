package collaboration_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	timelinemodule "github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func runIntentWriterScenarios(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	atomicIncidentID string,
	atomicIncidentUUID uuid.UUID,
	intents privatestream.IntentWriter,
) {
	t.Run("intent and source state commit or roll back together", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `
DELETE FROM collaboration_event_intents WHERE incident_id = $1
`, atomicIncidentUUID)
		})
		var originalTitle string
		if err := pool.QueryRow(ctx, `SELECT title FROM incidents WHERE id = $1`, atomicIncidentUUID).Scan(&originalTitle); err != nil {
			t.Fatalf("load incident title: %v", err)
		}

		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin failing source transaction: %v", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE incidents SET title = $2 WHERE id = $1`, atomicIncidentUUID, originalTitle+" changed"); err != nil {
			t.Fatalf("update source state: %v", err)
		}
		invalidIntent := requireJobIntent(t, uuid.New(), "atomic-failure", time.Now().UTC())
		if err := intents.AppendTx(ctx, tx, invalidIntent); err == nil {
			t.Fatal("intent with missing incident must fail its source transaction")
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("roll back failed intent transaction: %v", err)
		}

		var titleAfterRollback string
		if err := pool.QueryRow(ctx, `SELECT title FROM incidents WHERE id = $1`, atomicIncidentUUID).Scan(&titleAfterRollback); err != nil {
			t.Fatalf("reload incident title: %v", err)
		}
		if titleAfterRollback != originalTitle {
			t.Fatalf("source state survived failed intent insertion: got %q want %q", titleAfterRollback, originalTitle)
		}

		payload := map[string]any{
			"client_txn_id":                   "txn-collaboration-durable-exact-replay",
			"timeline.activity_synopsis_text": "Durable exact replay",
		}
		first := timelineroutetest.CreateRow(t, harness.Server, admin, atomicIncidentID, payload)
		replayResponse := httptestx.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+atomicIncidentID+"/views/"+timelinemodule.TimelineViewSchemaID+"/rows",
			payload,
			httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
		)
		replayed := httptestx.RequireSuccessEnvelope(t, replayResponse, http.StatusOK)["data"].(map[string]any)
		firstRecordID := first["row"].(map[string]any)["record_id"].(string)
		if replayed["row"].(map[string]any)["record_id"] != firstRecordID {
			t.Fatalf("exact replay changed record identity: first=%s replay=%v", firstRecordID, replayed["row"])
		}
		var intentCount int
		if err := pool.QueryRow(ctx, `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE incident_id = $1
   AND source_record_id = $2
`, atomicIncidentUUID, uuid.MustParse(firstRecordID)).Scan(&intentCount); err != nil {
			t.Fatalf("count exact replay intents: %v", err)
		}
		if intentCount != 1 {
			t.Fatalf("exact replay intent count = %d want 1", intentCount)
		}
	})

	t.Run("intent keys are immutable across exact and divergent replay", func(t *testing.T) {
		createdAt := time.Now().UTC()
		identity := "immutable-intent-key"
		intentKey := "job_progress:" + identity
		intent := requireJobIntent(t, atomicIncidentUUID, identity, createdAt)
		appendCommittedIntent(t, pool, intents, intent)
		appendCommittedIntent(t, pool, intents, intent)

		divergent, err := privatestream.NewJobProgressIntent(
			intentKey,
			atomicIncidentUUID,
			collabtestprotocol.NewIncidentJobProgressPayload(
				identity,
				atomicIncidentUUID,
				collabprotocol.JobStatusQueued,
				collabprotocol.JobProgress{Completed: 0},
				createdAt,
			),
			"job:"+identity+":divergent",
			createdAt,
		)
		if err != nil {
			t.Fatalf("create divergent intent: %v", err)
		}
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin divergent intent transaction: %v", err)
		}
		err = intents.AppendTx(ctx, tx, divergent)
		if err == nil || !strings.Contains(err.Error(), "collaboration intent key collision") {
			_ = tx.Rollback(ctx)
			t.Fatalf("divergent intent error = %v want private collision diagnostic", err)
		}
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			t.Fatalf("roll back divergent intent transaction: %v", rollbackErr)
		}

		var count int
		if err := pool.QueryRow(ctx, `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intentKey).Scan(&count); err != nil {
			t.Fatalf("count immutable-key intents: %v", err)
		}
		if count != 1 {
			t.Fatalf("immutable-key intent count = %d want 1", count)
		}
		if _, err := pool.Exec(ctx, `
DELETE FROM collaboration_event_intents
 WHERE intent_key = $1
`, intentKey); err != nil {
			t.Fatalf("clean up immutable-key intent: %v", err)
		}
	})
}
