package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	norm "github.com/JochiRaider/cartulary/internal/gen/networkflowunicode"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

// Each family has an explicit owner test; summaries are never evaluated.
// Runtime callers must match this registry. The integrity row checks exact
// inventory equality; only the individually selected comparisons qualify behavior.
type fixtureComparisonRegistration struct {
	Revision int
	TestName string
}

var behavioralFixtureComparisons = map[string]fixtureComparisonRegistration{
	"NF-FIX-029-cisco-sna-flow-search":       {5, "TestCiscoSNAFlowSearchExampleFixtureComparison"},
	"NF-FIX-001-cisco-sna-minimal":           {8, "TestNetworkFlowMinimalImportFixtureComparison"},
	"NF-FIX-002-cisco-sna-interface-fields":  {5, "TestNetworkFlowInterfacesImportFixtureComparison"},
	"NF-FIX-003-duplicate-headers":           {4, "TestNetworkFlowOrdinalFixtureComparison"},
	"NF-FIX-004-rejected-rows":               {6, "TestNetworkFlowRejectedImportFixtureComparison"},
	"NF-FIX-005-csv-parser-edges":            {3, "TestNetworkFlowCSVFixtureComparison"},
	"NF-FIX-006-cross-table-graph":           {3, "TestNetworkFlowCrossTableGraphFixtureComparison"},
	"NF-FIX-007-indicator-linking":           {4, "TestNetworkFlowIndicatorLinkingFixtureComparison"},
	"NF-FIX-008-large-limits":                {3, "TestNetworkFlowLargeLimitsFixtureComparison"},
	"NF-FIX-009-soft-delete-stale-graph":     {3, "TestNetworkFlowStaleGraphFixtureComparison"},
	"NF-FIX-010-json-admission":              {4, "TestNetworkFlowJSONFixtureComparison"},
	"NF-FIX-011-alias-collision":             {4, "TestNetworkFlowAliasFixtureComparison"},
	"NF-FIX-012-sys-uptime-timestamps":       {3, "TestNetworkFlowUptimeFixtureComparison"},
	"NF-FIX-013-filename-display":            {4, "TestNetworkFlowNamesFixtureComparison"},
	"NF-FIX-014-cursor-pagination":           {5, "TestNetworkFlowPagingFixtureComparison"},
	"NF-FIX-015-graph-adapter-input":         {4, "TestNetworkFlowGraphAdapterFixtureComparison"},
	"NF-FIX-016-redaction":                   {9, "TestNetworkFlowPrivacyFixtureComparison"},
	"NF-FIX-017-indicator-link-mismatch":     {3, "TestNetworkFlowIndicatorMismatchFixtureComparison"},
	"NF-FIX-018-resource-limits":             {5, "TestNetworkFlowResourceLimitsFixtureComparison"},
	"NF-FIX-019-canonical-json-unicode":      {5, "TestNetworkFlowCanonicalFixtureComparison"},
	"NF-FIX-020-atomic-import-commit":        {3, "TestNetworkFlowAtomicImportFixtureComparison"},
	"NF-FIX-021-preview-boundaries":          {3, "TestNetworkFlowPreviewFixtureComparison"},
	"NF-FIX-022-timestamp-rulesets":          {4, "TestNetworkFlowTimestampFixtureComparison"},
	"NF-FIX-023-import-facade-source-change": {3, "TestNetworkFlowSourceChangedImportFixtureComparison"},
	"NF-FIX-024-query-normalization-cursors": {5, "TestNetworkFlowQueryNormalizationFixtureComparison"},
	"NF-FIX-025-graph-contributors":          {3, "TestNetworkFlowContributorsFixtureComparison"},
	"NF-FIX-026-audit-and-replay":            {5, "TestNetworkFlowAuditReplayFixtureComparison"},
	"NF-FIX-027-retention-soft-delete":       {4, "TestNetworkFlowRetentionFixtureComparison"},
	"NF-FIX-028-graph-aggregate-bounds":      {3, "TestNetworkFlowAggregateBoundsFixtureComparison"},
}

