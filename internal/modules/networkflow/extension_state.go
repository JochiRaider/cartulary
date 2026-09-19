package networkflow

import (
	"context"
	"errors"
	"fmt"
	"sort"

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
	return []extensionstore.FamilyCounter{
		bindings,
		countExtensionFamily(extensionFamilyRejectedRowDiagnostics, `SELECT COUNT(*) FROM network_flow_rejected_row_diagnostics`),
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

		semantic, apiErr := decodeGraphSemanticRequestHTTP(semanticQuery, limits)
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
	return reader.ValidateFamilyState(ctx, extensionFamilyIndicatorBindings)
}
