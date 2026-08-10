package artifacts_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestArtifactRollbackProviderExecutionMatrix(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-rollback-execution")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-rollback-owner@example.test",
		"Artifact Rollback Owner",
		"ArtifactRollbackOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-rollback-incident",
		"IR-ARTIFACTS-ROLLBACK",
		"Artifact rollback execution",
	)
	now := time.Date(2026, 7, 30, 16, 0, 0, 0, time.UTC)
	contribution, err := artifacts.NewRevisionContribution()
	if err != nil {
		t.Fatalf("construct artifact revision contribution: %v", err)
	}
	provider := contribution.Records[0].RowRollbackProvider

	seedParent := func(t testing.TB, artifactType string) uuid.UUID {
		t.Helper()
		recordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "artifact", now)
		if _, err := harness.DB.Exec(ctx, `
INSERT INTO artifacts (
    record_id, incident_id, artifact_type, title, body, comm_id, comm_type,
    audience, channel_or_meeting, summary, handoff_id,
    incoming_owner_user_id, current_state_summary, next_checks,
    status_review_id, review_owner_user_id, active_risks_summary,
    lesson_id, owner_user_id, closure_state, created_by_user_id,
    created_at, updated_at
)
VALUES (
    $1, $2, $3, 'after title', 'after body', $4, 'briefing',
    'responders', 'bridge', 'after summary', $5,
    $6, 'after current state', 'after checks',
    $7, $6, 'after risks', $8, $6, 'open', $6, $9, $9
)
`, recordID, incident.ID, artifactType,
			"comm-"+recordID.String(), "handoff-"+recordID.String(), actor.ID,
			"status-"+recordID.String(), "lesson-"+recordID.String(), now,
		); err != nil {
			t.Fatalf("seed %s artifact parent: %v", artifactType, err)
		}
		return recordID
	}

	type rollbackCase struct {
		name         string
		artifactType string
		seedSubtype  func(testing.TB, uuid.UUID)
		source       func(uuid.UUID) map[string]any
		assert       func(testing.TB, uuid.UUID)
	}
	cases := []rollbackCase{
		{
			name: "note", artifactType: "note",
			source: func(uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "note", "title": "before title", "body": "before body"}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT title FROM artifacts WHERE record_id = $1`, recordID, "before title")
			},
		},
		{
			name: "comm_log", artifactType: "comm_log",
			source: func(recordID uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "comm_log", "comm_id": "comm-" + recordID.String(), "summary": "before summary"}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT summary FROM artifacts WHERE record_id = $1`, recordID, "before summary")
			},
		},
		{
			name: "handoff", artifactType: "handoff",
			source: func(recordID uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "handoff", "handoff_id": "handoff-" + recordID.String(), "next_checks": "before checks"}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT next_checks FROM artifacts WHERE record_id = $1`, recordID, "before checks")
			},
		},
		{
			name: "status_review", artifactType: "status_review",
			source: func(recordID uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "status_review", "status_review_id": "status-" + recordID.String(), "active_risks_summary": "before risks"}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT active_risks_summary FROM artifacts WHERE record_id = $1`, recordID, "before risks")
			},
		},
		{
			name: "lesson", artifactType: "lesson",
			source: func(recordID uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "lesson", "lesson_id": "lesson-" + recordID.String(), "summary": "before lesson"}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT summary FROM artifacts WHERE record_id = $1`, recordID, "before lesson")
			},
		},
		{
			name: "finding", artifactType: "finding",
			seedSubtype: func(t testing.TB, recordID uuid.UUID) {
				t.Helper()
				if _, err := harness.DB.Exec(ctx, `
INSERT INTO artifact_findings (record_id, incident_id, kind, statement, state, confidence_score, owner_user_id, created_at, updated_at)
VALUES ($1, $2, 'finding', 'after statement', 'open', 90, $3, $4, $4)
`, recordID, incident.ID, actor.ID, now); err != nil {
					t.Fatalf("seed finding subtype: %v", err)
				}
			},
			source: func(uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "finding", "kind": "hypothesis", "statement": "before statement", "state": "open", "confidence_score": 25, "owner_user_id": actor.ID.String()}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT statement FROM artifact_findings WHERE record_id = $1`, recordID, "before statement")
				requireArtifactTimestamp(t, harness, recordID, now.Add(time.Hour))
			},
		},
		{
			name: "investigative_query", artifactType: "investigative_query",
			seedSubtype: func(t testing.TB, recordID uuid.UUID) {
				t.Helper()
				if _, err := harness.DB.Exec(ctx, `
INSERT INTO artifact_investigative_queries (record_id, incident_id, query_id, platform, purpose, query_text, created_by_user_id, created_at, updated_at)
VALUES ($1, $2, $3, 'edr', 'after purpose', 'after query', $4, $5, $5)
`, recordID, incident.ID, "query-"+recordID.String(), actor.ID, now); err != nil {
					t.Fatalf("seed investigative query subtype: %v", err)
				}
			},
			source: func(recordID uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "investigative_query", "query_id": "query-" + recordID.String(), "purpose": "before purpose"}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT purpose FROM artifact_investigative_queries WHERE record_id = $1`, recordID, "before purpose")
				requireArtifactTimestamp(t, harness, recordID, now.Add(time.Hour))
			},
		},
		{
			name: "forensic_keyword", artifactType: "forensic_keyword",
			seedSubtype: func(t testing.TB, recordID uuid.UUID) {
				t.Helper()
				if _, err := harness.DB.Exec(ctx, `
INSERT INTO artifact_forensic_keywords (record_id, incident_id, keyword_id, pattern, reason, match_mode, case_sensitive, created_at, updated_at)
VALUES ($1, $2, $3, 'after-pattern', 'after reason', 'literal', false, $4, $4)
`, recordID, incident.ID, "keyword-"+recordID.String(), now); err != nil {
					t.Fatalf("seed forensic keyword subtype: %v", err)
				}
			},
			source: func(recordID uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "forensic_keyword", "keyword_id": "keyword-" + recordID.String(), "reason": "before reason", "match_mode": "regex", "case_sensitive": true}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT reason FROM artifact_forensic_keywords WHERE record_id = $1`, recordID, "before reason")
				requireArtifactTimestamp(t, harness, recordID, now.Add(time.Hour))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recordID := seedParent(t, tc.artifactType)
			if tc.seedSubtype != nil {
				tc.seedSubtype(t, recordID)
			}
			tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin %s rollback transaction: %v", tc.name, err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
				RecordID: recordID, ActorUserID: actor.ID, Now: now.Add(time.Hour),
				NextRowVersion: 3, RetainedValue: map[string]any{"source": tc.source(recordID)},
			})
			if err != nil {
				t.Fatalf("restore %s: %v", tc.name, err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatalf("commit %s rollback: %v", tc.name, err)
			}
			tc.assert(t, recordID)
		})
	}

	t.Run("wrong subtype is rejected", func(t *testing.T) {
		recordID := seedParent(t, "note")
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin wrong-subtype rollback: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
			RecordID: recordID, ActorUserID: actor.ID, Now: now.Add(2 * time.Hour),
			RetainedValue: map[string]any{"source": map[string]any{"artifact_type": "finding", "statement": "wrong subtype"}},
		})
		if !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("wrong-subtype restore error = %v, want ErrTargetNotReversible", err)
		}
	})

	t.Run("failure rolls back prior source restoration", func(t *testing.T) {
		noteID := seedParent(t, "note")
		missingSubtypeID := seedParent(t, "finding")
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin atomic rollback: %v", err)
		}
		if err := provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
			RecordID: noteID, ActorUserID: actor.ID, Now: now.Add(3 * time.Hour),
			RetainedValue: map[string]any{"source": map[string]any{"artifact_type": "note", "title": "must roll back"}},
		}); err != nil {
			t.Fatalf("first atomic restore: %v", err)
		}
		err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
			RecordID: missingSubtypeID, ActorUserID: actor.ID, Now: now.Add(3 * time.Hour),
			RetainedValue: map[string]any{"source": map[string]any{"artifact_type": "finding", "statement": "missing"}},
		})
		if !errors.Is(err, rollbackcontract.ErrTargetNotFound) {
			t.Fatalf("missing subtype restore error = %v, want ErrTargetNotFound", err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("roll back failed source sequence: %v", err)
		}
		requireArtifactText(t, harness, `SELECT title FROM artifacts WHERE record_id = $1`, noteID, "after title")
	})
}
