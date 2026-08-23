package indicators

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
)

func TestIndicatorPortablePrepareInvariantPartition(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		invariant string
		mutate    func(indicatorTestBundle)
	}{
		{
			name: "representation", invariant: "indicators.representation_legal",
			mutate: func(bundle indicatorTestBundle) {
				row := portableCanonicalIndicatorRow(t)
				delete(row, "updated_at")
				bundle["data/indicators.ndjson"] = marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "normalization", invariant: "indicators.normalization_exact",
			mutate: func(bundle indicatorTestBundle) {
				row := portableCanonicalIndicatorRow(t)
				row["display_value"] = "EXAMPLE.TEST"
				bundle["data/indicators.ndjson"] = marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "identity_unique", invariant: "indicators.identity_unique",
			mutate: func(bundle indicatorTestBundle) {
				first := portableCanonicalIndicatorRow(t)
				second := cloneAnyMap(first)
				second["record_id"] = "00000000-0000-4000-8000-000000000002"
				bundle["data/indicators.ndjson"] = marshalNDJSONRows(t, []map[string]any{first, second})
			},
		},
		{
			name: "observation_ordered", invariant: "indicators.observation_ordered",
			mutate: func(bundle indicatorTestBundle) {
				row := portableObservationRow()
				row["row_version"] = 0
				bundle["data/indicator_observations.ndjson"] = marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "observation_coherent", invariant: "indicators.observation_coherent",
			mutate: func(bundle indicatorTestBundle) {
				row := portableObservationRow()
				row["origin_kind"] = "interactive_cell"
				bundle["data/indicator_observations.ndjson"] = marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "interval_ordered", invariant: "indicators.interval_ordered",
			mutate: func(bundle indicatorTestBundle) {
				row := portableIntervalRow()
				row["row_version"] = 0
				bundle["data/indicator_state_intervals.ndjson"] = marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "interval_coherent", invariant: "indicators.interval_coherent",
			mutate: func(bundle indicatorTestBundle) {
				row := portableIntervalRow()
				row["confidence"] = 101
				bundle["data/indicator_state_intervals.ndjson"] = marshalNDJSONRows(t, []map[string]any{row})
			},
		},
	}
	port := mustIndicatorSourcePort(t)
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			bundle := portableIdentityBundle(t, portableCanonicalIndicatorRow(t))
			testCase.mutate(bundle)
			_, err := port.PrepareImport(
				context.Background(), bundle,
				portableImportContext(t, "indicator-portability-"+testCase.name),
			)
			assertIndicatorInvariantFailure(t, err, testCase.invariant)
		})
	}
}

func TestIndicatorPortableMultiDefectSelectionIsDeterministic(t *testing.T) {
	t.Parallel()
	port := mustIndicatorSourcePort(t)
	for permutation := 0; permutation < 3; permutation++ {
		bundle := portableIdentityBundle(t, portableCanonicalIndicatorRow(t))
		indicatorRow := portableCanonicalIndicatorRow(t)
		indicatorRow["display_value"] = "EXAMPLE.TEST"
		observationRow := portableObservationRow()
		observationRow["origin_kind"] = "auto_extract"
		intervalRow := portableIntervalRow()
		intervalRow["confidence"] = 101
		bundle["data/indicators.ndjson"] = marshalNDJSONRows(t, []map[string]any{indicatorRow})
		bundle["data/indicator_observations.ndjson"] = marshalNDJSONRows(t, []map[string]any{observationRow})
		bundle["data/indicator_state_intervals.ndjson"] = marshalNDJSONRows(t, []map[string]any{intervalRow})
		if permutation == 1 {
			bundle = indicatorTestBundle{
				"data/indicator_state_intervals.ndjson": bundle["data/indicator_state_intervals.ndjson"],
				"data/indicators.ndjson":                bundle["data/indicators.ndjson"],
				"data/indicator_observations.ndjson":    bundle["data/indicator_observations.ndjson"],
			}
		} else if permutation == 2 {
			bundle = indicatorTestBundle{
				"data/indicator_observations.ndjson":    bundle["data/indicator_observations.ndjson"],
				"data/indicator_state_intervals.ndjson": bundle["data/indicator_state_intervals.ndjson"],
				"data/indicators.ndjson":                bundle["data/indicators.ndjson"],
			}
		}
		_, err := port.PrepareImport(
			context.Background(), bundle,
			portableImportContext(t, "indicator-multi-defect-"+string(rune('0'+permutation))),
		)
		assertIndicatorInvariantFailure(t, err, "indicators.normalization_exact")
	}
}

func TestIndicatorPortableStrictDecodingAndSafeFailure(t *testing.T) {
	t.Parallel()
	validRow := portableCanonicalIndicatorRow(t)
	validPayload := marshalNDJSONRows(t, []map[string]any{validRow})
	duplicateMember := append(
		[]byte(`{"record_id":"00000000-0000-4000-8000-000000000001",`),
		validPayload[1:]...,
	)
	multipleValues := append(append([]byte(nil), validPayload[:len(validPayload)-1]...), []byte("{}\n")...)
	testCases := []struct {
		name    string
		payload func() []byte
	}{
		{
			name: "unknown_member",
			payload: func() []byte {
				row := cloneAnyMap(validRow)
				row["legacy_alias"] = "rejected"
				return marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "wrong_type",
			payload: func() []byte {
				row := cloneAnyMap(validRow)
				row["row_version"] = "1"
				return marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{name: "duplicate_member", payload: func() []byte { return duplicateMember }},
		{name: "blank_line", payload: func() []byte { return append([]byte{'\n'}, validPayload...) }},
		{name: "multiple_values", payload: func() []byte { return multipleValues }},
		{name: "trailing_content", payload: func() []byte { return append(append([]byte(nil), validPayload...), 'x') }},
		{
			name: "noncanonical_timestamp",
			payload: func() []byte {
				row := cloneAnyMap(validRow)
				row["created_at"] = "2025-01-02T03:04:05Z"
				return marshalNDJSONRows(t, []map[string]any{row})
			},
		},
		{
			name: "hostile_value_is_redacted",
			payload: func() []byte {
				row := cloneAnyMap(validRow)
				row["display_value"] = "hostile\x00secret relation constraint"
				return marshalNDJSONRows(t, []map[string]any{row})
			},
		},
	}
	port := mustIndicatorSourcePort(t)
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			bundle := portableIdentityBundle(t, validRow)
			bundle["data/indicators.ndjson"] = testCase.payload()
			_, err := port.PrepareImport(
				context.Background(), bundle,
				portableImportContext(t, "indicator-strict-"+testCase.name),
			)
			assertIndicatorInvariantFailure(t, err, "indicators.representation_legal")
			for _, forbidden := range []string{"hostile", "secret", "relation", "constraint", "data/"} {
				if strings.Contains(err.Error(), forbidden) {
					t.Fatalf("safe failure disclosed %q: %v", forbidden, err)
				}
			}
		})
	}
}

func TestIndicatorPortablePreparedValuesAreContextBound(t *testing.T) {
	t.Parallel()
	port := mustIndicatorSourcePort(t)
	importContext := portableImportContext(t, "indicator-prepared-binding")
	prepared, err := port.PrepareImport(
		context.Background(), portableIdentityBundle(t, portableCanonicalIndicatorRow(t)), importContext,
	)
	if err != nil {
		t.Fatalf("prepare portable value: %v", err)
	}
	wrongOperation := importContext
	wrongOperation.OperationID = "indicator-prepared-binding-other"
	if err := port.ApplyImportTx(context.Background(), nil, prepared, wrongOperation); !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("operation binding failure = %v, want ErrPreparedBinding", err)
	}
	wrongIncident := importContext
	wrongIncident.IncidentID = uuid.MustParse("00000000-0000-4000-8000-000000000199")
	assertIndicatorInvariantFailure(
		t, port.ApplyImportTx(context.Background(), nil, prepared, wrongIncident),
		"indicators.representation_legal",
	)
	wrongVersion := importContext
	wrongVersion.BundleVersion = 1
	assertIndicatorInvariantFailure(
		t, port.ApplyImportTx(context.Background(), nil, prepared, wrongVersion),
		"indicators.representation_legal",
	)
}

func portableCanonicalIndicatorRow(t testing.TB) map[string]any {
	t.Helper()
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: "domain_name", ValueKind: "atomic", DisplayValue: "example.test",
	})
	if err != nil {
		t.Fatalf("canonicalize portable fixture: %v", err)
	}
	return portableIdentityRow("00000000-0000-4000-8000-000000000001", canonical)
}
