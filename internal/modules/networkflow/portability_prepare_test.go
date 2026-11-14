package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
)

type portableFixtureMembers struct{ portabilitysupport.MapBundle }

func (m portableFixtureMembers) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	if err := ctx.Err(); err != nil {
		return incidentportability.Member{}, err
	}
	if _, exists := m.MapBundle[path]; exists {
		return incidentportability.Member{}, errPortableNetworkFlow
	}
	raw, err := io.ReadAll(io.LimitReader(source, 1<<20))
	if err != nil {
		return incidentportability.Member{}, err
	}
	m.MapBundle[path] = raw
	for _, member := range m.Members() {
		if member.Path == path {
			return member, nil
		}
	}
	panic("missing test member")
}

type portableFixtureSources struct {
	source imports.CommittedSource
	bytes  []byte
}

func (f portableFixtureSources) ReadSource(_ context.Context, id uuid.UUID) (imports.CommittedSource, error) {
	if id != f.source.CommittedSourceID {
		return imports.CommittedSource{}, errors.New("source not found")
	}
	return f.source, nil
}
func (f portableFixtureSources) OpenSource(ctx context.Context, id uuid.UUID) (imports.CommittedSource, io.ReadCloser, error) {
	value, err := f.ReadSource(ctx, id)
	return value, io.NopCloser(bytes.NewReader(f.bytes)), err
}

type portableFixtureIndicators struct{}

func (portableFixtureIndicators) FindIndicator(context.Context, uuid.UUID) (indicators.IndicatorReference, error) {
	return indicators.IndicatorReference{}, errors.New("indicator not found")
}

func portableAdmissionFixture(t *testing.T) (incidentbundles.ImportInvocation, portableFixtureMembers) {
	t.Helper()
	ctx := t.Context()
	limits := defaultLimits()
	f, request, sourceFixture := streamingApplyFixture(t, validPreviewRow+validPreviewRow, limits)
	request.IncidentID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	request.ActorUserID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	request.ImportSessionID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	request.ImportUnitID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	applied, err := f.prepareImportApply(ctx, nil, request)
	if err != nil {
		t.Fatal(err)
	}
	params := applied.params
	rowContext, err := validatedRowImportContext(params)
	if err != nil {
		t.Fatal(err)
	}
	table := tableRecord{TableID: "nft_" + strings.Repeat("a", 32), IncidentID: request.IncidentID, DisplayName: "Original flow source", TableVersion: 1, TableStatus: tableStatusActive, CommittedSourceID: params.CommittedSourceID, SourceInterpretationSHA256: params.SourceInterpretationSHA256, SourceRawCellScalarLimit: params.SourceRawCellScalarLimit, SourceContentSHA256: params.SourceContentSHA256, SourceFilenameDisplay: "source.csv", MappingFingerprint: params.MappingFingerprint, SourceProfileID: params.SourceProfileID, ParserProfileID: params.ParserProfileID, RowCountAccepted: 2, CreatedByUserID: request.ActorUserID, CreatedAt: params.Now, UpdatedAt: params.Now}
	var rows []flowRow
	if err := params.VisitRows(func(row normalizedRow) error {
		retained, err := rowContext.persistedRow(table.TableID, row)
		rows = append(rows, retained)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	sort.Slice(rows, func(i, j int) bool { return comparePortableRows(rows[i], rows[j]) < 0 })
	members := portableFixtureMembers{MapBundle: portabilitysupport.MapBundle{}}
	for _, path := range portabilityFamilyPaths {
		members.MapBundle[portabilityMemberPrefix+path] = []byte{}
	}
	encode := func(path string, value any) {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		members.MapBundle[portabilityMemberPrefix+path] = append(members.MapBundle[portabilityMemberPrefix+path], append(raw, '\n')...)
	}
	encode("tables.ndjson", portableTableRecord(table))
	for _, row := range rows {
		encode("observations.ndjson", portableFlowRow(row))
	}
	catalog, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: request.ActorUserID.String(), DisplayName: "Historical author"}})
	if err != nil {
		t.Fatal(err)
	}
	envelope := mappingTestJSON(t, map[string]any{"owner_mapping": json.RawMessage(request.OwnerMapping), "owner_mapping_schema_id": importMappingSchemaID})
	sources := portableFixtureSources{source: imports.CommittedSource{CommittedSourceID: params.CommittedSourceID, IncidentID: request.IncidentID, ContentSHA256: request.ExpectedSourceContentSHA256, MappingFingerprint: request.MappingFingerprint, TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID, SourceFileKind: "csv", OriginalFilename: "source.csv", HeaderRowRef: 1, DataStartRowRef: 2, LocatorKind: "csv_file", Locator: "file", CreatedByUserID: request.ActorUserID, ApprovedMapping: envelope, CommittedAt: params.Now}, bytes: []byte(sourceFixture.source)}
	dependencies, err := NewPortabilityDependencies(request.IncidentID, sources, portableFixtureIndicators{})
	if err != nil {
		t.Fatal(err)
	}
	invocation := incidentbundles.ImportInvocation{PortabilityImportContext: incidentbundles.PortabilityImportContext{OperationID: uuid.NewString(), IncidentID: request.IncidentID, ActorUserID: uuid.New(), Actors: catalog}, ProfileID: ProfileID, ContractMajor: 10, ClaimState: incidentbundles.ClaimStateClaimed, StateVersion: 7, ContentSchemaID: portabilityContentSchema, CatalogSHA256: strings.Repeat("b", 64), Members: members, Dependencies: dependencies, PreparationMembers: portableFixtureMembers{MapBundle: portabilitysupport.MapBundle{}}}
	return invocation, members
}