func runFixtureComparison(t *testing.T, id string, revision int, compare func(*networkflowfixtures.Bundle) ([]map[string]any, error)) {
	t.Helper()
	registration, ok := behavioralFixtureComparisons[id]
	if !ok || registration.Revision != revision || strings.SplitN(t.Name(), "/", 2)[0] != registration.TestName {
		t.Fatalf("unregistered fixture comparison: %s revision %d test %s", id, revision, t.Name())
	}
	runRoot, err := networkflowfixtures.SelectedRunRoot()
	if err != nil {
		t.Fatal(err)
	}
	summary, err := networkflowfixtures.Run(networkflowfixtures.Config{
		CorpusRoot: filepath.Join(RepoRoot(t), "fixtures", "network-flow"),
		RunRoot:    runRoot, FixtureID: id, Revision: revision,
	}, func(bundle *networkflowfixtures.Bundle) error {
		outcomes, err := compare(bundle)
		if err != nil {
			return err
		}
		if len(outcomes) == 0 {
			return errors.New("fixture adapter returned no observable outcomes")
		}
		actual := map[string]any{"fixture_id": id, "freeze_revision": revision, "outcomes": outcomes}
		data, err := encodeFixtureEvidence(actual)
		if err != nil {
			return err
		}
		if err := bundle.Record("actual.json", data); err != nil {
			return err
		}
		slug := strings.SplitN(id, "-", 4)[3]
		for _, path := range []string{"expected/" + slug + ".expected.json", "transcripts/" + slug + ".transcript.json"} {
			expected, err := bundle.Read(path)
			if err != nil {
				return err
			}
			if err := compareFixtureJSON(expected, data); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			// Preserve identity, revision, case labels and requests. Changing an
			// observed result must fail, even if identity checks still pass.
			wrong, err := wrongFixtureExpectation(expected)
			if err != nil {
				return err
			}
			if compareFixtureJSON(wrong, data) == nil {
				return errors.New("fixture comparator accepted wrong observable expectation")
			}
		}
		if id == "NF-FIX-029-cisco-sna-flow-search" {
			if err := ciscoExampleNegativeControls(bundle, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("fixture comparison failed: %v (summary %s)", err, summary)
	}
}

// These owner-specific controls target the contract members omitted by the
// former example summary. They are not an executable fixture-language layer.
func ciscoExampleNegativeControls(bundle *networkflowfixtures.Bundle, actual []byte) error {
	type object = map[string]any
	first := func(value any) object { return value.([]any)[0].(object) }
	graph := func(o object) object { return o["graph"].(object)["graph_projection_result"].(object) }
	controls := []struct {
		name   string
		mutate func(object)
	}{
		{"source_descriptor", func(o object) {
			first(o["inspection"].(object)["owner_result"].(object)["source_columns"])["raw_header_sha256"] = strings.Repeat("0", 64)
		}},
		{"mapping_default", func(o object) {
			first(o["preview"].(object)["owner_result"].(object)["materialized_mapping"].(object)["field_mappings"])["empty_value_policy"] = "empty_string_is_null"
		}},
		{"table_profile", func(o object) {
			o["table"].(object)["table"].(object)["source_profile_id"] = sourceProfileCiscoSNANetFlowCSV
		}},
		{"projected_identity", func(o object) { first(graph(o)["vertices"])["vertex_id"] = "vx_incorrect" }},
		{"endpoint_binding", func(o object) { first(graph(o)["edges"])["src_vertex_id"] = "vx_incorrect" }},
		{"selector", func(o object) {
			first(o["graph"].(object)["vertex_selectors"])["selector"].(object)["source_vertex_id"] = "nfe_incorrect"
		}},
		{"metadata", func(o object) {
			first(graph(o)["vertices"])["metadata"].(object)["mapped_metadata"].(object)["mapping_fingerprints"] = []any{}
		}},
		{"adapter_definition", func(o object) {
			first(first(o["composition"].(object)["projection_invocations"])["projection_input"].(object)["property_definitions"])["required"] = false
		}},
		{"missing_member", func(o object) { delete(o["table"].(object)["table"].(object), "mapping_fingerprint") }},
		{"extra_member", func(o object) { o["table"].(object)["table"].(object)["unexpected"] = true }},
		{"ordering", func(o object) { a := graph(o)["vertices"].([]any); a[0], a[1] = a[1], a[0] }},
		{"null_port_is_not_zero", func(o object) {
			for _, item := range graph(o)["edges"].([]any) {
				p := item.(object)["properties"].(object)
				if p["dst_port"] == nil {
					p["dst_port"] = json.Number("0")
					return
				}
			}
		}},
		{"cursor_binding", func(o object) { o["row_cursor"].(object)["query_hash"] = strings.Repeat("0", 64) }},
	}
	passed := []string{}
	for _, control := range controls {
		decoded, err := canonicaljson.DecodeStrict(actual)
		if err != nil {
			return err
		}
		control.mutate(first(decoded.(object)["outcomes"]))
		wrong, err := encodeFixtureEvidence(decoded)
		if err != nil {
			return err
		}
		if compareFixtureJSON(wrong, actual) == nil {
			return fmt.Errorf("example negative control did not fail: %s", control.name)
		}
		passed = append(passed, control.name)
	}
	evidence, err := encodeFixtureEvidence(map[string]any{"rejected_incorrect_expectations": passed})
	if err != nil {
		return err
	}
	return bundle.Record("negative-controls.json", evidence)
}

// Mutate one existing scalar under an observable outcome, without interpreting
// product semantics. Case labels and request stimuli are intentionally excluded.
func wrongFixtureExpectation(data []byte) ([]byte, error) {
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("fixture expectation is not an object")
	}
	outcomes, ok := object["outcomes"].([]any)
	if !ok || len(outcomes) == 0 {
		return nil, errors.New("fixture expectation has no outcomes")
	}
	for _, outcome := range outcomes {
		fields, ok := outcome.(map[string]any)
		if !ok {
			return nil, errors.New("fixture outcome is not an object")
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			if key != "case" && key != "request" && key != "kind" {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			if changed, ok := wrongFixtureValue(fields[key]); ok {
				fields[key] = changed
				return encodeFixtureEvidence(object)
			}
		}
	}
	return nil, errors.New("fixture has no observable value for negative control")
}

func wrongFixtureValue(value any) (any, bool) {
	switch value := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if changed, ok := wrongFixtureValue(value[key]); ok {
				value[key] = changed
				return value, true
			}
		}
	case []any:
		for index, entry := range value {
			if changed, ok := wrongFixtureValue(entry); ok {
				value[index] = changed
				return value, true
			}
		}
	case string:
		return value + "-incorrect", true
	case bool:
		return !value, true
	case json.Number:
		if value == "0" {
			return json.Number("1"), true
		}
		return json.Number("0"), true
	case nil:
		return "incorrect-null", true
	}
	return nil, false
}

// Preserve the literal harness redaction marker in retained JSON bytes.
func encodeFixtureEvidence(value any) ([]byte, error) {
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func compareFixtureJSON(expected, actual []byte) error {
	left, err := canonicaljson.DecodeStrict(expected)
	if err != nil {
		return err
	}
	right, err := canonicaljson.DecodeStrict(actual)
	if err != nil {
		return err
	}
	leftBytes, err := json.Marshal(left)
	if err != nil {
		return err
	}
	rightBytes, err := json.Marshal(right)
	if err != nil {
		return err
	}
	if !bytes.Equal(leftBytes, rightBytes) {
		return errors.New("complete output differs from independent expectation; see actual.json")
	}
	return nil
}

func TestNetworkFlowCanonicalFixtureComparison(t *testing.T) {
	t.Run("encoding_errors_do_not_return_substitute_bytes", func(t *testing.T) {
		cycle := map[string]any{}
		cycle["self"] = cycle
		for _, value := range []any{cycle, float64(1), json.Number("1.5"), json.Number("1e2"), json.Number("01"), "\xff", map[string]any{"\xff": 1}, json.RawMessage(`"\ud800"`), json.RawMessage(`{"n":1,"n":2}`), json.RawMessage(`{`), make(chan int)} {
			data, err := canonicalJSON(value)
			if err == nil || data != nil {
				t.Fatal("invalid canonical value produced bytes")
			}
		}
	})
	runFixtureComparison(t, "NF-FIX-019-canonical-json-unicode", 5, canonicalFixtureOutcomes)
}

func canonicalFixtureOutcomes(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
	data, err := bundle.Read("source/canonical-json-unicode.jsonl")
	if err != nil {
		return nil, err
	}
	outcomes := []map[string]any{}
	for i, line := range bytes.Split(bytes.TrimSuffix(data, []byte{'\n'}), []byte{'\n'}) {
		value, err := canonicaljson.DecodeStrict(line)
		outcome := map[string]any{"case": fmt.Sprintf("json-%02d", i+1), "request": string(line)}
		if err != nil {
			outcome["result"] = map[string]any{"status": "rejected", "reason": "invalid_json"}
		} else {
			canonical, err := canonicalJSON(value)
			if err != nil {
				outcome["result"] = map[string]any{"status": "rejected", "reason": "invalid_canonical_value"}
			} else {
				outcome["result"] = map[string]any{"status": "accepted", "canonical": string(canonical), "sha256": sha256Hex(canonical)}
			}
		}
		outcomes = append(outcomes, outcome)
	}
	var textCases []struct {
		Input string `json:"input"`
	}
	data, err = bundle.Read("source/text.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &textCases); err != nil {
		return nil, err
	}
	for i, test := range textCases {
		normalized := norm.NFC.String(trimUnicodeWhitespace(test.Input))
		outcomes = append(outcomes, map[string]any{"case": fmt.Sprintf("text-%02d", i+1), "request": test.Input, "result": map[string]any{"trimmed": trimUnicodeWhitespace(test.Input), "normalized": normalized}})
	}
	var digestCases []struct {
		Fingerprint string            `json:"fingerprint"`
		Values      map[string]any    `json:"values"`
		Unmapped    retainedRawValues `json:"unmapped"`
	}
	data, err = bundle.Read("source/digests.json")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&digestCases); err != nil {
		return nil, err
	}
	for i, test := range digestCases {
		outcomes = append(outcomes, map[string]any{"case": fmt.Sprintf("digest-%02d", i+1), "request": test, "result": mustNormalizedRowDigest(test.Fingerprint, test.Values, test.Unmapped)})
	}
	data, err = bundle.Read("source/timestamp-profile.json")
	if err != nil {
		return nil, err
	}
	var profile timestampProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	canonical := mustCanonicalJSON(profile)
	outcomes = append(outcomes, map[string]any{"case": "closed-profile", "request": json.RawMessage(data), "result": map[string]any{"canonical": string(canonical), "sha256": sha256Hex(canonical)}})
	data, err = bundle.Read("source/mapping-vectors.json")
	if err != nil {
		return nil, err
	}
	var mappings []struct {
		Name      string           `json:"name"`
		Field     *fieldMapping    `json:"field_mapping,omitempty"`
		Approved  *approvedMapping `json:"approved_mapping,omitempty"`
		SourceSHA string           `json:"source_sha256,omitempty"`
	}
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, err
	}
	for _, vector := range mappings {
		result := map[string]any{}
		if vector.Field != nil {
			encoded, err := canonicalJSON(*vector.Field)
			if err != nil {
				return nil, err
			}
			result["canonical"] = string(encoded)
		} else if vector.Approved != nil {
			if err := validateApprovedMapping(*vector.Approved); err != nil {
				return nil, err
			}
			fingerprint, err := mappingFingerprint(*vector.Approved, vector.SourceSHA)
			if err != nil {
				return nil, err
			}
			result["fingerprint"] = fingerprint
		} else {
			return nil, errors.New("missing mapping vector")
		}
		outcomes = append(outcomes, map[string]any{"case": vector.Name, "request": vector, "result": result})
	}
	data, err = bundle.Read("source/graph-vectors.json")
	if err != nil {
		return nil, err
	}
	var graphs []struct {
		Name        string          `json:"name"`
		Semantic    json.RawMessage `json:"semantic"`
		IncidentID  uuid.UUID       `json:"incident_id"`
		DisplayName string          `json:"display_name"`
	}
	if err := json.Unmarshal(data, &graphs); err != nil {
		return nil, err
	}
	for _, vector := range graphs {
		semantic, failure := decodeGraphSemanticRequest(vector.Semantic, defaultEffectiveLimits())
		if failure != nil {
			return nil, fmt.Errorf("graph vector admission: %v", failure)
		}
		canonical, err := canonicalGraphSemanticJSON(semantic.Raw)
		if err != nil {
			return nil, err
		}
		digest, err := graphQueryDigestForSemantic(vector.IncidentID, semantic.SelectedTableIDs, semantic)
		if err != nil {
			return nil, err
		}
		request, err := mutationRequestHash(routeKeyGraphViewsCreate, "collection", map[string]any{"display_name": vector.DisplayName, "semantic_query": semantic.Raw})
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, map[string]any{"case": vector.Name, "request": vector, "result": map[string]any{"canonical": string(canonical), "sha256": sha256Hex(canonical), "graph_query_digest": digest, "mutation_sha256": fmt.Sprintf("%x", request)}})
	}
	return outcomes, nil
}

