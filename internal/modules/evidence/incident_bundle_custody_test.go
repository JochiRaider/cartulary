package evidence_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
)

const portableCustodyTestPath = "data/evidence_custody_events.ndjson"

func portableCustodyTestRow(incident uuid.UUID, index int) map[string]any {
	return map[string]any{"custody_event_id": fmt.Sprintf("40000000-0000-4000-8000-%012d", index), "incident_id": incident.String(), "evidence_record_id": "10000000-0000-4000-8000-000000000001", "custody_event_type": "transferred", "actor_user_id": evidencePortableActor.String(), "occurred_at": "2024-01-01T00:00:00.123456+00:00", "location_text": "  Source location  ", "note": "Line one\nLine two", "metadata": map[string]any{"object_blob_id": "opaque-context", "nested_user_id": "historical-label", "counter": json.Number("9007199254740993"), "fraction": json.Number("1.0000000000000000000000000000000001"), "values": []any{true, nil, "e\u0301"}}}
}
func portableCustodyTestBundle(t testing.TB, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	b := evidencePortableBundle(t)
	var out bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&out).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	b[portableCustodyTestPath] = out.Bytes()
	return b
}
func TestEvidencePortableCustodyAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, local := uuid.New(), uuid.New()
	port := evidence.NewIncidentBundleSourcePort()
	input := evidencePortableInput(t, incident, local)
	prepare := func(rows ...map[string]any) error {
		b := portableCustodyTestBundle(t, rows...)
		before := maps.Clone(b)
		for k, v := range before {
			before[k] = bytes.Clone(v)
		}
		_, err := port.PrepareImport(ctx, b, input)
		if !reflect.DeepEqual(b, before) {
			t.Fatal("Prepare mutated custody input")
		}
		return err
	}
	base := portableCustodyTestRow(incident, 1)
	if err := prepare(base); err != nil {
		t.Fatal(err)
	}
	if err := prepare(); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"requested", "received", "made_available", "transferred", "quarantined", "released"} {
		row := maps.Clone(base)
		row["custody_event_type"] = kind
		if err := prepare(row); err != nil {
			t.Fatal(kind, err)
		}
	}
	for field := range base {
		t.Run(field, func(t *testing.T) {
			invariant := "row_shape_exact"
			if field == "custody_event_id" {
				invariant = "source_identity_admitted"
			}
			row := maps.Clone(base)
			delete(row, field)
			requireEvidencePortableFailure(t, prepare(row), invariant)
			row = maps.Clone(base)
			row[field] = false
			requireEvidencePortableFailure(t, prepare(row), invariant)
			row = maps.Clone(base)
			row[field] = nil
			if field == "actor_user_id" || field == "location_text" || field == "note" {
				if err := prepare(row); err != nil {
					t.Fatal(err)
				}
			} else {
				requireEvidencePortableFailure(t, prepare(row), invariant)
			}
		})
	}
	for _, tc := range []struct {
		name, field, invariant string
		value                  any
	}{
		{"unknown member", "extra", "row_shape_exact", true}, {"unknown event", "custody_event_type", "row_shape_exact", "unknown"}, {"foreign incident", "incident_id", "custody_same_incident", uuid.NewString()}, {"unknown actor", "actor_user_id", "attribution_exact", uuid.NewString()},
		{"NUL note", "note", "row_shape_exact", "bad\x00note"}, {"NUL metadata key", "metadata", "row_shape_exact", map[string]any{"bad\x00key": true}}, {"NUL metadata value", "metadata", "row_shape_exact", map[string]any{"nested": []any{map[string]any{"value": "bad\x00text"}}}},
		{"metadata array", "metadata", "row_shape_exact", []any{}}, {"metadata scalar", "metadata", "row_shape_exact", "{}"},
		{"lossy timestamp", "occurred_at", "custody_ordered", "2024-01-01T00:00:00.1234567Z"}, {"noncanonical timestamp", "occurred_at", "custody_ordered", "2024-01-01T00:00:00.0000000Z"}, {"year zero", "occurred_at", "custody_ordered", "0000-01-01T00:00:00Z"}, {"offset timestamp", "occurred_at", "custody_ordered", "2024-01-01T01:00:00+01:00"}, {"invalid timestamp", "occurred_at", "custody_ordered", "not-a-time"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := maps.Clone(base)
			row[tc.field] = tc.value
			requireEvidencePortableFailure(t, prepare(row), tc.invariant)
		})
	}
	for _, field := range []string{"custody_event_id", "incident_id", "evidence_record_id", "actor_user_id"} {
		for _, value := range []string{uuid.Nil.String(), "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "10000000000040008000000000000001", "bad"} {
			row := maps.Clone(base)
			row[field] = value
			invariant := "row_shape_exact"
			if field == "custody_event_id" {
				invariant = "source_identity_admitted"
			}
			requireEvidencePortableFailure(t, prepare(row), invariant)
		}
	}
	for _, text := range []string{"", "   ", "e\u0301", "\u2060text", "🧪"} {
		row := maps.Clone(base)
		row["location_text"], row["note"] = text, text
		if err := prepare(row); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []json.Number{"9007199254740993", "1.0000000000000000000000000000000001", "1.2300e2", "-0", "0e999999999999999999999999", "1e131071", "1e-16383"} {
		row := maps.Clone(base)
		row["metadata"] = map[string]any{"number": value}
		if err := prepare(row); err != nil {
			t.Fatalf("exact decimal %s: %v", value, err)
		}
	}
	for _, value := range []json.Number{"1e131072", "1e-16384", "1e999999999999999999999999", "1e-999999999999999999999999"} {
		row := maps.Clone(base)
		row["metadata"] = map[string]any{"number": value}
		requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
	}
	t.Run("bounded decimal expansion", func(t *testing.T) {
		values := make([]any, 128)
		for i := range values {
			values[i] = json.Number("1e131071")
		}
		row := maps.Clone(base)
		row["metadata"] = map[string]any{"values": values}
		requireEvidencePortableFailure(t, prepare(row), "row_shape_exact")
	})
	second := portableCustodyTestRow(incident, 2)
	second["occurred_at"] = "2023-01-01T00:00:00Z"
	if err := prepare(base, second); err != nil {
		t.Fatal("UUID order does not impose time order", err)
	}
	requireEvidencePortableFailure(t, prepare(second, base), "source_identity_admitted")
	requireEvidencePortableFailure(t, prepare(base, base), "source_identity_admitted")
	for _, payload := range []string{"[]\n", "\n", "{} {}\n", "{\"metadata\":{\"x\":1,\"x\":2}}\n", "{\"note\":\"\\ud800\"}\n", string([]byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'})} {
		b := portableCustodyTestBundle(t)
		b[portableCustodyTestPath] = []byte(payload)
		_, err := port.PrepareImport(ctx, b, input)
		requireEvidencePortableFailure(t, err, "row_shape_exact")
	}
	bad := maps.Clone(base)
	bad["custody_event_id"] = uuid.Nil.String()
	b := portableCustodyTestBundle(t, bad)
	b[portableCustodyTestPath] = append(b[portableCustodyTestPath], []byte("{\n")...)
	_, err := port.PrepareImport(ctx, b, input)
	requireEvidencePortableFailure(t, err, "row_shape_exact")
	early := maps.Clone(base)
	early["occurred_at"] = "bad"
	late := maps.Clone(second)
	delete(late, "note")
	requireEvidencePortableFailure(t, prepare(early, late), "row_shape_exact")
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = port.PrepareImport(cancelled, portableCustodyTestBundle(t, base), input)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel=%v", err)
	}
}
