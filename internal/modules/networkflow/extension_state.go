package networkflow

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

const (
	extensionFamilyIndicatorBindings      = "network_flow_activity.indicator_bindings"
	extensionFamilyRejectedRowDiagnostics = "network_flow_activity.rejected_row_diagnostics"
	extensionFamilyRows                   = "network_flow_activity.rows"
	extensionFamilyTables                 = "network_flow_activity.tables"
	extensionFamilyGraphViews             = "network_flow_activity.graph_views"
)

var networkFlowExtensionFamilies = []string{
	extensionFamilyGraphViews,
	extensionFamilyIndicatorBindings,
	extensionFamilyRejectedRowDiagnostics,
	extensionFamilyRows,
	extensionFamilyTables,
}

// ExtensionStateReader is the Network Flow owner's read-only logical view. The
// caller may back it with PostgreSQL, but no platform transaction type crosses
// this semantic validation boundary.
type ExtensionStateReader interface {
	FamilyCounts(context.Context, []string) (map[string]int64, error)
	ValidateFamilyState(context.Context, string) error
}

// ExtensionStateFamilyCounters is the Network Flow owner's physical adapter for
// the generated logical-family identities. The generic Extensions coordinator
// never receives table names or unrestricted SQL.
func ExtensionStateFamilyCounters() []extensionstore.FamilyCounter {
	graphViews := countExtensionFamily(extensionFamilyGraphViews, `SELECT COUNT(*) FROM network_flow_graph_views`)
	graphViews.Validate = validatePersistedGraphViewFamily
	bindings := countExtensionFamily(extensionFamilyIndicatorBindings, `SELECT COUNT(*) FROM network_flow_indicator_bindings`)
	bindings.Validate = validatePersistedIndicatorLinkFamily
	diagnostics := countExtensionFamily(extensionFamilyRejectedRowDiagnostics, `SELECT COUNT(*) FROM network_flow_rejected_row_diagnostics`)
	diagnostics.Validate = validatePersistedDiagnosticFamily
	return []extensionstore.FamilyCounter{
		bindings,
		diagnostics,
		countExtensionFamily(extensionFamilyRows, `SELECT COUNT(*) FROM network_flow_rows`),
		countExtensionFamily(extensionFamilyTables, `SELECT COUNT(*) FROM network_flow_tables`),
		graphViews,
	}
}

func countExtensionFamily(familyID, query string) extensionstore.FamilyCounter {
	return extensionstore.FamilyCounter{
		FamilyID: familyID,
		Count: func(ctx context.Context, querier extensionstore.Querier) (int64, error) {
			var count int64
			if err := querier.QueryRow(ctx, query).Scan(&count); err != nil {
				return 0, err
			}
			return count, nil
		},
	}
}

