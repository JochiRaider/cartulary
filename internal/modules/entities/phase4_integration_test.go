package entities_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

// I-4-01 / REQ-01-196..REQ-01-227, REQ-02-039..REQ-02-044 / AC-188..AC-190, AC-221..AC-225.
func TestPhase4_ResolveRoute_I_4_01_Red(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-i-4-01")
	phase4test.RequireRouteSurface(
		t,
		"I-4-01",
		harness.Server,
		http.MethodPost,
		"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
		fixtures.MentionResolveRoutePayload(7, "txn-phase4-i-4-01", golden.Phase4MentionActionResolve, uuidPointer(golden.Phase4CanonicalHostRecordID), nil),
	)
}

// I-4-02 / REQ-02-035..REQ-02-036, REQ-02-054..REQ-02-055, REQ-02-059..REQ-02-063 / AC-022, AC-186.
func TestPhase4_EntityOriginUpsert_I_4_02_Red(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-i-4-02")
	resp := phase4test.RequireRouteSurface(
		t,
		"I-4-02",
		harness.Server,
		http.MethodPost,
		"/api/v1/incidents/"+golden.Phase4IncidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
		fixtures.HostCreatePayload("txn-phase4-i-4-02"),
	)
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] == "unknown_view_schema" {
		t.Fatalf("Phase 4 I-4-02 expected Hosts entity_origin surface %s to be active, got reason_code=%v", golden.Phase4HostsViewSchemaID, details["reason_code"])
	}
}

// I-4-03 / REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitMergeRoute_I_4_03_Red(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-i-4-03")
	phase4test.RequireSchemaTables(t, harness.DB, "I-4-03", "hosts", "identities", "entity_mentions", "record_tags", "compromise_assessments")
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}
