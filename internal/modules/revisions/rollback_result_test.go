package revisions

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestRollbackStoredValuesPreserveJSONNumbers_Unit(t *testing.T) {
	t.Run("idempotency response", func(t *testing.T) {
		for _, want := range []json.Number{"1", "9007199254740993", "9223372036854775807"} {
			t.Run(string(want), func(t *testing.T) {
				raw := []byte(`{"affected_record_ids":["00000000-0000-4000-8000-000000000002"],"incident_id":"00000000-0000-4000-8000-000000000001","record_id":"00000000-0000-4000-8000-000000000002","rollback_change_set_id":"00000000-0000-4000-8000-000000000003","row_version":` + string(want) + `,"target":{"change_set_id":"00000000-0000-4000-8000-000000000004","kind":"change_set"},"target_change_set_id":"00000000-0000-4000-8000-000000000004"}`)
				payload, err := decodeStoredRollbackPayload(raw)
				if err != nil {
					t.Fatal(err)
				}
				result := (rollbackResultCoordinator{}).replayed(payload, "exact-rollback-replay")
				encoded, err := json.Marshal(result.Payload)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(encoded, raw) {
					t.Errorf("replay response = %s, want %s", encoded, raw)
				}
				version, err := want.Int64()
				if err != nil {
					t.Fatal(err)
				}
				if !result.Replayed || result.ClientTxnID != "exact-rollback-replay" || result.IncidentID != uuid.MustParse("00000000-0000-4000-8000-000000000001") {
					t.Fatalf("replay identity = %#v", result)
				}
				if len(result.Changes) != 1 || result.Changes[0].RecordID != uuid.MustParse("00000000-0000-4000-8000-000000000002") || result.Changes[0].ChangeSetID != uuid.MustParse("00000000-0000-4000-8000-000000000003") || result.Changes[0].RowVersion != version {
					t.Fatalf("replay changes = %#v, want exact version %d", result.Changes, version)
				}
			})
		}
		for _, raw := range []string{"", `{"record":`, `[]`, `{} {}`, `{} null`, `{} invalid`} {
			if _, err := decodeStoredRollbackPayload([]byte(raw)); err == nil {
				t.Errorf("invalid stored response %q accepted", raw)
			}
		}
	})
	for _, want := range []json.Number{"9007199254740993", "9223372036854775807"} {
		t.Run(string(want), func(t *testing.T) {
			value, err := decodeRollbackValue([]byte(`{"record":{"row_version":` + string(want) + `},"source":{"row_version":` + string(want) + `,"confidence_score":25},"nested":[` + string(want) + `]}`))
			if err != nil {
				t.Fatal(err)
			}
			for _, member := range []string{"record", "source"} {
				if got := value[member].(map[string]any)["row_version"]; got != want {
					t.Errorf("%s version = %#v, want %s", member, got, want)
				}
			}
			if got := value["nested"].([]any)[0]; got != want {
				t.Errorf("nested value = %#v, want %s", got, want)
			}
			if got := value["source"].(map[string]any)["confidence_score"]; got != json.Number("25") {
				t.Errorf("score = %#v, want exact 25", got)
			}
		})
	}
	for _, raw := range []string{"", "null"} {
		if value, err := decodeRollbackValue([]byte(raw)); err != nil || value != nil {
			t.Errorf("absent snapshot %q = %#v, %v", raw, value, err)
		}
	}
	for _, raw := range []string{`{"record":`, `[]`, `{} {}`, `{} null`, `{} invalid`} {
		if _, err := decodeRollbackValue([]byte(raw)); err == nil {
			t.Errorf("invalid stored snapshot %q accepted", raw)
		}
	}
}
