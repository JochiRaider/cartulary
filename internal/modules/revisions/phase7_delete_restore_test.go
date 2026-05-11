package revisions_test

import "testing"

func TestPhase7_SoftDeleteRoutePreconditions_U_7_03(t *testing.T) {
	requirePhase7LaterSprintScope(t, "U-7-03", "soft-delete route row_version, role gates, and route-owned failures")
}

func TestPhase7_RestoreTombstonePreconditions_U_7_04(t *testing.T) {
	requirePhase7LaterSprintScope(t, "U-7-04", "restore tombstone row_version and append-only history")
}
