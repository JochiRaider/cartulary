package revisions_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exerciseRevisionContextExportCapacity(t *testing.T, h revisionsPortabilityHarness, probe *revisionsPreparationEnvelopeProbe) {
	for _, scenario := range []string{"exact_ceiling", "one_byte_over", "escaping_expansion", "capacity_precedes_actor"} {
		t.Run("context_capacity_"+scenario, func(t *testing.T) {
			bundle := h.validBundle(t)
			row := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])[0]
			row["reason"] = ""
			expected := clonePortableRow(t, row)
			expected["actor_import_contexts"] = []any{map[string]any{"import_id": h.context.OperationID, "actor_user_id": h.actor.ID.String()}}
			empty, err := json.Marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			capacity := incidentportability.MaxNDJSONLogicalLineBytes - len(empty)
			var reason string
			switch scenario {
			case "exact_ceiling":
				reason = strings.Repeat("a", capacity)
			case "one_byte_over", "capacity_precedes_actor":
				reason = strings.Repeat("a", capacity+1)
			case "escaping_expansion":
				reason = strings.Repeat("<", capacity/6+1)
			}
			row["reason"], expected["reason"] = reason, reason
			if scenario == "capacity_precedes_actor" {
				row["actor_user_id"] = "ffffffff-ffff-4fff-8fff-ffffffffffff"
			}
			want, err := json.Marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "exact_ceiling" {
				if len(want) != incidentportability.MaxNDJSONLogicalLineBytes {
					t.Fatal("positive fixture is not exactly at the canonical ceiling")
				}
			} else if len(want) <= incidentportability.MaxNDJSONLogicalLineBytes {
				t.Fatal("negative fixture fits the canonical ceiling")
			}
			var input bytes.Buffer
			encoder := json.NewEncoder(&input)
			encoder.SetEscapeHTML(false)
			if err := encoder.Encode(row); err != nil {
				t.Fatal(err)
			}
			if input.Len()-1 > incidentportability.MaxNDJSONLogicalLineBytes {
				t.Fatal("fixture exceeds the incoming frame ceiling")
			}
			bundle[revisionsTestChangeSetsPath] = input.Bytes()
			before := sha256.Sum256(bundle[revisionsTestChangeSetsPath])
			_, err = h.port.PrepareImport(t.Context(), bundle, h.context)
			if scenario == "exact_ceiling" {
				if err != nil {
					t.Fatalf("exact export capacity rejected: %v", err)
				}
			} else {
				requireRevisionsInvariant(t, err, "revisions.references_complete")
			}
			if sha256.Sum256(bundle[revisionsTestChangeSetsPath]) != before || probe.calls != 0 {
				t.Fatal("capacity preparation changed source bytes or used the database")
			}
		})
	}
}
