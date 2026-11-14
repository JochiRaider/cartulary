package revisions_test

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"reflect"
	"sort"
	"testing"
)

func exerciseHistoryOrderAdmission(t *testing.T, h revisionsPortabilityHarness, probe *revisionsPreparationEnvelopeProbe) {
	for _, tc := range []struct {
		name  string
		value any
		valid bool
	}{
		{"one", json.Number("1"), true}, {"large_exact", json.Number("9007199254740993"), true}, {"maximum", json.Number("9223372036854775807"), true},
		{"zero", json.Number("0"), false}, {"negative", json.Number("-1"), false}, {"fraction", json.Number("1.0"), false}, {"exponent", json.Number("1e1"), false}, {"overflow", json.Number("9223372036854775808"), false}, {"string", "1", false}, {"null", nil, false}, {"missing", nil, false}, {"boolean", true, false},
	} {
		t.Run("history_order_scalar/"+tc.name, func(t *testing.T) {
			bundle := h.validBundle(t)
			row := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])[0]
			row["history_order"] = tc.value
			if tc.name == "missing" {
				delete(row, "history_order")
			}
			if !tc.valid {
				row["actor_user_id"] = uuid.NewString()
			}
			bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, []map[string]any{row})
			before := clonePortableBundle(bundle)
			_, err := h.port.PrepareImport(t.Context(), bundle, h.context)
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireRevisionsInvariant(t, err, "revisions.references_complete")
			}
			if !reflect.DeepEqual(bundle, before) || probe.calls != 0 {
				t.Fatal("preparation changed bytes or queried source state")
			}
		})
	}
}
func exercisePortableHistoryOrder(t *testing.T) {
	t.Run("duplicate_order", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "duplicate-order")
		bundle := h.validBundle(t)
		rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
		other := clonePortableRow(t, rows[0])
		other["change_set_id"] = uuid.NewString()
		rows = append(rows, other)
		sort.Slice(rows, func(i, j int) bool { return rows[i]["change_set_id"].(string) < rows[j]["change_set_id"].(string) })
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
		requireRevisionsInvariant(t, h.applyAndValidate(t, bundle), "revisions.history_reconstruction")
	})
	for _, want := range []json.Number{"9007199254740993", "9223372036854775807"} {
		t.Run("portable_order_"+string(want), func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "portable-order")
			bundle := h.validBundle(t)
			rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
			rows[0]["history_order"] = want
			bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
			input := h.freshImportContext()
			prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var stored int64
			expected, err := want.Int64()
			if err != nil {
				t.Fatal(err)
			}
			if err := tx.QueryRow(t.Context(), `SELECT history_order FROM change_sets WHERE change_set_id=$1`, h.changeSet).Scan(&stored); err != nil || stored != expected {
				t.Fatalf("stored=%d want=%d error=%v", stored, expected, err)
			}
			files, err := portabilitysupport.Export(h.port, t.Context(), sourceport.ExportContext{Query: tx, IncidentID: h.incidentID})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, file := range files {
				if file.Path == revisionsTestChangeSetsPath {
					found = true
					got := decodePortableRows(t, file.Payload)
					if len(got) != 1 || got[0]["history_order"] != want {
						t.Fatalf("export lost exact order: %v", got)
					}
				}
			}
			if !found {
				t.Fatal("export omitted change sets")
			}
		})
	}
}
