package openapicompat

import (
	"crypto/sha256"
	"encoding/hex"
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
