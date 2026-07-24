package networkflow

import "fmt"

// RecoveryPostgresTables resolves one owner-declared physical backup binding
// to its authoritative PostgreSQL tables. It performs no profile startup,
// claim, participant, worker, dependency, or migration work and is therefore
// safe for stopped-target backup and restore composition.
func RecoveryPostgresTables(bindingID string) ([]string, error) {
	switch bindingID {
	case "network_flow_activity.tables":
		return []string{"network_flow_tables"}, nil
	case "network_flow_activity.rows":
		return []string{"network_flow_rows"}, nil
	case "network_flow_activity.rejected_row_diagnostics":
		return []string{"network_flow_rejected_row_diagnostics"}, nil
	case "network_flow_activity.indicator_bindings":
		return []string{"network_flow_indicator_bindings"}, nil
	default:
		return nil, fmt.Errorf("unknown Network Flow recovery binding %q", bindingID)
	}
}
