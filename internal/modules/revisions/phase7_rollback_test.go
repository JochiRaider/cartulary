package revisions_test

import "testing"

func TestPhase7_RollbackSelectorUnion_U_7_05(t *testing.T) {
	requirePhase7LaterSprintScope(t, "U-7-05", "rollback selector union and source=rollback change set")
}
