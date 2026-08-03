package indicators

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
)

const (
	portableIncidentID = "00000000-0000-4000-8000-000000000100"
	portableActorID    = "00000000-0000-4000-8000-000000000101"
	portableSourceID   = "00000000-0000-4000-8000-000000000102"
	portableTimestamp  = "2025-01-02T03:04:05.000001+00:00"
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
	importContext := portableImportContext(t, "indicator-identity-portability")

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
	invalidObservation := portableObservationRow()
	invalidObservation["origin_kind"] = "auto_extract"
	invalidOriginBundle["data/indicator_observations.ndjson"] = marshalNDJSONRows(t, []map[string]any{invalidObservation})
	_, err = port.PrepareImport(context.Background(), invalidOriginBundle, importContext)
	assertIndicatorInvariantFailure(t, err, "indicators.observation_coherent")
}

func portableIdentityRow(recordID string, canonical identity.Canonical) map[string]any {
	return map[string]any{
		"record_id": recordID, "incident_id": portableIncidentID,
		"indicator_type": canonical.IndicatorType, "value_kind": canonical.ValueKind,
		"display_value": canonical.DisplayValue, "normalized_value": canonical.NormalizedValue,
		"dedupe_key": canonical.DedupeKey, "defanged_value": nil,
		"hash_algorithm": canonical.HashAlgorithm, "hash_value": canonical.HashValue,
		"stix_pattern": nil, "row_version": 1,
		"created_at": portableTimestamp, "updated_at": portableTimestamp,
		"created_by_user_id": portableActorID, "updated_by_user_id": portableActorID,
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
}

func portableObservationRow() map[string]any {
	return map[string]any{
		"indicator_observation_id": "00000000-0000-4000-8000-000000000010",
		"incident_id":              portableIncidentID, "source_record_id": portableSourceID,
		"source_field_key": "source_text", "origin_kind": "manual_entry",
		"origin_locator": "portable-fixture", "observed_text": "not parsed",
		"parsed_indicator_type": nil, "normalized_candidate": nil,
		"resolution_status": "unresolved", "resolved_indicator_record_id": nil,
		"row_version": 1, "created_by_user_id": portableActorID,
		"created_at": portableTimestamp, "resolved_by_user_id": nil,
		"resolved_at": nil, "resolution_method": nil,
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
}

func portableIntervalRow() map[string]any {
	return map[string]any{
		"indicator_state_interval_id": "00000000-0000-4000-8000-000000000020",
		"incident_id":                 portableIncidentID,
		"indicator_record_id":         "00000000-0000-4000-8000-000000000001",
		"lifecycle_state":             "active", "valid_from": portableTimestamp,
		"valid_to": nil, "confidence": nil, "rationale": nil,
		"support_refs": []any{}, "assessor": nil, "assessed_at": portableTimestamp,
		"row_version": 1, "created_by_user_id": portableActorID,
		"created_at": portableTimestamp, "deleted_at": nil, "deleted_by_user_id": nil,
	}
}

func portableImportContext(t testing.TB, operationID string) sourceport.ImportContext {
	t.Helper()
	actorID := uuid.MustParse(portableActorID)
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: actorID.String()}})
	if err != nil {
		t.Fatalf("actor catalog: %v", err)
	}
	return sourceport.ImportContext{
		IncidentID: uuid.MustParse(portableIncidentID), ActorUserID: actorID,
		BundleVersion: 2, OperationID: operationID, Actors: actors,
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
		"data/indicator_observations.ndjson":    marshalNDJSONRows(t, []map[string]any{portableObservationRow()}),
		"data/indicator_state_intervals.ndjson": marshalNDJSONRows(t, []map[string]any{portableIntervalRow()}),
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
