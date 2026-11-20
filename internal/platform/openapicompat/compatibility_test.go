package openapicompat

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionContractPassesReleaseCompatibilityGate(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	report, err := CheckRepository(root)
	if err != nil {
		t.Fatalf("check production compatibility: %v", err)
	}
	if report.BaselineVersion != "1.0.0" || report.TargetVersion == "" {
		t.Fatalf("unexpected production versions: %#v", report)
	}
	for _, change := range report.Changes {
		if change.Fingerprint == "" || change.Classification == "" {
			t.Fatalf("incomplete approved production change: %#v", change)
		}
	}
}

func TestCompatibilityClassificationAndVersionPolicy(t *testing.T) {
	baseline := []byte(`{
		"openapi":"3.1.0",
		"info":{"title":"fixture","version":"1.0.0"},
		"paths":{
			"/items":{"get":{
				"operationId":"listItems",
				"security":[{"sessionCookie":[]}],
				"responses":{"200":{"description":"ok"}}
			}}
		},
		"components":{"schemas":{"Mode":{"type":"string","enum":["one","two"]}}}
	}`)
	target := []byte(`{
		"openapi":"3.1.0",
		"info":{"title":"fixture changed","version":"2.0.0"},
		"paths":{
			"/items":{"get":{
				"operationId":"listItems",
				"security":[{"sessionCookie":[]}],
				"responses":{"200":{"description":"still ok"}}
			}}
		},
		"components":{"schemas":{"Mode":{"type":"string","enum":["one"]}}}
	}`)
	report, err := Compare(baseline, target)
	if err != nil {
		t.Fatalf("compare fixture: %v", err)
	}
	if err := validateVersionPolicy(report); err != nil {
		t.Fatalf("validate fixture version: %v", err)
	}
	foundBreaking := false
	foundNonBehavioral := false
	for _, change := range report.Changes {
		switch change.Classification {
		case Breaking:
			foundBreaking = true
		case NonBehavioral:
			foundNonBehavioral = true
		}
	}
	if !foundBreaking || !foundNonBehavioral {
		t.Fatalf("expected breaking and non-behavioral changes, got %#v", report.Changes)
	}
}

func TestCompatibilityRejectsBreakingChangeWithoutMajorVersion(t *testing.T) {
	baseline := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1.0.0"},"paths":{"/a":{"get":{"operationId":"a","security":[],"responses":{"200":{"description":"ok"}}}}}}`)
	target := []byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1.1.0"},"paths":{}}`)
	report, err := Compare(baseline, target)
	if err != nil {
		t.Fatalf("compare fixture: %v", err)
	}
	if err := validateVersionPolicy(report); err == nil {
		t.Fatal("breaking change with minor version unexpectedly passed")
	}
}

