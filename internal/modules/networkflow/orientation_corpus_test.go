package networkflow

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// The oracle was calculated with Python csv/ipaddress, independently of this
// parser and graph composer. It includes every endpoint and every contributor.
func testOrientationCorpus(t *testing.T) {
	const root = "../../../fixtures/network-flow-orientation/v1"
	var manifest struct {
		Datasets []struct {
			Name         string `json:"name"`
			CSV          string `json:"csv"`
			CSVHash      string `json:"csv_sha256"`
			Expected     string `json:"expected"`
			ExpectedHash string `json:"expected_sha256"`
			Observations int    `json:"observations"`
			Reversals    int    `json:"subject_server_reversals"`
			Unknown      int    `json:"unknown_direction"`
			Edges        int    `json:"graph_edges"`
		} `json:"datasets"`
	}
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, dataset := range manifest.Datasets {
		t.Run("corpus_"+dataset.Name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join(root, dataset.CSV))
			if err != nil {
				t.Fatal(err)
			}
			expectedBytes, err := os.ReadFile(filepath.Join(root, dataset.Expected))
			if err != nil {
				t.Fatal(err)
			}
			if sha256Hex(source) != dataset.CSVHash || sha256Hex(expectedBytes) != dataset.ExpectedHash {
				t.Fatal("fixture bytes changed")
			}
			var expected struct {
				Rows []struct {
					Number    int64         `json:"source_row_number"`
					A         flowEndpoint  `json:"endpoint_a"`
					B         flowEndpoint  `json:"endpoint_b"`
					Direction flowDirection `json:"direction"`
					Protocol  int32         `json:"protocol"`
					Bytes     string        `json:"bytes"`
					Packets   string        `json:"packets"`
				} `json:"rows"`
				Edges []struct {
					Kind     string       `json:"direction_kind"`
					First    flowEndpoint `json:"first"`
					Second   flowEndpoint `json:"second"`
					Protocol int32        `json:"protocol"`
					Rows     []int64      `json:"source_rows"`
					Bytes    string       `json:"bytes"`
					Packets  string       `json:"packets"`
				} `json:"edges"`
			}
			if err := json.Unmarshal(expectedBytes, &expected); err != nil {
				t.Fatal(err)
			}
			parsed, err := parseCSVApply(bytes.NewReader(source), "", defaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			profile, _ := mappingSourceProfile("cisco_sna_flow_search_csv_v2")
			fields := []fieldMapping{}
			for _, field := range profile.Fields {
				if field.Requirement != "required" {
					continue
				}
				for _, column := range parsed.SourceColumns {
					if column.RawHeaderText == field.Aliases[0] {
						fields = append(fields, fieldMapping{MappingKind: mappingKindSourceColumn, FieldKey: field.FieldKey, SourceColumnOrdinal: column.SourceColumnOrdinal, TransformID: field.TransformID, Combinability: "single_source_only"})
					}
				}
			}
			mapped := map[int]bool{}
			for _, field := range fields {
				mapped[field.SourceColumnOrdinal] = true
			}
			for _, column := range parsed.SourceColumns {
				if !mapped[column.SourceColumnOrdinal] {
					fields = append(fields, fieldMapping{MappingKind: mappingKindIgnoredSourceColumn, SourceColumnOrdinal: column.SourceColumnOrdinal, IgnoreReason: "policy_accounting"})
				}
			}
			candidate := mappingCandidate{EndpointDerivationID: profile.EndpointDerivationID, TargetKind: targetKindNetworkFlowTable, TargetTableSchemaID: targetTableSchemaID, SourceProfileID: profile.SourceProfileID, TimestampProfile: materializeTimestampProfile(timestampProfile{Mode: "cisco_sna_datetime_v1"}), UnknownColumnPolicy: unknownColumnPolicyIgnore, FieldMappings: candidateFieldMappingsForTest(t, fields)}
			mappingBytes, _ := json.Marshal(candidate)
			mapping, err := materializeApprovedMapping(mappingBytes, parsed.SourceColumns)
			if err != nil {
				t.Fatal(err)
			}
			fingerprint, err := mappingFingerprint(mapping, parsed.SourceContentSHA256)
			if err != nil {
				t.Fatal(err)
			}
			incident := uuid.MustParse("11111111-1111-4111-8111-111111111111")
			table := tableRecord{TableID: "nft_11111111111111111111111111111111", IncidentID: incident, SourceProfileID: profile.SourceProfileID, MappingFingerprint: fingerprint}
			context := rowImportContext{incidentID: incident, sourceProfile: profile.SourceProfileID, parserProfile: mapping.ParserProfileID, fingerprint: fingerprint, sourceSHA: parsed.SourceContentSHA256, createdAt: time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)}
			composition := graphComposition{Vertices: map[string]*graphVertex{}, Edges: map[string]*graphEdge{}, TableRanks: map[string]int{table.TableID: 0}, ResultLimits: graphResultLimits{MaxVertices: 10000, MaxEdges: 10000, MaxExampleRowRefsPerEdge: 5000, MaxAggregateCounterDigits: 100, MaxContributingRows: 10000}, IncludeExamples: true}
			if len(parsed.Records) != dataset.Observations || len(expected.Rows) != dataset.Observations {
				t.Fatal("observation count")
			}
			unknown, reversals := 0, 0
			identities := map[string]bool{}
			for i, record := range parsed.Records {
				normalized, diagnostics, err := validateCSVRecord(record, parsed.SourceColumns, mapping, fingerprint, parsed.SourceContentSHA256, defaultLimits(), sourceFieldMappings(mapping))
				if err != nil || len(diagnostics) != 0 {
					t.Fatalf("row %d rejected: %v %v", record.SourceRowNumber, diagnostics, err)
				}
				row, err := context.persistedRow(table.TableID, normalized)
				if err != nil {
					t.Fatal(err)
				}
				want := expected.Rows[i]
				a, b := row.endpoints()
				if !reflect.DeepEqual(a, want.A) || !reflect.DeepEqual(b, want.B) || row.Direction != want.Direction || row.SourceRowNumber != want.Number || row.BytesCount != want.Bytes || row.PacketsCount != want.Packets || row.IPProtocol != want.Protocol {
					t.Fatalf("observation %d disagrees with independent oracle", record.SourceRowNumber)
				}
				if identities[row.RowID] {
					t.Fatal("observations deduplicated")
				}
				identities[row.RowID] = true
				if row.Direction.Kind == directionUnknown {
					unknown++
				}
				var evidence struct {
					FirstRole string `json:"first_role"`
				}
				if json.Unmarshal(row.DirectionEvidence, &evidence) != nil {
					t.Fatal("missing evidence")
				}
				if evidence.FirstRole == "server" {
					reversals++
				}
				if failure := composeDefaultGraphRow(incident, row, map[string]tableRecord{table.TableID: table}, &composition); failure != nil {
					t.Fatal(failure)
				}
			}
			if unknown != dataset.Unknown || reversals != dataset.Reversals || len(composition.Edges) != dataset.Edges {
				t.Fatalf("counts: unknown %d reversals %d edges %d", unknown, reversals, len(composition.Edges))
			}
			for _, want := range expected.Edges {
				key := edgeSemantics{DirectionKind: want.Kind, CounterSemantics: "conversation_total", First: want.First, Second: want.Second, Protocol: want.Protocol}
				edge := composition.Edges[key.identity(incident, nil, nil)]
				if edge == nil || edge.BytesSum.String() != want.Bytes || edge.PacketsSum.String() != want.Packets || edge.FlowRowCount != len(want.Rows) {
					t.Fatalf("edge totals disagree: %#v", want)
				}
				got := []int64{}
				for _, row := range edge.ExampleRows {
					got = append(got, row.SourceRowNumber)
				}
				if !reflect.DeepEqual(got, want.Rows) {
					t.Fatalf("edge contributors disagree: %v != %v", got, want.Rows)
				}
			}
		})
	}
}
