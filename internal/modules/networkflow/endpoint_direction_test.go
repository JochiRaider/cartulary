package networkflow

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"reflect"
	"strings"
	"testing"
)

func testEndpointOrientation(t *testing.T) {
	t.Helper()
	serverPort, clientPort := int32(5432), int32(57455)
	server := flowEndpoint{"10.20.20.40", &serverPort}
	client := flowEndpoint{"10.20.20.30", &clientPort}
	for _, tc := range []struct {
		first, second string
		kind, reason  string
		bad           bool
	}{
		{"server", "client", directionClientServer, "", false},
		{"client", "server", directionClientServer, "", false},
		{"", "", directionUnknown, "missing_orientation", false},
		{"client", "", directionUnknown, "incomplete_orientation", false},
		{"", "server", directionUnknown, "incomplete_orientation", false},
		{"server", "server", "", "", true},
		{"client", "client", "", "", true},
		{"responder", "client", "", "", true},
	} {
		r, err := deriveEndpoints(deriveCiscoOrientation, server, client, tc.first, tc.second)
		if (err != nil) != tc.bad {
			t.Fatalf("roles %q/%q: %v", tc.first, tc.second, err)
		}
		if tc.bad {
			continue
		}
		if r.A.IP != client.IP || *r.A.Port != clientPort || r.B.IP != server.IP || *r.B.Port != serverPort || r.Direction.Kind != tc.kind || r.Direction.Reason != tc.reason || !r.Direction.valid() {
			t.Fatalf("lost tuple or direction: %#v", r)
		}
	}
	forward, _ := deriveEndpoints(deriveCiscoOrientation, client, server, "client", "server")
	reversed, _ := deriveEndpoints(deriveCiscoOrientation, server, client, "server", "client")
	if forward.Direction != reversed.Direction || reversed.Direction.FromEndpoint != "a" {
		t.Fatal("export position affected direction")
	}
	self, _ := deriveEndpoints(deriveCiscoOrientation, server, server, "server", "client")
	if self.Direction.FromEndpoint != "a" {
		t.Fatal("identical tuple tie is unstable")
	}
	for _, raw := range []string{"CLIENT", " client ", "Client"} {
		if role, ok := normalizeOrientation(raw); !ok || role != "client" {
			t.Fatalf("normalization: %q", raw)
		}
	}
	for _, raw := range []string{"\tclient", "clİent", "responder", "client\n"} {
		if _, ok := normalizeOrientation(raw); ok {
			t.Fatalf("unsupported token admitted: %q", raw)
		}
	}

	a, b := flowEndpoint{"192.0.2.1", nil}, flowEndpoint{"192.0.2.2", nil}
	evidence := declaredEvidenceFixture()
	direction := flowDirection{Kind: directionSourceDestination, FromEndpoint: "a"}
	if err := validateObservationDirection(a, b, direction, evidence, "cisco_sna_netflow_csv_v2"); err != nil {
		t.Fatal(err)
	}
	var reordered map[string]any
	if err := json.Unmarshal(evidence, &reordered); err != nil {
		t.Fatal(err)
	}
	inputs := reordered["inputs"].([]any)
	inputs[0], inputs[1] = inputs[1], inputs[0]
	rawEvidence, _ := json.Marshal(reordered)
	if validateObservationDirection(a, b, direction, rawEvidence, "cisco_sna_netflow_csv_v2") == nil {
		t.Fatal("noncanonical evidence input ordering admitted")
	}
	null := flowEndpoint{"2001:db8::1", nil}
	populated := flowEndpoint{"2001:db8::1", &serverPort}
	if compareFlowEndpoints(null, populated) >= 0 || compareFlowEndpoints(server, null) >= 0 {
		t.Fatal("family/null ordering")
	}
}