func TestReleaseRegistryRejectsModifiedSnapshot(t *testing.T) {
	root := t.TempDir()
	for _, relative := range []string{
		"contracts/openapi-releases",
		"contracts/openapi",
	} {
		if err := os.MkdirAll(filepath.Join(root, relative), 0o755); err != nil {
			t.Fatalf("mkdir fixture: %v", err)
		}
	}
	if err := os.WriteFile(
		filepath.Join(root, "contracts/openapi-releases/index.json"),
		[]byte(`{"schema_id":"cartulary.openapi_release_registry.v1","latest_released_version":"1.0.0","releases":[{"version":"1.0.0","document_path":"contracts/openapi-releases/1.0.0.openapi.json","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","byte_length":2,"source_commit":"53f4553150117d09535794e59a8f500485bdb94c","publication_state":"historical_baseline"}]}`),
		0o644,
	); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "contracts/openapi-releases/1.0.0.openapi.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write release: %v", err)
	}
	// The digest is checked before the target document, manifest or change
	// sets are read, so only the digest can reject this fixture.
	if _, err := CheckRepository(root); err == nil || err.Error() != "released document SHA-256 does not match registry" {
		t.Fatalf("modified release snapshot error = %v", err)
	}
	// Control: with the true digest the gate proceeds to the next stage.
	sum := sha256.Sum256([]byte(`{}`))
	registry := strings.Replace(`{"schema_id":"cartulary.openapi_release_registry.v1","latest_released_version":"1.0.0","releases":[{"version":"1.0.0","document_path":"contracts/openapi-releases/1.0.0.openapi.json","sha256":"DIGEST","byte_length":2,"source_commit":"53f4553150117d09535794e59a8f500485bdb94c","publication_state":"historical_baseline"}]}`, "DIGEST", hex.EncodeToString(sum[:]), 1)
	if err := os.WriteFile(filepath.Join(root, "contracts/openapi-releases/index.json"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	if _, err := CheckRepository(root); err == nil || !strings.HasPrefix(err.Error(), "read current document:") {
		t.Fatalf("true digest must pass the digest check, got %v", err)
	}
}

func TestMutationClassificationIsSecuritySensitive(t *testing.T) {
	for _, kind := range []string{"add", "remove", "change"} {
		if got := classify(kind, "/paths/~1a/post/x-cartulary-state-changing", false, true); got != SecuritySensitive {
			t.Fatalf("%s classified %s", kind, got)
		}
	}
}

// TestErrorCodeDeclarationsCompareAsSets covers REQ-01-702: declaring another
// code is additive, and losing a declared code is breaking.
// TestParameterSchemaNarrowingIsBreaking covers the S-1089 classifier rule:
// adding pattern, format, minLength or maxLength to a component parameter's
// schema narrows the values a request may carry, so it is breaking, while
// declaring other parameter metadata stays additive. Inline parameter arrays
// are compared whole and are breaking whenever they change.
func TestParameterSchemaNarrowingIsBreaking(t *testing.T) {
	const parameter = "/components/parameters/ProviderKeyPathParameter"
	for _, tc := range []struct {
		name, kind, pointer string
		before, after       any
		want                Classification
	}{
		{"add pattern", "add", parameter + "/schema/pattern", nil, "^[a-z]+$", Breaking},
		{"add format", "add", parameter + "/schema/format", nil, "uuid", Breaking},
		{"add minLength", "add", parameter + "/schema/minLength", nil, 1, Breaking},
		{"add maxLength", "add", parameter + "/schema/maxLength", nil, 64, Breaking},
		{"add nested item pattern", "add", parameter + "/schema/items/pattern", nil, "^[a-z]+$", Breaking},
		{"change pattern", "change", parameter + "/schema/pattern", "^[a-z]+$", "^[a-z0-9]+$", Breaking},
		{"add not-found code", "add", parameter + "/x-cartulary-not-found-code", nil, "auth_provider_not_found", Additive},
		{"add parameter component", "add", "/components/parameters/NewPathParameter", nil, map[string]any{"name": "x"}, Additive},
		{"add schema pattern outside parameters", "add", "/components/schemas/Thing/pattern", nil, "^[a-z]+$", Additive},
		{"inline parameter array change", "change", "/paths/~1a~1{id}/get/parameters", []any{map[string]any{"name": "id"}}, []any{map[string]any{"name": "id", "x-cartulary-not-found-code": "thing_not_found"}}, Breaking},
	} {
		if got := classify(tc.kind, tc.pointer, tc.before, tc.after); got != tc.want {
			t.Errorf("%s: classified %s, want %s", tc.name, got, tc.want)
		}
	}
	before := map[string]any{"openapi": "3.1.0", "info": map[string]any{"version": "1.0.0"}, "components": map[string]any{"parameters": map[string]any{
		"ProviderKeyPathParameter": map[string]any{"name": "provider_key", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
	}}}
	after := map[string]any{"openapi": "3.1.0", "info": map[string]any{"version": "2.0.0"}, "components": map[string]any{"parameters": map[string]any{
		"ProviderKeyPathParameter": map[string]any{"name": "provider_key", "in": "path", "required": true, "schema": map[string]any{"type": "string", "pattern": "^[a-z]+$"}, "x-cartulary-not-found-code": "auth_provider_not_found"},
	}}}
	beforeBytes, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	afterBytes, err := json.Marshal(after)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Compare(beforeBytes, afterBytes)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Classification{}
	for _, change := range report.Changes {
		got[change.Pointer] = change.Classification
	}
	want := map[string]Classification{
		parameter + "/schema/pattern":             Breaking,
		parameter + "/x-cartulary-not-found-code": Additive,
	}
	if len(got) != len(want) {
		t.Fatalf("changes = %#v, want %#v", got, want)
	}
	for pointer, classification := range want {
		if got[pointer] != classification {
			t.Fatalf("%s classified %s, want %s", pointer, got[pointer], classification)
		}
	}
}

func TestErrorCodeDeclarationsCompareAsSets(t *testing.T) {
	const pointer = "/paths/~1a/post/responses/409/x-cartulary-error-codes"
	for _, tc := range []struct {
		name          string
		kind          string
		before, after any
		want          Classification
	}{
		{"declare", "add", nil, []any{"client_txn_conflict"}, Additive},
		{"add a code", "change", []any{"client_txn_conflict"}, []any{"client_txn_conflict", "illegal_transition"}, Additive},
		{"lose a code", "change", []any{"client_txn_conflict", "illegal_transition"}, []any{"illegal_transition"}, Breaking},
		{"replace a code", "change", []any{"invalid_mutation_payload"}, []any{"user_email_conflict"}, Breaking},
		{"drop the declaration", "remove", []any{"client_txn_conflict"}, nil, Breaking},
	} {
		if got := classify(tc.kind, pointer, tc.before, tc.after); got != tc.want {
			t.Errorf("%s: classified %s, want %s", tc.name, got, tc.want)
		}
	}
}