type csvFixtureCase struct {
	Name           string `json:"name"`
	Source         string `json:"source"`
	Preview        bool   `json:"preview"`
	MaxRows        int64  `json:"max_rows"`
	ExpectedSHA256 string `json:"expected_sha256"`
}

func TestNetworkFlowCSVFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-005-csv-parser-edges", 3, csvFixtureOutcomes)
}

func TestNetworkFlowPreviewFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-021-preview-boundaries", 3, csvFixtureOutcomes)
}

func csvFixtureOutcomes(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
	data, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var cases []csvFixtureCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}
	outcomes := []map[string]any{}
	for _, test := range cases {
		data, err := bundle.Read(test.Source)
		if err != nil {
			return nil, err
		}
		limits := defaultLimits()
		if test.MaxRows > 0 {
			limits.MaxRowsPerCSV = test.MaxRows
		}
		var parsed parsedCSV
		if test.Preview {
			parsed, err = parseCSVPreview(bytes.NewReader(data), test.ExpectedSHA256, limits)
		} else {
			parsed, err = parseCSVApply(bytes.NewReader(data), test.ExpectedSHA256, limits)
		}
		var result any
		if err != nil {
			var sourceError *sourceValidationError
			if errors.As(err, &sourceError) {
				result = map[string]any{"status": "rejected", "error": sourceError}
			} else if errors.Is(err, errSourceChanged) {
				result = map[string]any{"status": "rejected", "error": "source_changed"}
			} else {
				return nil, err
			}
		} else {
			result = map[string]any{"status": "accepted", "source_sha256": parsed.SourceContentSHA256, "columns": parsed.SourceColumns, "records": parsed.Records}
		}
		outcomes = append(outcomes, map[string]any{"case": test.Name, "request": test, "result": result})
	}
	return outcomes, nil
}

func TestNetworkFlowJSONFixtureComparison(t *testing.T) {
	runFixtureComparison(t, "NF-FIX-010-json-admission", 4, jsonFixtureOutcomes)
}

func jsonFixtureOutcomes(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
	data, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var cases []struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}
	outcomes := []map[string]any{}
	for _, test := range cases {
		data, err := bundle.Read(test.Source)
		if err != nil {
			return nil, err
		}
		request, failure := decodeAcceptedRowQueryRequestHTTP(bytes.NewReader(data), schemaTableQueryRequest, schemaTableQueryContinuation, defaultLimits())
		var result any
		if failure != nil {
			result = map[string]any{"status": "rejected", "error": failure}
		} else {
			result = map[string]any{"status": "accepted", "request": request}
		}
		outcomes = append(outcomes, map[string]any{"case": test.Name, "request": string(data), "result": result})
	}
	return outcomes, nil
}