func declaredEvidenceFixture() json.RawMessage {
	inputs := []any{}
	for i, key := range []string{fieldInputFirstIP, fieldInputFirstPort, fieldInputSecondIP, fieldInputSecondPort} {
		inputs = append(inputs, map[string]any{"input_key": key, "source_column_ordinal": i + 1, "raw_header_sha256": strings.Repeat("a", 64), "raw_value_sha256": strings.Repeat("b", 64)})
	}
	raw, _ := json.Marshal(map[string]any{"derivation_id": deriveDeclaredDirection, "first_endpoint": "a", "first_role": "", "second_role": "", "inputs": inputs})
	return raw
}

func testOrientationImportBoundaries(t *testing.T) {
	t.Helper()
	_, preview, source := approvedVariantFixture(t, "cisco_sna_flow_search_csv_v2")
	baseline, err := decodeApprovedMapping(preview.OwnerMapping, defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	header := strings.Split(source, "\n")[0]
	base := strings.Split(strings.Split(source, "\n")[1], ",")
	base[2], base[3], base[4], base[5] = "10.20.20.40", "10.20.20.30", "5432", "57455"
	var firstRow, swappedRow normalizedRow
	for _, tc := range []struct {
		name, first, second, kind, reason string
		ordinals                          []int64
	}{
		{"reverse", " SERVER ", " Client ", directionClientServer, "", nil},
		{"missing", "", "", directionUnknown, "missing_orientation", nil},
		{"partial", "server", "", directionUnknown, "incomplete_orientation", nil},
		{"partial peer", "", "client", directionUnknown, "incomplete_orientation", nil},
		{"conflict client", "client", "client", "", "", []int64{10, 11}},
		{"conflict server", "server", "server", "", "", []int64{10, 11}},
		{"invalid peer", "client", "responder", "", "", []int64{11}},
		{"invalid subject", "responder", "server", "", "", []int64{10}},
	} {
		t.Run("orientation_"+tc.name, func(t *testing.T) {
			cells := append([]string(nil), base...)
			cells[9], cells[10] = tc.first, tc.second
			source := header + "\n" + strings.Join(cells, ",") + "\n"
			parsed, err := parseCSVPreview(strings.NewReader(source), "", defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			mapping := baseline
			mapping.SourceColumns = parsed.SourceColumns
			fingerprint, err := mappingFingerprint(mapping, parsed.SourceContentSHA256)
			if err != nil {
				t.Fatal(err)
			}
			row, diagnostics, err := validateCSVRecord(parsed.Records[0], parsed.SourceColumns, mapping, fingerprint, parsed.SourceContentSHA256, defaultLimits(), sourceFieldMappings(mapping))
			if err != nil {
				t.Fatal(err)
			}
			counts, err := validatePreview(parsed, mapping, fingerprint, defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			if len(tc.ordinals) > 0 {
				if len(diagnostics) != len(tc.ordinals) || counts.RejectedRowCount != 1 || counts.AcceptedRowCount != 0 {
					t.Fatalf("diagnostic multiplicity changed rejected observation count: %#v", counts)
				}
				for i, ordinal := range tc.ordinals {
					if diagnostics[i].SourceColumnOrdinal == nil || *diagnostics[i].SourceColumnOrdinal != ordinal || diagnostics[i].RawHeaderSHA256 == nil || *diagnostics[i].RawHeaderSHA256 != parsed.SourceColumns[ordinal-1].RawHeaderSHA256 {
						t.Fatalf("input attribution: %#v", diagnostics)
					}
				}
				return
			}
			if len(diagnostics) != 0 || row.Direction.Kind != tc.kind || row.Direction.Reason != tc.reason || row.EndpointAIP != "10.20.20.30" || *row.EndpointAPort != 57455 || row.EndpointBIP != "10.20.20.40" || *row.EndpointBPort != 5432 {
				t.Fatalf("incorrect normalization: %#v %#v", row, diagnostics)
			}
			if tc.kind == directionUnknown && counts.UnknownDirectionCount != 1 || tc.kind != directionUnknown && counts.ResolvedDirectionCount != 1 {
				t.Fatal("preview direction counts")
			}
			capability := imports.ImportSourceCapability{SourceStreamRef: "impsrc_" + strings.Repeat("1", 32), SourceContentSHA256: parsed.SourceContentSHA256, SourceMediaType: "text/csv", SourceByteSize: int64(len(source))}
			stream := &applySourceFixture{source: source, capability: capability}
			f := &importFacade{sourceStore: stream, limits: defaultLimits()}
			request := imports.ExtensionImportApplyRequest{SourceCapability: capability, ExpectedSourceContentSHA256: parsed.SourceContentSHA256, MappingFingerprint: fingerprint}
			var previous importValidationSummary
			for pass := 0; pass < 2; pass++ {
				seen := 0
				summary, err := f.validateImportPass(context.Background(), nil, request, mapping, pass == 0, func(actual normalizedRow) error {
					seen++
					if !reflect.DeepEqual(actual, row) {
						t.Fatal("preview/apply derivation diverged")
					}
					return nil
				})
				if err != nil || seen != 1 || summary.Accepted != 1 || summary.Rejected != 0 {
					t.Fatalf("bounded apply: %v %#v", err, summary)
				}
				if pass == 1 && !summary.agrees(previous) {
					t.Fatal("apply passes diverged")
				}
				previous = summary
			}
			if stream.opens != 2 || stream.closes != 2 {
				t.Fatal("source streams leaked")
			}
			if tc.name == "reverse" {
				firstRow = row
			}
		})
	}
	// Exchanged exporter positions are one connection, but remain distinct evidence.
	cells := append([]string(nil), base...)
	cells[2], cells[3], cells[4], cells[5], cells[9], cells[10] = "10.20.20.30", "10.20.20.40", "57455", "5432", "client", "server"
	parsed, err := parseCSVApply(strings.NewReader(header+"\n"+strings.Join(cells, ",")+"\n"), "", defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	baseline.SourceColumns = parsed.SourceColumns
	fingerprint := mustMappingFingerprint(baseline, parsed.SourceContentSHA256)
	swappedRow, diagnostics, err := validateCSVRecord(parsed.Records[0], parsed.SourceColumns, baseline, fingerprint, parsed.SourceContentSHA256, defaultLimits(), sourceFieldMappings(baseline))
	if err != nil || len(diagnostics) != 0 || swappedRow.Direction != firstRow.Direction || swappedRow.EndpointAIP != firstRow.EndpointAIP || swappedRow.EndpointBIP != firstRow.EndpointBIP || swappedRow.NormalizedRowDigestSHA256 == firstRow.NormalizedRowDigestSHA256 || swappedRow.SourceRowDigestSHA256 == firstRow.SourceRowDigestSHA256 {
		t.Fatal("connection semantics conflated with observation identity")
	}
	changed := baseline
	changed.FieldMappings = append([]fieldMapping(nil), baseline.FieldMappings...)
	for i := range changed.FieldMappings {
		m := &changed.FieldMappings[i]
		if m.FieldKey == fieldInputFirstOrientation {
			m.SourceColumnOrdinal = 11
		}
		if m.FieldKey == fieldInputSecondOrientation {
			m.SourceColumnOrdinal = 10
		}
	}
	if mustMappingFingerprint(changed, parsed.SourceContentSHA256) == fingerprint {
		t.Fatal("direction binding omitted from approval identity")
	}
	changed = baseline
	changed.EndpointDerivationID += "-changed"
	if mustMappingFingerprint(changed, parsed.SourceContentSHA256) == fingerprint {
		t.Fatal("algorithm omitted from approval identity")
	}
	changed = baseline
	changed.FieldMappings = append([]fieldMapping(nil), baseline.FieldMappings...)
	for i, m := range changed.FieldMappings {
		if m.FieldKey == fieldInputFirstOrientation {
			changed.FieldMappings = append(changed.FieldMappings[:i], changed.FieldMappings[i+1:]...)
			break
		}
	}
	if validateApprovedMapping(changed) == nil {
		t.Fatal("missing orientation input admitted")
	}
}
