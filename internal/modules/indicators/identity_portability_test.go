package indicators

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
)

func TestIndicatorPortablePreparationUsesCanonicalIdentity(t *testing.T) {
	t.Parallel()
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "example.test",
	})
	if err != nil {
		t.Fatalf("canonicalize fixture: %v", err)
	}
	row := portableIdentityRow("00000000-0000-4000-8000-000000000001", canonical)
	port := NewIncidentBundleContribution().SourcePort
	importContext := sourceport.ImportContext{BundleVersion: 2, OperationID: "indicator-identity-portability"}

	if _, err := port.PrepareImport(context.Background(), portableIdentityBundle(t, row), importContext); err != nil {
		t.Fatalf("prepare canonical identity: %v", err)
	}

	noncanonical := cloneAnyMap(row)
	noncanonical["display_value"] = "EXAMPLE.TEST"
	_, err = port.PrepareImport(context.Background(), portableIdentityBundle(t, noncanonical), importContext)
	assertIndicatorInvariantFailure(t, err, "indicators.normalization_exact")

	duplicate := cloneAnyMap(row)
	duplicate["record_id"] = "00000000-0000-4000-8000-000000000002"
	_, err = port.PrepareImport(context.Background(), portableIdentityBundle(t, row, duplicate), importContext)
	assertIndicatorInvariantFailure(t, err, "indicators.identity_unique")

	invalidOriginBundle := portableIdentityBundle(t, row)
	invalidOriginBundle["data/indicator_observations.ndjson"] = marshalNDJSONRows(t, []map[string]any{{
		"indicator_observation_id": "00000000-0000-4000-8000-000000000010",
		"origin_kind":              "auto_extract",
	}})
	_, err = port.PrepareImport(context.Background(), invalidOriginBundle, importContext)
	assertIndicatorInvariantFailure(t, err, "indicators.representation_legal")
}

func portableIdentityRow(recordID string, canonical identity.Canonical) map[string]any {
	return map[string]any{
		"record_id":        recordID,
		"indicator_type":   canonical.IndicatorType,
		"value_kind":       canonical.ValueKind,
		"display_value":    canonical.DisplayValue,
		"normalized_value": canonical.NormalizedValue,
		"dedupe_key":       canonical.DedupeKey,
		"defanged_value":   nil,
		"hash_algorithm":   nil,
		"hash_value":       nil,
		"stix_pattern":     nil,
	}
}

type indicatorTestBundle map[string][]byte

func (b indicatorTestBundle) File(path string) ([]byte, bool) {
	payload, ok := b[path]
	return payload, ok
}

func (b indicatorTestBundle) Paths() []string {
	paths := make([]string, 0, len(b))
	for path := range b {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func portableIdentityBundle(t testing.TB, indicatorRows ...map[string]any) indicatorTestBundle {
	t.Helper()
	return indicatorTestBundle{
		"data/indicators.ndjson":                marshalNDJSONRows(t, indicatorRows),
		"data/indicator_observations.ndjson":    marshalNDJSONRows(t, []map[string]any{{"indicator_observation_id": "00000000-0000-4000-8000-000000000010", "origin_kind": "manual_entry"}}),
		"data/indicator_state_intervals.ndjson": marshalNDJSONRows(t, []map[string]any{{"indicator_state_interval_id": "00000000-0000-4000-8000-000000000020"}}),
	}
}

func marshalNDJSONRows(t testing.TB, rows []map[string]any) []byte {
	t.Helper()
	payload := make([]byte, 0, len(rows)*128)
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("marshal portable row: %v", err)
		}
		payload = append(payload, encoded...)
		payload = append(payload, '\n')
	}
	return payload
}

func assertIndicatorInvariantFailure(t testing.TB, err error, invariantID string) {
	t.Helper()
	want := "incident bundle source family indicators failed invariant " + invariantID
	if err == nil || err.Error() != want {
		t.Fatalf("failure = %#v, want indicators/%s", err, invariantID)
	}
}

func cloneAnyMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
