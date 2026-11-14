package rollback

import (
	"errors"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/google/uuid"
)

func requireCompleteIndicatorSource(t *testing.T) {
	recordID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	incidentID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	source := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "indicator_type": "domain_name", "value_kind": "atomic",
		"display_value": "before.example", "normalized_value": "before.example", "dedupe_key": "domain_name:before.example", "defanged_value": "before[.]example",
		"hash_algorithm": nil, "hash_value": nil, "stix_pattern": "[domain-name:value = 'before.example']"}
	record := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String()}
	snapshot := map[string]any{"record": record, "source": source}
	state, err := parseCompleteIndicatorSource(snapshot)
	want := rowState{recordID: recordID, incidentID: incidentID, indicatorType: "domain_name", valueKind: "atomic", displayValue: "before.example", normalized: stringPointer("before.example"), dedupeKey: "domain_name:before.example", defanged: stringPointer("before[.]example"), stixPattern: stringPointer("[domain-name:value = 'before.example']")}
	if err != nil || !reflect.DeepEqual(state, want) {
		t.Fatalf("complete source=%#v want=%#v error=%v", state, want, err)
	}
	for _, text := range []any{nil, "", " \t", "retained"} {
		copy := maps.Clone(source)
		copy["defanged_value"], copy["stix_pattern"] = text, text
		got, err := parseCompleteIndicatorSource(map[string]any{"record": record, "source": copy})
		if err != nil {
			t.Fatalf("retained optional text %#v: %v", text, err)
		}
		if text == nil {
			if got.defanged != nil || got.stixPattern != nil {
				t.Fatal("explicit null was filled")
			}
		} else if got.defanged == nil || got.stixPattern == nil || *got.defanged != text || *got.stixPattern != text {
			t.Fatal("retained text changed")
		}
	}
	copy := maps.Clone(source)
	copy["hash_algorithm"], copy["hash_value"] = "sha256", strings.Repeat("a", 64)
	if got, err := parseCompleteIndicatorSource(map[string]any{"record": record, "source": copy}); err != nil || got.hashAlgorithm == nil || *got.hashAlgorithm != "sha256" || got.hashValue == nil || *got.hashValue != strings.Repeat("a", 64) {
		t.Fatalf("complete hash pair=%#v error=%v", got, err)
	}
	check := func(name string, value map[string]any) {
		t.Run(name, func(t *testing.T) {
			original := maps.Clone(value)
			if src, ok := value["source"].(map[string]any); ok {
				original["source"] = maps.Clone(src)
			}
			if _, err := parseCompleteIndicatorSource(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
				t.Fatalf("invalid complete source admitted: %v", err)
			}
			if !reflect.DeepEqual(value, original) {
				t.Fatal("retained source changed")
			}
		})
	}
	for _, key := range []string{"record_id", "incident_id", "indicator_type", "value_kind", "display_value", "normalized_value", "dedupe_key", "defanged_value", "hash_algorithm", "hash_value", "stix_pattern"} {
		for _, defect := range []string{"missing", "wrong_type"} {
			copy := maps.Clone(source)
			if defect == "missing" {
				delete(copy, key)
			} else {
				copy[key] = false
			}
			check(key+"_"+defect, map[string]any{"record": record, "source": copy})
		}
	}
	for _, bad := range []struct {
		name, key string
		value     any
	}{
		{"extra", "row_version", 1}, {"drift_record", "record_id", uuid.NewString()}, {"drift_incident", "incident_id", uuid.NewString()},
		{"zero_uuid", "record_id", uuid.Nil.String()}, {"upper_uuid", "incident_id", strings.ToUpper(incidentID.String())},
		{"unknown_type", "indicator_type", "domain"}, {"padded_type", "indicator_type", " domain_name"}, {"wrong_kind", "value_kind", "literal"},
		{"null_normalized", "normalized_value", nil}, {"blank_normalized", "normalized_value", " "}, {"blank_display", "display_value", " "}, {"blank_dedupe", "dedupe_key", " "},
		{"nul_text", "stix_pattern", "a\x00b"}, {"invalid_utf8", "defanged_value", string([]byte{255})},
		{"partial_hash", "hash_algorithm", "sha256"}, {"partial_hash_value", "hash_value", strings.Repeat("a", 64)},
	} {
		copy := maps.Clone(source)
		copy[bad.key] = bad.value
		check(bad.name, map[string]any{"record": record, "source": copy})
	}
	for _, value := range []map[string]any{nil, {}, {"record": record}, {"record": record, "source": nil}, {"record": record, "source": "wrong"}, {"source": source}, {"record": "wrong", "source": source}} {
		check("container", value)
	}
}
