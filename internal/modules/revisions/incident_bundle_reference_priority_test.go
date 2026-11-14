package revisions_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseRevisionsSourceReferencePriority(t *testing.T) {
	t.Helper()
	for _, defect := range []struct {
		name, invariant string
		afterApply      bool
	}{
		{"sequence", "revisions.mutation_sequence_contiguous", false},
		{"version", "revisions.record_version_unique", false},
		{"snapshot", "revisions.history_reconstruction", false},
		{"terminal", "revisions.history_reconstruction", true},
		{"chain", "revisions.history_reconstruction", true},
	} {
		for _, missingFirst := range []bool{false, true} {
			for _, missing := range []bool{false, true} {
				for _, mutationTarget := range []bool{false, true} {
					// One pair covers missing mutation-target source; all five
					// conditions independently cover record-revision references.
					if mutationTarget && defect.name != "sequence" {
						continue
					}
					t.Run(fmt.Sprintf("source_reference_%s_first=%t_missing=%t_mutation=%t", defect.name, missingFirst, missing, mutationTarget), func(t *testing.T) {
						h := newRevisionsPortabilityHarness(t, "source-reference-priority")
						ctx := t.Context()
						otherID := uuid.New()
						if _, err := h.db.Exec(ctx, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, created_at,
 updated_by_user_id, updated_at, row_version)
VALUES ($1, $2, 'host', $3, $4, $3, $4, 1)
`, otherID, h.incidentID, h.actor.ID, h.createdAt); err != nil {
							t.Fatal(err)
						}
						if _, err := h.db.Exec(ctx, `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, row_version,
 created_at, updated_at, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'Portable host', 'second-host', 'canonical', 1, $3, $3, $4, $4)
`, otherID, h.incidentID, h.createdAt, h.actor.ID); err != nil {
							t.Fatal(err)
						}
						otherSnapshot := clonePortableRow(t, h.snapshot)
						otherSnapshot["record"].(map[string]any)["record_id"] = otherID.String()
						otherSource := otherSnapshot["source"].(map[string]any)
						otherSource["record_id"], otherSource["hostname"] = otherID.String(), "second-host"
						bundle := h.validBundle(t)
						mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
						history := decodePortableRows(t, bundle[revisionsTestHistoryPath])
						otherHistory := clonePortableRow(t, history[0])
						otherHistory["record_id"], otherHistory["after_json"] = otherID.String(), otherSnapshot
						if mutationTarget {
							otherMutation := clonePortableRow(t, mutations[0])
							otherMutation["target_id"] = otherID.String()
							otherMutation["after_version_id"] = "host:" + otherID.String() + ":1"
							otherMutation["after_value"] = otherSnapshot
							if missingFirst {
								mutations = append([]map[string]any{otherMutation}, mutations...)
							} else {
								mutations = append(mutations, otherMutation)
							}
							for i, row := range mutations {
								row["sequence_no"] = i + 1
							}
						}
						switch defect.name {
						case "sequence":
							mutations[len(mutations)-1]["sequence_no"] = len(mutations) + 1
						case "version":
							history = append(history, clonePortableRow(t, history[0]))
						case "snapshot":
							delete(history[0]["after_json"].(map[string]any)["source"].(map[string]any), "hostname")
						case "terminal":
							history[0]["after_json"].(map[string]any)["source"].(map[string]any)["display_name"] = "Not current"
						case "chain":
							for _, query := range []string{`UPDATE records SET row_version=2 WHERE record_id=$1`, `UPDATE hosts SET row_version=2 WHERE record_id=$1`} {
								if _, err := h.db.Exec(ctx, query, h.recordID); err != nil {
									t.Fatal(err)
								}
							}
							current := clonePortableRow(t, h.snapshot)
							current["record"].(map[string]any)["row_version"] = 2
							current["source"].(map[string]any)["row_version"] = 2
							wrongBefore := clonePortableRow(t, h.snapshot)
							wrongBefore["source"].(map[string]any)["display_name"] = "Not previous"
							next := clonePortableRow(t, history[0])
							next["row_version"], next["before_json"], next["after_json"] = 2, wrongBefore, current
							history = append(history, next)
						}
						if missingFirst {
							history = append([]map[string]any{otherHistory}, history...)
						} else {
							history = append(history, otherHistory)
						}
						for i, row := range history {
							row["revision_id"] = i + 1
						}
						bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
						bundle[revisionsTestHistoryPath] = encodePortableRows(t, history)
						prepared, err := h.port.PrepareImport(ctx, bundle, h.context)
						if err != nil {
							t.Fatalf("prepare reference priority fixture: %v", err)
						}
						tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
						if err != nil {
							t.Fatal(err)
						}
						defer func() { _ = tx.Rollback(context.Background()) }()
						removeSource := func() {
							t.Helper()
							if !missing {
								return
							}
							if tag, err := tx.Exec(ctx, `DELETE FROM hosts WHERE record_id=$1`, otherID); err != nil || tag.RowsAffected() != 1 {
								t.Fatalf("remove authoritative source: affected=%d err=%v", tag.RowsAffected(), err)
							}
							var envelopes, sources int
							if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM records WHERE record_id=$1), (SELECT count(*) FROM hosts WHERE record_id=$1)`, otherID).Scan(&envelopes, &sources); err != nil || envelopes != 1 || sources != 0 {
								t.Fatalf("reference fixture envelope/source=%d/%d err=%v, want 1/0", envelopes, sources, err)
							}
						}
						if !defect.afterApply {
							removeSource()
						}
						err = h.port.ApplyImportTx(ctx, tx, prepared, h.context)
						if defect.afterApply {
							if err != nil {
								t.Fatalf("apply valid source references: %v", err)
							}
							removeSource()
							err = h.port.ValidateImportTx(ctx, tx, prepared, h.context)
						}
						want := defect.invariant
						if missing {
							want = "revisions.references_complete"
						}
						requireRevisionsInvariant(t, err, want)
						for _, table := range []struct {
							name  string
							count int
						}{
							{"change_sets", 1}, {"change_set_mutations", len(mutations)}, {"record_revisions", len(history)},
						} {
							wantCount := 0
							if defect.afterApply {
								wantCount = table.count
							}
							var count int
							if err := tx.QueryRow(ctx, "SELECT count(*) FROM "+table.name+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != wantCount {
								t.Fatalf("%s rows=%d want=%d err=%v", table.name, count, wantCount, err)
							}
						}
						if err := tx.Rollback(ctx); err != nil {
							t.Fatal(err)
						}
						for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
							var count int
							if err := h.db.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != 0 {
								t.Fatalf("rollback %s rows=%d want=0 err=%v", table, count, err)
							}
						}
						var restored int
						if err := h.db.QueryRow(ctx, `SELECT count(*) FROM hosts WHERE record_id=$1`, otherID).Scan(&restored); err != nil || restored != 1 {
							t.Fatalf("caller rollback source count=%d want=1 err=%v", restored, err)
						}
					})
				}
			}
		}
	}
}
