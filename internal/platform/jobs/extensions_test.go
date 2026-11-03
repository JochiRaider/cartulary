package jobs_test

import (
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
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
		"import", key,
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
	t.Run("canonical byte contract", checkTerminalCanonicalBytes)
	t.Run("aggregate bound", checkTerminalReferenceBound)
	t.Run("current public resources", checkCurrentPublicResources)
	definition := jobs.Definition{
		JobKind: "import.apply_v1", ProgressUnitID: "import.apply.import_unit.v1",
		HandlerName: "import.apply_worker_v1",
		Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"import_session_applied"},
			Resources: []jobcontract.ResourceContract{
				{Kind: "import_session", SchemaID: "cartulary.common_job_resource_ref_id.v1", MinRefs: 1, MaxRefs: 1},
				{Kind: "network_flow_table", SchemaID: "cartulary.common_job_resource_ref_id.v1", MaxRefs: 2},
			},
		},
		Extension: &jobs.ExtensionPolicy{IdempotencyPolicy: "required", CancellationPolicy: "precommit_observable", IdentitySchemaID: jobs.HumanRouteIdentitySchema,
			OwnerProfileID: "import", OperationKind: "import.apply",
			ContractSHA256: strings.Repeat("a", 64), ProofRequired: true, MaxProofBytes: 4096,
		},
	}
	normalized, terminal, refs, digest, err := jobs.CanonicalTerminalSuccess(definition, &jobs.ResultSummary{
		Code: "import_session_applied", Message: "Applied.",
		ResourceRefs: []jobs.ResourceRef{
			{Kind: "network_flow_table", ID: "nft_00000000000000000000000000000002", Route: "/api/v1/incidents/22222222-2222-4222-8222-222222222222/network-flow/tables/nft_00000000000000000000000000000002"},
			{Kind: "import_session", ID: "11111111-1111-4111-8111-111111111111", Route: "/api/v1/import-sessions/11111111-1111-4111-8111-111111111111"},
			{Kind: "network_flow_table", ID: "nft_00000000000000000000000000000001", Route: "/api/v1/incidents/22222222-2222-4222-8222-222222222222/network-flow/tables/nft_00000000000000000000000000000001"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if normalized.ResourceRefs[0].Kind != "import_session" ||
		normalized.ResourceRefs[1].ID != "nft_00000000000000000000000000000001" ||
		len(terminal) == 0 || len(refs) == 0 || len(digest) != 64 {
		t.Fatalf("unexpected canonical terminal result: %#v %s %s %s", normalized, terminal, refs, digest)
	}
	if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &jobs.ResultSummary{
		Code: "bad", Message: "Bad.",
		ResourceRefs: []jobs.ResourceRef{{Kind: "undeclared", ID: "1"}},
	}); err == nil {
		t.Fatal("expected undeclared resource reference rejection")
	}
	if _, err := jobs.NewCatalog([]jobs.Definition{definition}); err != nil {
		t.Fatal(err)
	}
	invalid := definition
	invalidPolicy := *definition.Extension
	invalid.Extension = &invalidPolicy
	invalid.Terminal.Resources = []jobcontract.ResourceContract{
		{Kind: "network_flow_table", SchemaID: "cartulary.common_job_resource_ref_id.v1", MaxRefs: 1},
		{Kind: "import_session", SchemaID: "cartulary.common_job_resource_ref_id.v1", MaxRefs: 1},
	}
	if _, err := jobs.NewCatalog([]jobs.Definition{invalid}); err == nil {
		t.Fatal("expected unsorted resource contract rejection")
	}
	invalid = definition
	invalid.ProgressUnitID = "caller selected unit"
	if _, err := jobs.NewCatalog([]jobs.Definition{invalid}); err == nil {
		t.Fatal("expected invalid progress unit rejection")
	}
}

func checkTerminalReferenceBound(t *testing.T) {
	definition := jobs.Definition{Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"done"}, Resources: []jobcontract.ResourceContract{{Kind: "incident", SchemaID: "cartulary.common_job_resource_ref_id.v1", MaxRefs: 1024}}}}
	result := jobs.ResultSummary{Code: "done", Message: "Done."}
	for i := 0; i < 1024; i++ {
		id := fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1)
		result.ResourceRefs = append(result.ResourceRefs, jobs.ResourceRef{Kind: "incident", ID: id, Route: "/api/v1/incidents/" + id})
	}
	if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err != nil {
		t.Fatal("1024 references", err)
	}
	result.ResourceRefs = append(result.ResourceRefs, jobs.ResourceRef{Kind: "incident", ID: "00000000-0000-4000-8000-000000002000"})
	if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err == nil {
		t.Fatal("accepted 1025 references")
	}
	result.ResourceRefs = result.ResourceRefs[:1]
	definition.Terminal.Resources[0].SchemaID = "cartulary.unsupported.v1"
	if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err == nil {
		t.Fatal("unsupported resource schema")
	}
	definition.Terminal.Resources[0].SchemaID = "cartulary.common_job_resource_ref_id.v1"
	definition.Terminal.SchemaID = "cartulary.unsupported.v1"
	if _, _, _, _, err := jobs.CanonicalTerminalSuccess(definition, &result); err == nil {
		t.Fatal("unsupported terminal schema")
	}
}
