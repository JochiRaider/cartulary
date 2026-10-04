package reference_data

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testConsumerMachineContracts(t *testing.T) {
	m, c, _ := consumerFixture(t)
	requests := map[string]map[string]any{
		"ResolveCurrentPackSet":  {},
		"GetPackEntry":           {"pack_set_id": m.set.ID, "pack_key": "type_registry.host", "entry_id": "unknown"},
		"GetPackProvenance":      {"pack_set_id": m.set.ID, "pack_key": "type_registry.host"},
		"LookupPackEntries":      {"pack_set_id": m.set.ID, "pack_key": "type_registry.host", "lookup_kind": "alias", "lookup_value": "unknown"},
		"EvaluateIndicatorValue": {"pack_set_id": m.set.ID, "indicator_type_id": "ipv4_addr", "value_kind": "atomic", "raw_value": "192.0.2.1"},
	}
	for operation, request := range requests {
		if err := packformat.ValidateConsumerRequest(operation, request); err != nil {
			t.Fatal(operation, err)
		}
		for _, field := range slices.Sorted(maps.Keys(request)) {
			original := request[field]
			delete(request, field)
			if err := packformat.ValidateConsumerRequest(operation, request); err == nil {
				t.Fatal("omitted request member", operation, field)
			}
			for _, wrong := range []any{nil, []any{}} {
				request[field] = wrong
				if err := packformat.ValidateConsumerRequest(operation, request); err == nil {
					t.Fatal("invalid request member", operation, field)
				}
			}
			request[field] = original
		}
		request["private_unknown"] = true
		if err := packformat.ValidateConsumerRequest(operation, request); err == nil {
			t.Fatal("unknown request member", operation)
		}
		delete(request, "private_unknown")
	}
	entry := requests["GetPackEntry"]
	entry["entry_id"] = strings.Repeat("é", 256)
	if err := packformat.ValidateConsumerRequest("GetPackEntry", entry); err != nil {
		t.Fatal("512 UTF-8 bytes", err)
	}
	entry["entry_id"] = strings.Repeat("é", 256) + "a"
	if err := packformat.ValidateConsumerRequest("GetPackEntry", entry); err == nil {
		t.Fatal("byte bound treated as scalar bound")
	}
	eval := requests["EvaluateIndicatorValue"]
	eval["raw_value"] = strings.Repeat("😀", 8192)
	if err := packformat.ValidateConsumerRequest("EvaluateIndicatorValue", eval); err != nil {
		t.Fatal("scalar bound treated as byte bound", err)
	}
	eval["raw_value"] = strings.Repeat("😀", 8193)
	eval["value_kind"] = strings.Repeat("x", 65)
	raw, _ := json.Marshal(eval)
	var decoded EvaluateIndicatorRequest
	err := json.Unmarshal(raw, &decoded)
	public, ok := err.(*ConsumerError)
	if !ok || public.ReasonCode != nil {
		t.Fatal("earlier field must precede raw limit", err)
	}
	eval["value_kind"] = "atomic"
	raw, _ = json.Marshal(eval)
	err = json.Unmarshal(raw, &decoded)
	public, ok = err.(*ConsumerError)
	if !ok || public.ReasonCode == nil || *public.ReasonCode != "indicator_input_too_long" {
		t.Fatal("raw limit lost safe reason", err)
	}
	limit := 1
	results := []struct {
		operation string
		value     any
	}{
		{"ResolveCurrentPackSet", c.ResolveCurrentPackSet(context.Background())},
		{"GetPackEntry", c.GetPackEntry(context.Background(), GetPackEntryRequest{PackSetID: m.set.ID, PackKey: "type_registry.indicator", EntryID: "ipv4_addr"})},
		{"GetPackProvenance", c.GetPackProvenance(context.Background(), GetPackProvenanceRequest{PackSetID: m.set.ID, PackKey: "type_registry.host"})},
		{"LookupPackEntries", c.LookupPackEntries(context.Background(), LookupPackEntriesRequest{PackSetID: m.set.ID, PackKey: "type_registry.host", LookupKind: "alias", LookupValue: "unknown", Limit: &limit})},
		{"EvaluateIndicatorValue", c.EvaluateIndicatorValue(context.Background(), EvaluateIndicatorRequest{PackSetID: m.set.ID, IndicatorTypeID: "ipv4_addr", ValueKind: "atomic", RawValue: "192.0.2.1"})},
		{"EvaluateIndicatorValue", c.EvaluateIndicatorValue(context.Background(), EvaluateIndicatorRequest{PackSetID: m.set.ID, IndicatorTypeID: "ipv4_addr", ValueKind: "atomic", RawValue: "invalid"})},
		{"EvaluateIndicatorValue", c.EvaluateIndicatorValue(context.Background(), EvaluateIndicatorRequest{PackSetID: m.set.ID, IndicatorTypeID: "unknown", ValueKind: "atomic", RawValue: "private input"})},
	}
	for _, result := range results {
		raw, err := json.Marshal(result.value)
		if err != nil {
			t.Fatal(err)
		}
		if err := packformat.ValidateConsumerResult(result.operation, raw); err != nil {
			t.Fatal("actual result", result.operation, err)
		}
		value, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			t.Fatal(err)
		}
		var visit func(any)
		visit = func(node any) {
			switch object := node.(type) {
			case map[string]any:
				for _, field := range slices.Sorted(maps.Keys(object)) {
					original := object[field]
					delete(object, field)
					mutated, _ := canonicaljson.Marshal(value)
					if err := packformat.ValidateConsumerResult(result.operation, mutated); err == nil {
						t.Fatal("omitted result member", result.operation, field)
					}
					object[field] = []any{}
					if _, array := original.([]any); array {
						object[field] = true
					}
					mutated, _ = canonicaljson.Marshal(value)
					if err := packformat.ValidateConsumerResult(result.operation, mutated); err == nil {
						t.Fatal("wrong result member type", result.operation, field)
					}
					if original != nil {
						object[field] = nil
						mutated, _ = canonicaljson.Marshal(value)
						err := packformat.ValidateConsumerResult(result.operation, mutated)
						// These populated fixture members may become explicit null.
						// All other populated members in these independently chosen
						// results must remain present and non-null.
						nullable := slices.Contains([]string{"source_as_of", "trust_valid_until", "next_cursor", "description", "replacement_entry_id", "stix_mapping"}, field)
						if (err == nil) != nullable {
							t.Fatal("result nullability", result.operation, field, err)
						}
					}
					object[field] = original
					visit(original)
				}
				object["private_unknown"] = true
				mutated, _ := canonicaljson.Marshal(value)
				if err := packformat.ValidateConsumerResult(result.operation, mutated); err == nil {
					t.Fatal("unknown result member", result.operation)
				}
				delete(object, "private_unknown")
			case []any:
				for index, child := range object {
					for _, wrong := range []any{nil, false} {
						object[index] = wrong
						mutated, _ := canonicaljson.Marshal(value)
						if err := packformat.ValidateConsumerResult(result.operation, mutated); err == nil {
							t.Fatal("invalid result array element", result.operation, index)
						}
					}
					object[index] = child
					visit(child)
				}
			}
		}
		visit(value)
	}
}