func validatePersistedGraphViewFamily(ctx context.Context, querier extensionstore.Querier) error {
	rows, err := querier.Query(ctx, graphViewDeclarationSelect+` ORDER BY graph_view_id ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	limits := defaultEffectiveLimits()
	limits.MaxSelectedTablesPerQuery = 64
	limits.MaxGraphVertices = 100000
	limits.MaxGraphEdges = 250000
	limits.MaxExampleRowRefsPerEdge = 100
	limits.MaxContributingRowsPerGraph = 5000000
	limits.MaxTimeBucketsPerGraph = 1024
	for rows.Next() {
		declaration, err := scanGraphViewDeclaration(rows)
		if err != nil {
			return err
		}
		if !validGraphViewDeclaration(declaration) {
			return errSavedGraphCutoverIncompatible
		}
		graphViewID, semanticQuery, semanticDigest := declaration.GraphViewID, declaration.SemanticQueryJSON, declaration.SemanticQuerySHA256
		selectedProjectionVersion := ""
		if declaration.SelectedResult != nil {
			selectedProjectionVersion = declaration.SelectedResult.ProjectionVersion
		}

		semantic, apiErr := decodeGraphSemanticRequest(semanticQuery, limits)
		if apiErr != nil {
			return fmt.Errorf("saved graph %s has an unsupported semantic query", graphViewID)
		}
		canonical := canonicalJSON(semantic.Raw)
		if graphViewSemanticQuerySHA256(canonical) != semanticDigest {
			return fmt.Errorf("saved graph %s semantic query digest mismatch", graphViewID)
		}
		wantProjectionVersion := "network_flow_activity.v1"
		if semantic.SchemaID == schemaGraphSemanticQueryV2 && semantic.Aggregation.Mode == "time_bucket_v1" {
			wantProjectionVersion = "network_flow_activity.time_bucket.v1"
		}
		if selectedProjectionVersion != "" && selectedProjectionVersion != wantProjectionVersion {
			return fmt.Errorf("saved graph %s selected projection version does not match its semantic query", graphViewID)
		}
	}
	return rows.Err()
}

// ValidateExtensionState is the digest-bound Network Flow final validator. The
// database owns structural constraints; this owner-level check closes family
// presence relationships without exposing physical storage to Extensions.
func ValidateExtensionState(ctx context.Context, reader ExtensionStateReader) error {
	if reader == nil {
		return errors.New("network flow state reader unavailable")
	}
	families := append([]string(nil), networkFlowExtensionFamilies...)
	sort.Strings(families)
	counts, err := reader.FamilyCounts(ctx, families)
	if err != nil {
		return err
	}
	if len(counts) != len(families) {
		return errors.New("network flow state family set incomplete")
	}
	if counts[extensionFamilyTables] == 0 &&
		(counts[extensionFamilyRows] != 0 ||
			counts[extensionFamilyRejectedRowDiagnostics] != 0 ||
			counts[extensionFamilyIndicatorBindings] != 0 ||
			counts[extensionFamilyGraphViews] != 0) {
		return fmt.Errorf("network flow dependent state exists without table state")
	}
	if err := reader.ValidateFamilyState(ctx, extensionFamilyGraphViews); err != nil {
		return err
	}
	if err := reader.ValidateFamilyState(ctx, extensionFamilyIndicatorBindings); err != nil {
		return err
	}
	return reader.ValidateFamilyState(ctx, extensionFamilyRejectedRowDiagnostics)
}

// Retained diagnostics use the same semantic boundary as preview publication.
// Iterate rows without accumulating source material or rewriting stored facts.
func validatePersistedDiagnosticFamily(ctx context.Context, querier extensionstore.Querier) error {
	rows, err := querier.Query(ctx, `SELECT
 t.source_content_sha256, t.mapping_fingerprint,
 d.diagnostic_id, d.source_row_number, d.source_column_ordinal,
 d.raw_header_sha256, d.field_key, d.error_code, d.reason_code,
 d.safe_sample, d.raw_value_sha256, d.message_key, d.message_args,
 d.message, d.limit_name, d.limit_value, d.actual_value,
 d.incident_id = t.incident_id
 FROM network_flow_rejected_row_diagnostics d
 JOIN network_flow_tables t ON t.network_flow_table_id = d.network_flow_table_id
 ORDER BY d.network_flow_table_id, d.source_row_number,
 d.source_column_ordinal ASC NULLS LAST, d.field_key ASC NULLS LAST, d.error_code`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var source, fingerprint string
		var d rejectedRowDiagnostic
		var scoped bool
		if err := rows.Scan(&source, &fingerprint, &d.DiagnosticID, &d.SourceRowNumber, &d.SourceColumnOrdinal, &d.RawHeaderSHA256, &d.FieldKey, &d.ErrorCode, &d.ReasonCode, &d.SafeSample, &d.RawValueSHA256, &d.MessageKey, &d.MessageArgs, &d.Message, &d.LimitName, &d.LimitValue, &d.ActualValue, &scoped); err != nil {
			return err
		}
		if !scoped || validateDiagnostic(d, source, fingerprint) != nil {
			return errors.New("network flow retained diagnostics are incompatible")
		}
	}
	return rows.Err()
}

// ValidateRestoredExtensionState rejects legacy metadata before Recovery commits
// restored owner facts. Missing metadata is admitted only for empty owner state.
func ValidateRestoredExtensionState(ctx context.Context, reader extensionstore.Querier) error {
	var version int
	var lineage string
	err := reader.QueryRow(ctx, `SELECT state_version,migration_lineage_id FROM extension_state_metadata WHERE profile_id = 'network_flow_activity'`).Scan(&version, &lineage)
	if errors.Is(err, pgx.ErrNoRows) {
		for _, family := range ExtensionStateFamilyCounters() {
			count, err := family.Count(ctx, reader)
			if err != nil {
				return err
			}
			if count != 0 {
				return errors.New("network flow restored state has no compatible metadata")
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	if version != 5 || lineage != "network_flow_activity.state_v1" {
		return errors.New("network flow restored state version is incompatible")
	}
	return ValidateRetainedExtensionState(ctx, reader)
}
