import assert from "node:assert/strict";
import { copyFileSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { resolveRowSelector } from "../../test-support/owner-fixtures.mjs";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
import { assertPostgresCatalogClosure, assertPostgresPolicyFixtures, assertFixtureBuilderClosure, assertSelectorResolutionScope, assertPrivateRowSelection, assertSourceSnapshotIdentity } from "./assertions.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
const tiers = ["fast", "standard", "full", "release"];
function assertGoSelectorBuildContextContract() {
  const sourceRoot = mkdtempSync(path.join(tmpdir(), "cartulary-selector-build-context."));
  const fixtureRoot = path.join(sourceRoot, "internal/fixture");
  mkdirSync(fixtureRoot, { recursive: true });
  mkdirSync(path.join(sourceRoot, "tools"));
  copyFileSync(path.join(root, "tools/toolchain_pins.json"), path.join(sourceRoot, "tools/toolchain_pins.json"));
  copyFileSync(path.join(root, "tools/generated_artifact_policy.json"), path.join(sourceRoot, "tools/generated_artifact_policy.json"));
  const relativePackage = "./internal/fixture";
  const runner = { approved_roots: ["internal"] };
  const row = (testName) => ({
    row_id: "harness.test_catalog.behavior.fixture",
    runner: "go",
    selector: { package: relativePackage, tests: [testName] },
  });
  try {
    writeFileSync(path.join(fixtureRoot, "ordinary_test.go"), "package fixture\nfunc TestOrdinary() {}\n");
    writeFileSync(path.join(fixtureRoot, "applicable_linux_amd64_test.go"), "//go:build linux && amd64\n\npackage fixture\nfunc TestApplicablePlatform() {}\n");
    writeFileSync(path.join(fixtureRoot, "private_test.go"), "//go:build cartulary_harness\n\npackage fixture\nfunc TestPrivateProfile() {}\n");
    writeFileSync(path.join(fixtureRoot, "excluded_windows_test.go"), "//go:build windows\n\npackage fixture\nfunc TestExcludedPlatform() {}\n");
    assert.deepEqual(resolveRowSelector({ root: sourceRoot, row: row("TestOrdinary"), runner, taskSurfaceCommandIDs: new Set() }), [`go:${relativePackage}:TestOrdinary`]);
    assert.deepEqual(resolveRowSelector({ root: sourceRoot, row: row("TestApplicablePlatform"), runner, taskSurfaceCommandIDs: new Set() }), [`go:${relativePackage}:TestApplicablePlatform`]);
    assert.throws(() => resolveRowSelector({ root: sourceRoot, row: row("TestPrivateProfile"), runner, taskSurfaceCommandIDs: new Set() }), /TestPrivateProfile is excluded from the Go runner build context linux\/amd64/u);
    assert.throws(() => resolveRowSelector({ root: sourceRoot, row: row("TestExcludedPlatform"), runner, taskSurfaceCommandIDs: new Set() }), /TestExcludedPlatform is excluded from the Go runner build context linux\/amd64/u);
  }
  finally {
    rmSync(sourceRoot, { recursive: true, force: true });
  }
}
export function registerBoundariesTests() {
  test("catalog_tier_closure: catalog rows use only current minimum tiers", async () => {
    const context = createContractTestContext(root);
    assert.ok(context.catalog.rows.length > 0);
    assert.ok(context.catalog.rows.every((row) => [...tiers, "explicit_only"].includes(row.aggregate_placement)));
  });
  test("fixture_profile_closure: fixture profiles, retired rows, and canonical construction are closed", async () => {
    const context = createContractTestContext(root);
    assert.ok(context.catalog.rows.every((row) => typeof row.fixture_capability === "string"));
    assert.ok(context.catalog.rows.every((row) => !("default_check" in row)));
    assertFixtureBuilderClosure(context);
    const migration = context.rowMigrations.migrations.find((entry) => entry.retired_row_id ===
      "harness.browser.integration.performance_fixture_source_owner_assembly");
    assert.deepEqual(migration?.replacement_row_ids, [
      "harness.browser.integration.performance_fixture_snapshot_lifecycle",
      "harness.browser.unit.source_owner_contribution_assembler",
    ]);
  });
}
export function registerCommandSurfaceTests() {
  test("source_snapshot_identity: source snapshots preserve byte identity and reject changes during hashing", async () => {
    const context = createContractTestContext(root);
    assertSourceSnapshotIdentity(context);
  });
  test("private_row_selection_boundary: private row selection preserves metadata closure and exact source checks", async () => {
    const context = createContractTestContext(root);
    assertPrivateRowSelection(context);
  });
}
export function registerEvidenceTests() {
  test("postgres_catalog_closure: Postgres fixture counts reconcile with active policy partitions", async () => {
    const context = createContractTestContext(root);
    assertPostgresCatalogClosure(context);
  });
  test("postgres_policy_growth_and_rejection: Postgres policy admits growth and rejects invalid isolation approvals", async () => {
    const context = createContractTestContext(root);
    assertPostgresPolicyFixtures(context);
  });
  test("go_selector_build_context: Go selectors honor the canonical build context", async () => {
    assertGoSelectorBuildContextContract();
  });
  test("catalog_selector_scope: selector facts refresh between loads and preserve path checks", async () => {
    const context = createContractTestContext(root);
    assertSelectorResolutionScope(context);
  });
  test("active_owner_row_coverage: every active owner retains current row coverage", async () => {
    const context = createContractTestContext(root);
    assert.ok(context.catalogMetadata.registry.owners.filter((owner) => owner.status === "active").every((owner) => context.catalogMetadata.rows.some((row) => row.owner_id === owner.owner_id)));
  });
}
