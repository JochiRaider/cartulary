package revisions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Integer storage is independent of full retained-chain acceptance: these
// deliberately partial fixtures cannot represent quadrillions of revisions.
func exerciseRevisionIntegerStorage(t *testing.T) {
	for _, want := range []json.Number{"9007199254740993", "9223372036854775807"} {
		t.Run(string(want), func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "exact-integer-storage")
			version, err := want.Int64()
			if err != nil {
				t.Fatal(err)
			}
			before := clonePortableRow(t, h.snapshot)
			for _, member := range []string{"record", "source"} {
				h.snapshot[member].(map[string]any)["row_version"] = want
				before[member].(map[string]any)["row_version"] = json.Number(fmt.Sprint(version - 1))
			}
			bundle := h.validBundle(t)
			history := decodePortableRows(t, bundle[revisionsTestHistoryPath])
			history[0]["row_version"], history[0]["before_json"] = want, before
			bundle[revisionsTestHistoryPath] = encodePortableRows(t, history)
			mutations := decodePortableRows(t, bundle[revisionsTestMutationsPath])
			mutations[0]["operation_kind"], mutations[0]["before_value"] = "patch", before
			mutations[0]["before_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, version-1)
			mutations[0]["after_version_id"] = fmt.Sprintf("host:%s:%d", h.recordID, version)
			bundle[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
			input := h.freshImportContext()
			prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var stored int64
			var revisionRaw, mutationRaw []byte
			if err := tx.QueryRow(t.Context(), `SELECT row_version,after_json FROM record_revisions WHERE change_set_id=$1`, h.changeSet).Scan(&stored, &revisionRaw); err != nil {
				t.Fatal(err)
			}
			if err := tx.QueryRow(t.Context(), `SELECT after_value FROM change_set_mutations WHERE change_set_id=$1`, h.changeSet).Scan(&mutationRaw); err != nil {
				t.Fatal(err)
			}
			if stored != version {
				t.Fatalf("stored version=%d want=%d", stored, version)
			}
			for _, raw := range [][]byte{revisionRaw, mutationRaw} {
				snapshot := decodePortableRows(t, append(raw, '\n'))[0]
				for _, member := range []string{"record", "source"} {
					if snapshot[member].(map[string]any)["row_version"] != want {
						t.Fatalf("lost exact %s integer", member)
					}
				}
			}
			if _, err := tx.Exec(t.Context(), `UPDATE record_revisions SET after_json=jsonb_set(after_json,'{source,row_version}',to_jsonb($2::bigint)) WHERE change_set_id=$1`, h.changeSet, version-1); err != nil {
				t.Fatal(err)
			}
			requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.source_state_exact")
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
				var count int
				if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s cleanup count=%d error=%v", table, count, err)
				}
			}
		})
	}
}
