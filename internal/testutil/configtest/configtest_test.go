package configtest

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEffectiveFixture(t *testing.T) {
	cfg := LoadEffectiveFixture(t, []string{"config", "valid.toml"}, nil)

	if cfg.ConfigSchemaID != "cartulary.deployment_config.v1" {
		t.Fatalf("unexpected config schema id: %q", cfg.ConfigSchemaID)
	}
	if cfg.DeploymentProfile != "disconnected" {
		t.Fatalf("unexpected deployment profile: %q", cfg.DeploymentProfile)
	}
}

func TestOverlayAppliesToFixture(t *testing.T) {
	cfg := LoadEffectiveFixture(t, []string{"config", "valid.toml"}, Overlay(
		"CARTULARY__ROOTS__TEMPORARY_WORK__PATH", "/tmp/cartulary-test",
		"CARTULARY__LIMITS__IMPORTS__MAX_ROWS", "42",
	))

	if cfg.Roots.TemporaryWork.Path != "/tmp/cartulary-test" {
		t.Fatalf("unexpected temporary work path: %q", cfg.Roots.TemporaryWork.Path)
	}
	if cfg.Limits.Imports.MaxRows != 42 {
		t.Fatalf("unexpected max rows: %d", cfg.Limits.Imports.MaxRows)
	}
}

func TestInvalidFixtureMatchesGolden(t *testing.T) {
	err := LoadInvalidFixture(t, []string{"config", "invalid_missing_required.toml"}, nil)
	RequireDiagnosticsMatchGolden(t, err, []string{"config", "invalid_missing_required.json"})
}

func TestSetupTempRootsCreatesDeterministicSubpaths(t *testing.T) {
	roots := SetupTempRoots(t)

	if filepath.Base(roots.Base) != "cartulary-runtime-roots" {
		t.Fatalf("unexpected base directory: %q", roots.Base)
	}
	if !strings.HasSuffix(roots.Paths["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"], filepath.Join("cartulary-runtime-roots", "temporary-work")) {
		t.Fatalf("unexpected temporary work suffix: %q", roots.Paths["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"])
	}
}
