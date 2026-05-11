package revisions_test

import "testing"

func requirePhase7LaterSprintScope(t testing.TB, id string, feature string) {
	t.Helper()
	if id == "" || feature == "" {
		t.Fatalf("phase 7 deferred scope sentinel must name both row id and feature")
	}
	t.Logf("%s remains later Phase 7 work for %s; Sprint 1 completion stays limited to GET /api/v1/records/{record_id}/history", id, feature)
}
