package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

const selectedSourceSchema = "cartulary.network_flow.selected_source_binding.v1"
const selectedSourceAlgorithm = "network_flow_activity.selected_source.v1"

// This is the complete table metadata consumed by Graph computation. Catalog
// names, deletion state, actors and deployment-keyed digests are not inputs.
type selectedSourceTable struct {
	TableID              string    `json:"network_flow_table_id"`
	TableVersion         int64     `json:"table_version"`
	CommittedSourceID    uuid.UUID `json:"committed_source_id"`
	ContentSHA256        string    `json:"source_content_sha256"`
	InterpretationSHA256 string    `json:"source_interpretation_sha256"`
	MappingFingerprint   string    `json:"mapping_fingerprint"`
	SourceProfileID      string    `json:"source_profile_id"`
	ParserProfileID      string    `json:"parser_profile_id"`
	Accepted             int64     `json:"row_count_accepted"`
	Rejected             int64     `json:"row_count_rejected"`
	Fingerprint          string    `json:"content_fingerprint"`
}

type selectedSourceBinding struct {
	SchemaID      string                         `json:"schema_id"`
	AlgorithmID   string                         `json:"algorithm_id"`
	SemanticQuery json.RawMessage                `json:"semantic_query"`
	Limits        graphResultLimits              `json:"computation_limits"`
	Tables        []selectedSourceTable          `json:"tables"`
	Result        graphViewSelectedResultBinding `json:"expected_result_binding"`
}

func selectedTable(table tableRecord) selectedSourceTable {
	return selectedSourceTable{TableID: table.TableID, TableVersion: table.TableVersion,
		CommittedSourceID: table.CommittedSourceID, ContentSHA256: table.SourceContentSHA256,
		InterpretationSHA256: table.SourceInterpretationSHA256, MappingFingerprint: table.MappingFingerprint,
		SourceProfileID: table.SourceProfileID, ParserProfileID: table.ParserProfileID,
		Accepted: table.RowCountAccepted, Rejected: table.RowCountRejected, Fingerprint: portableTableFingerprint(table)}
}

func portableTableFingerprint(table tableRecord) string {
	// Fixed struct field order is part of selected_source.v1. No keyed value or
	// mutable table version participates in portable content identity.
	raw, _ := json.Marshal(struct {
		Algorithm      string    `json:"algorithm"`
		Source         uuid.UUID `json:"committed_source_id"`
		Content        string    `json:"content_sha256"`
		Interpretation string    `json:"interpretation_sha256"`
		Mapping        string    `json:"mapping_fingerprint"`
		Profile        string    `json:"source_profile_id"`
		Parser         string    `json:"parser_profile_id"`
		Accepted       int64     `json:"accepted"`
		Rejected       int64     `json:"rejected"`
	}{selectedSourceAlgorithm, table.CommittedSourceID, table.SourceContentSHA256, table.SourceInterpretationSHA256,
		table.MappingFingerprint, table.SourceProfileID, table.ParserProfileID, table.RowCountAccepted, table.RowCountRejected})
	return sha256Hex(raw)
}

func bindSelectedSource(composition graphComposition, selected graphViewSelectedResultBinding) (json.RawMessage, error) {
	semantic, err := canonicalGraphSemanticJSON(composition.SemanticQuery)
	if err != nil {
		return nil, err
	}
	binding := selectedSourceBinding{SchemaID: selectedSourceSchema, AlgorithmID: selectedSourceAlgorithm,
		SemanticQuery: semantic, Limits: composition.ResultLimits, Tables: make([]selectedSourceTable, 0, len(composition.SourceTables)), Result: selected}
	for _, table := range composition.SourceTables {
		binding.Tables = append(binding.Tables, selectedTable(table))
	}
	return json.Marshal(binding)
}

