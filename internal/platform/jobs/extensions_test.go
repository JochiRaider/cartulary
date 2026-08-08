package jobs_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestExtensionJobAdmissionMetadataIsClosedAndInternal_Unit(t *testing.T) {
	actorID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	key := jobs.RouteIdempotencyKey{
		RouteKey: "imports.sessions.create", ActorUserID: actorID,
		ScopeKey: incidentID.String(), ClientTxnID: "txn-1",
	}
	admission, err := jobs.NewExtensionJobAdmission(
		"import", "import.discovery_v1", key,
		jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		[]byte(`{"client_txn_id":"txn-1"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	var identity map[string]any
	if err := json.Unmarshal(admission.IdempotencyIdentity, &identity); err != nil {
		t.Fatal(err)
	}
	if len(identity) != 6 ||
		identity["schema_id"] != "cartulary.route_scoped_idempotency_identity.v1" ||
		identity["actor_user_id"] != actorID.String() ||
		identity["route_identity"] != "imports.sessions.create:"+incidentID.String() ||
		identity["scope_kind"] != jobs.ScopeKindIncident ||
		identity["scope_id"] != incidentID.String() ||
		identity["client_txn_id"] != "txn-1" {
		t.Fatalf("unexpected idempotency identity: %#v", identity)
	}
	if len(admission.NormalizedRequestSHA256) != 64 {
		t.Fatalf("request digest = %q", admission.NormalizedRequestSHA256)
	}
	publicJSON, err := json.Marshal(jobs.Resource{JobID: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(publicJSON), "extension_") || strings.Contains(string(publicJSON), "job_kind") ||
		strings.Contains(string(publicJSON), "progress_unit") {
		t.Fatalf("internal extension metadata leaked into public job: %s", publicJSON)
	}
}

func TestCanonicalExtensionTerminalSuccessValidatesResourceContracts_Unit(t *testing.T) {
	contract := jobs.ExtensionJobContract{
		OwnerProfileID: "import", JobKind: "import.apply_v1",
		ProgressUnitID: "import.apply.import_unit.v1",
		OperationKind:  "import.apply", WorkerKind: "import.apply_worker_v1",
		ContractSHA256: strings.Repeat("a", 64), ProofRequired: true, MaxProofBytes: 4096,
		ResourceRefs: []jobs.ExtensionResourceRefContract{
			{Kind: "import_session", MaxRefs: 1},
			{Kind: "network_flow_table", MaxRefs: 2},
		},
	}
	normalized, terminal, refs, digest, err := jobs.CanonicalExtensionTerminalSuccess(contract, &jobs.ResultSummary{
		Code: "import_session_applied", Message: "Applied.",
		ResourceRefs: []jobs.ResourceRef{
			{Kind: "network_flow_table", ID: "table-2", Route: "/tables/table-2"},
			{Kind: "import_session", ID: "session-1", Route: "/imports/session-1"},
			{Kind: "network_flow_table", ID: "table-1", Route: "/tables/table-1"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if normalized.ResourceRefs[0].Kind != "import_session" ||
		normalized.ResourceRefs[1].ID != "table-1" ||
		len(terminal) == 0 || len(refs) == 0 || len(digest) != 64 {
		t.Fatalf("unexpected canonical terminal result: %#v %s %s %s", normalized, terminal, refs, digest)
	}
	if _, _, _, _, err := jobs.CanonicalExtensionTerminalSuccess(contract, &jobs.ResultSummary{
		Code: "bad", Message: "Bad.",
		ResourceRefs: []jobs.ResourceRef{{Kind: "undeclared", ID: "1"}},
	}); err == nil {
		t.Fatal("expected undeclared resource reference rejection")
	}
	manager := jobs.NewManager()
	if err := manager.ConfigureExtensionContracts([]jobs.ExtensionJobContract{contract}); err != nil {
		t.Fatal(err)
	}
	invalid := contract
	invalid.ResourceRefs = []jobs.ExtensionResourceRefContract{
		{Kind: "network_flow_table", MaxRefs: 1},
		{Kind: "import_session", MaxRefs: 1},
	}
	if err := manager.ConfigureExtensionContracts([]jobs.ExtensionJobContract{invalid}); err == nil {
		t.Fatal("expected unsorted resource contract rejection")
	}
	invalid = contract
	invalid.ProgressUnitID = "caller selected unit"
	if err := manager.ConfigureExtensionContracts([]jobs.ExtensionJobContract{invalid}); err == nil {
		t.Fatal("expected invalid progress unit rejection")
	}
}
