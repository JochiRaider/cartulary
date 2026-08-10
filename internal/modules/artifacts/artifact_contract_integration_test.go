package artifacts_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestArtifactWorkbookMutationContractMatrix(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-workbook-mutation-contract")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-contract-owner@example.test",
		"Artifact Contract Owner",
		"ArtifactContractOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-contract-incident",
		"IR-ARTIFACTS-CONTRACT",
		"Artifact mutation contract",
	)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-contract"),
	)
	now := time.Date(2026, 7, 30, 14, 0, 0, 0, time.UTC)

	cases := []struct {
		name         string
		viewSchemaID string
		artifactType string
		values       func() map[string]artifacts.FieldValue
		patchField   string
		patchValue   string
	}{
		{"note", artifacts.NotesViewSchemaID, "note", func() map[string]artifacts.FieldValue {
			return artifactTextValues("note.title", "Contract note")
		}, "note.body", "Patched note body"},
		{"comm_log", artifacts.CommLogViewSchemaID, "comm_log", func() map[string]artifacts.FieldValue {
			return artifactTextValues(
				"comm_log.comm_type", "briefing",
				"comm_log.audience", "responders",
				"comm_log.channel_or_meeting", "bridge",
				"comm_log.summary", "Initial briefing",
			)
		}, "comm_log.summary", "Updated briefing"},
		{"handoff", artifacts.HandoffViewSchemaID, "handoff", func() map[string]artifacts.FieldValue {
			return map[string]artifacts.FieldValue{
				"handoff.incoming_owner_user_id": {UUID: &actor.ID},
				"handoff.current_state_summary":  artifactTextValue("Containment is stable"),
			}
		}, "handoff.next_checks", "Review telemetry"},
		{"status_review", artifacts.StatusReviewViewSchemaID, "status_review", func() map[string]artifacts.FieldValue {
			return artifactTextValues("status_review.current_state_summary", "Recovery in progress")
		}, "status_review.active_risks_summary", "One residual risk"},
		{"lesson", artifacts.LessonViewSchemaID, "lesson", func() map[string]artifacts.FieldValue {
			return artifactTextValues("lesson.summary", "Capture volatile data early")
		}, "lesson.summary", "Capture volatile data immediately"},
		{"finding", artifacts.FindingsViewSchemaID, "finding", func() map[string]artifacts.FieldValue {
			return artifactTextValues("finding.statement", "Credential misuse observed")
		}, "finding.statement", "Credential misuse confirmed"},
		{"investigative_query", artifacts.InvestigativeQueriesViewSchemaID, "investigative_query", func() map[string]artifacts.FieldValue {
			return artifactTextValues(
				"investigative_query.platform", "edr",
				"investigative_query.purpose", "scope",
				"investigative_query.query_text", "synthetic query",
			)
		}, "investigative_query.purpose", "containment"},
		{"forensic_keyword", artifacts.ForensicKeywordsViewSchemaID, "forensic_keyword", func() map[string]artifacts.FieldValue {
			return artifactTextValues(
				"forensic_keyword.pattern", "synthetic-pattern",
				"forensic_keyword.reason", "Synthetic fixture",
			)
		}, "forensic_keyword.reason", "Updated synthetic fixture"},
	}

	for index, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if err := artifacts.ValidateCreateParams(artifacts.CreateParams{ViewSchemaID: tc.viewSchemaID}); err == nil {
				t.Fatalf("empty %s create unexpectedly passed validation", tc.viewSchemaID)
			}
			clientTxnID := fmt.Sprintf("txn-artifacts-contract-create-%02d", index)
			command := artifacts.WorkbookCreateCommand{
				ActorUserID: actor.ID,
				IncidentID:  incident.ID,
				Request: artifacts.WorkbookCreateRequest{
					ViewSchemaID: tc.viewSchemaID,
					ClientTxnID:  clientTxnID,
					Values:       tc.values(),
				},
				RequestHash: []byte("hash-" + clientTxnID),
				RequestID:   "req-" + clientTxnID,
				OperationID: artifacts.OperationWorkbookCreate,
				Now:         now.Add(time.Duration(index) * time.Minute),
			}
			created, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("create %s: %v", tc.viewSchemaID, err)
			}
			if created.RecordID == uuid.Nil || created.ChangeSetID == uuid.Nil || created.RowVersion != 1 {
				t.Fatalf("create result is incomplete: %#v", created)
			}
			requireArtifactRecordChangedIntent(t, harness, created, actor.ID)
			requireCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN artifacts a USING (incident_id, record_id)
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND r.row_version = 1
   AND a.artifact_type = $2
   AND a.created_by_user_id = $3
