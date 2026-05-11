package revisions_test

import "testing"

func TestPhase7_DestructiveOperationLocks_U_7_06(t *testing.T) {
	requirePhase7LaterSprintScope(t, "U-7-06", "restore, rollback, and merge destructive-operation lock precedence")
}