func decodeSelectedSource(raw json.RawMessage, selected *graphViewSelectedResultBinding, limits EffectiveLimits) (selectedSourceBinding, graphSemanticRequest, error) {
	var binding selectedSourceBinding
	object, err := decodeClosedImportObject(raw, "schema_id algorithm_id semantic_query computation_limits tables expected_result_binding", "", "", &binding)
	if err != nil ||
		binding.SchemaID != selectedSourceSchema || binding.AlgorithmID != selectedSourceAlgorithm || selected == nil ||
		binding.Result != *selected || !validSelectedGraphViewResult(binding.Result) {
		return binding, graphSemanticRequest{}, errGraphViewDeclarationInvalid
	}
	if _, err := decodeClosedImportObject(object["expected_result_binding"], "projection_result_id source_snapshot_id projection_schema_id projection_version normalized_configuration_sha256 normalized_source_sha256 canonical_output_sha256", "", "", &binding.Result); err != nil {
		return binding, graphSemanticRequest{}, errGraphViewDeclarationInvalid
	}
	if _, err := decodeClosedImportObject(object["computation_limits"], "max_vertices max_edges max_example_row_refs_per_edge max_aggregate_counter_digits max_contributing_rows max_time_buckets", "", "", &binding.Limits); err != nil || !validSelectedComputationLimits(binding.Limits) {
		return binding, graphSemanticRequest{}, errGraphViewDeclarationInvalid
	}
	semantic, failure := decodeGraphSemanticRequest(binding.SemanticQuery, limits)
	if failure != nil || len(binding.Tables) == 0 || len(binding.Tables) != len(semantic.SelectedTableIDs) {
		return binding, semantic, errGraphViewDeclarationInvalid
	}
	semantic.ResultLimits = binding.Limits
	var tableObjects []json.RawMessage
	if err := json.Unmarshal(object["tables"], &tableObjects); err != nil {
		return binding, semantic, errGraphViewDeclarationInvalid
	}
	ids := make([]string, 0, len(binding.Tables))
	seen := map[string]bool{}
	for i, table := range binding.Tables {
		if _, err := decodeClosedImportObject(tableObjects[i], "network_flow_table_id table_version committed_source_id source_content_sha256 source_interpretation_sha256 mapping_fingerprint source_profile_id parser_profile_id row_count_accepted row_count_rejected content_fingerprint", "", "", &table); err != nil {
			return binding, semantic, errGraphViewDeclarationInvalid
		}
		if seen[table.TableID] || !linkTableIDPattern.MatchString(table.TableID) || table.TableVersion < 1 ||
			table.CommittedSourceID == uuid.Nil || !validLowerSHA256(table.ContentSHA256) ||
			!validLowerSHA256(table.InterpretationSHA256) || !validLowerSHA256(table.MappingFingerprint) ||
			!validLowerSHA256(table.Fingerprint) || table.Accepted < 1 || table.Rejected < 0 {
			return binding, semantic, errGraphViewDeclarationInvalid
		}
		profile, ok := mappingSourceProfile(table.SourceProfileID)
		if !ok || profile.ParserProfileID != table.ParserProfileID {
			return binding, semantic, errGraphViewDeclarationInvalid
		}
		if portableTableFingerprint(tableRecord{CommittedSourceID: table.CommittedSourceID, SourceContentSHA256: table.ContentSHA256,
			SourceInterpretationSHA256: table.InterpretationSHA256, MappingFingerprint: table.MappingFingerprint,
			SourceProfileID: table.SourceProfileID, ParserProfileID: table.ParserProfileID, RowCountAccepted: table.Accepted, RowCountRejected: table.Rejected}) != table.Fingerprint {
			return binding, semantic, errGraphViewDeclarationInvalid
		}
		seen[table.TableID] = true
		ids = append(ids, table.TableID)
	}
	// Tables retain computation/workspace order; semantic query identity sorts
	// its selection as a set. Validate membership without rewriting that order.
	slices.Sort(ids)
	if !slices.Equal(ids, semantic.SelectedTableIDs) {
		return binding, semantic, errGraphViewDeclarationInvalid
	}
	return binding, semantic, nil
}

// Retained input admission uses the contract ceilings rather than current query
// defaults. A later configuration change must not rewrite selected computation.
func retainedSemanticLimits() EffectiveLimits {
	limits := defaultEffectiveLimits()
	limits.MaxSelectedTablesPerQuery = 64
	limits.MaxFiltersPerQuery = 16
	limits.MaxRawCellScalarLength = 16384
	return limits
}