func TestNetworkFlowPortableMemberAdmission(t *testing.T) {
	t.Run("prepared descriptor binds closed owner members", func(t *testing.T) {
		invocation, _ := portableAdmissionFixture(t)
		result, err := (&portabilityParticipant{limits: defaultLimits()}).PrepareImport(t.Context(), invocation, nil)
		if err != nil {
			t.Fatal(err)
		}
		decode := func() map[string]any {
			var value map[string]any
			if err := json.Unmarshal(result.ParticipantInput, &value); err != nil {
				t.Fatal(err)
			}
			return value
		}
		if !validatePortabilityTransactionInput(decode()) {
			t.Fatal("valid prepared input rejected")
		}
		for name, change := range map[string]func(map[string]any){
			"foreign member": func(v map[string]any) {
				v["members"].([]any)[0].(map[string]any)["path"] = "ext/extensions/other/rows.ndjson"
			},
			"traversal": func(v map[string]any) {
				v["members"].([]any)[0].(map[string]any)["path"] = portabilityMemberPrefix + "../rows.ndjson"
			},
			"nested unknown field": func(v map[string]any) { v["members"].([]any)[0].(map[string]any)["extra"] = true },
			"omitted size":         func(v map[string]any) { delete(v["members"].([]any)[0].(map[string]any), "byte_size") },
			"duplicate member":     func(v map[string]any) { members := v["members"].([]any); members[1] = members[0] },
			"false count":          func(v map[string]any) { v["counts"].(map[string]any)["observations"] = float64(0) },
			"wrong empty hash":     func(v map[string]any) { v["members"].([]any)[0].(map[string]any)["sha256"] = strings.Repeat("f", 64) },
			"unbound graph": func(v map[string]any) {
				v["prepared_graphs"] = []any{map[string]any{"path": "prepared/network_flow_activity/graphs/nfgv_" + strings.Repeat("a", 32) + ".json", "byte_size": float64(2), "sha256": strings.Repeat("a", 64)}}
			},
		} {
			t.Run(name, func(t *testing.T) {
				v := decode()
				change(v)
				if validatePortabilityTransactionInput(v) {
					t.Fatal("altered prepared descriptor accepted")
				}
			})
		}
	})
	t.Run("retained normalization ignores changed destination defaults", func(t *testing.T) {
		invocation, _ := portableAdmissionFixture(t)
		limits := defaultLimits()
		limits.MaxRawCellScalarLength = 1
		if _, err := prepareNetworkFlow(t.Context(), invocation, limits); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("exact retained observations and immutable member reopening", func(t *testing.T) {
		invocation, members := portableAdmissionFixture(t)
		prepared, err := prepareNetworkFlow(t.Context(), invocation, defaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		if prepared.counts["observations"] != 2 || prepared.counts["tables"] != 1 {
			t.Fatal("retained multiplicity changed")
		}
		members.MapBundle[portabilityMemberPrefix+"observations.ndjson"] = []byte("{}\n")
		if err := prepared.IterateRowsForTables(t.Context(), invocation.IncidentID, []string{"nft_" + strings.Repeat("a", 32)}, func(flowRow) error { return nil }); err == nil {
			t.Fatal("changed sealed input accepted")
		}
	})
	for _, name := range []string{"extra member", "missing member", "missing field", "extra field", "wrong incident", "unknown actor", "wrong source", "changed bytes counter", "changed row identity", "duplicate observation", "missing observation", "unsupported version", "altered mapping", "altered interpretation", "altered normalization limit", "nested unknown member", "changed committed bytes", "changed committed locator"} {
		t.Run(name, func(t *testing.T) {
			invocation, members := portableAdmissionFixture(t)
			change := func(path string, edit func(map[string]any)) {
				rows, err := incidentportability.DecodeStrictNDJSONObjects(members.MapBundle[portabilityMemberPrefix+path], path)
				if err != nil {
					t.Fatal(err)
				}
				edit(rows[0])
				var out bytes.Buffer
				for _, row := range rows {
					if err := json.NewEncoder(&out).Encode(row); err != nil {
						t.Fatal(err)
					}
				}
				members.MapBundle[portabilityMemberPrefix+path] = out.Bytes()
			}
			switch name {
			case "extra member":
				members.MapBundle[portabilityMemberPrefix+"unexpected.ndjson"] = nil
			case "missing member":
				delete(members.MapBundle, portabilityMemberPrefix+"diagnostics.ndjson")
			case "missing field":
				change("tables.ndjson", func(v map[string]any) { delete(v, "diagnostics_truncated") })
			case "extra field":
				change("tables.ndjson", func(v map[string]any) { v["source_filename_digest"] = "deployment-keyed" })
			case "wrong incident":
				change("tables.ndjson", func(v map[string]any) { v["incident_id"] = uuid.NewString() })
			case "unknown actor":
				change("observations.ndjson", func(v map[string]any) { v["created_by_user_id"] = uuid.NewString() })
			case "wrong source":
				change("tables.ndjson", func(v map[string]any) { v["committed_source_id"] = uuid.NewString() })
			case "changed bytes counter":
				change("observations.ndjson", func(v map[string]any) { v["bytes_count"] = "999999" })
			case "changed row identity":
				change("observations.ndjson", func(v map[string]any) { v["network_flow_row_id"] = "nfr_" + strings.Repeat("e", 64) })
			case "duplicate observation":
				raw := members.MapBundle[portabilityMemberPrefix+"observations.ndjson"]
				members.MapBundle[portabilityMemberPrefix+"observations.ndjson"] = append(append([]byte(nil), raw...), raw...)
			case "missing observation":
				members.MapBundle[portabilityMemberPrefix+"observations.ndjson"] = nil
			case "unsupported version":
				invocation.StateVersion = 6
			case "altered mapping":
				change("tables.ndjson", func(v map[string]any) { v["mapping_fingerprint"] = strings.Repeat("e", 64) })
			case "altered interpretation":
				change("tables.ndjson", func(v map[string]any) { v["source_interpretation_sha256"] = strings.Repeat("e", 64) })
			case "altered normalization limit":
				change("tables.ndjson", func(v map[string]any) { v["source_raw_cell_scalar_limit"] = 1 })
			case "nested unknown member":
				change("observations.ndjson", func(v map[string]any) { v["direction"].(map[string]any)["unsupported"] = true })
			case "changed committed bytes", "changed committed locator":
				dependencies := invocation.Dependencies.(*portabilityDependencies)
				sources := dependencies.sources.(portableFixtureSources)
				if name == "changed committed bytes" {
					sources.bytes = append(append([]byte(nil), sources.bytes...), '\n')
				} else {
					sources.source.DataStartRowRef = 3
				}
				dependencies.sources = sources
			}
			if _, err := prepareNetworkFlow(t.Context(), invocation, defaultLimits()); err == nil {
				t.Fatal("invalid portability input admitted")
			}
		})
	}
}

func TestNetworkFlowPortableSelectedGraphHistory(t *testing.T) {
	invocation, members := portableAdmissionFixture(t)
	prepared, err := prepareNetworkFlow(t.Context(), invocation, defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	tableID := "nft_" + strings.Repeat("a", 32)
	table := prepared.tables[tableID]
	semanticJSON := mappingTestJSON(t, map[string]any{"schema_id": schemaGraphSemanticQuery, "selected_table_ids": []string{tableID}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}})
	semantic, failure := decodeGraphSemanticRequest(semanticJSON, defaultLimits())
	if failure != nil {
		t.Fatal(failure)
	}
	composer, err := newGraphSourceComposer(prepared, defaultLimits(), newGraphProjectionAdapter(), nil)
	if err != nil {
		t.Fatal(err)
	}
	composition, failure := composer.composeGraphSourceFromSemantic(t.Context(), invocation.IncidentID, semantic)
	if failure != nil {
		t.Fatal(failure)
	}
	t.Run("bounded examples do not depend on input order", func(t *testing.T) {
		var rows []flowRow
		if err := prepared.IterateRowsForTables(t.Context(), invocation.IncidentID, []string{tableID}, func(row flowRow) error { rows = append(rows, row); return nil }); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 2 {
			t.Fatal("fixture needs two observations")
		}
		compose := func(order []flowRow) graphComposition {
			value := composition
			value.Vertices, value.Edges = map[string]*graphVertex{}, map[string]*graphEdge{}
			value.ContributingRows = 0
			value.ResultLimits.MaxExampleRowRefsPerEdge = 1
			for _, row := range order {
				if failure := composeGraphRow(invocation.IncidentID, row, map[string]tableRecord{tableID: table}, &value); failure != nil {
					t.Fatal(failure)
				}
			}
			return value
		}
		forward := compose(rows)
		reverse := compose([]flowRow{rows[1], rows[0]})
		if len(forward.Edges) != 1 || len(reverse.Edges) != 1 {
			t.Fatalf("two matching observations must yield one edge: forward=%d reverse=%d", len(forward.Edges), len(reverse.Edges))
		}
		if !canonicalEqual(networkFlowProjectionInput("nfsnap_test", forward), networkFlowProjectionInput("nfsnap_test", reverse)) {
			t.Fatal("transport order changed selected examples")
		}
		for _, edge := range reverse.Edges {
			if len(edge.ExampleRows) != 1 || edge.ExampleRows[0].RowID != rows[0].RowID {
				t.Fatal("earliest canonical example was lost")
			}
		}
	})
	snapshot := graphSourceSnapshotDigest(invocation.IncidentID, composition.SourceTables, composition.Digest)
	graphID := "nfgv_" + strings.Repeat("d", 32)
	projection, err := graphprojection.ProjectV2(t.Context(), graphprojection.InvocationContextV2{GraphViewID: graphID, SourceOwnerID: ProfileID}, mappingTestJSON(t, networkFlowProjectionInput(snapshot, composition)))
	if err != nil {
		t.Fatal(err)
	}
	selected := graphViewSelectedResultFromBinding(projection.ResultBindingV2())
	binding, err := bindSelectedSource(composition, selected)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicalGraphSemanticJSON(semantic.Raw)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := encodeRefreshIntent("all_active_tables", canonical)
	if err != nil {
		t.Fatal(err)
	}
	retiredAt := table.CreatedAt.Add(time.Hour)
	declaration := graphViewDeclaration{GraphViewID: graphID, IncidentID: invocation.IncidentID, DisplayName: "Historical graph", NormalizedDisplayName: "historical graph", DeclarationState: graphViewDeclarationStateRetired, RefreshIntentJSON: intent, SemanticQueryJSON: canonical, SemanticQuerySHA256: graphViewSemanticQuerySHA256(canonical), DesiredSourceSnapshotID: snapshot, SelectedSourceBindingJSON: binding, SelectedResult: &selected, GraphViewVersion: 3, MaterializationGeneration: 2, CreatedByUserID: table.CreatedByUserID, CreatedAt: table.CreatedAt, UpdatedAt: retiredAt, RetiredAt: &retiredAt}
	t.Run("duplicate graph names preserve distinct declarations", func(t *testing.T) {
		duplicateInvocation, duplicateMembers := portableAdmissionFixture(t)
		var rows []byte
		for _, id := range []string{graphID, "nfgv_" + strings.Repeat("e", 32)} {
			unmaterialized := declaration
			unmaterialized.GraphViewID = id
			unmaterialized.DeclarationState = graphViewDeclarationStateActive
			unmaterialized.RetiredAt = nil
			unmaterialized.SelectedResult = nil
			unmaterialized.SelectedSourceBindingJSON = nil
			rows = append(rows, mappingTestJSON(t, portableGraphViewDeclaration(unmaterialized))...)
			rows = append(rows, '\n')
		}
		duplicateMembers.MapBundle[portabilityMemberPrefix+"graph_views.ndjson"] = rows
		got, err := prepareNetworkFlow(t.Context(), duplicateInvocation, defaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		if len(got.graphs) != 2 || got.graphs[0].declaration.GraphViewID == got.graphs[1].declaration.GraphViewID || got.graphs[0].declaration.DisplayName != got.graphs[1].declaration.DisplayName {
			t.Fatal("duplicate display names merged or changed graph declarations")
		}
		for _, graph := range got.graphs {
			if graph.declaration.SelectedResult != nil || graph.declaration.LatestJobID != nil {
				t.Fatal("unmaterialized declaration acquired a result or job")
			}
		}
		limits := defaultLimits()
		limits.MaxActiveGraphViewsPerIncident = 1
		if _, err := prepareNetworkFlow(t.Context(), duplicateInvocation, limits); err == nil {
			t.Fatal("duplicate-name admission bypassed active graph quota")
		}
	})
	table.DisplayName = "Renamed after selection"
	table.TableStatus = "soft_deleted"
	table.TableVersion = 3
	table.UpdatedAt = retiredAt
	table.DeletedAt = &retiredAt
	members.MapBundle[portabilityMemberPrefix+"tables.ndjson"] = append(mappingTestJSON(t, portableTableRecord(table)), '\n')
	members.MapBundle[portabilityMemberPrefix+"graph_views.ndjson"] = append(mappingTestJSON(t, portableGraphViewDeclaration(declaration)), '\n')
	reconstructed, err := prepareNetworkFlow(t.Context(), invocation, defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if len(reconstructed.graphs) != 1 || reconstructed.graphs[0].declaration.SelectedResult == nil || *reconstructed.graphs[0].declaration.SelectedResult != selected || reconstructed.graphs[0].members == nil {
		t.Fatal("selected historical graph was not prepared exactly")
	}
	declaration.SelectedResult.CanonicalOutputSHA256 = strings.Repeat("e", 64)
	members.MapBundle[portabilityMemberPrefix+"graph_views.ndjson"] = append(mappingTestJSON(t, portableGraphViewDeclaration(declaration)), '\n')
	if _, err := prepareNetworkFlow(t.Context(), invocation, defaultLimits()); err == nil {
		t.Fatal("altered graph binding admitted")
	}
}
