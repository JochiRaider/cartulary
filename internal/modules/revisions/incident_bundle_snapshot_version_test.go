package revisions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func exerciseRevisionSnapshotVersionAgreement(t *testing.T) {
	for _, test := range []struct {
		name    string
		version int64
		defect  string
	}{
		{"creation", 1, ""},
		{"complete_matching_pair", 3, ""},
		{"partial_exact_large_pair", 9007199254740993, ""},
		{"partial_max_pair", math.MaxInt64, ""},
		{"outer_disagrees", 3, "outer"},
		{"before_disagrees", 3, "before"},
		{"after_disagrees", 3, "after"},
		{"before_has_no_positive_predecessor", 1, "before"},
		{"large_before_disagrees", 9007199254740993, "before"},
		{"references_precede", 3, "references"},
		{"attribution_precedes", 3, "attribution"},
		{"sequence_precedes", 3, "sequence"},
		{"version_uniqueness_precedes", 3, "version_unique"},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "snapshot-version")
			for _, table := range []string{"records", "hosts"} {
				if _, err := h.db.Exec(t.Context(), "UPDATE "+table+" SET row_version=$2 WHERE record_id=$1", h.recordID, test.version); err != nil {
					t.Fatal(err)
				}
			}
			setVersion := func(snapshot map[string]any, version int64) {
				for _, key := range []string{"record", "source"} {
					snapshot[key].(map[string]any)["row_version"] = json.Number(strconv.FormatInt(version, 10))
				}
			}
			setVersion(h.snapshot, test.version)
			base := h.validBundle(t)
			history := decodePortableRows(t, base[revisionsTestHistoryPath])
			history[0]["row_version"] = json.Number(strconv.FormatInt(test.version, 10))
			if test.version > 1 {
				before := clonePortableRow(t, h.snapshot)
				setVersion(before, test.version-1)
				history[0]["before_json"] = before
			}
			withPrefix := func(rows []map[string]any) []map[string]any {
				if test.version != 3 {
					return rows
				}
				var prefix []map[string]any
				var previous map[string]any
				for version := int64(1); version < test.version; version++ {
					revision := clonePortableRow(t, rows[0])
					after := clonePortableRow(t, h.snapshot)
					setVersion(after, version)
					revision["revision_id"], revision["row_version"], revision["before_json"], revision["after_json"] = version, version, previous, after
					prefix = append(prefix, revision)
					previous = after
				}
				return append(prefix, rows...)
			}
			targetRevisionID := int64(1)
			if test.version == 3 {
				targetRevisionID = 3
			}
			history[0]["revision_id"] = targetRevisionID
			base[revisionsTestHistoryPath] = encodePortableRows(t, withPrefix(history))
			mutations := decodePortableRows(t, base[revisionsTestMutationsPath])
			mutations[0]["after_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, test.version)
			if test.version > 1 {
				mutations[0]["operation_kind"] = "patch"
				mutations[0]["before_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, test.version-1)
				mutations[0]["before_value"] = history[0]["before_json"]
			}
			base[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
			candidate := portabilitysupport.MapBundle{}
			for path, payload := range base {
				candidate[path] = append([]byte(nil), payload...)
			}
			wantInvariant := "revisions.history_reconstruction"
			switch test.defect {
			case "outer", "references", "attribution", "sequence", "version_unique":
				history[0]["row_version"] = test.version + 1
			case "before":
				history[0]["before_json"] = clonePortableRow(t, h.snapshot)
			case "after":
				after := clonePortableRow(t, h.snapshot)
				setVersion(after, test.version-1)
				history[0]["after_json"] = after
			}
			input := h.freshImportContext()
			switch test.defect {
			case "references":
				mutations[0]["target_id"] = uuid.NewString()
				wantInvariant = "revisions.references_complete"
			case "attribution":
				input.Attributions = nil
				wantInvariant = "revisions.actor_references_complete"
			case "sequence":
				mutations[0]["sequence_no"] = 2
				wantInvariant = "revisions.mutation_sequence_contiguous"
			case "version_unique":
				duplicate := clonePortableRow(t, history[0])
				duplicate["revision_id"] = 4
				history = append(history, duplicate)
				wantInvariant = "revisions.record_version_unique"
			}
			candidate[revisionsTestHistoryPath] = encodePortableRows(t, withPrefix(history))
			candidate[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
			prepared, err := h.port.PrepareImport(t.Context(), candidate, input)
			if err != nil {
				t.Fatalf("prepare scalar-valid candidate: %v", err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			counts := func(query interface {
				QueryRow(context.Context, string, ...any) pgx.Row
			}, want int) {
				t.Helper()
				for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
					var got int
					expected := want
					if table == "record_revisions" && test.version == 3 {
						expected *= 3
					}
					if err := query.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != expected {
						t.Fatalf("%s count=%d want=%d error=%v", table, got, expected, err)
					}
				}
			}
			err = h.port.ApplyImportTx(t.Context(), tx, prepared, input)
			if test.defect == "" {
				if err != nil {
					t.Fatalf("matching snapshot versions rejected: %v", err)
				}
				counts(tx, 1)
				validationErr := h.port.ValidateImportTx(t.Context(), tx, prepared, input)
				if test.version > 3 {
					requireRevisionsInvariant(t, validationErr, "revisions.history_reconstruction")
				} else if validationErr != nil {
					t.Fatalf("complete creation history rejected: %v", validationErr)
				}
			} else {
				requireRevisionsInvariant(t, err, wantInvariant)
				counts(tx, 0)
				if wantInvariant == "revisions.history_reconstruction" && test.version <= 3 {
					valid, err := h.port.PrepareImport(t.Context(), base, input)
					if err != nil {
						t.Fatal(err)
					}
					if err := h.port.ApplyImportTx(t.Context(), tx, valid, input); err != nil {
						t.Fatalf("apply unmodified baseline after rejected candidate: %v", err)
					}
					if err := h.port.ValidateImportTx(t.Context(), tx, valid, input); err != nil {
						t.Fatalf("complete unmodified baseline rejected: %v", err)
					}
					encode := func(value any) any {
						if value == nil {
							return nil
						}
						payload, err := json.Marshal(value)
						if err != nil {
							t.Fatal(err)
						}
						return payload
					}
					// Make storage equal the admitted candidate to isolate history
					// validation from the higher-priority source-state comparison.
					storedVersion, err := strconv.ParseInt(fmt.Sprint(history[0]["row_version"]), 10, 64)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := tx.Exec(t.Context(), `UPDATE record_revisions SET row_version=$2,before_json=$3,after_json=$4 WHERE change_set_id=$1 AND revision_id=$5`, h.changeSet, storedVersion, encode(history[0]["before_json"]), encode(history[0]["after_json"]), targetRevisionID); err != nil {
						t.Fatal(err)
					}
					counts(tx, 1)
					requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), wantInvariant)
				}
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			counts(h.db, 0)
		})
	}
}