`, created.RecordID, tc.artifactType, actor.ID, 1)
			requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1 AND artifact_type = $2`, created.RecordID, tc.artifactType, 1)
			requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE change_set_id = $1 AND source = $2`, created.ChangeSetID, string(command.OperationID), 1)
			requireCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1`, created.ChangeSetID, 1)
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND row_version = 1`, created.RecordID, 1)

			replayed, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("replay %s create: %v", tc.viewSchemaID, err)
			}
			if !replayed.Replayed || replayed.RecordID != created.RecordID {
				t.Fatalf("create replay = %#v, want original record %s", replayed, created.RecordID)
			}
			requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1`, created.RecordID, 1)
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
			requireCount(t, harness, `SELECT count(*) FROM collaboration_event_intents WHERE source_change_set_id = $1 AND source_record_id = $2 AND event_family = 'record_changed'`, created.ChangeSetID, created.RecordID, 1)

			conflicting := command
			conflicting.RequestHash = []byte("changed-" + clientTxnID)
			if _, err := facade.Create(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
				t.Fatalf("changed create replay error = %v, want client transaction conflict", err)
			}

			patchTxnID := fmt.Sprintf("txn-artifacts-contract-patch-%02d", index)
			patchCommand := artifacts.WorkbookPatchCommand{
				ActorUserID: actor.ID,
				RecordID:    created.RecordID,
				Request: artifacts.WorkbookPatchRequest{
					ViewSchemaID:   tc.viewSchemaID,
					BaseRowVersion: 1,
					ClientTxnID:    patchTxnID,
					Changes: []artifacts.WorkbookPatchChange{{
						FieldKey: tc.patchField,
						Value:    artifactFieldValuePtr(artifactTextValue(tc.patchValue)),
					}},
				},
				RequestHash:         []byte("hash-" + patchTxnID),
				RequestID:           "req-" + patchTxnID,
				OperationID:         artifacts.OperationWorkbookPatch,
				ConflictOperationID: artifacts.OperationConflictResolve,
				Now:                 now.Add(2*time.Hour + time.Duration(index)*time.Minute),
			}
			patched, err := facade.Patch(ctx, patchCommand)
			if err != nil {
				t.Fatalf("patch %s: %v", tc.viewSchemaID, err)
			}
			if patched.RowVersion != 2 || !slices.Contains(patched.ChangedFieldKeys, tc.patchField) {
				t.Fatalf("patch result = %#v, want row version 2 and changed field %s", patched, tc.patchField)
			}
			requireArtifactRecordChangedIntent(t, harness, patched, actor.ID)
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 2)
			requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1 AND row_version = 2`, created.RecordID, 1)

			patchReplay, err := facade.Patch(ctx, patchCommand)
			if err != nil || !patchReplay.Replayed || patchReplay.RecordID != created.RecordID {
				t.Fatalf("patch replay = %#v, %v; want original record", patchReplay, err)
			}
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 2)
			requireCount(t, harness, `SELECT count(*) FROM collaboration_event_intents WHERE source_change_set_id = $1 AND source_record_id = $2 AND event_family = 'record_changed'`, patched.ChangeSetID, created.RecordID, 1)
		})
	}
}

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
	provider := artifacts.RevisionProviderContribution().Records[0].RowRollbackProvider

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

func TestArtifactPatchEnvelopeOutcomes(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-patch-envelope-outcomes")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-envelope-owner@example.test",
		"Artifact Envelope Owner",
		"ArtifactEnvelopeOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-envelope-incident",
		"IR-ARTIFACTS-ENVELOPE",
		"Artifact envelope outcomes",
	)
	now := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-envelope-outcomes"),
	)

	create := func(t testing.TB, suffix string) artifacts.WorkbookMutationResult {
		t.Helper()
		result, err := facade.Create(ctx, artifacts.WorkbookCreateCommand{
			ActorUserID: actor.ID, IncidentID: incident.ID,
			Request: artifacts.WorkbookCreateRequest{
				ViewSchemaID: artifacts.NotesViewSchemaID,
				ClientTxnID:  "txn-artifacts-envelope-create-" + suffix,
				Values:       artifactTextValues("note.title", "Envelope "+suffix),
			},
			RequestHash: []byte("hash-artifacts-envelope-create-" + suffix),
			RequestID:   "req-artifacts-envelope-create-" + suffix,
			OperationID: artifacts.OperationWorkbookCreate,
			Now:         now,
		})
		if err != nil {
			t.Fatalf("create %s envelope fixture: %v", suffix, err)
		}
		return result
	}
	patch := func(recordID uuid.UUID, baseVersion int64, suffix string) (artifacts.WorkbookMutationResult, error) {
		value := artifactTextValue("Patched " + suffix)
		return facade.Patch(ctx, artifacts.WorkbookPatchCommand{
			ActorUserID: actor.ID, RecordID: recordID,
			Request: artifacts.WorkbookPatchRequest{
				ViewSchemaID:   artifacts.NotesViewSchemaID,
				BaseRowVersion: baseVersion,
				ClientTxnID:    "txn-artifacts-envelope-patch-" + suffix,
				Changes: []artifacts.WorkbookPatchChange{{
					FieldKey: "note.title", Value: &value,
				}},
			},
			RequestHash:         []byte("hash-artifacts-envelope-patch-" + suffix),
			RequestID:           "req-artifacts-envelope-patch-" + suffix,
			OperationID:         artifacts.OperationWorkbookPatch,
			ConflictOperationID: artifacts.OperationConflictResolve,
			Now:                 now.Add(time.Minute),
		})
	}

	t.Run("future base version fails with authoritative current version", func(t *testing.T) {
		created := create(t, "stale")
		_, err := patch(created.RecordID, 2, "stale")
		var conflict *artifacts.RowVersionConflictError
		if !errors.As(err, &conflict) || conflict.CurrentRowVersion != 1 || conflict.BaseRowVersion != 2 {
			t.Fatalf("future base-version error = %#v, %v", conflict, err)
		}
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
	})

	t.Run("deleted record requires restore", func(t *testing.T) {
		created := create(t, "deleted")
		if _, err := harness.DB.Exec(ctx, `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = $3
 WHERE record_id = $1
