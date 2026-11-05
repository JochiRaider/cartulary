package networkflow

import (
	"context"
	"database/sql"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// ValidateMigrationAdmission runs before the migration provider can touch an
// older database. Historical observations remain in their matching environment.
func ValidateMigrationAdmission(ctx context.Context, conn *sql.Conn) error {
	var current bool
	if err := conn.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='network_flow_rows' AND column_name='direction_evidence')`).Scan(&current); err != nil {
		return err
	}
	if current {
		return nil
	}
	blocked := orientationMigrationBlocked{}
	for _, probe := range []struct{ table, query string }{
		{"network_flow_tables", "SELECT EXISTS (SELECT 1 FROM public.network_flow_tables)"},
		{"network_flow_rows", "SELECT EXISTS (SELECT 1 FROM public.network_flow_rows)"},
		{"network_flow_graph_views", "SELECT EXISTS (SELECT 1 FROM public.network_flow_graph_views)"},
		{"network_flow_indicator_bindings", "SELECT EXISTS (SELECT 1 FROM public.network_flow_indicator_bindings)"},
		{"network_flow_rejected_row_diagnostics", "SELECT EXISTS (SELECT 1 FROM public.network_flow_rejected_row_diagnostics)"},
		{"extension_state_metadata", "SELECT EXISTS (SELECT 1 FROM public.extension_state_metadata WHERE profile_id='network_flow_activity')"},
	} {
		var present bool
		if err := conn.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+probe.table).Scan(&present); err != nil {
			return err
		}
		if !present {
			continue
		}
		var retained bool
		if err := conn.QueryRowContext(ctx, probe.query).Scan(&retained); err != nil {
			return err
		}
		if retained {
			return blocked
		}
	}
	retained, err := imports.RetainedAnalyticalStatePresent(ctx, conn, ProfileID)
	if err != nil {
		return err
	}
	if retained {
		return blocked
	}
	retained, err = authn.RetainedRouteStatePresent(ctx, conn, []string{routeKeyTablesPatch, routeKeyTablesDelete, routeKeyIndicatorLinksCreate, routeKeyGraphViewsCreate, routeKeyGraphViewsPatch, routeKeyGraphViewsDelete, routeKeyGraphViewsRefresh})
	if err != nil {
		return err
	}
	if retained {
		return blocked
	}
	retained, err = jobs.RetainedExtensionStatePresent(ctx, conn, ProfileID, GraphViewMaterializationJobKind, graphViewWorkerKind)
	if err != nil {
		return err
	}
	if retained {
		return blocked
	}
	return nil
}

type orientationMigrationBlocked struct{}

func (orientationMigrationBlocked) Error() string {
	return "network flow orientation requires fresh state; retain the previous environment and re-import original sources"
}
func (orientationMigrationBlocked) ReasonCode() string {
	return "network_flow_orientation_state_incompatible"
}
