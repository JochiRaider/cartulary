package extensions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestCoordinatorGeneratedRegistry_Unit(t *testing.T) {
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		if validation, ok := err.(*ValidationError); ok {
			t.Fatalf("admit generated registry: %#v", validation.Findings())
		}
		t.Fatalf("admit generated registry: %v", err)
	}
	descriptors := coordinator.Descriptors()
	if len(descriptors) != 6 || len(coordinator.RegistrySHA256()) != 64 {
		t.Fatalf("generated registry identity = %d/%q", len(descriptors), coordinator.RegistrySHA256())
	}
	if descriptors[0].ProfileID != "enterprise_authentication" || descriptors[5].ProfileID != "snapshot_reporting" {
		t.Fatalf("descriptors are not canonical: %#v", descriptors)
	}
	descriptors[0].RouteFamilies[0] = "mutated"
	again, _ := coordinator.Descriptor("enterprise_authentication")
	if again.RouteFamilies[0] == "mutated" {
		t.Fatal("descriptor query leaked mutable coordinator state")
	}
}

func TestCoordinatorClaimResolution_Unit(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	resolution, err := coordinator.ResolveClaims([]string{"network_flow_activity", "reference_pack", "import"})
	if err != nil {
		t.Fatalf("resolve explicit dependency set: %v", err)
	}
	if got, want := resolution.Claims().ProfileIDs(), []string{"import", "network_flow_activity", "reference_pack"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("claim set = %v; want %v", got, want)
	}
	if got, want := resolution.AdmissionOrder(), []string{"import", "network_flow_activity", "reference_pack"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("admission order = %v; want %v", got, want)
	}
	_, err = coordinator.ResolveClaims([]string{"network_flow_activity"})
	requireFinding(t, err, "extension_dependency_not_claimed", "dependency_validation")
	_, err = coordinator.ResolveClaims([]string{"future_profile"})
	requireFinding(t, err, "extension_profile_unrecognized", "claim_configuration")
}

func TestCoordinatorCollisionAdmission_Unit(t *testing.T) {
	source := newMutableGeneratedSource()
	basePath := "contracts/extensions/build/base-route-reservations.json"
	base := decodeMutableObject(t, source[basePath].JSON)
	reservations := base["reservations"].([]any)
	reservations = append(reservations, map[string]any{
		"reservation_id": "base.test.capture", "path_template": "/api/v1/import-sessions", "match_scope": "descendants", "owner_contract_ref": "docs/spec/01_architecture_storage_and_view_contracts.md#req:REQ-01-151.1",
	})
	base["reservations"] = reservations
	source.replace(t, basePath, base)
	integrityPath := generatedExtensionsRoot + "registry-integrity.json"
	integrity := decodeMutableObject(t, source[integrityPath].JSON)
	for _, raw := range integrity["supporting_contract_artifact_digests"].([]any) {
		row := raw.(map[string]any)
		if row["artifact_id"] == "build.base-route-reservations" {
			row["artifact_sha256"] = source[basePath].SHA256
		}
	}
	source.replace(t, integrityPath, integrity)
	_, err := NewCoordinator(source)
	requireCollision(t, err, "base_route_capture")
}

func TestCoordinatorBindingAdmission_Unit(t *testing.T) {
	source := newMutableGeneratedSource()
	bindingPath := generatedExtensionsRoot + "implementation-bindings/import.json"
	binding := decodeMutableObject(t, source[bindingPath].JSON)
	binding["contract_major"] = json.Number("99")
	source.replace(t, bindingPath, binding)

	integrityPath := generatedExtensionsRoot + "registry-integrity.json"
	integrity := decodeMutableObject(t, source[integrityPath].JSON)
	for _, raw := range integrity["implementation_binding_digests"].([]any) {
		row := raw.(map[string]any)
		if row["profile_id"] == "import" {
			row["binding_sha256"] = source[bindingPath].SHA256
		}
	}
	source.replace(t, integrityPath, integrity)
	_, err := NewCoordinator(source)
	requireFinding(t, err, "extension_implementation_unavailable", "implementation_binding")
}

