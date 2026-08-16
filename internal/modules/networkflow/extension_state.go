package networkflow

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

const (
	ExtensionFamilyIndicatorBindings      = "network_flow_activity.indicator_bindings"
	ExtensionFamilyRejectedRowDiagnostics = "network_flow_activity.rejected_row_diagnostics"
	ExtensionFamilyRows                   = "network_flow_activity.rows"
	ExtensionFamilyTables                 = "network_flow_activity.tables"
	ExtensionFamilyGraphViews             = "network_flow_activity.graph_views"
)

var networkFlowExtensionFamilies = []string{
	ExtensionFamilyGraphViews,
	ExtensionFamilyIndicatorBindings,
	ExtensionFamilyRejectedRowDiagnostics,
	ExtensionFamilyRows,
	ExtensionFamilyTables,
}

// ExtensionStateReader is the Network Flow owner's read-only logical view. The
// caller may back it with PostgreSQL, but no platform transaction type crosses
// this semantic validation boundary.
type ExtensionStateReader interface {
	FamilyCounts(context.Context, []string) (map[string]int64, error)
}

// ExtensionStateFamilyCounters is the Network Flow owner's physical adapter for
// the generated logical-family identities. The generic Extensions coordinator
// never receives table names or unrestricted SQL.
func ExtensionStateFamilyCounters() []extensionstore.FamilyCounter {
	return []extensionstore.FamilyCounter{
		countExtensionFamily(ExtensionFamilyIndicatorBindings, `SELECT COUNT(*) FROM network_flow_indicator_bindings`),
		countExtensionFamily(ExtensionFamilyRejectedRowDiagnostics, `SELECT COUNT(*) FROM network_flow_rejected_row_diagnostics`),
		countExtensionFamily(ExtensionFamilyRows, `SELECT COUNT(*) FROM network_flow_rows`),
		countExtensionFamily(ExtensionFamilyTables, `SELECT COUNT(*) FROM network_flow_tables`),
		countExtensionFamily(ExtensionFamilyGraphViews, `SELECT COUNT(*) FROM network_flow_graph_views`),
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
	if counts[ExtensionFamilyTables] == 0 &&
		(counts[ExtensionFamilyRows] != 0 ||
			counts[ExtensionFamilyRejectedRowDiagnostics] != 0 ||
			counts[ExtensionFamilyIndicatorBindings] != 0 ||
			counts[ExtensionFamilyGraphViews] != 0) {
		return fmt.Errorf("network flow dependent state exists without table state")
	}
	return nil
}