`, created.RecordID, now.Add(30*time.Second), actor.ID); err != nil {
			t.Fatalf("mark artifact envelope deleted: %v", err)
		}
		if _, err := patch(created.RecordID, 1, "deleted"); !errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
			t.Fatalf("deleted artifact patch error = %v, want ErrRecordDeletedUseRestore", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
	})

	t.Run("wrong record type remains concealed", func(t *testing.T) {
		recordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "host", now)
		if _, err := patch(recordID, 1, "wrong-type"); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("wrong-type artifact patch error = %v, want pgx.ErrNoRows", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND source = 'workbook.rows.patch'`, incident.ID, 0)
	})
}

func TestArtifactCollectionMutationContractMatrix(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-collection-contract")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-collection-owner@example.test",
		"Artifact Collection Owner",
		"ArtifactCollectionOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-collection-incident",
		"IR-ARTIFACTS-COLLECTION",
		"Artifact collection contract",
	)
	now := time.Date(2026, 7, 30, 15, 0, 0, 0, time.UTC)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-collection"),
	)

	wantPolicies := map[string]artifacts.CollectionFamily{
		"note.tags":                          artifacts.CollectionFamilyRecordTag,
		"comm_log.decision_ids":              artifacts.CollectionFamilyRecordRef,
		"comm_log.action_task_ids":           artifacts.CollectionFamilyRecordRef,
		"comm_log.audience_party_ids":        artifacts.CollectionFamilyPartyRef,
		"comm_log.attendee_party_ids":        artifacts.CollectionFamilyPartyRef,
		"handoff.open_task_ids":              artifacts.CollectionFamilyRecordRef,
		"handoff.open_decision_ids":          artifacts.CollectionFamilyRecordRef,
		"handoff.open_risk_refs":             artifacts.CollectionFamilyRiskRef,
		"status_review.blocked_task_ids":     artifacts.CollectionFamilyRecordRef,
		"status_review.pending_evidence_ids": artifacts.CollectionFamilyRecordRef,
		"status_review.open_decision_ids":    artifacts.CollectionFamilyRecordRef,
		"lesson.follow_up_task_ids":          artifacts.CollectionFamilyRecordRef,
		"lesson.evidence_refs":               artifacts.CollectionFamilyRecordRef,
		"finding.supporting_refs":            artifacts.CollectionFamilyRecordRef,
		"finding.contradictory_refs":         artifacts.CollectionFamilyRecordRef,
	}
	for fieldKey, family := range wantPolicies {
		policy, ok := artifacts.LookupCollectionPolicy(fieldKey)
		if !ok || policy.FieldKey != fieldKey || policy.Family != family || len(policy.AllowedOps) != 2 {
			t.Fatalf("collection policy %s = %#v, %v; want family %s with two operations", fieldKey, policy, ok, family)
		}
	}
	if _, ok := artifacts.LookupCollectionPolicy("note.body"); ok {
		t.Fatal("scalar field was admitted as an artifact collection")
	}

	targetRecordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "evidence", now)
	partyRecordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "party", now)
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO parties (record_id, incident_id, display_name, party_kind, created_at, updated_at)
VALUES ($1, $2, 'Synthetic responder', 'person', $3, $3)
`, partyRecordID, incident.ID, now); err != nil {
		t.Fatalf("seed party subtype: %v", err)
	}

	cases := []struct {
		name         string
		viewSchemaID string
		values       map[string]artifacts.FieldValue
		collections  map[string]artifacts.WorkbookCollectionActionPayload
		assert       func(artifacts.WorkbookMutationResult)
	}{
		{
			name:         "record_tag",
			viewSchemaID: artifacts.NotesViewSchemaID,
			values:       artifactTextValues("note.title", "Tagged note"),
			collections: map[string]artifacts.WorkbookCollectionActionPayload{
				"note.tags": {Actions: []artifacts.WorkbookCollectionAction{
					{Op: "add_tag", RawText: "Synthetic", NormalizedText: "synthetic"},
					{Op: "add_tag", RawText: "synthetic", NormalizedText: "synthetic"},
				}},
			},
			assert: func(result artifacts.WorkbookMutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_tags WHERE record_id = $1 AND normalized_tag_name = 'synthetic' AND deleted_at IS NULL`, result.RecordID, 1)
			},
		},
		{
			name:         "record_ref",
			viewSchemaID: artifacts.FindingsViewSchemaID,
			values:       artifactTextValues("finding.statement", "Synthetic relationship"),
			collections: map[string]artifacts.WorkbookCollectionActionPayload{
				"finding.supporting_refs": {Actions: []artifacts.WorkbookCollectionAction{
					{Op: "add_record_ref", LinkedRecordID: &targetRecordID},
					{Op: "add_record_ref", LinkedRecordID: &targetRecordID},
				}},
			},
			assert: func(result artifacts.WorkbookMutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND link_type = 'supported_by' AND deleted_at IS NULL`, result.RecordID, targetRecordID, 1)
			},
		},
		{
			name:         "party_ref",
			viewSchemaID: artifacts.CommLogViewSchemaID,
			values: artifactTextValues(
				"comm_log.comm_type", "briefing",
				"comm_log.audience", "responders",
				"comm_log.channel_or_meeting", "bridge",
				"comm_log.summary", "Synthetic party reference",
			),
			collections: map[string]artifacts.WorkbookCollectionActionPayload{
				"comm_log.audience_party_ids": {Actions: []artifacts.WorkbookCollectionAction{
					{Op: "add_party_ref", PartyID: &partyRecordID},
					{Op: "add_party_ref", PartyID: &partyRecordID},
				}},
			},
			assert: func(result artifacts.WorkbookMutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND field_key = 'comm_log.audience_party_ids' AND deleted_at IS NULL`, result.RecordID, partyRecordID, 1)
			},
		},
		{
			name:         "risk_ref",
			viewSchemaID: artifacts.HandoffViewSchemaID,
			values: map[string]artifacts.FieldValue{
				"handoff.incoming_owner_user_id": {UUID: &actor.ID},
				"handoff.current_state_summary":  artifactTextValue("Synthetic handoff"),
			},
			collections: map[string]artifacts.WorkbookCollectionActionPayload{
				"handoff.open_risk_refs": {Actions: []artifacts.WorkbookCollectionAction{
					{Op: "add_risk_ref", RiskRefText: "Synthetic risk", NormalizedText: "synthetic risk"},
					{Op: "add_risk_ref", RiskRefText: "synthetic risk", NormalizedText: "synthetic risk"},
				}},
			},
			assert: func(result artifacts.WorkbookMutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM handoff_risk_refs WHERE handoff_record_id = $1 AND normalized_risk_ref_text = 'synthetic risk' AND deleted_at IS NULL`, result.RecordID, 1)
			},
		},
	}

	for index, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clientTxnID := fmt.Sprintf("txn-artifacts-collection-%02d", index)
			command := artifacts.WorkbookCreateCommand{
				ActorUserID: actor.ID, IncidentID: incident.ID,
				Request: artifacts.WorkbookCreateRequest{
					ViewSchemaID: tc.viewSchemaID,
					ClientTxnID:  clientTxnID,
					Values:       tc.values,
					Collections:  tc.collections,
				},
				RequestHash: []byte("hash-" + clientTxnID),
				RequestID:   "req-" + clientTxnID,
				OperationID: artifacts.OperationWorkbookCreate,
				Now:         now.Add(time.Duration(index) * time.Minute),
			}
			result, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("create collection family %s: %v", tc.name, err)
			}
			tc.assert(result)
			replayed, err := facade.Create(ctx, command)
			if err != nil || !replayed.Replayed || replayed.RecordID != result.RecordID {
				t.Fatalf("collection replay = %#v, %v; want original result", replayed, err)
			}
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, result.RecordID, 1)

			conflicting := command
			conflicting.RequestHash = []byte("changed-" + clientTxnID)
			if _, err := facade.Create(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
				t.Fatalf("changed collection replay error = %v, want client transaction conflict", err)
			}
		})
	}

	beforeRecords := artifactContractCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID)
	_, err := facade.Create(ctx, artifacts.WorkbookCreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID,
		Request: artifacts.WorkbookCreateRequest{
			ViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:  "txn-artifacts-collection-invalid",
			Values:       artifactTextValues("note.title", "Rejected collection"),
			Collections: map[string]artifacts.WorkbookCollectionActionPayload{
				"note.tags": {Actions: []artifacts.WorkbookCollectionAction{{Op: "replace", RawText: "invalid"}}},
			},
		},
		RequestHash: []byte("hash-artifacts-collection-invalid"),
		RequestID:   "req-artifacts-collection-invalid",
		OperationID: artifacts.OperationWorkbookCreate,
		Now:         now.Add(time.Hour),
	})
	if err == nil {
		t.Fatal("invalid collection operation unexpectedly succeeded")
	}
	if got := artifactContractCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID); got != beforeRecords {
		t.Fatalf("invalid collection changed record count: got %d want %d", got, beforeRecords)
	}
}

func TestArtifactImportCreateFacadeContract(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-import-create-contract")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-import-owner@example.test",
		"Artifact Import Owner",
		"ArtifactImportOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-import-incident",
		"IR-ARTIFACTS-IMPORT",
		"Artifact import contract",
	)
	appender := revisionsupport.MustAppender(t)
	now := time.Date(2026, 7, 30, 17, 0, 0, 0, time.UTC)

	artifactTargets := make(map[string]importtargetregistry.Target)
	for _, target := range importtargetregistry.Targets {
		if target.TargetViewSchemaID == nil || target.OwnerContractRef != "module.artifacts@1" {
			continue
		}
		artifactTargets[*target.TargetViewSchemaID] = target
	}
	if len(artifactTargets) != 8 {
		t.Fatalf("generated artifact import targets = %d, want 8", len(artifactTargets))
	}

	cases := []struct {
		viewSchemaID string
		fields       []ownerfacade.ImportFieldValue
	}{
		{artifacts.NotesViewSchemaID, importTextFields("note.title", "Imported note")},
		{artifacts.CommLogViewSchemaID, importTextFields(
			"comm_log.comm_type", "briefing",
			"comm_log.audience", "responders",
			"comm_log.channel_or_meeting", "bridge",
			"comm_log.summary", "Imported briefing",
		)},
		{artifacts.HandoffViewSchemaID, append(
			importTextFields("handoff.current_state_summary", "Imported handoff"),
			ownerfacade.ImportFieldValue{
				FieldKey: "handoff.incoming_owner_user_id",
				NormalizedValue: ownerfacade.ImportScalarValue{
					Kind: "uuid", UUID: &actor.ID,
				},
			},
		)},
		{artifacts.StatusReviewViewSchemaID, importTextFields("status_review.current_state_summary", "Imported status")},
		{artifacts.LessonViewSchemaID, importTextFields("lesson.summary", "Imported lesson")},
	}

	for index, tc := range cases {
		target := artifactTargets[tc.viewSchemaID]
		if target.AvailabilityKind != "enabled" || target.FacadeID == nil {
			t.Fatalf("%s target = %#v, want enabled owner-create facade", tc.viewSchemaID, target)
		}
		facade, err := artifacts.NewImportCreateFacade(
			tc.viewSchemaID,
			*target.FacadeID,
			appender,
			appsupport.ArtifactProjectionRows(harness.DB),
		)
		if err != nil {
			t.Fatalf("construct %s import facade: %v", tc.viewSchemaID, err)
		}
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin %s import transaction: %v", tc.viewSchemaID, err)
		}
		changeSetID, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID: incident.ID, ActorUserID: actor.ID,
			Source:    "imports.unit.apply",
			CreatedAt: now.Add(time.Duration(index) * time.Minute),
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("append %s import change set: %v", tc.viewSchemaID, err)
		}
		response, err := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID: incident.ID, ActorUserID: actor.ID,
				TargetViewSchemaID: tc.viewSchemaID,
				ImportSessionID:    uuid.New(),
				ImportUnitID:       uuid.New(),
				MappingFingerprint: fmt.Sprintf("synthetic-map-%02d", index),
				SourceFileKind:     "csv",
				ParserProfileID:    "synthetic",
				ParserVersion:      "1",
				LocatorKind:        "row",
				Locator:            fmt.Sprintf("%d", index+1),
				ClientTxnID:        fmt.Sprintf("txn-artifacts-import-%02d", index),
				FieldValues:        tc.fields,
			},
			ChangeSetID: changeSetID,
			SequenceNo:  1,
			Now:         now.Add(time.Duration(index) * time.Minute),
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("create %s import row: %v", tc.viewSchemaID, err)
		}
		if response.RecordID == uuid.Nil || response.RowVersion != 1 ||
			response.CreatedOrReused != "created" || response.OwnerResultCode != "created" ||
			response.ChangeSetMutationRef == "" || response.RowRefresh["record_id"] == nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("%s import response is incomplete: %#v", tc.viewSchemaID, response)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit %s import transaction: %v", tc.viewSchemaID, err)
		}
		requireCount(t, harness, `SELECT count(*) FROM artifacts WHERE record_id = $1 AND artifact_type = $2`, response.RecordID, artifacts.ArtifactTypeForView(tc.viewSchemaID), 1)
		requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1`, response.RecordID, 1)
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND change_set_id = $2`, response.RecordID, changeSetID, 1)
	}

	for _, viewSchemaID := range []string{
		artifacts.FindingsViewSchemaID,
		artifacts.InvestigativeQueriesViewSchemaID,
		artifacts.ForensicKeywordsViewSchemaID,
	} {
		target := artifactTargets[viewSchemaID]
		if target.AvailabilityKind != "reserved" || target.FacadeID != nil || target.FacadeBindingID != nil {
			t.Fatalf("%s target = %#v, want reserved with no dispatch facade", viewSchemaID, target)
		}
	}

	target := artifactTargets[artifacts.NotesViewSchemaID]
	facade, err := artifacts.NewImportCreateFacade(
		artifacts.NotesViewSchemaID,
		*target.FacadeID,
		appender,
		appsupport.ArtifactProjectionRows(harness.DB),
	)
	if err != nil {
		t.Fatalf("construct rollback import facade: %v", err)
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin rollback import transaction: %v", err)
	}
	changeSetID, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: incident.ID, ActorUserID: actor.ID, Source: "imports.unit.apply", CreatedAt: now.Add(time.Hour),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("append rollback import change set: %v", err)
	}
	rolledBack, err := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
		Request: ownerfacade.ImportOwnerCreateRequest{
			IncidentID: incident.ID, ActorUserID: actor.ID,
			TargetViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:        "txn-artifacts-import-rollback",
			FieldValues:        importTextFields("note.title", "Rolled back note"),
		},
		ChangeSetID: changeSetID, SequenceNo: 1, Now: now.Add(time.Hour),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("create rollback import row: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback caller-owned import transaction: %v", err)
	}
	requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1`, rolledBack.RecordID, 0)
	requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, changeSetID, 0)
	requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1`, rolledBack.RecordID, 0)
}

func TestArtifactConflictSourceRevalidation(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-conflict-source-revalidation")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-conflict-owner@example.test",
		"Artifact Conflict Owner",
		"ArtifactConflictOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-conflict-incident",
		"IR-ARTIFACTS-CONFLICT",
		"Artifact conflict contract",
	)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-conflict"),
	)
	now := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	created, err := facade.Create(ctx, artifacts.WorkbookCreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID,
		Request: artifacts.WorkbookCreateRequest{
			ViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:  "txn-artifacts-conflict-create",
			Values: artifactTextValues(
				"note.title", "Conflict title",
				"note.body", "Base body",
			),
		},
		RequestHash: []byte("hash-artifacts-conflict-create"),
		RequestID:   "req-artifacts-conflict-create",
		OperationID: artifacts.OperationWorkbookCreate,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create conflict source: %v", err)
	}

	serverBody := artifactTextValue("Server body")
	serverPatch := artifacts.WorkbookPatchCommand{
		ActorUserID: actor.ID, RecordID: created.RecordID,
		Request: artifacts.WorkbookPatchRequest{
			ViewSchemaID:   artifacts.NotesViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-artifacts-conflict-server",
			Changes: []artifacts.WorkbookPatchChange{{
				FieldKey: "note.body", Value: &serverBody,
			}},
		},
		RequestHash:         []byte("hash-artifacts-conflict-server"),
		RequestID:           "req-artifacts-conflict-server",
		OperationID:         artifacts.OperationWorkbookPatch,
		ConflictOperationID: artifacts.OperationConflictResolve,
		Now:                 now.Add(time.Minute),
	}
	if result, err := facade.Patch(ctx, serverPatch); err != nil || result.RowVersion != 2 {
		t.Fatalf("server-side patch = %#v, %v; want row version 2", result, err)
	}

	clientBody := artifactTextValue("Client body")
	stalePatch := artifacts.WorkbookPatchCommand{
		ActorUserID: actor.ID, RecordID: created.RecordID,
		Request: artifacts.WorkbookPatchRequest{
			ViewSchemaID:   artifacts.NotesViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-artifacts-conflict-stale",
			Changes: []artifacts.WorkbookPatchChange{{
				FieldKey: "note.body", Value: &clientBody,
			}},
		},
		RequestHash:         []byte("hash-artifacts-conflict-stale"),
		RequestID:           "req-artifacts-conflict-stale",
		OperationID:         artifacts.OperationWorkbookPatch,
		ConflictOperationID: artifacts.OperationConflictResolve,
		Now:                 now.Add(2 * time.Minute),
	}
	beforeChangeSets := artifactContractCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incident.ID)
	_, err = facade.Patch(ctx, stalePatch)
	var conflict *artifacts.SameFieldConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("stale same-field patch error = %v, want SameFieldConflictError", err)
	}
	if conflict.Conflict["conflict_token"] == nil ||
		conflict.Conflict["field_key"] != "note.body" ||
		conflict.Conflict["current_row_version"] == nil {
		t.Fatalf("same-field conflict payload is incomplete: %#v", conflict.Conflict)
	}
	requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1 AND row_version = 2`, created.RecordID, 1)
	requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 2)
	if got := artifactContractCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incident.ID); got != beforeChangeSets {
		t.Fatalf("stale conflict changed change-set count: got %d want %d", got, beforeChangeSets)
	}

	title := artifactTextValue("Rebased title")
	rebased, err := facade.Patch(ctx, artifacts.WorkbookPatchCommand{
		ActorUserID: actor.ID, RecordID: created.RecordID,
		Request: artifacts.WorkbookPatchRequest{
			ViewSchemaID:   artifacts.NotesViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-artifacts-conflict-rebase",
			Changes: []artifacts.WorkbookPatchChange{{
				FieldKey: "note.title", Value: &title,
			}},
		},
		RequestHash:         []byte("hash-artifacts-conflict-rebase"),
		RequestID:           "req-artifacts-conflict-rebase",
		OperationID:         artifacts.OperationWorkbookPatch,
		ConflictOperationID: artifacts.OperationConflictResolve,
		Now:                 now.Add(3 * time.Minute),
	})
	if err != nil || rebased.RowVersion != 3 {
		t.Fatalf("different-field stale patch = %#v, %v; want row version 3", rebased, err)
	}

	wrongView := stalePatch
	wrongView.Request.ViewSchemaID = artifacts.FindingsViewSchemaID
	wrongView.Request.ClientTxnID = "txn-artifacts-conflict-wrong-view"
	wrongView.RequestHash = []byte("hash-artifacts-conflict-wrong-view")
	if _, err := facade.Patch(ctx, wrongView); err == nil {
		t.Fatal("conflict source revalidation admitted a mismatched artifact view")
	}
	requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1 AND row_version = 3`, created.RecordID, 1)
}

func importTextFields(entries ...string) []ownerfacade.ImportFieldValue {
	fields := make([]ownerfacade.ImportFieldValue, 0, len(entries)/2)
	for index := 0; index < len(entries); index += 2 {
		value := entries[index+1]
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey: entries[index],
			NormalizedValue: ownerfacade.ImportScalarValue{
				Kind: "text", Text: &value,
			},
		})
	}
	return fields
}

func seedArtifactContractRecord(
	t testing.TB,
	harness *appsupport.StoreHarness,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	recordType string,
	now time.Time,
) uuid.UUID {
	t.Helper()
	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin record seed: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	recordID, err := records.NewStore().InsertTx(context.Background(), tx, records.InsertParams{
		IncidentID: incidentID, RecordType: recordType,
		CreatedByUserID: actorID, CreatedAt: now,
		UpdatedByUserID: actorID, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("insert %s record envelope: %v", recordType, err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit %s record envelope: %v", recordType, err)
	}
	return recordID
}

func artifactContractCount(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("count artifact contract state: %v", err)
	}
	return count
}

func requireArtifactRecordChangedIntent(
	t testing.TB,
	harness *appsupport.StoreHarness,
	result artifacts.WorkbookMutationResult,
	actorID uuid.UUID,
) {
	t.Helper()
	requireCount(
		t,
		harness,
		`SELECT count(*) FROM collaboration_event_intents WHERE source_change_set_id = $1 AND source_record_id = $2 AND source_row_version = $3 AND event_family = 'record_changed'`,
		result.ChangeSetID,
		result.RecordID,
		result.RowVersion,
		1,
	)
	var raw []byte
	if err := harness.DB.QueryRow(context.Background(), `