func TestCoordinatorPublicationPlan_Unit(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	summary := plan.Summary()
	if summary.SchemaID != "cartulary.extension_publication_plan.v1" || summary.RegistrySHA256 != coordinator.RegistrySHA256() {
		t.Fatalf("unexpected publication summary: %#v", summary)
	}
	for _, digest := range []string{summary.ResolvedClaimSetSHA256, summary.ContributionRegistrySHA256, summary.RouteDispatchPlanSHA256, summary.WorkspaceRegistrySHA256, summary.WorkerPlanSHA256, summary.ListenerPlanSHA256, summary.ClientSupportRegistrySHA256, summary.ImplementationBindingSetSHA256} {
		if len(digest) != 64 {
			t.Fatalf("publication component digest %q is invalid", digest)
		}
	}
	routes := plan.Routes()
	claimed, inactive := 0, 0
	for _, route := range routes {
		if route.DispatchState == "claimed" {
			claimed++
			if route.ContributionID == nil {
				t.Fatal("claimed route has no contribution")
			}
		} else {
			inactive++
			if route.ContributionID != nil {
				t.Fatal("inactive route exposes a contribution")
			}
		}
	}
	if claimed != 2 || inactive != 9 || len(plan.Workspaces()) != 1 {
		t.Fatalf("publication route/workspace counts = %d/%d/%d", claimed, inactive, len(plan.Workspaces()))
	}
	canonical := plan.CanonicalJSON()
	canonical[0] = 'X'
	if plan.CanonicalJSON()[0] == 'X' {
		t.Fatal("publication plan leaked mutable canonical bytes")
	}
}

func requireGeneratedCoordinator(t testing.TB) *Coordinator {
	t.Helper()
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		if validation, ok := err.(*ValidationError); ok {
			t.Fatalf("admit generated coordinator: %#v", validation.Findings())
		}
		t.Fatalf("admit generated coordinator: %v", err)
	}
	return coordinator
}

func requireFinding(t testing.TB, err error, code, phase string) {
	t.Helper()
	validation, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("error = %T %v; want ValidationError", err, err)
	}
	for _, finding := range validation.Findings() {
		if finding.Code == code && finding.Phase == phase {
			return
		}
	}
	t.Fatalf("findings = %#v; want %s/%s", validation.Findings(), code, phase)
}

func requireCollision(t testing.TB, err error, class string) {
	t.Helper()
	validation, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("error = %T %v; want ValidationError", err, err)
	}
	for _, finding := range validation.Findings() {
		if finding.Code == "extension_registry_conflict" && finding.CollisionClass == class {
			return
		}
	}
	t.Fatalf("findings = %#v; want collision %s", validation.Findings(), class)
}

type mutableArtifactSource map[string]PackagedArtifact

func newMutableGeneratedSource() mutableArtifactSource {
	source := mutableArtifactSource{}
	for path, artifact := range contractsgen.ExtensionArtifactsIndex {
		if len(path) >= len("contracts/extensions/") && path[:len("contracts/extensions/")] == "contracts/extensions/" {
			source[path] = PackagedArtifact{JSON: []byte(artifact.JSON), SHA256: artifact.SHA256}
		}
	}
	return source
}

func (source mutableArtifactSource) Artifact(path string) (PackagedArtifact, bool) {
	artifact, ok := source[path]
	artifact.JSON = append([]byte(nil), artifact.JSON...)
	return artifact, ok
}

func (source mutableArtifactSource) replace(t testing.TB, path string, object map[string]any) {
	t.Helper()
	encoded, err := canonicalJSON(object, true)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(encoded)
	source[path] = PackagedArtifact{JSON: encoded, SHA256: hex.EncodeToString(digest[:])}
}

func decodeMutableObject(t testing.TB, data []byte) map[string]any {
	t.Helper()
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	return object
}
