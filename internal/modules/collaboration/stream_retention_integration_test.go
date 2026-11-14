package collaboration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/intenttest"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
)

func runReplayRetentionScenarios(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	harness *appsupport.ServerHarness,
	admin authcookietest.AuthCookies,
	adminID uuid.UUID,
	dispatchIncidentUUID uuid.UUID,
	replay *privatestream.PostgresStream,
	intents privatestream.IntentWriter,
) {
	t.Run("resume tokens are hashed durable and retention is conjunctive", func(t *testing.T) {
		var (
			sessionID        uuid.UUID
			sessionExpiresAt time.Time
		)
		if err := pool.QueryRow(ctx, `
SELECT id, session_expires_at
  FROM user_sessions
 WHERE user_id = $1
   AND revoked_at IS NULL
 ORDER BY created_at DESC
 LIMIT 1
`, adminID).Scan(&sessionID, &sessionExpiresAt); err != nil {
			t.Fatalf("load active session: %v", err)
		}
		tokenNow := time.Now().UTC()
		token, _, err := replay.IssueResumeToken(
			ctx,
			sessionID,
			dispatchIncidentUUID,
			"durable-restart-client",
			sessionExpiresAt,
			tokenNow,
		)
		if err != nil {
			t.Fatalf("issue durable resume token: %v", err)
		}
		var storedHash []byte
		if err := pool.QueryRow(ctx, `
SELECT token_hash
  FROM collaboration_resume_tokens
 WHERE session_id = $1
   AND incident_id = $2
   AND client_instance_id = $3
`, sessionID, dispatchIncidentUUID, "durable-restart-client").Scan(&storedHash); err != nil {
			t.Fatalf("load stored resume-token hash: %v", err)
		}
		wantHash := sha256.Sum256([]byte(token))
		if !bytes.Equal(storedHash, wantHash[:]) || bytes.Equal(storedHash, []byte(token)) {
			t.Fatalf("resume token was not stored exclusively by hash")
		}

		restartedReplay := newPostgresStreamForTest(t, pool)
		type replayObservation struct {
			Status   string
			Messages []collabprotocol.Message
		}
		observeReplay := func(client string) (replayObservation, error) {
			var observation replayObservation
			prepared, err := restartedReplay.PrepareReplay(ctx, sessionID, dispatchIncidentUUID, client, token, 0, tokenNow)
			if err != nil {
				return observation, err
			}
			defer prepared.Close()
			observation.Status = prepared.Status()
			err = prepared.Visit(ctx, func(message collabprotocol.Message) error {
				observation.Messages = append(observation.Messages, message)
				return nil
			})
			return observation, err
		}
		replay, err := observeReplay("durable-restart-client")
		if err != nil {
			t.Fatalf("replay after store restart: %v", err)
		}
		if replay.Status != collabprotocol.ResumeStatusReplayed || len(replay.Messages) != 2 {
			t.Fatalf("restart replay result = status %q messages %d want replayed/2", replay.Status, len(replay.Messages))
		}
		if replay.Messages[0].StreamSeq == nil || replay.Messages[1].StreamSeq == nil ||
			*replay.Messages[0].StreamSeq != 1 || *replay.Messages[1].StreamSeq != 2 {
			t.Fatalf("restart replay order = %#v", replay.Messages)
		}
		var (
			corruptEventID  uuid.UUID
			originalPayload []byte
		)
		if err := pool.QueryRow(ctx, `
SELECT event_id, canonical_payload
  FROM collaboration_replay_events
 WHERE incident_id = $1
   AND stream_seq = 1
`, dispatchIncidentUUID).Scan(&corruptEventID, &originalPayload); err != nil {
			t.Fatalf("load authenticated replay corruption fixture: %v", err)
		}
		intenttest.ReplacePersistedReplayPayload(t, pool, corruptEventID, []byte(`{"job_id":"corrupt-replay"}`))
		corruptReplay, err := observeReplay("durable-restart-client")
		if err == nil || len(corruptReplay.Messages) != 0 {
			t.Fatalf("corrupt authenticated replay result = %#v error=%v", corruptReplay, err)
		}
		intenttest.ReplacePersistedReplayPayload(t, pool, corruptEventID, originalPayload)
		repairedReplay, err := observeReplay("durable-restart-client")
		if err != nil || len(repairedReplay.Messages) != 2 ||
			repairedReplay.Messages[0].StreamSeq == nil || *repairedReplay.Messages[0].StreamSeq != 1 ||
			repairedReplay.Messages[1].StreamSeq == nil || *repairedReplay.Messages[1].StreamSeq != 2 {
			t.Fatalf("repaired authenticated replay result = %#v error=%v", repairedReplay, err)
		}
		reset, err := observeReplay("another-client")
		if err != nil {
			t.Fatalf("validate mismatched token: %v", err)
		}
		if reset.Status != collabprotocol.ResumeStatusResetNeeded || len(reset.Messages) != 0 {
			t.Fatalf("mismatched token replayed data: %#v", reset)
		}

		t.Run("prepared replay covers large ranges and keeps its snapshot through pruning", func(t *testing.T) {
			incidentID := uuid.MustParse(createDurableStreamIncident(t, harness, admin, "prepared-replay"))
			t.Cleanup(func() {
				if _, err := pool.Exec(ctx, `DELETE FROM collaboration_replay_events WHERE incident_id = $1`, incidentID); err != nil {
					t.Errorf("remove replay fixture: %v", err)
				}
			})
			payload := collabtestprotocol.NewIncidentJobProgressPayload("large-replay", incidentID, collabprotocol.JobStatusRunning, collabprotocol.JobProgress{}, tokenNow)
			payload.Message = strings.Repeat("x", 2048)
			expectedPayload := collabtestprotocol.RawPayload(t, payload)
			// 10001 independently authored events exceed both retired limits.
			fixtures := make([]intenttest.PersistedReplayEventFixture, 0, 10001)
			for sequence := int64(1); sequence <= 10001; sequence++ {
				fixtures = append(fixtures, intenttest.PersistedReplayEventFixture{
					EventID: uuid.New(), IncidentID: incidentID, StreamSeq: sequence,
					IntentKey: fmt.Sprintf("prepared-replay:%d", sequence), EventFamily: privatestream.EventFamilyJobProgress,
					CanonicalPayload: expectedPayload, EmittedAt: tokenNow.Add(-time.Hour),
				})
			}
			intenttest.InsertPersistedReplayEventFixtures(t, pool, fixtures...)
			replayToken, _, err := restartedReplay.IssueResumeToken(ctx, sessionID, incidentID, "prepared-client", sessionExpiresAt, tokenNow)
			if err != nil {
				t.Fatal(err)
			}
			prepare := func(after int64) *privatestream.PreparedReplay {
				prepared, err := restartedReplay.PrepareReplay(ctx, sessionID, incidentID, "prepared-client", replayToken, after, tokenNow)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := prepared.Close(); err != nil {
						t.Errorf("close replay snapshot: %v", err)
					}
				})
				return prepared
			}
			prepared := prepare(0)
			if prepared.Status() != collabprotocol.ResumeStatusReplayed || prepared.HighWater() != 10001 {
				t.Fatalf("prepared replay = %s/%d", prepared.Status(), prepared.HighWater())
			}
			if err := restartedReplay.PruneReplay(ctx, tokenNow); err != nil {
				t.Fatal(err)
			}
			var remaining int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM collaboration_replay_events WHERE incident_id = $1`, incidentID).Scan(&remaining); err != nil || remaining != 10000 {
				t.Fatalf("concurrent prune = %d / %v", remaining, err)
			}
			seen := 0
			err = prepared.Visit(ctx, func(message collabprotocol.Message) error {
				if seen >= len(fixtures) || message.StreamSeq == nil || *message.StreamSeq != int64(seen+1) || message.EventID != fixtures[seen].EventID.String() || !jsonValuesEqual(message.Payload, expectedPayload) {
					return fmt.Errorf("replay mismatch at %d", seen)
				}
				seen++
				return nil
			})
			if err != nil || seen != 10001 {
				t.Fatalf("large snapshot delivered %d / %v", seen, err)
			}
			if err := prepared.Visit(ctx, func(collabprotocol.Message) error { t.Fatal("closed replay emitted"); return nil }); err == nil {
				t.Fatal("consumed snapshot remained open")
			}
			if err := prepared.Close(); err != nil {
				t.Fatal(err)
			}
			reset := prepare(0)
			if reset.Status() != collabprotocol.ResumeStatusResetNeeded {
				t.Fatal("new snapshot ignored pruned prefix")
			}
			if err := reset.Visit(ctx, func(collabprotocol.Message) error { t.Fatal("reset emitted partial replay"); return nil }); err != nil {
				t.Fatal(err)
			}
			for _, failure := range []string{"visitor", "cancellation", "unused"} {
				prepared := prepare(1)
				if prepared.Status() != collabprotocol.ResumeStatusReplayed {
					t.Fatal("retained suffix did not prepare")
				}
				injected := errors.New("stop replay visitor")
				visitCtx, cancel := context.WithCancel(ctx)
				calls := 0
				if failure == "unused" {
					if err := prepared.Close(); err != nil {
						t.Fatal(err)
					}
				} else {
					err := prepared.Visit(visitCtx, func(collabprotocol.Message) error {
						calls++
						if failure == "visitor" {
							return injected
						}
						cancel()
						return nil
					})
					want := injected
					if failure == "cancellation" {
						want = context.Canceled
					}
					if !errors.Is(err, want) || calls != 1 {
						t.Fatalf("%s = calls%d/error%v", failure, calls, err)
					}
				}
				cancel()
				if err := prepared.Close(); err != nil {
					t.Fatal(err)
				}
				if err := prepared.Visit(ctx, func(collabprotocol.Message) error { return nil }); err == nil {
					t.Fatalf("%s snapshot was not closed", failure)
				}
			}
			for _, missing := range []int64{5000, 10001} {
				if _, err := pool.Exec(ctx, `DELETE FROM collaboration_replay_events WHERE incident_id = $1 AND stream_seq = $2`, incidentID, missing); err != nil {
					t.Fatal(err)
				}
				reset := prepare(1)
				if reset.Status() != collabprotocol.ResumeStatusResetNeeded {
					t.Fatalf("missing sequence %d was accepted", missing)
				}
				if err := reset.Visit(ctx, func(collabprotocol.Message) error { t.Fatal("hole emitted partial replay"); return nil }); err != nil {
					t.Fatal(err)
				}
				intenttest.InsertPersistedReplayEventFixtures(t, pool, fixtures[missing-1])
				// The fixture helper advances to its supplied maximum; restore the known high water.
				if _, err := pool.Exec(ctx, `UPDATE collaboration_incident_stream_cursors SET high_water_stream_seq = 10001 WHERE incident_id = $1`, incidentID); err != nil {
					t.Fatal(err)
				}
			}
		})

		retentionNow := time.Now().UTC().Add(time.Hour)
		for _, scenario := range []struct {
			name        string
			count       int
			recentStart int
			recentEnd   int
			recentAge   time.Duration
			wantCount   int
			wantFirst   int64
		}{
			{"low-volume", 3, 0, 0, 0, 3, 1},
			{"exact-count", 10000, 0, 0, 0, 10000, 1},
			{"count-window", 10002, 0, 0, 0, 10000, 3},
			{"time-boundary", 10002, 2, 2, 5 * time.Minute, 10001, 2},
			{"outside-time", 10002, 2, 2, 5*time.Minute + time.Microsecond, 10000, 3},
			{"recent-burst", 10002, 1, 10002, 0, 10002, 1},
			{"clock-regression", 10002, 1, 1, time.Minute, 10002, 1},
		} {
			t.Run(scenario.name, func(t *testing.T) {
				retentionIncidentUUID := uuid.MustParse(createDurableStreamIncident(t, harness, admin, "retention-"+scenario.name))
				retentionPayload := collabtestprotocol.RawPayload(t, collabtestprotocol.NewIncidentJobProgressPayload(
					"retention", retentionIncidentUUID, collabprotocol.JobStatusSucceeded,
					collabprotocol.JobProgress{}, retentionNow,
				))
				if _, err := pool.Exec(ctx, `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, updated_at
) VALUES ($1, $3, $2)
`, retentionIncidentUUID, retentionNow, scenario.count); err != nil {
					t.Fatalf("seed retention cursor: %v", err)
				}
				if _, err := pool.Exec(ctx, `
INSERT INTO collaboration_replay_events (
    event_id, incident_id, stream_seq, intent_key, event_family, canonical_payload, emitted_at
)
SELECT gen_random_uuid(),
       $1::uuid,
       sequence,
       'retention:' || $1::uuid::text || ':' || sequence::text,
       'job_progress',
       $3::jsonb,
       CASE
           WHEN sequence BETWEEN $5 AND $6 THEN $7::timestamptz
           ELSE $2::timestamptz - interval '48 hours'
       END
  FROM generate_series(1, $4::integer) AS sequence
`, retentionIncidentUUID, retentionNow, retentionPayload, scenario.count, scenario.recentStart, scenario.recentEnd, retentionNow.Add(-scenario.recentAge)); err != nil {
					t.Fatalf("seed retention events: %v", err)
				}
				retentionDispatcher := newDispatcherForTest(pool, &recordingBroadcaster{}, func() time.Time { return retentionNow })
				if _, err := retentionDispatcher.RunOnce(ctx); err != nil {
					t.Fatalf("prune retained replay events: %v", err)
				}
				var retainedCount int
				var firstSequence, lastSequence int64
				if err := pool.QueryRow(ctx, `
SELECT count(*), min(stream_seq), max(stream_seq)
  FROM collaboration_replay_events
 WHERE incident_id = $1
`, retentionIncidentUUID).Scan(&retainedCount, &firstSequence, &lastSequence); err != nil {
					t.Fatalf("load retained replay range: %v", err)
				}
				if retainedCount != scenario.wantCount || firstSequence != scenario.wantFirst || lastSequence != int64(scenario.count) {
					t.Fatalf("retained range = %d/%d/%d want %d/%d/%d", retainedCount, firstSequence, lastSequence, scenario.wantCount, scenario.wantFirst, scenario.count)
				}
			})
		}
	})
}
