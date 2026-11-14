package revisions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

func exerciseCompleteRowHistory(t *testing.T) {
	for _, defect := range []string{"", "missing_creation", "only_terminal", "gap", "creation_before", "later_before_null", "creation_after_null", "later_after_null", "terminal_after_null"} {
		name := defect
		if name == "" {
			name = "complete_three_revision_chain"
		}
		t.Run(name, func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "chain-origin")
			for _, table := range []string{"records", "hosts"} {
				if _, err := h.db.Exec(t.Context(), "UPDATE "+table+" SET row_version=3 WHERE record_id=$1", h.recordID); err != nil {
					t.Fatal(err)
				}
			}
			base := h.validBundle(t)
			firstHistory := decodePortableRows(t, base[revisionsTestHistoryPath])[0]
			firstMutation := decodePortableRows(t, base[revisionsTestMutationsPath])[0]
			var history, mutations []map[string]any
			var previous map[string]any
			for version := 1; version <= 3; version++ {
				snapshot := clonePortableRow(t, h.snapshot)
				for _, member := range []string{"record", "source"} {
					snapshot[member].(map[string]any)["row_version"] = version
				}
				revision := clonePortableRow(t, firstHistory)
				revision["revision_id"], revision["row_version"], revision["before_json"], revision["after_json"] = version, version, previous, snapshot
				mutation := clonePortableRow(t, firstMutation)
				mutation["sequence_no"], mutation["before_value"], mutation["after_value"] = version, previous, snapshot
				mutation["after_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, version)
				if version > 1 {
					mutation["operation_kind"] = "patch"
					mutation["before_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, version-1)
				}
				history, mutations = append(history, revision), append(mutations, mutation)
				previous = snapshot
			}
			base[revisionsTestHistoryPath], base[revisionsTestMutationsPath] = encodePortableRows(t, history), encodePortableRows(t, mutations)
			candidate := h.validBundle(t)
			for path, payload := range base {
				candidate[path] = append([]byte(nil), payload...)
			}
			history = decodePortableRows(t, candidate[revisionsTestHistoryPath])
			switch defect {
			case "missing_creation":
				history = history[1:]
			case "only_terminal":
				history = history[2:]
			case "gap":
				history = append(history[:1], history[2:]...)
			case "creation_before":
				history[0]["before_json"] = clonePortableRow(t, h.snapshot)
			case "later_before_null":
				history[1]["before_json"] = nil
			case "creation_after_null":
				history[0]["after_json"] = nil
			case "later_after_null":
				history[1]["after_json"] = nil
			case "terminal_after_null":
				history[2]["after_json"] = nil
			}
			candidate[revisionsTestHistoryPath] = encodePortableRows(t, history)
			input := h.freshImportContext()
			prepared, err := h.port.PrepareImport(t.Context(), candidate, input)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			counts := func(wantChanges, wantMutations, wantRevisions int) {
				t.Helper()
				for table, want := range map[string]int{"change_sets": wantChanges, "change_set_mutations": wantMutations, "record_revisions": wantRevisions} {
					var got int
					if err := tx.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != want {
						t.Fatalf("%s count=%d want=%d error=%v", table, got, want, err)
					}
				}
			}
			localDefect := defect == "creation_before" || defect == "later_before_null" || defect == "creation_after_null" || defect == "later_after_null" || defect == "terminal_after_null"
			err = h.port.ApplyImportTx(t.Context(), tx, prepared, input)
			if localDefect {
				requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
				counts(0, 0, 0)
				valid, err := h.port.PrepareImport(t.Context(), base, input)
				if err != nil {
					t.Fatal(err)
				}
				if err := h.port.ApplyImportTx(t.Context(), tx, valid, input); err != nil {
					t.Fatal(err)
				}
				if err := h.port.ValidateImportTx(t.Context(), tx, valid, input); err != nil {
					t.Fatalf("complete baseline rejected: %v", err)
				}
				if _, err := tx.Exec(t.Context(), `DELETE FROM record_revisions WHERE change_set_id=$1`, h.changeSet); err != nil {
					t.Fatal(err)
				}
				encode := func(value any) any {
					if value == nil {
						return nil
					}
					raw, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					return raw
				}
				for _, row := range history {
					revisionID, err := row["revision_id"].(json.Number).Int64()
					if err != nil {
						t.Fatal(err)
					}
					version, err := row["row_version"].(json.Number).Int64()
					if err != nil {
						t.Fatal(err)
					}
					if _, err := tx.Exec(t.Context(), `INSERT INTO record_revisions(revision_id,change_set_id,record_id,row_version,before_json,after_json,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, revisionID, h.changeSet, h.recordID, version, encode(row["before_json"]), encode(row["after_json"]), h.createdAt); err != nil {
						t.Fatal(err)
					}
				}
			} else if err != nil {
				t.Fatalf("row-local valid history rejected before aggregate validation: %v", err)
			}
			counts(1, 3, len(history))
			err = h.port.ValidateImportTx(t.Context(), tx, prepared, input)
			if defect == "" {
				if err != nil {
					t.Fatalf("complete chain rejected: %v", err)
				}
			} else {
				requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
				var got int
				if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != 0 {
					t.Fatalf("%s rollback count=%d error=%v", table, got, err)
				}
			}
		})
	}
}