func validSelectedComputationLimits(l graphResultLimits) bool {
	return l.MaxVertices >= 1 && l.MaxVertices <= 100000 && l.MaxEdges >= 0 && l.MaxEdges <= 250000 &&
		l.MaxExampleRowRefsPerEdge >= 0 && l.MaxExampleRowRefsPerEdge <= 100 &&
		l.MaxAggregateCounterDigits >= 1 && l.MaxAggregateCounterDigits <= 128 &&
		l.MaxContributingRows >= 1 && l.MaxContributingRows <= 5000000 && l.MaxTimeBuckets >= 1 && l.MaxTimeBuckets <= 1024
}

type historicalGraphSourceReader interface {
	graphCompositionSourceReader
	GetRetainedTable(context.Context, uuid.UUID, string) (tableRecord, error)
}

// Only an admitted selected binding grants access to retained historical tables.
// The ordinary active-table reader remains unchanged.
type selectedGraphSourceReader struct {
	graphCompositionSourceReader
	incidentID uuid.UUID
	tables     []tableRecord
}

func (s *selectedGraphSourceReader) ListActiveTables(_ context.Context, incidentID uuid.UUID) ([]tableRecord, error) {
	if incidentID != s.incidentID {
		return nil, errTableNotFound
	}
	return append([]tableRecord(nil), s.tables...), nil
}
func (s *selectedGraphSourceReader) GetActiveTable(_ context.Context, incidentID uuid.UUID, id string) (tableRecord, error) {
	if incidentID == s.incidentID {
		for _, table := range s.tables {
			if table.TableID == id {
				return table, nil
			}
		}
	}
	return tableRecord{}, errTableNotFound
}

func (s *graphSourceComposer) forSelectedSource(ctx context.Context, declaration graphViewDeclaration) (*graphSourceComposer, graphSemanticRequest, error) {
	binding, semantic, err := decodeSelectedSource(declaration.SelectedSourceBindingJSON, declaration.SelectedResult, retainedSemanticLimits())
	if err != nil {
		return nil, semantic, err
	}
	source, ok := s.store.(historicalGraphSourceReader)
	if !ok {
		return nil, semantic, fmt.Errorf("historical graph source capability unavailable")
	}
	tables := make([]tableRecord, 0, len(binding.Tables))
	for _, retained := range binding.Tables {
		table, err := source.GetRetainedTable(ctx, declaration.IncidentID, retained.TableID)
		if err != nil {
			return nil, semantic, err
		}
		current := selectedTable(table)
		current.TableVersion = retained.TableVersion
		if current != retained || retained.TableVersion > table.TableVersion {
			return nil, semantic, errGraphViewDeclarationInvalid
		}
		table.TableVersion = retained.TableVersion
		tables = append(tables, table)
	}
	result := *s
	result.limits = retainedSemanticLimits()
	result.store = &selectedGraphSourceReader{graphCompositionSourceReader: source, incidentID: declaration.IncidentID, tables: tables}
	return &result, semantic, nil
}

func (s *graphSourceComposer) composeSelectedSource(ctx context.Context, declaration graphViewDeclaration) (graphComposition, graphSemanticRequest, error) {
	historical, semantic, err := s.forSelectedSource(ctx, declaration)
	if err != nil {
		return graphComposition{}, semantic, err
	}
	composition, failure := historical.composeGraphSourceFromSemantic(ctx, declaration.IncidentID, semantic)
	if failure != nil {
		return graphComposition{}, semantic, failure
	}
	if graphSourceSnapshotDigest(declaration.IncidentID, composition.SourceTables, composition.Digest) != declaration.SelectedResult.SourceSnapshotID {
		return graphComposition{}, semantic, errGraphViewDeclarationInvalid
	}
	return composition, semantic, nil
}

func (s *selectedGraphSourceReader) IterateGraphContributorRows(ctx context.Context, incidentID uuid.UUID, tableIDs []string, predicate graphContributorPredicate, visit func(flowRow) error) error {
	if s == nil || incidentID != s.incidentID {
		return errTableNotFound
	}
	for _, id := range tableIDs {
		if _, err := s.GetActiveTable(ctx, incidentID, id); err != nil {
			return err
		}
	}
	source, ok := s.graphCompositionSourceReader.(graphContributorSourceReader)
	if !ok {
		return fmt.Errorf("historical contributor capability unavailable")
	}
	return source.IterateGraphContributorRows(ctx, incidentID, tableIDs, predicate, visit)
}