SELECT canonical_payload
  FROM collaboration_event_intents
 WHERE source_change_set_id = $1
   AND source_record_id = $2
   AND source_row_version = $3
   AND event_family = 'record_changed'
`, result.ChangeSetID, result.RecordID, result.RowVersion).Scan(&raw); err != nil {
		t.Fatalf("load artifact record_changed intent: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode artifact record_changed intent: %v", err)
	}
	if payload["record_id"] != result.RecordID.String() ||
		payload["change_set_id"] != result.ChangeSetID.String() ||
		payload["client_txn_id"] != result.ClientTxnID ||
		payload["actor_user_id"] != actorID.String() ||
		payload["row_version"] != float64(result.RowVersion) {
		t.Fatalf("artifact record_changed identity = %#v, want result %#v actor %s", payload, result, actorID)
	}
	gotKeys := jsonStringSlice(t, payload["changed_field_keys"])
	wantKeys := append([]string(nil), result.ChangedFieldKeys...)
	slices.Sort(gotKeys)
	slices.Sort(wantKeys)
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("artifact record_changed changed keys = %#v, want %#v", gotKeys, wantKeys)
	}
	affected, ok := payload["affected_views"].([]any)
	if !ok || len(affected) != 1 {
		t.Fatalf("artifact record_changed affected views = %#v, want exactly one", payload["affected_views"])
	}
	view, ok := affected[0].(map[string]any)
	if !ok || view["view_schema_id"] != result.ViewSchemaID || view["change_kind"] != "patch" {
		t.Fatalf("artifact record_changed affected view = %#v, want patch for %s", affected[0], result.ViewSchemaID)
	}
}

func jsonStringSlice(t testing.TB, raw any) []string {
	t.Helper()
	values, ok := raw.([]any)
	if !ok {
		t.Fatalf("JSON string list has type %T", raw)
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("JSON string list contains %T", value)
		}
		result = append(result, text)
	}
	return result
}

func requireArtifactText(
	t testing.TB,
	harness *appsupport.StoreHarness,
	query string,
	recordID uuid.UUID,
	want string,
) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRow(context.Background(), query, recordID).Scan(&got); err != nil {
		t.Fatalf("query restored artifact text: %v", err)
	}
	if got != want {
		t.Fatalf("restored artifact text = %q, want %q", got, want)
	}
}

func requireArtifactTimestamp(
	t testing.TB,
	harness *appsupport.StoreHarness,
	recordID uuid.UUID,
	want time.Time,
) {
	t.Helper()
	var got time.Time
	if err := harness.DB.QueryRow(context.Background(), `SELECT updated_at FROM artifacts WHERE record_id = $1`, recordID).Scan(&got); err != nil {
		t.Fatalf("query restored artifact parent timestamp: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("restored artifact parent updated_at = %s, want %s", got, want)
	}
}

func requireCount(
	t testing.TB,
	harness *appsupport.StoreHarness,
	query string,
	argsAndWant ...any,
) {
	t.Helper()
	if len(argsAndWant) == 0 {
		t.Fatal("count assertion requires an expected value")
	}
	want := argsAndWant[len(argsAndWant)-1].(int)
	args := argsAndWant[:len(argsAndWant)-1]
	if got := artifactContractCount(t, harness, query, args...); got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func artifactTextValues(entries ...string) map[string]artifacts.FieldValue {
	values := make(map[string]artifacts.FieldValue, len(entries)/2)
	for index := 0; index < len(entries); index += 2 {
		values[entries[index]] = artifactTextValue(entries[index+1])
	}
	return values
}

func artifactTextValue(value string) artifacts.FieldValue {
	return artifacts.FieldValue{Text: &value}
}

func artifactFieldValuePtr(value artifacts.FieldValue) *artifacts.FieldValue {
	return &value
}

func mustArtifactMutationFacade(
	t testing.TB,
	pool postgres.DB,
	codec conflicts.ConflictTokenCodec,
) *artifacts.MutationFacade {
	t.Helper()
	facade, err := workbookassembly.NewArtifactMutationContribution(
		pool,
		codec,
		revisionsupport.MustAppender(t),
		mustConflictFieldResolver(t),
		appsupport.ArtifactProjectionRows(pool),
	)
	if err != nil {
		t.Fatalf("compose Artifacts mutation facade: %v", err)
	}
	return facade
}

func mustConflictFieldResolver(t testing.TB) conflicts.FieldResolver {
	t.Helper()
	resolver, err := revisionassembly.CurrentConflictFieldResolver()
	if err != nil {
		t.Fatalf("compose conflict field resolver: %v", err)
	}
	return resolver
}
