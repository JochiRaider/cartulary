package artifacts_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
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
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "artifact-rollback-owner@example.test", DisplayName: "Artifact Rollback Owner"})
	childActor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "artifact-child-owner@example.test", DisplayName: "Artifact Child Owner"})
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
    created_at, updated_at, timestamp_utc, outgoing_owner_user_id
)
VALUES (
    $1, $2, $3, 'after title', 'after body', $4, 'briefing',
    'responders', 'bridge', 'after summary', $5,
    $6, 'after current state', 'after checks',
    $7, $6, 'after risks', $8, $6, 'open', $6, $9, $9, $9, $6
)
`, recordID, incident.ID, artifactType,
			"comm-"+recordID.String(), "handoff-"+recordID.String(), actor.ID,
			"status-"+recordID.String(), "lesson-"+recordID.String(), now,
		); err != nil {
			t.Fatalf("seed %s artifact parent: %v", artifactType, err)
		}
		return recordID
	}

	captureRetained := func(t testing.TB, tx pgx.Tx, id uuid.UUID) map[string]any {
		t.Helper()
		value, err := contribution.Records[0].DeleteRestoreSource.SnapshotTx(ctx, tx, id)
		if err != nil {
			t.Fatal(err)
		}
		value["snapshot_schema_id"] = "cartulary.revisions.snapshot.artifact.v2"
		return value
	}
	retainedFinding := func(id uuid.UUID) map[string]any {
		return map[string]any{"record_id": id.String(), "incident_id": incident.ID.String(), "kind": "finding", "statement": "retained statement", "state": "open", "confidence_score": json.Number("25"), "owner_user_id": actor.ID.String(), "closed_at": nil, "created_at": now.Format(time.RFC3339Nano), "updated_at": now.Format(time.RFC3339Nano)}
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
VALUES ($1, $2, 'finding', 'after statement', 'open', 90, $3, $4, $5)
`, recordID, incident.ID, childActor.ID, now.Add(10*time.Minute), now.Add(20*time.Minute)); err != nil {
					t.Fatalf("seed finding subtype: %v", err)
				}
			},
			source: func(uuid.UUID) map[string]any {
				return map[string]any{"artifact_type": "finding", "kind": "hypothesis", "statement": "before statement", "state": "open", "confidence_score": json.Number("25"), "owner_user_id": actor.ID.String()}
			},
			assert: func(t testing.TB, recordID uuid.UUID) {
				requireArtifactText(t, harness, `SELECT statement FROM artifact_findings WHERE record_id = $1`, recordID, "before statement")
				var score int
				if err := harness.DB.QueryRow(ctx, `SELECT confidence_score FROM artifact_findings WHERE record_id=$1`, recordID).Scan(&score); err != nil {
					t.Fatal(err)
				}
				if score != 25 {
					t.Fatalf("restored finding confidence = %d, want 25", score)
				}
				requireArtifactTimestamp(t, harness, recordID, now.Add(time.Hour))
			},
		},
		{
			name: "investigative_query", artifactType: "investigative_query",
			seedSubtype: func(t testing.TB, recordID uuid.UUID) {
				t.Helper()
				if _, err := harness.DB.Exec(ctx, `
INSERT INTO artifact_investigative_queries (record_id, incident_id, query_id, platform, purpose, query_text, created_by_user_id, created_at, updated_at)
VALUES ($1, $2, $3, 'edr', 'after purpose', 'after query', $4, $5, $6)
`, recordID, incident.ID, "query-"+recordID.String(), childActor.ID, now.Add(10*time.Minute), now.Add(20*time.Minute)); err != nil {
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
VALUES ($1, $2, $3, 'after-pattern', 'after reason', 'literal', false, $4, $5)
`, recordID, incident.ID, "keyword-"+recordID.String(), now.Add(10*time.Minute), now.Add(20*time.Minute)); err != nil {
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
			snapshot, err := contribution.Records[0].DeleteRestoreSource.SnapshotTx(ctx, tx, recordID)
			if err != nil {
				t.Fatal(err)
			}
			record, recordOK := snapshot["record"].(map[string]any)
			source, sourceOK := snapshot["source"].(map[string]any)
			if !recordOK || !sourceOK {
				t.Fatalf("invalid captured shape: %#v", snapshot)
			}
			validate := contribution.Records[0].ValidateSnapshotSource
			if validate == nil {
				t.Fatal("missing source validator")
			}
			if err := validate(record, source); err != nil {
				t.Fatalf("%s full source rejected: %v", tc.artifactType, err)
			}

			if len(source) != 31 || source["created_at"] != "2026-07-30T16:00:00+00:00" || source["updated_at"] != "2026-07-30T16:00:00+00:00" || source["created_by_user_id"] != actor.ID.String() || source["owner_user_id"] != actor.ID.String() {
				t.Fatalf("parent facts lost in capture: %#v", source)
			}
			for _, name := range []string{"finding", "investigative_query", "forensic_keyword"} {
				if name != tc.artifactType {
					if source[name] != nil {
						t.Fatalf("unexpected %s child", name)
					}
					continue
				}
				child, ok := source[name].(map[string]any)
				if !ok || child["record_id"] != recordID.String() || child["incident_id"] != incident.ID.String() || child["created_at"] != "2026-07-30T16:10:00+00:00" || child["updated_at"] != "2026-07-30T16:20:00+00:00" {
					t.Fatalf("child facts lost: %#v", child)
				}
				if name == "finding" && child["owner_user_id"] != childActor.ID.String() {
					t.Fatal("finding owner replaced with parent owner")
				}
				if name == "investigative_query" && child["created_by_user_id"] != childActor.ID.String() {
					t.Fatal("query creator replaced with parent creator")
				}
				for _, key := range slices.Sorted(maps.Keys(child)) {
					for _, missing := range []bool{true, false} {
						value, nested := maps.Clone(source), maps.Clone(child)
						value[name] = nested
						if missing {
							delete(nested, key)
						} else {
							nested[key] = []any{"wrong type"}
						}
						if err := validate(record, value); err == nil {
							t.Errorf("invalid nested %s.%s admitted", name, key)
						}
					}
				}
				for _, key := range []string{"record_id", "incident_id"} {
					value, nested := maps.Clone(source), maps.Clone(child)
					value[name], nested[key] = nested, "ffffffff-ffff-4fff-8fff-ffffffffffff"
					if err := validate(record, value); err == nil {
						t.Errorf("mismatched nested %s accepted", key)
					}
				}
				value := maps.Clone(source)
				value["updated_at"] = "2000-01-01T00:00:00Z"
				if err := validate(record, value); err == nil {
					t.Fatal("parent clock validation hidden by child")
				}
			}

			requireArtifactHistoricalSemantics(t, tc.artifactType, record, source, validate)
			for _, key := range slices.Sorted(maps.Keys(source)) {
				missing := maps.Clone(source)
				delete(missing, key)
				if err := validate(record, missing); err == nil {
					t.Errorf("missing %s accepted", key)
				}
				mistyped := maps.Clone(source)
				mistyped[key] = []any{"not a scalar"}
				if err := validate(record, mistyped); err == nil {
					t.Errorf("mistyped %s accepted", key)
				}
			}
			extra := maps.Clone(source)
			extra["projection_summary"] = "derived"
			if err := validate(record, extra); err == nil {
				t.Error("extra source field accepted")
			}
			retainedSource := maps.Clone(source)
			target := retainedSource
			if child, ok := source[tc.artifactType].(map[string]any); ok {
				target = maps.Clone(child)
				retainedSource[tc.artifactType] = target
			}
			for key, value := range tc.source(recordID) {
				if key != "artifact_type" {
					target[key] = value
				}
			}
			err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
				RecordID: recordID, ActorUserID: actor.ID, Now: now.Add(time.Hour),
				NextRowVersion: 3, RetainedValue: map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.artifact.v2", "record": record, "source": retainedSource},
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
		retained := captureRetained(t, tx, recordID)
		retained["source"].(map[string]any)["artifact_type"] = "finding"
		retained["source"].(map[string]any)["finding"] = retainedFinding(recordID)
		if err := contribution.Records[0].ValidateSnapshotSource(retained["record"].(map[string]any), retained["source"].(map[string]any)); err != nil {
			t.Fatalf("wrong-subtype fixture must itself be valid: %v", err)
		}
		err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
			RecordID: recordID, ActorUserID: actor.ID, Now: now.Add(2 * time.Hour), RetainedValue: retained,
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
		defer func() { _ = tx.Rollback(ctx) }()
		note := captureRetained(t, tx, noteID)
		note["source"].(map[string]any)["title"] = "must roll back"
		if err := provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
			RecordID: noteID, ActorUserID: actor.ID, Now: now.Add(3 * time.Hour), RetainedValue: note,
		}); err != nil {
			t.Fatalf("first atomic restore: %v", err)
		}
		missing := captureRetained(t, tx, missingSubtypeID)
		missing["source"].(map[string]any)["finding"] = retainedFinding(missingSubtypeID)
		if err := contribution.Records[0].ValidateSnapshotSource(missing["record"].(map[string]any), missing["source"].(map[string]any)); err != nil {
			t.Fatalf("missing-row fixture must itself be valid: %v", err)
		}
		err = provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
			RecordID: missingSubtypeID, ActorUserID: actor.ID, Now: now.Add(3 * time.Hour), RetainedValue: missing,
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

func requireArtifactHistoricalSemantics(t testing.TB, artifactType string, record, source map[string]any, validate func(map[string]any, map[string]any) error) {
	t.Helper()
	active := source
	if child, ok := source[artifactType].(map[string]any); ok {
		active = child
	}
	check := func(name string, changes map[string]any, wantValid bool) {
		t.Helper()
		value := maps.Clone(source)
		child, nested := source[artifactType].(map[string]any)
		if nested {
			child = maps.Clone(child)
			value[artifactType] = child
		}
		for key, replacement := range changes {
			if _, belongsToChild := child[key]; nested && belongsToChild {
				child[key] = replacement
			} else {
				value[key] = replacement
			}
		}
		before, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		err = validate(record, value)
		if (err == nil) != wantValid {
			t.Errorf("%s/%s admission=%v; want valid=%t", artifactType, name, err, wantValid)
		}
		after, marshalErr := json.Marshal(value)
		if marshalErr != nil || string(before) != string(after) {
			t.Fatalf("%s/%s changed retained bytes", artifactType, name)
		}
	}

	type textBinding struct {
		field               string
		limit               int
		multiline, nullable bool
	}
	bindings := map[string][]textBinding{
		"note":                {{"title", 512, false, true}, {"body", 16384, true, true}},
		"comm_log":            {{"audience", 256, false, false}, {"channel_or_meeting", 512, false, false}, {"summary", 512, false, false}, {"privilege_tag", 512, false, true}},
		"handoff":             {{"current_state_summary", 16384, true, false}, {"next_checks", 16384, true, true}},
		"status_review":       {{"current_state_summary", 16384, true, false}, {"active_risks_summary", 16384, true, true}},
		"lesson":              {{"summary", 512, false, false}},
		"finding":             {{"statement", 16384, true, false}},
		"investigative_query": {{"platform", 512, false, false}, {"purpose", 512, false, false}, {"query_text", 16384, true, false}},
		"forensic_keyword":    {{"pattern", 512, false, false}, {"reason", 4096, true, false}},
	}
	for _, b := range bindings[artifactType] {
		for _, value := range []string{"é  preserved\u200d text", strings.Repeat("😀", b.limit)} {
			check("canonical "+b.field, map[string]any{b.field: value}, true)
		}
		check("nullable "+b.field, map[string]any{b.field: nil}, b.nullable)
		for _, value := range []string{"", " ", " padded", "padded ", "decomposed e\u0301", strings.Repeat("😀", b.limit+1), "before\r\nafter", "before\rafter"} {
			check("noncanonical "+b.field, map[string]any{b.field: value}, false)
		}
		check("multiline "+b.field, map[string]any{b.field: "before\n\tafter"}, b.multiline)
		for r := rune(0); r <= 159; r++ {
			if r > 31 && r < 127 {
				continue
			}
			if b.multiline && (r == '\n' || r == '\t') {
				continue
			}
			check("control "+b.field, map[string]any{b.field: "before" + string(r) + "after"}, false)
		}
	}
	check("reversed source clocks", map[string]any{"updated_at": "2000-01-01T00:00:00Z"}, false)
	check("unpersistable text", map[string]any{"body": "retained\x00text"}, false)
	requiredText, requiredValues := []string{}, []string{}
	switch artifactType {
	case "note":
		check("cleared historical note", map[string]any{"title": nil, "body": nil}, true)
	case "comm_log":
		requiredText = []string{"comm_id", "audience", "channel_or_meeting", "summary"}
		requiredValues = []string{"timestamp_utc"}
		for _, value := range []string{"meeting", "notification", "approval", "briefing", "handoff"} {
			check("communication token", map[string]any{"comm_type": value}, true)
		}
		check("unknown communication", map[string]any{"comm_type": "other"}, false)
	case "handoff":
		requiredText = []string{"handoff_id", "current_state_summary"}
		requiredValues = []string{"timestamp_utc", "outgoing_owner_user_id", "incoming_owner_user_id"}
		check("future explicit acknowledgement", map[string]any{"acknowledged_at": "2030-01-01T00:00:00Z"}, true)
	case "status_review":
		requiredText = []string{"status_review_id", "current_state_summary"}
		requiredValues = []string{"timestamp_utc", "review_owner_user_id"}
	case "lesson":
		requiredText = []string{"lesson_id", "summary"}
		requiredValues = []string{"timestamp_utc", "owner_user_id"}
		for _, value := range []any{nil, "open", "closed"} {
			check("lesson closure", map[string]any{"closure_state": value}, true)
		}
		check("unknown lesson closure", map[string]any{"closure_state": "done"}, false)
	case "finding":
		requiredText = []string{"statement"}
		requiredValues = []string{"owner_user_id"}
		for _, kind := range []string{"finding", "hypothesis"} {
			for _, score := range []any{nil, json.Number("0"), json.Number("100")} {
				check("open finding", map[string]any{"kind": kind, "confidence_score": score}, true)
			}
		}
		check("closed finding", map[string]any{"state": "closed", "closed_at": active["updated_at"]}, true)
		for _, field := range []string{"kind", "state"} {
			check("unknown "+field, map[string]any{field: "unknown"}, false)
		}
		for _, score := range []string{"-1", "101", "9007199254740993"} {
			check("invalid confidence", map[string]any{"confidence_score": json.Number(score)}, false)
		}
		check("closed without instant", map[string]any{"state": "closed"}, false)
		check("open with instant", map[string]any{"closed_at": active["updated_at"]}, false)
		for _, value := range []string{"2000-01-01T00:00:00Z", "2030-01-01T00:00:00Z"} {
			check("closure outside lifetime", map[string]any{"state": "closed", "closed_at": value}, false)
		}
	case "investigative_query":
		requiredText = []string{"query_id", "platform", "purpose", "query_text"}
		requiredValues = []string{"created_by_user_id"}
	case "forensic_keyword":
		requiredText = []string{"keyword_id", "pattern", "reason"}
		for _, value := range []string{"literal", "regex"} {
			check("match mode", map[string]any{"match_mode": value}, true)
		}
		check("unknown match mode", map[string]any{"match_mode": "glob"}, false)
	}
	for _, field := range requiredText {
		for _, value := range []any{nil, "", " \t "} {
			check("required "+field, map[string]any{field: value}, false)
		}
	}
	for _, field := range requiredValues {
		check("required "+field, map[string]any{field: nil}, false)
	}
}
