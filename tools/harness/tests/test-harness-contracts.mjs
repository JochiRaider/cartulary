#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath, pathToFileURL } from "node:url";

import {
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
} from "../generated-artifacts/index.mjs";
import {
  expandServiceBackedSchedule,
  expandServiceBackedScheduleForCheck,
} from "../execution/service-backed/index.mjs";
import {
  collectTaskSurfaceManifestErrors,
  renderTaskSurfaceMake,
} from "../generated-artifacts/index.mjs";
import {
  fileArtifactRef,
  buildToolRunSummary,
  machineOutput,
  terminalArtifactPath,
  toolRunSummarySchemaID,
} from "../output/index.mjs";
import {
  HarnessConfigError,
  generateTestRouteToken,
  normalizeFailureClass,
  normalizeFailureRecord,
  primaryPublicFailure,
  preflightPublicTarget,
  redactString,
  redactValue,
  runCleanup,
  testRouteTokenValid,
  validatePreparedArtifactIdentity,
  validateSchema,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  collectGoShardsForTargetFromRows,
} from "../backend/backend-shard-plan.mjs";
import { collectTargetPlanRows } from "../backend/backend-target-plan.mjs";
import { renderServiceBackedScheduleManifest } from "../generated-artifacts/index.mjs";
import {
  loadHarnessHelperOwnership,
  ownerFacadePathLists,
} from "../static-analysis/harness-helper-ownership.mjs";
import { collectHarnessImportBoundaryViolations } from "../static-analysis/harness-import-boundary.mjs";
import { validateSchedulerEventOrderFile } from "../scheduler/scheduler/event-order.mjs";
import { runCommand as runScheduledCommand } from "../scheduler/process-executor.mjs";
import { runNormalizedSchedule } from "../scheduler/scheduler-runner.mjs";
import { priorityAdmissiblePendingUnitIndex } from "../scheduler/scheduler/state.mjs";
import {
  schedulerCapacityProfilesByFamily,
  schedulerCapacityProfileValues,
  schedulerFamilyForCapacityProfile,
  schedulerFamilyValues,
} from "../scheduler/scheduler-family-contract.mjs";
import { schedulerSummaryLine } from "../scheduler/scheduler-reporting.mjs";
import { validateSchedulerResourceRegistrySemantics } from "../scheduler/scheduler-resources.mjs";
import { resolveRetainedLogArtifacts } from "../diagnostics/retained-artifact-resolver.mjs";
import { loadVerificationContracts } from "../test-catalog/verification-contracts.mjs";
import { loadTestCatalog } from "../test-catalog/test-catalog.mjs";
import { parseStrictJSON, semanticJSONDigest } from "../test-catalog/semantic-json.mjs";
import { collectTestCatalogImportViolations } from "../test-catalog/import-boundary.mjs";
import { resolveRowSelector } from "../test-catalog/selector-resolution.mjs";
import { validateSemanticIdentities } from "../test-catalog/semantic-identity-check-cli.mjs";
import { commandTargetForEvidenceTarget } from "../test-catalog/target-routing.mjs";
import {
  auditOwnerEvidence,
  accountingRowsForTarget,
  buildTestEvidenceAccounting,
  buildTestOwnerSummary,
  deriveRequiredEvidencePartitions,
  evidenceTargetForCatalogRow,
  finalizeTargetOwnerEvidence,
  loadOwnerAccountingSelection,
} from "../evidence-accounting/index.mjs";
import { adaptGoInvocation } from "../execution/runners/go.mjs";
import { adaptPlaywrightReport } from "../execution/runners/playwright.mjs";
import { adaptShellInvocation } from "../execution/runners/shell.mjs";
import { adaptVitestInvocation } from "../execution/runners/vitest.mjs";
import {
  buildBrowserStageSchedule,
  selectedBrowserGroupRowIDs,
} from "../browser/index.mjs";
import { resolveBrowserBatchStage } from "../browser/browser-batch-manifest.mjs";
import {
  assertDocumentationAccessAllowed,
  scanDocumentationReadSource,
} from "../test-catalog/documentation-boundary.mjs";
import {
  buildOwnerSlicePlan,
  OwnerSliceUsageError,
  resolveOwnerSliceSelection,
} from "../owner-slice/index.mjs";
import {
  buildPlaywrightInvocations,
  ownerChildRunID,
} from "../owner-slice/execution.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { ownerSliceChildEnvironment } from "../owner-slice/scheduler.mjs";
import { buildModuleAuthorTaskGuide, explainTestOwner } from "../diagnostics/owner-diagnostics.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(repoRoot, relativePath), "utf8"));
}

function writeJSONFile(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function writeFixtureFile(root, relativePath, content) {
  const file = path.join(root, relativePath);
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, content);
}

function writeOwnerCatalogFixture(root, mutate = () => {}) {
  const verificationRegistry = {
    schema_id: "cartulary.verification_registry.v1",
    owners: [
      {
        owner_id: "module.fixture",
        contract_path: "contracts/verification/owners/module.fixture.json",
        status: "active",
      },
    ],
  };
  const verificationContract = {
    schema_id: "cartulary.verification_contract.v1",
    owner_id: "module.fixture",
    verifications: [
      {
        verification_id: "module.fixture.verification.behavior",
        behavior_class: "product",
        profile: "base",
        requirement: "Fixture-owned behavior remains exact.",
        evidence_kinds: ["go_test"],
        status: "active",
      },
    ],
  };
  const ownerRegistry = {
    schema_id: "cartulary.test_owner_registry.v1",
    owners: [
      {
        owner_id: "module.fixture",
        manifest_path: "tools/test_families/module.fixture.json",
        status: "active",
      },
    ],
  };
  const row = {
    row_id: "module.fixture.behavior.owned",
    owner_id: "module.fixture",
    family_id: "module.fixture.behavior",
    collaborator_ids: [],
    verification_ids: ["module.fixture.verification.behavior"],
    runner: "go",
    selector: { package: "./internal/fixture", tests: ["TestOwned"] },
    evidence_class: "unit",
    runtime_profile_id: "none",
    resource_profile_id: "go_balanced",
    fixture_profile_id: "none",
    default_check: true,
    claim_posture: "implementation",
    status: "active",
  };
  const familyManifest = {
    schema_id: "cartulary.test_family_manifest.v1",
    owner_id: "module.fixture",
    rows: [row],
  };
  const runnerRegistry = readJSON("tools/test_runner_registry.json");
  const topology = readJSON("tools/execution_topology_manifest.json");
  const schedulerResources = readJSON("tools/scheduler_resource_registry.json");
  const fixture = {
    verificationRegistry,
    verificationContract,
    ownerRegistry,
    familyManifest,
    runnerRegistry,
    topology: {
      schema_id: topology.schema_id,
      runtime_profiles: topology.runtime_profiles,
      resource_profiles: topology.resource_profiles,
      fixture_profiles: topology.fixture_profiles,
    },
    schedulerResources,
    taskSurface: { targets: [] },
  };
  mutate(fixture);
  writeJSONFile(path.join(root, "contracts/verification/registry.json"), fixture.verificationRegistry);
  writeJSONFile(
    path.join(root, "contracts/verification/owners/module.fixture.json"),
    fixture.verificationContract,
  );
  writeJSONFile(path.join(root, "tools/test_catalog_owner.json"), fixture.ownerRegistry);
  writeJSONFile(
    path.join(root, "tools/test_families/module.fixture.json"),
    fixture.familyManifest,
  );
  writeJSONFile(path.join(root, "tools/test_runner_registry.json"), fixture.runnerRegistry);
  writeJSONFile(path.join(root, "tools/execution_topology_manifest.json"), fixture.topology);
  writeJSONFile(path.join(root, "tools/scheduler_resource_registry.json"), fixture.schedulerResources);
  writeJSONFile(path.join(root, "tools/task_surface_manifest.json"), fixture.taskSurface);
  for (const runner of fixture.runnerRegistry.runners) {
    writeFixtureFile(root, runner.adapter_path, "export const fixture = true;\n");
  }
  writeFixtureFile(root, "cmd/fixture.txt", "fixture\n");
  writeFixtureFile(root, "internal/fixture/owned_test.go", "package fixture\n\nfunc TestOwned(t *testing.T) {}\n");
  return fixture;
}

test("owner catalog closes identities, selectors, profiles, and semantic digests", () => {
  const catalog = loadTestCatalog(repoRoot);
  assert.equal(catalog.summary.status, "pass");
  assert.equal(catalog.summary.owner_count, catalog.registry.owners.length);
  assert.equal(catalog.summary.owner_count, 48);
  assert.equal(catalog.summary.family_count, 175);
  assert.equal(catalog.summary.row_count, 862);
  assert.equal(catalog.summary.selector_count, 1410);
  assert.equal(
    Object.values(catalog.summary.runner_counts).reduce((sum, count) => sum + count, 0),
    catalog.summary.row_count,
  );
  assert.match(catalog.semantic_digest, /^sha256:[0-9a-f]{64}$/u);
  assert.match(catalog.verification.semantic_digest, /^sha256:[0-9a-f]{64}$/u);
  assert.ok(
    catalog.rowByID.has("module.graphprojection.engine.canonical_behavior"),
    "Graph Projection must be absorbed by the unified catalog",
  );
});

test("owner evidence accounting projects exact catalog rows without delivery metadata", () => {
  const selection = loadOwnerAccountingSelection(repoRoot, {
    ownerID: "module.networkflow",
  });
  assert.ok(selection.selected_rows.length > 0);
  assert.deepEqual(selection.selected_rows, [...selection.selected_rows].sort());
  assert.ok(selection.expected_rows.every((row) => row.owner_id === "module.networkflow"));
  assert.ok(selection.expected_rows.every((row) => row.selector_digest.startsWith("sha256:")));
  assert.ok(selection.expected_rows.some((row) => row.runner === "go"));
  assert.ok(selection.expected_rows.some((row) => row.runner === "playwright"));
  assert.ok(selection.expected_rows.some((row) => row.runner === "vitest"));
  assert.doesNotMatch(JSON.stringify(selection), /(?:guide_path|guide_digest)/u);

  const accessibility = accountingRowsForTarget(repoRoot, {
    ownerID: "module.networkflow",
    targetName: "browser-e2e-a11y",
  });
  assert.ok(accessibility);
  assert.ok(accessibility.expected_rows.length > 0);
  assert.ok(accessibility.expected_rows.every((row) => row.target_name === "browser-e2e-a11y"));
  assert.ok(accessibility.expected_rows.every((row) => row.runtime_profile_id === "network_flow_claimed"));

  const catalog = loadTestCatalog(repoRoot);
  const shellRow = catalog.rows.find((row) => row.runner === "shell");
  const goRow = catalog.rows.find((row) => row.runner === "go");
  assert.ok(shellRow);
  assert.ok(goRow);
  const commandTargets = new Map(
    readJSON("tools/task_surface_manifest.json").targets.map((entry) => [entry.command_id, entry.name]),
  );
  assert.notEqual(evidenceTargetForCatalogRow(shellRow, { commandTargetByID: commandTargets }), "");
  assert.notEqual(evidenceTargetForCatalogRow(goRow, { commandTargetByID: commandTargets }), "");
  for (const [familyID, targetName] of [
    ["module.graphprojection.engine", "backend-unit"],
    ["module.graphprojection.fixtures", "backend-unit"],
    ["module.graphprojection.storage", "backend-store"],
    ["harness.runtime.support_integration", "backend-integration-support"],
    ["app.server.process", "backend-process"],
  ]) {
    assert.equal(
      evidenceTargetForCatalogRow(
        { row_id: `${familyID}.fixture`, runner: "go", family_id: familyID },
        { commandTargetByID: commandTargets },
      ),
      targetName,
      `${familyID} must have one shared catalog target route`,
    );
  }
  assert.equal(
    commandTargetForEvidenceTarget("backend-integration-support"),
    "backend-integration",
  );
  assert.equal(
    commandTargetForEvidenceTarget("backend-unit"),
    "backend-unit",
  );
});

test("owner slice selection is exact, owner-qualified, and independent of default-check filtering", () => {
  const catalog = loadTestCatalog(repoRoot);
  const ownerRows = catalog.rows.filter((row) => row.owner_id === "platform.config");
  const exactRow = ownerRows.find((row) => row.runner === "go");
  assert.ok(exactRow);

  const omitted = resolveOwnerSliceSelection(repoRoot, {
    ownerID: "platform.config",
    dependencyScope: "all",
    rowsProvided: false,
  });
  assert.deepEqual(omitted.selection.resolved_row_ids, ownerRows.map((row) => row.row_id).sort());
  assert.equal(omitted.selection.selection_mode, "default_owner");
  assert.equal(omitted.selection.completion_scope, "full_owner");

  const exact = resolveOwnerSliceSelection(repoRoot, {
    ownerID: "platform.config",
    dependencyScope: "all",
    rows: exactRow.row_id,
    rowsProvided: true,
  });
  assert.deepEqual(exact.selection.requested_row_ids, [exactRow.row_id]);
  assert.deepEqual(exact.selection.resolved_row_ids, [exactRow.row_id]);
  assert.equal(exact.selection.selection_mode, "exact_rows");
  assert.equal(exact.selection.completion_scope, "selected_subset");
  assert.equal(exact.workUnits.length, 1);
});

test("owner slice input and service-backed selection fail closed before planning", () => {
  const catalog = loadTestCatalog(repoRoot);
  const nonServiceRow = catalog.rows.find((row) => row.owner_id === "platform.config");
  const serviceOwner = catalog.rows.find((row) => row.runtime_profile_id === "default")?.owner_id;
  assert.ok(nonServiceRow);
  assert.ok(serviceOwner);

  const serviceSelection = resolveOwnerSliceSelection(repoRoot, {
    ownerID: serviceOwner,
    dependencyScope: "service_backed",
    rowsProvided: false,
  });
  assert.ok(serviceSelection.rows.length > 0);
  assert.ok(serviceSelection.workUnits.every((unit) => unit.managed_service_ids.length > 0));

  for (const options of [
    { ownerID: "", dependencyScope: "all", rowsProvided: false },
    { ownerID: "unknown.owner", dependencyScope: "all", rowsProvided: false },
    { ownerID: "platform.config", dependencyScope: "all", rows: "", rowsProvided: true },
    {
      ownerID: "platform.config",
      dependencyScope: "all",
      rows: `${nonServiceRow.row_id},${nonServiceRow.row_id}`,
      rowsProvided: true,
    },
    {
      ownerID: "platform.config",
      dependencyScope: "service_backed",
      rows: nonServiceRow.row_id,
      rowsProvided: true,
    },
    { ownerID: "platform.config", dependencyScope: "all", rowsProvided: false, vitestWorkers: "17" },
    { ownerID: "platform.config", dependencyScope: "all", rowsProvided: false, jsonValue: "true" },
  ]) {
    assert.throws(
      () => resolveOwnerSliceSelection(repoRoot, options),
      (error) => error instanceof OwnerSliceUsageError && error.exitCode === 2,
    );
  }
});

test("owner slice plan is deterministic and schema-valid for semantic inputs", async () => {
  const catalog = loadTestCatalog(repoRoot);
  const row = catalog.rows.find((entry) => entry.owner_id === "platform.config" && entry.runner === "go");
  assert.ok(row);
  const options = {
    ownerID: row.owner_id,
    dependencyScope: "all",
    rows: row.row_id,
    rowsProvided: true,
    target: "test-slice",
    commandID: "cartulary.harness.command.test_slice.v1",
    runID: "fixture-run",
    timestamp: "2026-07-18T00:00:00.000Z",
  };
  const first = buildOwnerSlicePlan(repoRoot, options);
  const second = buildOwnerSlicePlan(repoRoot, options);
  const differentInvocation = buildOwnerSlicePlan(repoRoot, {
    ...options,
    runID: "fixture-run-2",
    timestamp: "2026-07-18T01:00:00.000Z",
  });
  assert.deepEqual(first, second);
  assert.equal(first.plan_semantic_digest, differentInvocation.plan_semantic_digest);
  assert.equal(first.scheduler_semantic_digest, differentInvocation.scheduler_semantic_digest);
  await validateSchema(first.schema_id, first);
  assert.deepEqual(first.selection.resolved_row_ids, [row.row_id]);
  assert.match(first.source_snapshot_digest, /^sha256:[0-9a-f]{64}$/u);

  const processRow = catalog.rows.find(
    (entry) => entry.family_id === "module.recovery.process",
  );
  assert.ok(processRow);
  const processPlan = buildOwnerSlicePlan(repoRoot, {
    ...options,
    ownerID: processRow.owner_id,
    dependencyScope: "service_backed",
    rows: processRow.row_id,
  });
  assert.deepEqual(processPlan.work_units[0].runtime_binary_ids, ["operator"]);
  await validateSchema(processPlan.schema_id, processPlan);
});

test("runner adapters require one exact terminal observation per selected row", () => {
  const goInvocation = {
    rows: [
      { row_id: "row.go.pass", selectors: ["TestPass"] },
      { row_id: "row.go.missing", selectors: ["TestMissing"] },
    ],
  };
  const goRows = adaptGoInvocation(goInvocation, {
    status: 1,
    stdout: `${JSON.stringify({ Action: "pass", Test: "TestPass", Elapsed: 0.01 })}\n`,
  });
  assert.deepEqual(goRows.map((row) => row.terminal_state), ["passed", "infrastructure_failed"]);

  const vitestRows = adaptVitestInvocation(
    { rows: [{ row_id: "row.vitest", selectors: ["exact title"] }] },
    {
      status: 0,
      stdout: JSON.stringify({
        testResults: [{ assertionResults: [{
          title: "exact title",
          fullName: "suite exact title",
          status: "passed",
        }] }],
      }),
    },
  );
  assert.equal(vitestRows[0].terminal_state, "passed");
  assert.equal(
    adaptVitestInvocation(
      { rows: [{ row_id: "row.vitest", selectors: ["other title"] }] },
      { status: 0, stdout: JSON.stringify({ testResults: [] }) },
    )[0].terminal_state,
    "infrastructure_failed",
  );

  const playwrightRow = {
    row_id: "row.playwright",
    selector: {
      file: "apps/web/e2e/exact.spec.ts",
      titles: ["exact browser title"],
    },
  };
  const playwrightReport = {
    suites: [{
      specs: [{
        file: "apps/web/e2e/exact.spec.ts",
        title: "exact browser title",
        tests: [{ status: "expected", results: [{ status: "passed", duration: 9 }] }],
      }],
    }],
  };
  assert.deepEqual(
    adaptPlaywrightReport([playwrightRow], playwrightReport, 0)[0],
    {
      row_id: "row.playwright",
      terminal_state: "passed",
      duration_ms: 9,
      exit_code: 0,
      failure_reason: null,
    },
  );
  assert.equal(
    adaptPlaywrightReport([playwrightRow], { suites: [] }, 0)[0].terminal_state,
    "infrastructure_failed",
    "aggregate Playwright success cannot close a row without its exact selector observation",
  );

  assert.equal(
    adaptShellInvocation(
      { rows: [{ row_id: "row.shell", selectors: ["command"] }] },
      { status: 7 },
    )[0].terminal_state,
    "failed",
  );
});

test("prepared artifact identity is atomic and validates before artifact creation", () => {
  const resultsRoot = mkdtempSync(path.join(repoRoot, "tmp", "prepared-identity-test."));
  try {
    const complete = {
      CARTULARY_HARNESS_IDENTITY_PREPARED: "1",
      CARTULARY_TEST_RESULTS_DIR: resultsRoot,
      CARTULARY_TEST_RUN_ID: "prepared-run",
      CARTULARY_TEST_TARGET: "backend-unit",
    };
    assert.equal(validatePreparedArtifactIdentity("backend-unit", complete), true);
    assert.equal(validatePreparedArtifactIdentity("backend-unit", {}), false);
    assert.throws(
      () => validatePreparedArtifactIdentity("backend-unit", {
        ...complete,
        CARTULARY_HARNESS_IDENTITY_PREPARED: "true",
      }),
      /must be exactly 1/u,
    );
    assert.throws(
      () => validatePreparedArtifactIdentity("backend-unit", {
        ...complete,
        CARTULARY_TEST_TARGET: "frontend-unit",
      }),
      /does not match/u,
    );
    assert.throws(
      () => preflightPublicTarget("backend-unit", {
        ...complete,
        CARTULARY_TEST_RUN_ID: "",
      }),
      /prepared harness artifact identity is incomplete/u,
    );
    assert.equal(
      existsSync(path.join(resultsRoot, "prepared-run")),
      false,
      "partial identity must fail before creating its run root",
    );
    const prepared = preflightPublicTarget("backend-unit", complete);
    const invocationStart = JSON.parse(readFileSync(
      path.join(prepared.run_root, "_shared", "harness-invocation-start.json"),
      "utf8",
    ));
    validateSchemaSync("cartulary.harness_invocation_start.v1", invocationStart);
    assert.equal(invocationStart.target, "backend-unit");
    assert.ok(
      invocationStart.invocation_edges.every((edge, index, rows) =>
        index === 0 ||
        rows[index - 1].parent_target.localeCompare(edge.parent_target) < 0 ||
        (
          rows[index - 1].parent_target === edge.parent_target &&
          rows[index - 1].child_target.localeCompare(edge.child_target) < 0
        )),
      "retained invocation edges must be sorted",
    );
    writeFileSync(path.join(prepared.run_root, "retained.txt"), "retained\n", "utf8");
    assert.equal(
      preflightPublicTarget("backend-unit", complete).run_root,
      prepared.run_root,
      "complete prepared identity may reuse its run root",
    );
    assert.throws(
      () => preflightPublicTarget("backend-unit", {
        ...complete,
        CARTULARY_HARNESS_IDENTITY_PREPARED: "",
      }),
      /non-empty run root/u,
      "an unprepared caller cannot collide with retained artifacts",
    );
  } finally {
    rmSync(resultsRoot, { recursive: true, force: true });
  }
});

test("owner slice children cannot inherit parent identity, selectors, or Make overrides", () => {
  const child = ownerSliceChildEnvironment({
    PATH: "/bin",
    OWNER: "harness.generated_artifacts",
    ROWS: "row.one",
    JSON: "1",
    PLAYWRIGHT_WORKERS: "7",
    VITEST_MAX_WORKERS: "8",
    MAKEFLAGS: " -- OWNER=harness.generated_artifacts",
    MFLAGS: "--no-print-directory",
    GNUMAKEFLAGS: "--warn-undefined-variables",
    MAKEOVERRIDES: "${-*-command-variables-*-}",
    CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli ROWS=cli",
    CARTULARY_TEST_CATALOG_ROW_IDS: "row.one",
    CARTULARY_TEST_OWNER: "harness.generated_artifacts",
    CARTULARY_TEST_RESULTS_DIR: "/tmp/parent-results",
    CARTULARY_TEST_RUN_ID: "parent-run",
    CARTULARY_TEST_TARGET: "test-slice",
    CARTULARY_HARNESS_IDENTITY_PREPARED: "1",
  }, {
    childResultsRoot: "/tmp/parent-results/parent-run/test-slice/child-results",
  });
  assert.deepEqual(child, {
    PATH: "/bin",
    CARTULARY_TEST_RESULTS_DIR: "/tmp/parent-results/parent-run/test-slice/child-results",
  });
  assert.match(ownerChildRunID("unit.playwright.browser", 0), /^owner-[a-f0-9]{16}-001$/u);
  assert.notEqual(
    ownerChildRunID("unit.playwright.browser", 0),
    ownerChildRunID("unit.playwright.browser", 1),
    "each child invocation receives a distinct run ID",
  );
  assert.notEqual(
    ownerChildRunID("unit.playwright.browser", 0),
    ownerChildRunID("unit.go.backend", 0),
    "work units cannot collide on child run ID",
  );
});

test("stateful owner browser rows execute as isolated single-worker partitions", () => {
  const artifactRoot = mkdtempSync(path.join(repoRoot, "tmp", "owner-stateful-plan-test."));
  try {
    const invocations = buildPlaywrightInvocations(
      repoRoot,
      { work_unit_id: "stateful.auth", target_name: "browser-e2e-stateful" },
      [
        {
          row_id: "row.stateful.one",
          selector: {
            file: "apps/web/e2e/one.spec.ts",
            project_id: "chromium",
            titles: ["stateful one"],
          },
        },
        {
          row_id: "row.stateful.two",
          selector: {
            file: "apps/web/e2e/two.spec.ts",
            project_id: "chromium",
            titles: ["stateful two"],
          },
        },
      ],
      { playwright: 3 },
      artifactRoot,
    );
    assert.equal(invocations.length, 2);
    assert.deepEqual(invocations.map((invocation) => invocation.rows.length), [1, 1]);
    assert.deepEqual(invocations.map((invocation) =>
      invocation.args[invocation.args.indexOf("--workers") + 1]), ["1", "1"]);
    assert.equal(new Set(invocations.map((invocation) => invocation.reportPath)).size, 2);
    assert.equal(new Set(invocations.map((invocation) => invocation.browserSessionGroup)).size, 2);
  } finally {
    rmSync(artifactRoot, { recursive: true, force: true });
  }
});

test("owner accounting closes exact rows and preserves subset completion scope", async () => {
  const catalog = loadTestCatalog(repoRoot);
  const row = catalog.rows.find((entry) => entry.owner_id === "platform.config" && entry.runner === "go");
  assert.ok(row);
  const plan = buildOwnerSlicePlan(repoRoot, {
    ownerID: row.owner_id,
    dependencyScope: "all",
    rows: row.row_id,
    rowsProvided: true,
    target: "test-slice",
    commandID: "cartulary.harness.command.test_slice.v1",
    runID: "accounting-fixture",
    timestamp: "2026-07-18T00:00:00.000Z",
  });
  const execution = {
    status: "pass",
    duration_ms: 4,
    row_results: [{
      row_id: row.row_id,
      terminal_state: "passed",
      duration_ms: 4,
      exit_code: 0,
      failure_reason: null,
      attempt: 1,
    }],
  };
  const logs = [{
    work_unit_id: plan.work_units[0].work_unit_id,
    stdout_path: "test-slice/work-units/fixture.stdout.log",
    stderr_path: "test-slice/work-units/fixture.stderr.log",
  }];
  const accounting = buildTestEvidenceAccounting(
    plan,
    execution,
    logs,
    "2026-07-18T00:00:00.000Z",
    "2026-07-18T00:00:00.004Z",
  );
  assert.equal(accounting.status, "pass");
  assert.deepEqual(accounting.expected_rows.map((entry) => entry.row_id), [row.row_id]);
  assert.deepEqual(accounting.observed_rows.map((entry) => entry.row_id), [row.row_id]);
  const summary = buildTestOwnerSummary(plan, accounting, {
    evidence_accounting: "test-slice/owners/platform.config/test-evidence-accounting.json",
    owner_summary: "test-slice/owners/platform.config/test-owner-summary.json",
    plan: "test-slice/test-slice-plan.json",
    scheduler_summary: "test-slice/test-slice-scheduler-summary.json",
    tool_run_summary: "test-slice/tool-run-summary.json",
  });
  assert.equal(summary.completion_scope, "selected_subset");
  assert.equal(summary.counts.passed, 1);
  assert.equal(summary.primary_failure, null);
  await validateSchema(accounting.schema_id, accounting);
  await validateSchema(summary.schema_id, summary);
  assert.throws(
    () => buildTestEvidenceAccounting(plan, { ...execution, row_results: [] }, logs, accounting.started_at, accounting.finished_at),
    /do not exactly match/u,
  );
});

test("target evidence finalization closes exact runner observations and scope", async () => {
  const resultsRoot = mkdtempSync(
    path.join(repoRoot, "tmp", "target-evidence-finalization."),
  );
  try {
    const catalog = loadTestCatalog(repoRoot);
    const taskSurface = readJSON("tools/task_surface_manifest.json");
    const targetByCommand = new Map(
      taskSurface.targets.map((entry) => [entry.command_id, entry.name]),
    );
    const targetForRow = (row) =>
      evidenceTargetForCatalogRow(row, { commandTargetByID: targetByCommand });
    const cases = [
      ["go", "go", "backend-unit"],
      ["go-support", "go", "backend-integration-support"],
      ["vitest", "vitest", "frontend-unit"],
      ["playwright", "playwright", "browser-e2e-webserver-backed"],
      ["shell", "shell", "generate-drift"],
    ].map(([caseID, runner, targetID]) => {
      const row = catalog.rows.find(
        (entry) => entry.runner === runner && targetForRow(entry) === targetID,
      );
      assert.ok(row, `${runner} target evidence fixture row`);
      return { caseID, row, runner, targetID };
    });

    for (const { caseID, row, runner, targetID } of cases) {
      const runID = `target-evidence-${caseID}`;
      const targetDir = path.join(resultsRoot, runID, targetID);
      if (runner === "go") {
        writeJSONFile(path.join(targetDir, "fixture", "step-summary.json"), {
          started_at: "2026-07-18T00:00:00.000Z",
          finished_at: "2026-07-18T00:00:00.010Z",
          inventory: row.selector.tests.map((symbol) => ({
            id: row.row_id,
            symbol_or_title: symbol,
          })),
        });
      } else if (runner === "vitest") {
        writeJSONFile(
          path.join(targetDir, "raw", "frontend-unit", "runner.json"),
          {
            startTime: Date.parse("2026-07-18T00:00:00.000Z"),
            testResults: [{
              name: path.join(repoRoot, row.selector.file),
              endTime: Date.parse("2026-07-18T00:00:00.010Z"),
              assertionResults: row.selector.titles.map((title) => ({
                fullName: title,
                status: "passed",
                duration: 1,
              })),
            }],
          },
        );
      } else if (runner === "playwright") {
        writeJSONFile(
          path.join(
            targetDir,
            "browser-groups",
            "fixture",
            "browser-group-result.json",
          ),
          {
            schema_id: "cartulary.browser_group_result.v1",
            target_id: targetID,
            stage_id: "webserver-backed",
            group_id: "target-evidence-fixture",
            runtime_profile_id: row.runtime_profile_id,
            fixture_profile_ids: [row.fixture_profile_id],
            resource_profile_ids: [row.resource_profile_id],
            selected_rows: [row.row_id],
            started_at: "2026-07-18T00:00:00.000Z",
            finished_at: "2026-07-18T00:00:00.010Z",
            duration_ms: 10,
            status: "pass",
            exit_code: 0,
            row_results: [{
              row_id: row.row_id,
              terminal_state: "passed",
              duration_ms: 10,
              exit_code: 0,
              failure_reason: null,
            }],
            artifacts: {
              playwright_report: `${targetID}/browser-groups/fixture/report.json`,
              stdout: `${targetID}/browser-groups/fixture/stdout.log`,
              stderr: `${targetID}/browser-groups/fixture/stderr.log`,
            },
          },
        );
      }

      const options = {
        targetID,
        requestedStatus: "pass",
        resultsDir: resultsRoot,
        runID,
        env: {
          CARTULARY_TARGET_EVIDENCE_SCOPE: "rows",
          CARTULARY_TARGET_EVIDENCE_ROW_IDS: row.row_id,
        },
      };
      const finalized = finalizeTargetOwnerEvidence(repoRoot, options);
      assert.equal(finalized.status, "pass");
      assert.equal(finalized.reused, false);
      assert.deepEqual(finalized.shards.flatMap((shard) => shard.selected_rows), [
        row.row_id,
      ]);
      const ownerDir = path.join(targetDir, "owners", row.owner_id);
      const accounting = JSON.parse(
        readFileSync(path.join(ownerDir, "test-evidence-accounting.json")),
      );
      assert.equal(accounting.status, "pass");
      assert.equal(accounting.target_id, targetID);
      if (targetID === "backend-integration-support") {
        assert.equal(
          accounting.command_id,
          "cartulary.harness.command.backend_integration.v1",
        );
      }
      assert.equal(
        JSON.parse(readFileSync(path.join(ownerDir, "test-owner-summary.json")))
          .status,
        "pass",
      );
      assert.equal(finalizeTargetOwnerEvidence(repoRoot, options).reused, true);
    }

    const goCase = cases.find((entry) => entry.runner === "go");
    assert.ok(goCase);
    assert.throws(
      () =>
        finalizeTargetOwnerEvidence(repoRoot, {
          targetID: goCase.targetID,
          requestedStatus: "pass",
          resultsDir: resultsRoot,
          runID: "target-evidence-missing",
          env: {
            CARTULARY_TARGET_EVIDENCE_SCOPE: "rows",
            CARTULARY_TARGET_EVIDENCE_ROW_IDS: goCase.row.row_id,
          },
        }),
      /Go selector mismatch/u,
    );
    assert.equal(
      existsSync(
        path.join(
          resultsRoot,
          "target-evidence-missing",
          goCase.targetID,
          "owners",
        ),
      ),
      false,
    );
    assert.throws(
      () =>
        finalizeTargetOwnerEvidence(repoRoot, {
          targetID: goCase.targetID,
          requestedStatus: "pass",
          resultsDir: resultsRoot,
          runID: "target-evidence-contradictory",
          env: {
            CARTULARY_TARGET_EVIDENCE_SCOPE: "rows",
            CARTULARY_TARGET_EVIDENCE_ROW_IDS: goCase.row.row_id,
            CARTULARY_GO_SCHEDULE_SCOPE: "all",
          },
        }),
      /selection scopes disagree/u,
    );
    assert.equal(
      existsSync(
        path.join(
          resultsRoot,
          "target-evidence-contradictory",
          goCase.targetID,
        ),
      ),
      false,
    );
    assert.throws(
      () =>
        finalizeTargetOwnerEvidence(repoRoot, {
          targetID: goCase.targetID,
          requestedStatus: "pass",
          resultsDir: resultsRoot,
          runID: "target-evidence-go",
          env: { CARTULARY_TARGET_EVIDENCE_SCOPE: "all" },
        }),
      /do not match the selected row scope/u,
    );
  } finally {
    rmSync(resultsRoot, { recursive: true, force: true });
  }
});

test("owner evidence audit accepts exact target partitions and rejects duplicate row evidence", () => {
  const resultRoot = mkdtempSync(path.join(repoRoot, "tmp", "owner-audit."));
  try {
    const ownerID = "platform.config";
    const partitions = deriveRequiredEvidencePartitions(repoRoot, ownerID);
    const catalog = loadTestCatalog(repoRoot);
    const source = buildSourceSnapshot(repoRoot);
    const entries = [];
    const writeAccounting = (runRoot, targetID, rowIDs) => {
      const ownerDir = path.join(runRoot, targetID, "owners", ownerID);
      mkdirSync(ownerDir, { recursive: true });
      const identity = loadOwnerAccountingSelection(repoRoot, { ownerID, rowIDs });
      writeJSONFile(path.join(ownerDir, "test-evidence-accounting.json"), {
        schema_id: "cartulary.test_evidence_accounting.v1",
        command_id: targetID === "test-slice"
          ? "cartulary.harness.command.test_slice.v1"
          : "cartulary.harness.command.fixture_target.v1",
        run_id: `fixture-${targetID}`,
        target_id: targetID,
        owner_id: ownerID,
        selected_rows: rowIDs,
        source_snapshot_digest: source.digest,
        catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
        verification_semantic_digest: catalog.summary.verification_semantic_digest,
        runtime_profile_digest: identity.runtime_profile_digest,
        resource_profile_digest: identity.resource_profile_digest,
        fixture_profile_digest: identity.fixture_profile_digest,
        started_at: "2026-07-18T00:00:00.000Z",
        finished_at: "2026-07-18T00:00:00.001Z",
        duration_ms: 1,
        status: "pass",
        expected_rows: identity.expected_rows.map((row) => ({
          row_id: row.row_id,
          owner_id: row.owner_id,
          family_id: row.family_id,
          verification_ids: row.verification_ids,
          runner: row.runner,
          selector_digest: row.selector_digest,
          evidence_class: row.evidence_class,
          evidence_target_id: row.target_name,
          runtime_profile_id: row.runtime_profile_id,
          resource_profile_id: row.resource_profile_id,
          fixture_profile_id: row.fixture_profile_id,
        })),
        observed_rows: rowIDs.map((rowID) => ({
          row_id: rowID,
          terminal_state: "passed",
          logical_duration_ms: 1,
          executed_duration_ms: 1,
          attempts: [{
            attempt: 1,
            terminal_state: "passed",
            exit_code: 0,
            duration_ms: 1,
            artifact_refs: [],
          }],
          failure: null,
        })),
      });
    };
    for (const [targetID, rowIDs] of [...partitions].filter(([targetID]) => targetID !== "test-slice")) {
      const runRoot = path.join(resultRoot, `run-${targetID}`);
      writeAccounting(runRoot, targetID, rowIDs);
      entries.push({ target_id: targetID, run_root: path.relative(repoRoot, runRoot) });
    }
    entries.sort((left, right) => left.target_id < right.target_id ? -1 : left.target_id > right.target_id ? 1 : 0);
    const manifestFile = path.join(resultRoot, "evidence-roots.json");
    writeJSONFile(manifestFile, {
      schema_id: "cartulary.test_evidence_root_manifest.v1",
      owner_id: ownerID,
      entries,
    });
    const summary = auditOwnerEvidence(repoRoot, {
      ownerID,
      manifestPath: path.relative(repoRoot, manifestFile),
      timestamp: "2026-07-18T00:00:00.000Z",
    });
    assert.equal(summary.status, "pass");
    assert.equal(summary.accepted_artifacts.length, partitions.size - 1);

    const sliceRows = partitions.get("test-slice");
    const sliceRunRoot = path.join(resultRoot, "run-test-slice");
    writeAccounting(sliceRunRoot, "test-slice", sliceRows);
    const sliceManifestFile = path.join(resultRoot, "slice-evidence-roots.json");
    writeJSONFile(sliceManifestFile, {
      schema_id: "cartulary.test_evidence_root_manifest.v1",
      owner_id: ownerID,
      entries: [{
        target_id: "test-slice",
        run_root: path.relative(repoRoot, sliceRunRoot),
      }],
    });
    const sliceSummary = auditOwnerEvidence(repoRoot, {
      ownerID,
      manifestPath: path.relative(repoRoot, sliceManifestFile),
      timestamp: "2026-07-18T00:00:00.000Z",
    });
    assert.equal(sliceSummary.status, "pass");
    assert.equal(sliceSummary.counts.required_target_partitions, 1);
    assert.deepEqual(sliceSummary.accepted_artifacts[0].row_ids, sliceRows);

    const cliResults = path.join(resultRoot, "cli-results");
    const cli = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/evidence-accounting/evidence-audit-cli.mjs"),
        "--owner",
        ownerID,
        "--evidence-roots-file",
        path.relative(repoRoot, manifestFile),
      ],
      {
        cwd: repoRoot,
        encoding: "utf8",
        env: {
          ...process.env,
          CARTULARY_TEST_RESULTS_DIR: cliResults,
          CARTULARY_TEST_RUN_ID: "audit-pass",
        },
      },
    );
    assert.equal(cli.status, 0, cli.stderr);
    const retainedAudit = JSON.parse(
      readFileSync(
        path.join(cliResults, "audit-pass", "test-evidence-audit", "test-evidence-audit-summary.json"),
        "utf8",
      ),
    );
    assert.equal(retainedAudit.schema_id, "cartulary.test_evidence_audit_summary.v1");
    assert.equal(retainedAudit.status, "pass");
    assert.ok(retainedAudit.duration_ms > 0);
    assert.equal(
      JSON.parse(
        readFileSync(
          path.join(cliResults, "audit-pass", "test-evidence-audit", "tool-run-summary.json"),
          "utf8",
        ),
      ).schema_id,
      "cartulary.tool_run_summary.v5",
    );

    const first = entries[0];
    const artifactFile = path.join(repoRoot, first.run_root, first.target_id, "owners", ownerID, "test-evidence-accounting.json");
    const artifact = JSON.parse(readFileSync(artifactFile, "utf8"));
    artifact.observed_rows.push(artifact.observed_rows[0]);
    writeJSONFile(artifactFile, artifact);
    const rejected = auditOwnerEvidence(repoRoot, {
      ownerID,
      manifestPath: path.relative(repoRoot, manifestFile),
      timestamp: "2026-07-18T00:00:00.000Z",
    });
    assert.equal(rejected.status, "fail");
    assert.ok(rejected.rejected_artifacts[0].reasons.includes("observed_row_inventory_mismatch"));
  } finally {
    rmSync(resultRoot, { recursive: true, force: true });
  }
});

test("suppressed Make node-tool children do not capture top-level observability", () => {
  const resultRoot = mkdtempSync(path.join(repoRoot, "tmp", "suppressed-node-tool."));
  const runID = "nested-finalizer-child";
  try {
    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/execution/run-make-node-tool-cli.mjs"),
        "go-test-duration-baseline-coverage",
      ],
      {
        cwd: repoRoot,
        encoding: "utf8",
        env: {
          ...process.env,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
          CARTULARY_TEST_RESULTS_DIR: resultRoot,
          CARTULARY_TEST_RUN_ID: runID,
          CARTULARY_TEST_TARGET: "go-test-duration-baseline-coverage",
        },
      },
    );
    assert.equal(result.status, 0, result.stderr);
    assert.equal(
      existsSync(path.join(resultRoot, runID, "_shared", "harness-observability")),
      false,
    );
    assert.equal(
      existsSync(
        path.join(
          resultRoot,
          runID,
          "go-test-duration-baseline-coverage",
          "tool-run-summary.json",
        ),
      ),
      true,
    );
  } finally {
    rmSync(resultRoot, { recursive: true, force: true });
  }
});

test("owner diagnostics project exact catalog topology and evidence-derived guidance", async () => {
  const catalog = loadTestCatalog(repoRoot);
  const ownerRows = catalog.rows.filter((row) => row.owner_id === "module.networkflow");
  const explanation = explainTestOwner(repoRoot, "module.networkflow");
  assert.equal(explanation.owner_id, "module.networkflow");
  assert.equal(explanation.row_count, ownerRows.length);
  assert.equal(
    explanation.service_backed_row_count,
    ownerRows.filter((row) => {
      const profile = catalog.profiles.semantic.runtime_profiles.find(
        (entry) => entry.id === row.runtime_profile_id,
      );
      return profile.managed_service_ids.length > 0;
    }).length,
  );
  assert.equal(
    explanation.families.reduce((sum, family) => sum + family.row_count, 0),
    ownerRows.length,
  );
  assert.equal(
    Object.values(explanation.runner_counts).reduce((sum, count) => sum + count, 0),
    ownerRows.length,
  );
  assert.equal(
    explanation.default_check.included + explanation.default_check.excluded,
    ownerRows.length,
  );
  assert.equal(explanation.commands.full_owner, "make test-slice OWNER=module.networkflow");
  assert.equal(
    explanation.commands.service_backed,
    "make service-backed-test-slice OWNER=module.networkflow",
  );
  assert.doesNotMatch(JSON.stringify(explanation), /(?:guide_path|guide_digest)/u);
  await validateSchema(explanation.schema_id, explanation);

  const guide = buildModuleAuthorTaskGuide(repoRoot, "module.networkflow", "module-author");
  assert.deepEqual(guide.focused_commands, [
    "make test-slice OWNER=module.networkflow",
    "make service-backed-test-slice OWNER=module.networkflow",
  ]);
  assert.ok(guide.broader_commands.includes("make test-fast"));
  assert.ok(guide.broader_commands.includes("make browser-e2e-stateful"));
  assert.doesNotMatch(
    JSON.stringify(guide),
    /(?:guide_path|guide_digest)/u,
  );
  await validateSchema(guide.schema_id, guide);
  assert.throws(
    () => buildModuleAuthorTaskGuide(repoRoot, "module.networkflow", "phase-author"),
    /module-author/u,
  );
  assert.throws(() => explainTestOwner(repoRoot, "module.unknown"), /unknown active test owner/u);
});

test("owner diagnostic CLI emits one JSON object and retains no artifacts", () => {
  const resultRoot = mkdtempSync(path.join(repoRoot, "tmp", "owner-diagnostics."));
  try {
    const cli = path.join(repoRoot, "tools/harness/diagnostics/owner-diagnostics-cli.mjs");
    const explanation = spawnSync(
      process.execPath,
      [cli, "--mode", "explain", "--owner", "platform.config", "--json"],
      {
        cwd: repoRoot,
        encoding: "utf8",
        env: { ...process.env, CARTULARY_TEST_RESULTS_DIR: resultRoot },
      },
    );
    assert.equal(explanation.status, 0, explanation.stderr);
    assert.equal(explanation.stdout.split("\n").filter(Boolean).length, 1);
    assert.equal(JSON.parse(explanation.stdout).schema_id, "cartulary.test_owner_explanation.v1");

    const guide = spawnSync(
      process.execPath,
      [
        cli,
        "--mode",
        "task-guide",
        "--role",
        "module-author",
        "--owner",
        "platform.config",
        "--json",
      ],
      {
        cwd: repoRoot,
        encoding: "utf8",
        env: { ...process.env, CARTULARY_TEST_RESULTS_DIR: resultRoot },
      },
    );
    assert.equal(guide.status, 0, guide.stderr);
    assert.equal(guide.stdout.split("\n").filter(Boolean).length, 1);
    assert.equal(JSON.parse(guide.stdout).schema_id, "cartulary.task_guide_summary.v2");

    for (const fixture of [
      {
        args: ["--mode", "explain", "--owner", "platform.config", "--json-value", "true"],
        env: process.env,
        expected: /JSON accepts only exact 1/u,
      },
      {
        args: ["--mode", "explain", "--owner", "platform.config", "--json"],
        env: { ...process.env, CARTULARY_OUTPUT_MODE: "machine" },
        expected: /cannot be combined/u,
      },
    ]) {
      const invalid = spawnSync(process.execPath, [cli, ...fixture.args], {
        cwd: repoRoot,
        encoding: "utf8",
        env: fixture.env,
      });
      assert.equal(invalid.status, 2);
      assert.match(invalid.stderr, fixture.expected);
      assert.equal(invalid.stdout, "");
    }
    assert.deepEqual(readdirSync(resultRoot), []);
  } finally {
    rmSync(resultRoot, { recursive: true, force: true });
  }
});

test("owner evidence accounting rejects duplicate, foreign, and target-incompatible rows", () => {
  const catalog = loadTestCatalog(repoRoot);
  const owned = catalog.rows.find((row) => row.owner_id === "module.networkflow" && row.runner === "vitest");
  const foreign = catalog.rows.find((row) => row.owner_id !== "module.networkflow");
  assert.ok(owned);
  assert.ok(foreign);
  assert.throws(
    () => loadOwnerAccountingSelection(repoRoot, {
      ownerID: "module.networkflow",
      rowIDs: [owned.row_id, owned.row_id],
    }),
    /duplicate/u,
  );
  assert.throws(
    () => loadOwnerAccountingSelection(repoRoot, {
      ownerID: "module.networkflow",
      rowIDs: [foreign.row_id],
    }),
    /does not belong/u,
  );
  assert.throws(
    () => loadOwnerAccountingSelection(repoRoot, {
      ownerID: "module.networkflow",
      rowIDs: [owned.row_id],
      targetName: "browser-e2e-a11y",
    }),
    /is not selected by evidence target/u,
  );
});

test("runner selector resolvers preserve exact closed shapes across all runners", () => {
  const runnerRegistry = readJSON("tools/test_runner_registry.json");
  const runnerByID = new Map(runnerRegistry.runners.map((entry) => [entry.runner, entry]));
  const commandIDs = new Set(
    readJSON("tools/task_surface_manifest.json").targets.map((entry) => entry.command_id),
  );
  const fixtures = [
    {
      row: {
        row_id: "module.fixture.behavior.go",
        runner: "go",
        selector: {
          package: "./internal/modules/graphprojection/fixturetest",
          tests: ["TestContainedPathRejectsTraversal"],
        },
      },
      expected: ["go:./internal/modules/graphprojection/fixturetest:TestContainedPathRejectsTraversal"],
    },
    {
      row: {
        row_id: "web.fixture.behavior.vitest",
        runner: "vitest",
        selector: {
          file: "apps/web/src/networkFlow/networkFlowClient.test.ts",
          titles: ["networkFlowClient route identity uses extension workspace route state for Network Analysis"],
        },
      },
      expected: [
        "vitest:apps/web/src/networkFlow/networkFlowClient.test.ts:networkFlowClient route identity uses extension workspace route state for Network Analysis",
      ],
    },
    {
      row: {
        row_id: "web.fixture.behavior.playwright",
        runner: "playwright",
        selector: {
          file: "apps/web/e2e/workbook.visual.spec.ts",
          project_id: "chromium",
          stage: "visual",
          scenario_ids: ["timeline_default_viewport"],
          titles: ["captures the Timeline default viewport with stable row version and save-state strip"],
        },
      },
      expected: ["playwright:chromium:visual:timeline_default_viewport"],
    },
    {
      row: {
        row_id: "harness.fixture.behavior.shell",
        runner: "shell",
        selector: { command_id: "cartulary.harness.command.help.v1" },
      },
      expected: ["shell:cartulary.harness.command.help.v1"],
    },
  ];
  for (const fixture of fixtures) {
    assert.deepEqual(
      resolveRowSelector({
        root: repoRoot,
        row: fixture.row,
        runner: runnerByID.get(fixture.row.runner),
        taskSurfaceCommandIDs: commandIDs,
      }),
      fixture.expected,
    );
  }
});

test("owner catalog rejects structural, reference, selector, and path ambiguity", () => {
  const cases = [
    {
      name: "zero-row owner",
      mutate: ({ familyManifest }) => { familyManifest.rows = []; },
      pattern: /must NOT have fewer than 1 items|must not be empty/iu,
    },
    {
      name: "duplicate owner",
      mutate: ({ ownerRegistry }) => { ownerRegistry.owners.push(structuredClone(ownerRegistry.owners[0])); },
      pattern: /must not contain duplicates/iu,
    },
    {
      name: "delivery-phase row ID",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].row_id = "module.fixture.semantic.owned"; },
      pattern: /must match pattern|row_id/iu,
    },
    {
      name: "unresolved verification",
      mutate: ({ familyManifest }) => {
        familyManifest.rows[0].verification_ids = ["module.fixture.verification.missing"];
      },
      pattern: /references unknown/iu,
    },
    {
      name: "unresolved collaborator",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].collaborator_ids = ["module.missing"]; },
      pattern: /collaborator_ids references unknown/iu,
    },
    {
      name: "unknown profile",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].runtime_profile_id = "unknown"; },
      pattern: /must be equal to one of the allowed values|runtime_profile_id/iu,
    },
    {
      name: "postgres fixture without postgres runtime",
      mutate: ({ familyManifest }) => {
        familyManifest.rows[0].fixture_profile_id = "postgres_group_clone";
      },
      pattern: /fixture_profile_id requires a postgres runtime profile/iu,
    },
    {
      name: "mutated profile definition",
      mutate: ({ topology }) => { topology.resource_profiles[1].resource_claims.go_cpu = 2; },
      pattern: /closed profile definitions/iu,
    },
    {
      name: "zero-match selector",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].selector.tests = ["TestMissing"]; },
      pattern: /must resolve exactly once/iu,
    },
    {
      name: "multiple-match selector",
      mutate: () => {},
      setup: (root) => {
        writeFixtureFile(root, "internal/fixture/duplicate_test.go", "package fixture\n\nfunc TestOwned(t *testing.T) {}\n");
      },
      pattern: /must resolve exactly once/iu,
    },
    {
      name: "overlapping selector",
      mutate: ({ familyManifest }) => {
        const duplicate = structuredClone(familyManifest.rows[0]);
        duplicate.row_id = "module.fixture.behavior.overlap";
        familyManifest.rows.push(duplicate);
        familyManifest.rows.sort((left, right) => left.row_id.localeCompare(right.row_id));
      },
      pattern: /selector overlaps/iu,
    },
    {
      name: "glob selector",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].selector.package = "./internal/*"; },
      pattern: /must match pattern|exact repository package/iu,
    },
    {
      name: "traversal selector",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].selector.package = "./internal/other/../fixture"; },
      pattern: /traversal|normalization drift/iu,
    },
    {
      name: "regex selector",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].selector.tests = ["/Test.*/"]; },
      pattern: /must match pattern|must resolve exactly once/iu,
    },
    {
      name: "unsupported runner",
      mutate: ({ familyManifest }) => { familyManifest.rows[0].runner = "python"; },
      pattern: /must be equal to one of the allowed values|unsupported/iu,
    },
    {
      name: "runner adapter mismatch",
      mutate: ({ runnerRegistry }) => {
        runnerRegistry.runners.find((entry) => entry.runner === "go").adapter_path =
          "tools/harness/execution/runners/shell.mjs";
      },
      pattern: /closed runner definition/iu,
    },
  ];
  for (const fixtureCase of cases) {
    const root = mkdtempSync(path.join(repoRoot, "tmp", "owner-catalog."));
    try {
      writeOwnerCatalogFixture(root, fixtureCase.mutate);
      fixtureCase.setup?.(root);
      assert.throws(() => loadTestCatalog(root), fixtureCase.pattern, fixtureCase.name);
    } finally {
      rmSync(root, { recursive: true, force: true });
    }
  }
});

test("owner catalog rejects symlinked selector roots", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "owner-catalog-symlink."));
  try {
    writeOwnerCatalogFixture(root);
    writeFixtureFile(root, "internal/fixture-real/owned_test.go", "package fixture\n\nfunc TestOwned(t *testing.T) {}\n");
    rmSync(path.join(root, "internal/fixture"), { recursive: true, force: true });
    symlinkSync(path.join(root, "internal/fixture-real"), path.join(root, "internal/fixture"));
    assert.throws(() => loadTestCatalog(root), /symbolic link|non-symlink directory/iu);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("semantic JSON rejects ambiguous encodings and ignores display metadata", () => {
  assert.throws(() => parseStrictJSON('{"a":1,"a":2}', "fixture"), /duplicate object member/iu);
  assert.throws(() => parseStrictJSON('{"a":-0}', "fixture"), /negative zero/iu);
  assert.throws(() => parseStrictJSON('{"a":9007199254740992}', "fixture"), /interoperable range/iu);
  assert.equal(
    semanticJSONDigest({ owner_id: "module.fixture", status: "active" }),
    semanticJSONDigest({ status: "active", owner_id: "module.fixture" }),
  );
});

test("semantic identity violations fail without an exception registry", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "semantic-identity-negative."));
  try {
    const deliveryShapedName = ["pha", "se2_test.go"].join("");
    writeFixtureFile(
      root,
      `internal/fixture/${deliveryShapedName}`,
      "package fixture\n\nfunc TestProductStage(t *testing.T) {}\n",
    );
    assert.deepEqual(validateSemanticIdentities(root), [
      {
        location: `internal/fixture/${deliveryShapedName}`,
        locator_kind: "filename",
        locator: deliveryShapedName,
        reason: "test filename encodes a delivery phase or sprint",
      },
    ]);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("semantic identity validation distinguishes fixture routing from product payloads", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "semantic-fixture-identity-negative."));
  try {
    writeFixtureFile(
      root,
      "apps/web/e2e/fixture.spec.ts",
      [
        'test("incident phase transition remains product vocabulary", async () => {',
        '  const productPayload = { label: "Phase 2" };',
        '  uniqueTxn("module.fixture.integration");',
        '  return productPayload;',
        '});',
        "",
      ].join("\n"),
    );
    assert.deepEqual(validateSemanticIdentities(root), [
      {
        location: "apps/web/e2e/fixture.spec.ts",
        locator_kind: "fixture_identity",
        locator: "uniqueTxn=module.fixture.integration",
        reason: "identity-bearing fixture metadata embeds a catalog identity",
      },
    ]);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("semantic identity validation rejects Go fixture routing without banning product payloads", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "semantic-go-fixture-identity-negative."));
  try {
    writeFixtureFile(
      root,
      "internal/fixture/identity_test.go",
      [
        "package fixture",
        "",
        "func exerciseFixture(t *testing.T) {",
        '  productPayload := "Phase 2"',
        '  StartStore(t, "module.fixture.integration")',
        '  request := Request{ClientTxnID: "txn-sprint8-conflict"}',
        "  _, _ = productPayload, request",
        "}",
        "",
      ].join("\n"),
    );
    assert.deepEqual(validateSemanticIdentities(root), [
      {
        location: "internal/fixture/identity_test.go",
        locator_kind: "fixture_identity",
        locator: "line 5:module.fixture.integration",
        reason: "identity-bearing Go fixture metadata embeds a catalog identity",
      },
      {
        location: "internal/fixture/identity_test.go",
        locator_kind: "fixture_identity",
        locator: "line 6:txn-sprint8-conflict",
        reason: "identity-bearing Go fixture metadata encodes a delivery phase or sprint",
      },
    ]);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("test catalog implementation cannot depend on execution or accounting layers", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "owner-catalog-imports."));
  try {
    writeFixtureFile(
      root,
      "tools/harness/test-catalog/invalid.mjs",
      'import "../scheduler/scheduler-resources.mjs";\n',
    );
    assert.deepEqual(collectTestCatalogImportViolations(root), [
      {
        file: "tools/harness/test-catalog/invalid.mjs",
        specifier: "../scheduler/scheduler-resources.mjs",
      },
    ]);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("documentation boundary rejects direct, computed, and symlinked reads", () => {
  const documentationDir = ["do", "cs"].join("");
  const direct = scanDocumentationReadSource(
    "tools/fixture/direct.mjs",
    `readFileSync(path.join(root, "${documentationDir}", "spec", "owner.md"), "utf8");`,
  );
  assert.equal(direct.length, 1);
  assert.equal(direct[0].operation, "read_file");

  const computed = scanDocumentationReadSource(
    "tools/fixture/computed.mjs",
    `const ownerPath = path.join(root, "${documentationDir}", "owner.md");\nstatSync(ownerPath);`,
  );
  assert.equal(computed.length, 1);
  assert.equal(computed[0].operation, "stat_path");

  const helperMediated = scanDocumentationReadSource(
    "tools/fixture/helper.mjs",
    `function loadOwner(file) { return readFileSync(file, "utf8"); }\nconst ownerPath = path.join(root, "${documentationDir}", "owner.md");\nloadOwner(ownerPath);`,
  );
  assert.equal(helperMediated.length, 1);
  assert.equal(helperMediated[0].operation, "read_file");

  const root = mkdtempSync(path.join(repoRoot, "tmp", "documentation-boundary."));
  try {
    writeFixtureFile(root, `${documentationDir}/spec/owner.md`, "# Owner\n");
    symlinkSync(path.join(root, documentationDir), path.join(root, "machine-link"));
    assert.throws(
      () =>
        assertDocumentationAccessAllowed({
          root,
          consumerPath: "tools/fixture/reader.mjs",
          operation: "read_file",
          candidatePath: `${documentationDir}/spec/owner.md`,
          exceptions: { exceptions: [] },
        }),
      /boundary_policy_violation/u,
    );
    assert.throws(
      () =>
        assertDocumentationAccessAllowed({
          root,
          consumerPath: "tools/fixture/reader.mjs",
          operation: "resolve_realpath",
          candidatePath: "machine-link/spec/owner.md",
          exceptions: { exceptions: [] },
        }),
      /boundary_policy_violation/u,
    );
    assert.doesNotThrow(() =>
      assertDocumentationAccessAllowed({
        root,
        consumerPath: "tools/docs/lint.mjs",
        operation: "read_file",
        candidatePath: `${documentationDir}/spec/owner.md`,
        exceptions: {
          exceptions: [
            {
              consumer_path: "tools/docs/lint.mjs",
              documentation_pattern: `^${documentationDir}/.*$`,
              operations: ["read_file"],
            },
          ],
        },
      }),
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

const fixtureImportKeyword = ["im", "port"].join("");
const fixtureExportKeyword = ["ex", "port"].join("");

function fixtureImport(specifier) {
  return `${fixtureImportKeyword} "${specifier}";\n`;
}

function fixtureDynamicImport(specifier) {
  return `${fixtureImportKeyword}("${specifier}")`;
}

function fixtureExportFrom(symbol, specifier) {
  return `${fixtureExportKeyword} { ${symbol} } from "${specifier}";\n`;
}

function legacyPlanningImport(file) {
  return ["..", "planning", file].join("/");
}

function legacyPlanningPath(file) {
  return ["tools", "harness", "planning", file].join("/");
}

test("network flow fixture manifest schema is closed and byte-addressed", async () => {
  function sha256(content) {
    return createHash("sha256").update(content).digest("hex");
  }

  function bundleHash(entries) {
    const hash = createHash("sha256");
    for (const entry of entries) {
      hash.update(entry.logical_path, "utf8");
      hash.update(Buffer.from([0]));
      hash.update(entry.sha256, "utf8");
      hash.update(Buffer.from([0]));
      hash.update(String(entry.size_bytes), "utf8");
      hash.update("\n", "utf8");
    }
    return hash.digest("hex");
  }

  const sourceContent = "src_ip,dst_ip\n10.0.0.1,10.0.0.2\n";
  const expectedContent = "{\"rows\":[]}\n";
  const transcriptContent = "{\"status\":\"pass\"}\n";
  const sourceFile = {
    logical_path: "source/cisco-sna-minimal.csv",
    media_type: "text/csv",
    size_bytes: Buffer.byteLength(sourceContent),
    sha256: sha256(sourceContent),
    role: "input",
    newline_policy: "lf_required",
  };
  const expectedFile = {
    logical_path: "expected/rows.json",
    media_type: "application/json",
    size_bytes: Buffer.byteLength(expectedContent),
    sha256: sha256(expectedContent),
    role: "expected_rows",
  };
  const transcriptFile = {
    logical_path: "transcripts/apply.jsonl",
    media_type: "application/jsonl",
    size_bytes: Buffer.byteLength(transcriptContent),
    sha256: sha256(transcriptContent),
    transcript_kind: "apply",
  };
  const manifest = {
    schema_id: "cartulary.network_flow_fixture_manifest.v1",
    manifest_version: 1,
    fixture_id: "NF-FIX-001-cisco-sna-minimal",
    profile_id: "network_flow_activity",
    freeze: {
      status: "frozen",
      revision: 1,
      change_policy: "new_fixture_revision_required",
    },
    verification_ids: [
      "module.networkflow.verification.contract_accounting",
    ],
    source_files: [sourceFile],
    expected_artifacts: [expectedFile],
    transcript_files: [transcriptFile],
    acceptance_ids: ["NF-AC-052"],
    execution_selectors: ["network-flow/acceptance/NF-AC-052"],
    source_bundle_sha256: bundleHash([sourceFile]),
    expected_bundle_sha256: bundleHash([expectedFile, transcriptFile]),
  };

  await validateSchema("cartulary.network_flow_fixture_manifest.v1", manifest);

  await assert.rejects(
    validateSchema("cartulary.network_flow_fixture_manifest.v1", {
      ...manifest,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const invalidDigest = structuredClone(manifest);
  invalidDigest.source_files[0].sha256 = "ABC";
  await assert.rejects(
    validateSchema("cartulary.network_flow_fixture_manifest.v1", invalidDigest),
    /must match pattern/u,
  );

  const root = mkdtempSync(path.join(repoRoot, "tmp", "network-flow-fixture."));
  try {
    const fixtureDir = path.join(root, manifest.fixture_id);
    writeFixtureFile(fixtureDir, sourceFile.logical_path, sourceContent);
    writeFixtureFile(fixtureDir, expectedFile.logical_path, expectedContent);
    writeFixtureFile(fixtureDir, transcriptFile.logical_path, transcriptContent);
    const manifestPath = path.join(fixtureDir, "manifest.json");
    writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);

    const checker = path.join(
      repoRoot,
      "tools/harness/generated-artifacts/check-json-shapes.mjs",
    );
    const pass = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-fixture-manifest", "--file", manifestPath],
      { encoding: "utf8" },
    );
    assert.equal(pass.status, 0, pass.stderr);

    const mismatched = structuredClone(manifest);
    mismatched.source_files[0].sha256 = "0".repeat(64);
    mismatched.source_bundle_sha256 = bundleHash(mismatched.source_files);
    writeFileSync(manifestPath, `${JSON.stringify(mismatched, null, 2)}\n`);
    const fail = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-fixture-manifest", "--file", manifestPath],
      { encoding: "utf8" },
    );
    assert.notEqual(fail.status, 0);
    assert.match(fail.stderr, /sha256 does not match file bytes/u);

    writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
    writeFixtureFile(fixtureDir, "source/unlisted.csv", "extra\n");
    const unlisted = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-fixture-manifest", "--file", manifestPath],
      { encoding: "utf8" },
    );
    assert.notEqual(unlisted.status, 0);
    assert.match(unlisted.stderr, /unlisted fixture file source\/unlisted\.csv/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("contract family registry schema is closed and restricts planned families", async () => {
  const registry = readJSON("contracts/index.json");

  await validateSchema("cartulary.contract_family_registry.v1", registry);

  await assert.rejects(
    validateSchema("cartulary.contract_family_registry.v1", {
      ...registry,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const activatedNetworkFlow = structuredClone(registry);
  activatedNetworkFlow.families = activatedNetworkFlow.families.map((family) =>
    family.family_id === "network-flow"
      ? { ...family, generation_status: "active", activation_dependency_ids: [] }
      : family,
  );
  await validateSchema(
    "cartulary.contract_family_registry.v1",
    activatedNetworkFlow,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(repoRoot, "contracts/index.json");
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "contract-family-registry", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "contract-family."));
  try {
    const mutableArtifactPath = path.join(root, "index.json");
    writeFileSync(
      mutableArtifactPath,
      `${JSON.stringify(activatedNetworkFlow, null, 2)}\n`,
    );
    const activePass = spawnSync(
      process.execPath,
      [checker, "--kind", "contract-family-registry", "--file", mutableArtifactPath],
      { encoding: "utf8" },
    );
    assert.equal(activePass.status, 0, activePass.stderr);

    const unexpectedPlanned = structuredClone(registry);
    unexpectedPlanned.families = unexpectedPlanned.families.map((family) =>
      family.family_id === "errors"
        ? {
            ...family,
            generation_status: "planned",
            activation_dependency_ids: ["NFA-GEN-002"],
          }
        : family,
    );
    writeFileSync(
      mutableArtifactPath,
      `${JSON.stringify(unexpectedPlanned, null, 2)}\n`,
    );
    const fail = spawnSync(
      process.execPath,
      [checker, "--kind", "contract-family-registry", "--file", mutableArtifactPath],
      { encoding: "utf8" },
    );
    assert.notEqual(fail.status, 0);
    assert.match(fail.stderr, /active output_order must be openapi/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow activity accounting is closed and fails drift gaps", async () => {
  const accounting = readJSON("tools/network_flow_activity_accounting.json");

  await validateSchema("cartulary.network_flow_activity_accounting.v2", accounting);

  await assert.rejects(
    validateSchema("cartulary.network_flow_activity_accounting.v2", {
      ...accounting,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );
  const fixtureOnly = structuredClone(accounting);
  fixtureOnly.acceptance_accounting.rows[0].exact_selectors = [];
  await assert.rejects(
    validateSchema("cartulary.network_flow_activity_accounting.v2", fixtureOnly),
    /must NOT have fewer than 1 items/u,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(
    repoRoot,
    "tools/network_flow_activity_accounting.json",
  );
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "network-flow-activity-accounting", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "network-flow-accounting."));
  try {
    const unresolvedSelector = structuredClone(accounting);
    unresolvedSelector.acceptance_accounting.rows[0].exact_selectors[0].title =
      "missing Network Flow browser selector";
    const unresolvedSelectorFile = path.join(root, "unresolved-selector.json");
    writeFileSync(
      unresolvedSelectorFile,
      `${JSON.stringify(unresolvedSelector, null, 2)}\n`,
    );
    const unresolvedSelectorResult = spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-activity-accounting", "--file", unresolvedSelectorFile],
      { encoding: "utf8" },
    );
    assert.notEqual(unresolvedSelectorResult.status, 0);
    assert.match(unresolvedSelectorResult.stderr, /does not resolve/u);

    const missingCopyPath = structuredClone(accounting);
    missingCopyPath.drift_accounting.required_copy_paths = [
      ...missingCopyPath.drift_accounting.required_copy_paths,
      "tools/missing-network-flow-accounting-input.json",
    ];
    const missingCopyPathFile = path.join(root, "missing-copy-path.json");
    writeFileSync(
      missingCopyPathFile,
      `${JSON.stringify(missingCopyPath, null, 2)}\n`,
    );
    const missingCopyPathResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        missingCopyPathFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(missingCopyPathResult.status, 0);
    assert.match(
      missingCopyPathResult.stderr,
      /copy_paths must include tools\/missing-network-flow-accounting-input\.json/u,
    );

    const plannedRegistry = readJSON("contracts/index.json");
    plannedRegistry.families = plannedRegistry.families.map((family) =>
      family.family_id === "network-flow"
        ? {
            ...family,
            generation_status: "planned",
            activation_dependency_ids:
              accounting.contract_registry.planned_activation_dependency_ids,
          }
        : family,
    );
    const plannedRegistryFile = path.join(root, "contracts-index.json");
    writeFileSync(
      plannedRegistryFile,
      `${JSON.stringify(plannedRegistry, null, 2)}\n`,
    );
    const stalePlannedRegistry = structuredClone(accounting);
    stalePlannedRegistry.contract_registry.path = path
      .relative(repoRoot, plannedRegistryFile)
      .split(path.sep)
      .join(path.posix.sep);
    const stalePlannedRegistryFile = path.join(root, "stale-planned-registry.json");
    writeFileSync(
      stalePlannedRegistryFile,
      `${JSON.stringify(stalePlannedRegistry, null, 2)}\n`,
    );
    const stalePlannedRegistryResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        stalePlannedRegistryFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(stalePlannedRegistryResult.status, 0);
    assert.match(
      stalePlannedRegistryResult.stderr,
      /planned but generated outputs contain Network Flow markers/u,
    );

    const activeWithDependencies = readJSON("contracts/index.json");
    activeWithDependencies.families = activeWithDependencies.families.map((family) =>
      family.family_id === "network-flow"
        ? { ...family, activation_dependency_ids: ["NFA-GEN-004"] }
        : family,
    );
    const activeWithDependenciesFile = path.join(
      root,
      "active-with-dependencies.json",
    );
    writeFileSync(
      activeWithDependenciesFile,
      `${JSON.stringify(activeWithDependencies, null, 2)}\n`,
    );
    const activeWithDependenciesAccounting = structuredClone(accounting);
    activeWithDependenciesAccounting.contract_registry.path = path
      .relative(repoRoot, activeWithDependenciesFile)
      .split(path.sep)
      .join(path.posix.sep);
    const activeWithDependenciesAccountingFile = path.join(
      root,
      "active-with-dependencies-accounting.json",
    );
    writeFileSync(
      activeWithDependenciesAccountingFile,
      `${JSON.stringify(activeWithDependenciesAccounting, null, 2)}\n`,
    );
    const activeWithDependenciesResult = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-activity-accounting",
        "--file",
        activeWithDependenciesAccountingFile,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(activeWithDependenciesResult.status, 0);
    assert.match(
      activeWithDependenciesResult.stderr,
      /active but activation dependencies remain/u,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow authored contracts are closed and index-owned", async () => {
  const contractIndex = readJSON("contracts/network-flow/index.json");

  await validateSchema("cartulary.network_flow_contract_index.v1", contractIndex);

  await assert.rejects(
    validateSchema("cartulary.network_flow_contract_index.v1", {
      ...contractIndex,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(repoRoot, "contracts/network-flow/index.json");
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "network-flow-contract-index", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(
    path.join(repoRoot, "contracts/network-flow/tmp-contract."),
  );
  const relativePath = (file) =>
    path.relative(repoRoot, file).split(path.sep).join(path.posix.sep);
  const routes = readJSON("contracts/network-flow/routes.v1.json");
  const schemas = readJSON("contracts/network-flow/schemas.v1.json");
  const errors = readJSON("contracts/network-flow/errors.v1.json");
  const tempIndexPath = path.join(root, "index.json");
  const tempRoutesPath = path.join(root, "routes.v1.json");
  const tempSchemasPath = path.join(root, "schemas.v1.json");
  const tempErrorsPath = path.join(root, "errors.v1.json");
  const tempIndex = {
    ...contractIndex,
    contract_files: {
      ...contractIndex.contract_files,
      routes: relativePath(tempRoutesPath),
      schemas: relativePath(tempSchemasPath),
      errors: relativePath(tempErrorsPath),
    },
  };
  const writeTempContracts = ({
    routeContracts = routes,
    schemaBundle = schemas,
    errorContracts = errors,
  } = {}) => {
    writeJSONFile(tempRoutesPath, routeContracts);
    writeJSONFile(tempSchemasPath, schemaBundle);
    writeJSONFile(tempErrorsPath, errorContracts);
    writeJSONFile(tempIndexPath, tempIndex);
  };
  const runChecker = () =>
    spawnSync(
      process.execPath,
      [checker, "--kind", "network-flow-contract-index", "--file", tempIndexPath],
      { encoding: "utf8" },
    );

  try {
    writeTempContracts();
    const tempPass = runChecker();
    assert.equal(tempPass.status, 0, tempPass.stderr);

    const routePathDrift = structuredClone(routes);
    routePathDrift.routes[0].path =
      "/api/v1/incidents/{incident_id}/network-flow/profiles";
    writeTempContracts({ routeContracts: routePathDrift });
    const routePathResult = runChecker();
    assert.notEqual(routePathResult.status, 0);
    assert.match(routePathResult.stderr, /routes\[1\]\.path/u);

    const unknownPrimaryError = structuredClone(routes);
    unknownPrimaryError.routes[0].primary_errors = ["network_flow_missing"];
    writeTempContracts({ routeContracts: unknownPrimaryError });
    const unknownPrimaryErrorResult = runChecker();
    assert.notEqual(unknownPrimaryErrorResult.status, 0);
    assert.match(unknownPrimaryErrorResult.stderr, /primary_errors/u);

    const openSchema = structuredClone(schemas);
    openSchema.$defs.TableQueryRequest.additionalProperties = true;
    writeTempContracts({ schemaBundle: openSchema });
    const openSchemaResult = runChecker();
    assert.notEqual(openSchemaResult.status, 0);
    assert.match(openSchemaResult.stderr, /additionalProperties must not be true/u);

    const mismatchedSchemaID = structuredClone(schemas);
    mismatchedSchemaID.$defs.TableQueryRequest.properties.schema_id.const =
      "cartulary.network_flow.table_query_request.v0";
    writeTempContracts({ schemaBundle: mismatchedSchemaID });
    const mismatchedSchemaIDResult = runChecker();
    assert.notEqual(mismatchedSchemaIDResult.status, 0);
    assert.match(mismatchedSchemaIDResult.stderr, /schema_id\.const must match/u);

    const missingPublicSchemaID = structuredClone(schemas);
    delete missingPublicSchemaID.$defs.TableQueryRequest.x_schema_id;
    writeTempContracts({ schemaBundle: missingPublicSchemaID });
    const missingPublicSchemaIDResult = runChecker();
    assert.notEqual(missingPublicSchemaIDResult.status, 0);
    assert.match(missingPublicSchemaIDResult.stderr, /public schema IDs mismatch/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow timezone provenance schema is closed and immutable-source scoped", async () => {
  const provenance = readJSON(
    "contracts/network-flow/timezone/tzdb-2026c.provenance.json",
  );

  await validateSchema(
    "cartulary.network_flow_timezone_ruleset_provenance.v1",
    provenance,
  );

  await assert.rejects(
    validateSchema("cartulary.network_flow_timezone_ruleset_provenance.v1", {
      ...provenance,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  const mutableSource = structuredClone(provenance);
  mutableSource.source_archive.url =
    "https://data.iana.org/time-zones/repository/tzdata-latest.tar.gz";
  await assert.rejects(
    validateSchema(
      "cartulary.network_flow_timezone_ruleset_provenance.v1",
      mutableSource,
    ),
    /must be equal to constant/u,
  );

  const hostAuthoritative = structuredClone(provenance);
  hostAuthoritative.conformance_policy.host_timezone_database_authoritative = true;
  await assert.rejects(
    validateSchema(
      "cartulary.network_flow_timezone_ruleset_provenance.v1",
      hostAuthoritative,
    ),
    /must be equal to constant/u,
  );

  const checker = path.join(
    repoRoot,
    "tools/harness/generated-artifacts/check-json-shapes.mjs",
  );
  const artifactPath = path.join(
    repoRoot,
    "contracts/network-flow/timezone/tzdb-2026c.provenance.json",
  );
  const pass = spawnSync(
    process.execPath,
    [checker, "--kind", "network-flow-timezone-provenance", "--file", artifactPath],
    { encoding: "utf8" },
  );
  assert.equal(pass.status, 0, pass.stderr);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "network-flow-tzdb."));
  try {
    const mutableArtifact = structuredClone(provenance);
    mutableArtifact.source_archive.sha256 = "0".repeat(64);
    const mutableArtifactPath = path.join(root, "tzdb.provenance.json");
    writeFileSync(
      mutableArtifactPath,
      `${JSON.stringify(mutableArtifact, null, 2)}\n`,
    );
    const fail = spawnSync(
      process.execPath,
      [
        checker,
        "--kind",
        "network-flow-timezone-provenance",
        "--file",
        mutableArtifactPath,
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(fail.status, 0);
    assert.match(fail.stderr, /source_archive\.sha256/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("network flow fault-control schema is closed and boundary-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_fault_control.v1",
    fault_id: "fault-1",
    boundary: "network_flow.import.before_transaction_commit",
    fault_kind: "return_error",
    error_code: "network_flow_fault_probe",
    correlation_key: "apply:job-1",
    consume_once: true,
  };

  await validateSchema("cartulary.test.network_flow_fault_control.v1", response);

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_fault_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_fault_control.v1", {
      ...response,
      boundary: "network_flow.import.unknown",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_fault_control.v1", {
      ...response,
      consume_once: false,
    }),
    /must be equal to constant/u,
  );
});

test("network flow randomness-control schema is closed and stream-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_randomness_control.v1",
    control_id: "random-1",
    stream: "network_flow.table_id",
    value_kind: "uuid",
    value_count: 2,
    remaining_count: 2,
    consume_once: true,
    exhaustion: "fail_closed",
  };

  await validateSchema(
    "cartulary.test.network_flow_randomness_control.v1",
    response,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_randomness_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_randomness_control.v1", {
      ...response,
      stream: "network_flow.unknown",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_randomness_control.v1", {
      ...response,
      consume_once: false,
    }),
    /must be equal to constant/u,
  );
});

test("network flow auth-transition-control schema is closed and hidden-resource scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_auth_transition_control.v1",
    control_id: "auth-transition-1",
    boundary: "network_flow.route.after_authorization_before_lookup",
    transition_kind: "incident_membership_revoked",
    actor_ref: "actor:analyst-1",
    incident_ref: "incident:alpha",
    resource_kind: "network_flow_table",
    resource_ref: "network-flow-table:table-1",
    hidden_response_kind: "not_found",
    must_not_disclose_resource: true,
    correlation_key: "query:page-1",
    consume_once: true,
  };

  await validateSchema(
    "cartulary.test.network_flow_auth_transition_control.v1",
    response,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_auth_transition_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_auth_transition_control.v1", {
      ...response,
      hidden_response_kind: "raw_resource",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_auth_transition_control.v1", {
      ...response,
      must_not_disclose_resource: false,
    }),
    /must be equal to constant/u,
  );
});

test("network flow audit-assertion-control schema is closed and count-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.network_flow_audit_assertion_control.v1",
    assertion_id: "audit-assertion-1",
    assertion_kind: "no_audit_replay",
    event_code: "network_flow_table_created",
    operation_ref: "import:apply-1",
    actor_ref: "actor:analyst-1",
    incident_ref: "incident:alpha",
    resource_kind: "network_flow_table",
    resource_ref: "network-flow-table:table-1",
    baseline_count: 1,
    expected_final_count: 1,
    expected_replay_increment: 0,
    correlation_key: "apply:job-1",
    consume_once: true,
  };

  await validateSchema(
    "cartulary.test.network_flow_audit_assertion_control.v1",
    response,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_audit_assertion_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_audit_assertion_control.v1", {
      ...response,
      event_code: "network_flow_secret_viewed",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.network_flow_audit_assertion_control.v1", {
      ...response,
      consume_once: false,
    }),
    /must be equal to constant/u,
  );
});

test("test clock-control schema is closed and mode-scoped", async () => {
  const response = {
    schema_id: "cartulary.test.clock_control.v1",
    mode: "fixed",
    now: "2026-11-01T05:30:00.123456789Z",
    offset_seconds: 0,
    fixed_now: "2026-11-01T05:30:00.123456789Z",
  };

  await validateSchema("cartulary.test.clock_control.v1", response);

  await assert.rejects(
    validateSchema("cartulary.test.clock_control.v1", {
      ...response,
      unexpected: true,
    }),
    /must NOT have additional properties/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.clock_control.v1", {
      ...response,
      mode: "frozen",
    }),
    /must be equal to one of the allowed values/u,
  );

  await assert.rejects(
    validateSchema("cartulary.test.clock_control.v1", {
      ...response,
      schema_id: "cartulary.test.clock_control.v2",
    }),
    /must be equal to constant/u,
  );
});

function runVitestStepSummaryFixture({ root, runnerJSON, sidecarJSON = "" }) {
  const stepDir = path.join(root, sidecarJSON ? "step-sidecar" : "step-fallback");
  const resultsDir = path.relative(repoRoot, path.join(root, "results"));
  const result = spawnSync(
    process.execPath,
    [path.join(repoRoot, "tools/harness/output/test-output.mjs"), "vitest-step"],
    {
      cwd: repoRoot,
      encoding: "utf8",
      env: {
        ...process.env,
        CARTULARY_TEST_RESULTS_DIR: resultsDir,
        CARTULARY_TEST_RUN_ID: "vitest-sidecar-fixture",
        CARTULARY_TEST_TARGET: "frontend-unit",
        CARTULARY_STEP_DIR: stepDir,
        CARTULARY_STEP_LABEL: "frontend-unit vitest",
        CARTULARY_STEP_COMMAND: "pnpm --dir apps/web exec vitest run",
        CARTULARY_STEP_START_TIME: "2026-01-01T00:00:00.000Z",
        CARTULARY_STEP_END_TIME: "2026-01-01T00:00:01.000Z",
        CARTULARY_STEP_LOGICAL_DURATION_MS: "1000",
        CARTULARY_STEP_EXECUTED_DURATION_MS: "1000",
        CARTULARY_STEP_EXIT_STATUS: "1",
        CARTULARY_STEP_RUNNER_LOG: runnerJSON,
        CARTULARY_STEP_STDOUT_LOG: path.join(root, "stdout.log"),
        CARTULARY_STEP_STDERR_LOG: path.join(root, "stderr.log"),
        ...(sidecarJSON
          ? { CARTULARY_STEP_VITEST_FAILURE_DETAILS: sidecarJSON }
          : {}),
      },
    },
  );
  assert.equal(
    result.status,
    1,
    `vitest step fixture should fail for the synthetic assertion: ${result.stderr}${result.stdout}`,
  );
  return JSON.parse(
    readFileSync(path.join(stepDir, "step-summary.json"), "utf8"),
  );
}

function browserWorkerSlotCount(group) {
  if (group?.kind === "functional_shard" || group?.kind === "support") {
    return 1;
  }
  const workers = group?.workers ?? "1";
  if (workers === "default") {
    return 1;
  }
  const parsed = Number.parseInt(String(workers), 10);
  assert.ok(
    Number.isInteger(parsed) && parsed > 0 && String(parsed) === String(workers),
    `browser group ${group?.id} workers must be a positive integer or default`,
  );
  return parsed;
}

function assertBrowserWorkerSlots(units, label) {
  assert.ok(units.length > 0, `${label} must include browser groups`);
  const expectedTotal = units.reduce(
    (sum, unit) => sum + browserWorkerSlotCount(unit.browser_group),
    0,
  );
  const occupied = new Set();
  for (const unit of units) {
    const env = unit.env ?? {};
    assert.equal(
      env.CARTULARY_PLAYWRIGHT_WORKER_COUNT,
      String(expectedTotal),
      `${label} ${unit.id} must receive the service-session worker count`,
    );
    assert.match(
      env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET ?? "",
      /^(0|[1-9][0-9]*)$/,
      `${label} ${unit.id} must receive an explicit worker offset`,
    );
    const offset = Number.parseInt(
      env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET,
      10,
    );
    const slots = browserWorkerSlotCount(unit.browser_group);
    for (let slot = offset; slot < offset + slots; slot += 1) {
      assert.ok(!occupied.has(slot), `${label} worker slot ${slot} overlaps`);
      occupied.add(slot);
    }
    if (unit.browser_group?.kind === "support") {
      assert.equal(env.PLAYWRIGHT_WORKERS, "1");
    }
  }
  assert.deepEqual(
    [...occupied].sort((left, right) => left - right),
    Array.from({ length: expectedTotal }, (_value, index) => index),
    `${label} worker slots must be contiguous`,
  );
}

test("Vitest failure sidecar overrides STACK_TRACE_ERROR summary fallback", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "vitest-sidecar."));
  try {
    const ownerPath = "apps/web/src/SyntheticSidecar.test.tsx";
    const title = "reports the retained assertion message";
    const runnerJSON = path.join(root, "runner.json");
    const sidecarJSON = path.join(root, "vitest-failure-details.json");
    writeFileSync(path.join(root, "stdout.log"), "");
    writeFileSync(path.join(root, "stderr.log"), "");
    writeFileSync(
      runnerJSON,
      JSON.stringify({
        testResults: [
          {
            name: path.join(repoRoot, ownerPath),
            status: "failed",
            assertionResults: [
              {
                title,
                status: "failed",
                failureMessages: [
                  `Error: STACK_TRACE_ERROR\n    at ${path.join(repoRoot, ownerPath)}:10:1`,
                ],
              },
            ],
          },
        ],
      }),
    );
    writeFileSync(
      sidecarJSON,
      JSON.stringify({
        schema_id: "cartulary.vitest_failure_details.v1",
        runner_json: path.relative(repoRoot, runnerJSON),
        stdout_log: "",
        stderr_log: "",
        generated_at: "2026-01-01T00:00:00.000Z",
        failures: [
          {
            owner_path: ownerPath,
            title,
            status: "failed",
            message: "expected fetchMock to have exactly 2 calls, got 3",
            message_source: "synthetic_sidecar_assertion_message",
            raw_messages: [
              "expected fetchMock to have exactly 2 calls, got 3",
            ],
            diagnostic_tags: ["vitest_stack_trace_error"],
            first_app_frame: `${ownerPath}:10`,
          },
        ],
      }),
    );

    const sidecarSummary = runVitestStepSummaryFixture({
      root,
      runnerJSON,
      sidecarJSON,
    });
    assert.equal(
      sidecarSummary.dossiers[0].message,
      "expected fetchMock to have exactly 2 calls, got 3",
    );
    assert.ok(
      sidecarSummary.dossiers[0].diagnostic_tags.includes(
        "vitest_failure_sidecar",
      ),
      "sidecar-backed failures must carry a diagnostic tag",
    );
    assert.match(
      sidecarSummary.dossiers[0].raw,
      /vitest-failure-details\.json/,
      "sidecar-backed failures must retain the sidecar artifact ref",
    );

    const fallbackSummary = runVitestStepSummaryFixture({
      root,
      runnerJSON,
    });
    assert.match(
      fallbackSummary.dossiers[0].message,
      /Vitest reporter emitted STACK_TRACE_ERROR/,
      "missing sidecar must keep the current fallback diagnostic",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

function renderedArtifacts() {
  const topology = loadExecutionTopology();
  const taskSurface = renderTaskSurfaceManifest(topology);
  const browserBatch = renderBrowserBatchManifest(topology);
  const serviceBacked = renderServiceBackedScheduleManifest({
    topology: "tools/execution_topology_manifest.json",
    topologyObject: topology,
  });
  return {
    topology,
    taskSurface,
    browserBatch,
    serviceBacked,
    checkSchedule: renderCheckScheduleManifest(topology),
    expandedCheckSchedule: renderCheckScheduleManifest(topology, {
      serviceBackedScheduleManifest: serviceBacked,
      expandServiceBackedScheduleForCheck,
    }),
    taskSurfaceMake: renderTaskSurfaceMake(taskSurface),
  };
}

function machineOwnedPublicRegistry() {
  const owner = readJSON("tools/task_surface_owner.json");
  const rows = new Map();
  for (const target of owner.targets) {
    if (target.target_class !== "public") {
      continue;
    }
    rows.set(target.name, {
      outputClass: target.output_policy.output_class,
      sideEffects: target.side_effects
        .map((entry) => entry.class)
        .sort((left, right) => left.localeCompare(right)),
    });
  }
  return rows;
}

function machineOwnedInputMatrix() {
  const owner = readJSON("tools/task_surface_owner.json");
  const byTarget = new Map();
  for (const target of owner.targets) {
    byTarget.set(target.name, target.input_contract?.inputs ?? []);
  }
  return byTarget;
}

function normalizeManifestInput(input) {
  const normalized = {
    name: input.name,
    binding: input.binding,
    allowed_sources: input.allowed_sources,
    required: input.required,
    type: input.type,
    default: input.default,
    empty_string: input.empty_string,
    normalization: input.normalization,
    invalid_reason: input.invalid_reason,
    summary_emission: input.summary_emission,
    child_forwarding: input.child_forwarding,
  };
  if (input.values !== undefined) {
    normalized.values = input.values;
  }
  if (input.min !== undefined) {
    normalized.min = input.min;
  }
  if (input.max !== undefined) {
    normalized.max = input.max;
  }
  return normalized;
}

function normalizeInputList(inputs = []) {
  return inputs
    .map((input) => ({ ...input }))
    .sort((left, right) => left.name.localeCompare(right.name));
}

test("fast harness smoke is role-complete and intentionally small", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const checksByName = new Map(
    manifest.harness_checks.map((check) => [check.name, check]),
  );
  const fastChecks = manifest.harness_tiers.fast.checks;
  assert.deepEqual(fastChecks, [
    "harness-smoke-public-make-wrapper",
    "harness-smoke-check-scheduler-smoke",
    "harness-smoke-service-backed-scheduler-smoke",
  ]);
  assert.deepEqual(
    fastChecks.map((name) => checksByName.get(name)?.gate_smoke_role),
    [
      "public_make_wrapper",
      "check_scheduler_semantic",
      "service_backed_scheduler_semantic",
    ],
  );
});

test("fast harness smoke scratch fixtures stay outside the repository", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const checksByName = new Map(
    manifest.harness_checks.map((check) => [check.name, check]),
  );
  const fastShellScripts = new Set(
    manifest.harness_tiers.fast.checks.flatMap((name) =>
      (checksByName.get(name)?.backing_scripts ?? []).filter((script) =>
        script.endsWith(".sh"),
      ),
    ),
  );
  const requiredScratchHelpers = new Map([
    [
      "tools/harness/tests/test-public-make-wrapper-smoke.sh",
      ["cartulary_harness_mktemp_dir \"public-make-wrapper.XXXXXX\""],
    ],
    [
      "tools/harness/scheduler/tests/test-check-scheduler.sh",
      [
        "cartulary_harness_mktemp_dir \"check-scheduler-smoke.XXXXXX\"",
        "cartulary_harness_mktemp_dir \"check-scheduler-smoke-service-timing.XXXXXX\"",
      ],
    ],
    [
      "tools/harness/scheduler/tests/test-service-backed-scheduler.sh",
      [
        "cartulary_harness_mktemp_dir \"service-backed-scheduler-smoke.XXXXXX\"",
      ],
    ],
  ]);
  for (const script of fastShellScripts) {
    const content = readFileSync(path.join(repoRoot, script), "utf8");
    assert.match(
      content,
      /tools\/harness\/test-support\/harness-scratch\.sh|test-support\/harness-scratch\.sh/,
      `${script} must source harness-scratch.sh for fast smoke scratch`,
    );
    for (const required of requiredScratchHelpers.get(script) ?? []) {
      assert.ok(content.includes(required), `${script} missing ${required}`);
    }
    assert.doesNotMatch(
      content,
      /mktemp -d "\$\{?ROOT_DIR\}?\/(?:tmp|\.cartulary\/tmp)\/(?:public-make-wrapper|check-scheduler-smoke|check-scheduler-smoke-service-timing|service-backed-scheduler-smoke)\.XXXXXX"/,
      `${script} must not create fast smoke fixtures with raw mktemp templates`,
    );
    assert.doesNotMatch(
      content,
      /CARTULARY_HARNESS_SCRATCH_ROOT:-\$\{ROOT_DIR\}\/\.cartulary\/tmp/,
      `${script} must not default harness scratch inside the repository`,
    );
  }
});

test("dev service lifecycle guards are mutation-safe", () => {
  const result = spawnSync("bash", ["./tools/harness/readiness/tests/test-dev-services-lifecycle.sh"], {
    cwd: repoRoot,
    encoding: "utf8",
    env: { ...process.env },
    maxBuffer: 16 * 1024 * 1024,
  });
  assert.equal(
    result.status,
    0,
    [
      "tools/harness/readiness/tests/test-dev-services-lifecycle.sh failed",
      "--- stdout ---",
      result.stdout,
      "--- stderr ---",
      result.stderr,
    ].join("\n"),
  );
});

test("task-surface validation rejects fast harness smoke drift", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  const invalid = structuredClone(taskSurface);
  invalid.harness_tiers.fast.checks.push("harness-smoke-execution-topology");
  assert.match(
    collectTaskSurfaceManifestErrors(invalid, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /harness_tiers\.fast\.checks must contain exactly 3 gate smoke checks/,
  );

  const missingRole = structuredClone(taskSurface);
  delete missingRole.harness_checks.find(
    (check) => check.name === "harness-smoke-public-make-wrapper",
  ).gate_smoke_role;
  assert.match(
    collectTaskSurfaceManifestErrors(missingRole, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /harness-smoke-public-make-wrapper\.gate_smoke_role is required for fast harness smoke/,
  );
});

test("task-surface validation rejects Node-backed wrappers without Node readiness", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  for (const target of ["check", "check-harness-smoke"]) {
    const invalid = structuredClone(taskSurface);
    invalid.make_recipes[target].prerequisites = [];
    assert.match(
      collectTaskSurfaceManifestErrors(invalid, {
        browserBatchManifest: browserBatch,
        serviceBackedScheduleManifest: serviceBacked,
      }).join("\n"),
      new RegExp(
        `make_recipes\\.${target}\\.prerequisites must include \\$\\(NODE_BIN\\)`,
      ),
      `${target} must require explicit Node readiness`,
    );
  }
});

test("full harness tier composes fast, extended, lifecycle, and full-only diagnostics", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const fullOnlyChecks = new Set(["harness-smoke-tool-output-real-targets"]);
  const expectedFullBase = [
    ...manifest.harness_tiers.fast.checks,
    ...manifest.harness_tiers.extended.checks,
    ...manifest.harness_tiers.lifecycle.checks,
  ];
  const filteredFull = manifest.harness_tiers.full.checks.filter(
    (check) => !fullOnlyChecks.has(check),
  );
  assert.deepEqual(filteredFull, expectedFullBase);
  for (const check of fullOnlyChecks) {
    assert.ok(manifest.harness_tiers.full.checks.includes(check));
  }
});

test("execution harness smoke is a narrow execution-wrapper subset", () => {
  const manifest = readJSON("tools/task_surface_manifest.json");
  const expectedExecutionChecks = [
    "harness-smoke-run-make-sequence-fast",
    "harness-smoke-cartulary-runner-service-backed-target",
    "harness-smoke-make-node-tools",
  ];
  assert.deepEqual(manifest.harness_tiers.execution.checks, expectedExecutionChecks);
  const extendedChecks = new Set(manifest.harness_tiers.extended.checks);
  for (const check of expectedExecutionChecks) {
    assert.ok(extendedChecks.has(check), `${check} must remain in extended smoke`);
  }
  const target = manifest.targets.find((entry) => entry.name === "run-harness-smoke-execution");
  assert.equal(target?.target_class, "internal_helper");
  assert.deepEqual(target?.default_inclusion_sets, []);
});

test("check scheduler restores node packages before run-step validation", () => {
  const { checkSchedule, taskSurface } = renderedArtifacts();
  const schedule = checkSchedule.schedules.find(
    (entry) => entry.target === "check",
  );
  assert.deepEqual(
    taskSurface.make_recipes.check.prerequisites,
    ["$(NODE_BIN)"],
    "check wrapper must only bootstrap the scheduler Node runtime",
  );
  const checkFrontendInstall = schedule.work_units.find(
    (unit) => unit.target === "check-frontend-install",
  );
  const toolchainDrift = schedule.work_units.find(
    (unit) => unit.target === "toolchain-drift",
  );
  const jsonShapeCheck = schedule.work_units.find(
    (unit) => unit.target === "json-shape-check",
  );
  assert.deepEqual(
    schedule.work_units
      .filter((unit) => (unit.needs ?? []).length === 0)
      .map((unit) => unit.target)
      .sort(),
    ["check-frontend-install"],
    "check-frontend-install must be the only dependency-free check unit",
  );
  assert.ok(
    checkFrontendInstall,
    "check schedule must include check-frontend-install",
  );
  assert.deepEqual(
    checkFrontendInstall.needs ?? [],
    [],
    "check-frontend-install must be able to run before run-step children",
  );
  assert.ok(toolchainDrift, "check schedule must include toolchain-drift");
  assert.ok(
    (toolchainDrift.needs ?? []).includes("check-frontend-install"),
    "toolchain-drift must wait for installed node package dependencies",
  );
  assert.ok(jsonShapeCheck, "check schedule must include json-shape-check");
  assert.ok(
    (jsonShapeCheck.needs ?? []).includes("check-frontend-install"),
    "json-shape-check must wait for installed node package dependencies",
  );
  for (const target of [
    "harness-contract",
    "harness-contract-tests",
    "toolchain-drift",
    "json-shape-check",
  ]) {
    assert.ok(
      taskSurface.make_recipes[target].prerequisites.includes(
        "$(FRONTEND_INSTALL_STAMP)",
      ),
      `${target} direct wrapper must bootstrap installed node package dependencies`,
    );
  }
});

test("check scheduler defers own schema validation until package readiness", () => {
  const schedulerScript = readFileSync(
    path.join(repoRoot, "tools/harness/scheduler/check-schedule-cli.mjs"),
    "utf8",
  );
  const schedulerEngine = readFileSync(
    path.join(repoRoot, "tools/harness/scheduler/scheduler/engine.mjs"),
    "utf8",
  );
  assert.match(
    schedulerScript,
    /schemaValidationEnabled:\s*!deferSchemaValidationForPackageReadiness/u,
  );
  assert.match(
    schedulerScript,
    /context\.reporter\.setSchemaValidationEnabled\(true\)/u,
  );
  assert.match(schedulerEngine, /deferredSchemaRecords/u);
});

test("generated task surface and Make wrapper keep harness projection wiring", () => {
  const { taskSurface, browserBatch, serviceBacked, taskSurfaceMake } =
    renderedArtifacts();
  assert.deepEqual(
    collectTaskSurfaceManifestErrors(taskSurface, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }),
    [],
  );
  assert.deepEqual(taskSurface.make_recipes.check?.prerequisites, [
    "$(NODE_BIN)",
  ]);
  assert.deepEqual(
    taskSurface.make_recipes["check-harness-smoke"]?.prerequisites,
    ["$(NODE_BIN)"],
  );
  const checkBlock = targetRecipeBlock(taskSurfaceMake, "check").join("\n");
  const preflightIndex = checkBlock.indexOf(
    "$(call RUN_PUBLIC_PREFLIGHT,check)",
  );
  const prerequisiteIndex = checkBlock.indexOf(
    "$(MAKE) --silent --no-print-directory $(NODE_BIN); fi",
  );
  const schedulerIndex = checkBlock.indexOf(
    "$(NODE_BIN) $(RUN_CHECK_SCHEDULE_SCRIPT)",
  );
  assert.ok(preflightIndex >= 0, "check must render public preflight");
  assert.ok(
    prerequisiteIndex >= 0,
    "check must render Node readiness before owner CLI preflight",
  );
  assert.ok(
    preflightIndex > prerequisiteIndex,
    "check must run owner CLI preflight after Node readiness",
  );
  assert.ok(
    schedulerIndex > preflightIndex,
    "check must launch the scheduler after preflight",
  );
  assert.equal(
    taskSurface.make_recipes["check-harness-smoke"]?.child_target,
    "run-harness-smoke-fast",
  );
  assert.equal(
    taskSurface.make_recipes["check-harness-smoke"]?.projection,
    "check-harness-smoke",
  );
  assert.match(
    taskSurfaceMake,
    /\$\(RUN_HARNESS_SMOKE_SCRIPT\) --tier fast --jobs "\$\(HARNESS_SMOKE_JOBS\)"/,
  );
  assert.match(
    taskSurfaceMake,
    /summary-target --target check-harness-smoke --child-target run-harness-smoke-fast --status pass/,
  );
  const targetEntries = new Map(
    taskSurface.targets.map((entry) => [entry.name, entry]),
  );
});

test("machine task-surface owner defines public output classes and side effects", () => {
  const { taskSurface } = renderedArtifacts();
  const ownerRows = machineOwnedPublicRegistry();
  const verificationContracts = loadVerificationContracts(repoRoot);
  assert.ok(
    verificationContracts.verificationByID.has(
      "harness.command_surface.verification.public_registry_parity",
    ),
    "public task-surface parity must have a machine verification owner",
  );
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  assert.equal(
    ownerRows.size,
    publicTargets.length,
    "authored owner public target count must match rendered manifest",
  );
  const publicIdentityBytes = `${publicTargets
    .map((target) => `${target.name}\t${target.command_id}`)
    .join("\n")}\n`;
  assert.equal(
    createHash("sha256").update(publicIdentityBytes).digest("hex"),
    "a6f7f954a565d90f0f28597ba4549aa1e4c7fe3eae6457f0e049da2e0ac80b1f",
    "public target and command ID inventory changed; revise the authored owner and this explicit interface digest together",
  );
  for (const target of publicTargets) {
    const spec = ownerRows.get(target.name);
    assert.ok(spec, `${target.name} must appear in the authored task-surface owner`);
    assert.equal(
      spec.outputClass,
      target.output_policy.output_class,
      `${target.name} output class must match the authored owner`,
    );
    assert.deepEqual(
      spec.sideEffects,
      target.side_effects
        .map((entry) => entry.class)
        .sort((left, right) => left.localeCompare(right)),
      `${target.name} side effects must match the authored owner`,
    );
  }
});

test("observability dispositions and sequence identities fail closed", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  const errorsFor = (candidate) => collectTaskSurfaceManifestErrors(candidate, {
    browserBatchManifest: browserBatch,
    serviceBackedScheduleManifest: serviceBacked,
  }).join("\n");

  const omitted = structuredClone(taskSurface);
  omitted.observability_policy.required_targets = omitted.observability_policy.required_targets
    .filter((target) => target !== "backend-unit");
  omitted.observability_policy.target_measurement_profiles = omitted.observability_policy.target_measurement_profiles
    .filter((binding) => binding.target !== "backend-unit");
  assert.match(errorsFor(omitted), /observability_policy omits public target backend-unit/);

  const overlap = structuredClone(taskSurface);
  overlap.observability_policy.excluded_targets.push({
    target: "backend-unit",
    owner_section: "Section 4",
    reason: "invalid overlap fixture",
  });
  assert.match(errorsFor(overlap), /excluded_targets\[5\]\.target overlaps required disposition/);

  const unknown = structuredClone(taskSurface);
  unknown.observability_policy.out_of_scope_targets.push({
    target: "unknown-public-command",
    owner_section: "Section 4",
    reason: "invalid unknown fixture",
  });
  assert.match(errorsFor(unknown), /target must name a public target/);

  const unowned = structuredClone(taskSurface);
  unowned.observability_policy.excluded_targets[0].owner_section = "";
  assert.match(errorsFor(unowned), /owner_section must be a Section reference/);

  const unownedMeasurementIdentity = structuredClone(taskSurface);
  delete unownedMeasurementIdentity.targets.find(
    (target) => target.name === "release-browser-readiness",
  ).command_id;
  assert.match(errorsFor(unownedMeasurementIdentity), /target must declare a valid command_id/);

  const duplicateSequence = structuredClone(taskSurface);
  duplicateSequence.sequences.lint.steps[1].target = duplicateSequence.sequences.lint.steps[0].target;
  assert.match(errorsFor(duplicateSequence), /occurrence aliases are unsupported/);
});

test("owner task surface exposes only the v2 command family and private catalog check", () => {
  const { taskSurface } = renderedArtifacts();
  const targets = new Map(taskSurface.targets.map((entry) => [entry.name, entry]));
  const expected = new Map([
    ["explain-test-owner", ["OWNER", "JSON"]],
    ["service-backed-test-slice", ["OWNER", "ROWS", "VITEST_MAX_WORKERS", "PLAYWRIGHT_WORKERS", "JSON"]],
    ["task-guide", ["ROLE", "OWNER", "JSON"]],
    ["test-evidence-audit", ["OWNER", "EVIDENCE_ROOTS_FILE"]],
    ["test-slice", ["OWNER", "ROWS", "VITEST_MAX_WORKERS", "PLAYWRIGHT_WORKERS", "JSON"]],
  ]);
  for (const [name, inputNames] of expected) {
    const target = targets.get(name);
    assert.ok(target, `${name} must be generated from the task-surface owner`);
    assert.equal(target.target_class, "public", `${name} must remain public`);
    assert.deepEqual(
      target.input_contract.inputs.map((input) => input.name),
      inputNames,
      `${name} must expose only its v2 inputs`,
    );
    assert.equal(
      taskSurface.make_recipes[name]?.type,
      "owner_command",
      `${name} must use the generated owner-command recipe`,
    );
  }
  const catalogCheck = targets.get("test-catalog-check");
  assert.equal(catalogCheck?.target_class, "check_internal", "catalog check must be private");
  assert.deepEqual(
    catalogCheck?.default_inclusion_sets,
    ["check"],
    "private catalog check must remain in default check",
  );
});

test("machine task-surface owner defines public target input contracts", () => {
  const { taskSurface } = renderedArtifacts();
  const ownerInputs = machineOwnedInputMatrix();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  for (const target of publicTargets) {
    const expected = normalizeInputList(ownerInputs.get(target.name) ?? []);
    const actual = normalizeInputList(
      (target.input_contract?.inputs ?? []).map(normalizeManifestInput),
    );
    assert.deepEqual(
      actual,
      expected,
      `${target.name} input_contract must match the authored owner`,
    );
  }

  const synthetic = structuredClone(taskSurface);
  const drift = synthetic.targets.find(
    (target) => target.name === "scheduler-summary-timing-drift",
  );
  drift.input_contract.inputs.find(
    (input) => input.name === "SCHEDULER_WARM_CHECK_BUDGET_MS",
  ).default = null;
  const expected = normalizeInputList(
    ownerInputs.get("scheduler-summary-timing-drift") ?? [],
  );
  assert.notDeepEqual(
    normalizeInputList(drift.input_contract.inputs.map(normalizeManifestInput)),
    expected,
    "matrix parity check must detect implementation default drift",
  );
});

test("scheduler manifest exercises every required command type", () => {
  const schedulerManifest = readJSON("tools/scheduler_manifest.json");
  const expected = new Set([
    "make_target",
    "service_session_start",
    "browser_stage_session_start",
    "browser_group",
    "browser_stage_complete",
    "go_shard",
    "go_shard_finalize",
    "service_complete",
  ]);
  for (const schedule of schedulerManifest.schedules ?? []) {
    for (const unit of schedule.work_units ?? []) {
      expected.delete(unit.command?.type);
    }
  }
  assert.deepEqual([...expected].sort(), [], "every required scheduler command type must have a live fixture");
});

test("scheduler family facade matches schema registry and generated manifests", () => {
  const expectedFamilies = [...schedulerFamilyValues];
  const expectedFamilySet = new Set(expectedFamilies);
  const schedulerSchema = readJSON(
    "tools/schemas/cartulary.scheduler_manifest.v2.schema.json",
  );
  assert.deepEqual(
    schedulerSchema.$defs.schedule.properties.scheduler_kind.enum,
    expectedFamilies,
    "scheduler manifest schema enum must match scheduler family facade",
  );

  const resourceRegistry = readJSON("tools/scheduler_resource_registry.json");
  validateSchedulerResourceRegistrySemantics(
    resourceRegistry,
    "tools/scheduler_resource_registry.json",
  );
  const registryFamilyReferences = [];
  for (const resource of resourceRegistry.resources ?? []) {
    for (const scheduler of resource.schedulers ?? []) {
      registryFamilyReferences.push([`resource ${resource.name}`, scheduler]);
    }
  }
  for (const template of resourceRegistry.templates ?? []) {
    for (const scheduler of template.schedulers ?? []) {
      registryFamilyReferences.push([`template ${template.name}`, scheduler]);
    }
  }
  for (const profile of resourceRegistry.capacity_profiles ?? []) {
    registryFamilyReferences.push([
      `capacity profile ${profile.name}`,
      profile.scheduler,
    ]);
  }
  for (const [source, scheduler] of registryFamilyReferences) {
    assert.ok(
      expectedFamilySet.has(scheduler),
      `${source} references unknown scheduler family ${scheduler}`,
    );
  }
  assert.deepEqual(
    (resourceRegistry.capacity_profiles ?? [])
      .map((profile) => profile.name)
      .sort(),
    [...schedulerCapacityProfileValues].sort(),
    "registry capacity profiles must match the scheduler facade",
  );
  for (const [family, profiles] of Object.entries(schedulerCapacityProfilesByFamily)) {
    for (const profileName of profiles) {
      assert.equal(
        schedulerFamilyForCapacityProfile(profileName),
        family,
        `${profileName} must map back to ${family}`,
      );
    }
  }

  const schedulerManifest = readJSON("tools/scheduler_manifest.json");
  for (const schedule of schedulerManifest.schedules ?? []) {
    assert.ok(
      expectedFamilySet.has(schedule.scheduler_kind),
      `${schedule.target} references unknown scheduler family ${schedule.scheduler_kind}`,
    );
    assert.equal(
      schedulerFamilyForCapacityProfile(schedule.capacity_profile),
      schedule.scheduler_kind,
      `${schedule.target} capacity profile must match scheduler_kind`,
    );
  }
});

function lifecycleFixtureUnit({ id, command, needs = [], kind = "test_work", timeoutMs = 0 }) {
  return {
    id,
    label: id,
    kind,
    class: kind === "finalizer" ? "cleanup" : "fixture",
    target: id,
    aggregateTarget: "test-scheduler-lifecycle",
    needs,
    completionKeys: [id],
    failureKeys: [id],
    resourceClaims: new Map(kind === "finalizer" ? [] : [["process", 1]]),
    priority: 0,
    weightMs: 1,
    order: 0,
    timeoutMs,
    countInTotal: kind === "finalizer" ? false : undefined,
    command: { command: process.execPath, args: ["-e", command], env: process.env },
  };
}

function lifecycleFixtureSchedule(target, workUnits) {
  return {
    target,
    kind: "test_slice",
    prefix: "TEST-SCHEDULER",
    eventSchemaID: "cartulary.scheduler_event.v7",
    summarySchemaID: "cartulary.service_backed_scheduler_summary.v10",
    resourceScheduler: "test_slice",
    stopOnFirstFailure: false,
    showFinalizing: true,
    validateSummaryTiming: false,
    resourceLimits: new Map([["process", 1]]),
    resourceLimitSources: new Map([["process", "fixture"]]),
    workUnits,
    finalizerCount: workUnits.filter((unit) => unit.kind === "finalizer").length,
    shouldReplayLog: () => false,
  };
}

async function withFixtureSchedulerArtifacts(fixture, callback) {
  const names = ["CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID"];
  const previous = new Map(names.map((name) => [name, process.env[name]]));
  process.env.CARTULARY_TEST_RESULTS_DIR = path.join(fixture, "results");
  process.env.CARTULARY_TEST_RUN_ID = "fixture-run";
  try {
    return await callback();
  } finally {
    for (const [name, value] of previous) {
      if (value === undefined) delete process.env[name];
      else process.env[name] = value;
    }
  }
}

test("generic scheduler drains independent work, skips failed dependencies, and always finalizes", async () => {
  const fixture = mkdtempSync(path.join(repoRoot, "tmp", "scheduler-lifecycle."));
  const independentMarker = path.join(fixture, "independent");
  const finalizerMarker = path.join(fixture, "finalizer");
  const target = `test-scheduler-lifecycle-${path.basename(fixture)}`;
  const units = [
    lifecycleFixtureUnit({ id: "product-failure", command: "process.exit(10)" }),
    lifecycleFixtureUnit({
      id: "independent",
      command: `require('node:fs').writeFileSync(${JSON.stringify(independentMarker)}, 'done')`,
    }),
    lifecycleFixtureUnit({ id: "dependent", needs: ["product-failure"], command: "process.exit(99)" }),
    lifecycleFixtureUnit({
      id: "cleanup",
      kind: "finalizer",
      needs: ["product-failure", "independent", "dependent"],
      command: `require('node:fs').writeFileSync(${JSON.stringify(finalizerMarker)}, 'done')`,
    }),
  ];
  const result = await withFixtureSchedulerArtifacts(fixture, async () =>
    await runNormalizedSchedule({
      repoRoot,
      schedule: lifecycleFixtureSchedule(target, units),
      testOutputScript: "",
    }));
  assert.equal(result.status, 10, "product failure must remain primary after cleanup");
  assert.equal(existsSync(independentMarker), true, "independent work must drain");
  assert.equal(existsSync(finalizerMarker), true, "finalizer must run after failed and skipped work");
  assert.deepEqual(
    result.reporter.skippedWork.map((entry) => [entry.id, entry.reason]),
    [["dependent", "dependency_failure"]],
  );
  assert.equal(
    result.reporter.completedWork.filter((entry) => entry.id === "product-failure").length,
    1,
    "product assertions must not retry",
  );
  rmSync(fixture, { recursive: true, force: true });
});

test("scheduler watchdog and cancellation terminate work while preserving finalization", async () => {
  const fixture = mkdtempSync(path.join(repoRoot, "tmp", "scheduler-watchdog."));
  const finalizerMarker = path.join(fixture, "finalizer");
  const target = `test-scheduler-watchdog-${path.basename(fixture)}`;
  const units = [
    lifecycleFixtureUnit({
      id: "timed-work",
      command: "setTimeout(() => {}, 10000)",
      timeoutMs: 50,
    }),
    lifecycleFixtureUnit({
      id: "cleanup",
      kind: "finalizer",
      needs: ["timed-work"],
      command: `require('node:fs').writeFileSync(${JSON.stringify(finalizerMarker)}, 'done')`,
    }),
  ];
  const timed = await withFixtureSchedulerArtifacts(fixture, async () =>
    await runNormalizedSchedule({
      repoRoot,
      schedule: lifecycleFixtureSchedule(target, units),
      testOutputScript: "",
    }));
  assert.equal(timed.status, 13);
  assert.equal(existsSync(finalizerMarker), true, "timeout must not suppress cleanup");

  const controller = new AbortController();
  const cancelLog = path.join(fixture, "cancel.log");
  const cancelled = runScheduledCommand(
    repoRoot,
    process.execPath,
    ["-e", "setTimeout(() => {}, 10000)"],
    cancelLog,
    { env: process.env, signal: controller.signal },
  );
  controller.abort({ exitCode: 130, signal: "SIGINT", reason: "cancelled_or_interrupted" });
  assert.deepEqual(await cancelled, { status: 130, terminationReason: "cancelled_or_interrupted" });
  rmSync(fixture, { recursive: true, force: true });
});

test("scheduler interruption cancels ordinary work and still runs finalizers", () => {
  const fixture = mkdtempSync(path.join(repoRoot, "tmp", "scheduler-interrupt."));
  const finalizerMarker = path.join(fixture, "finalizer");
  const runner = path.join(fixture, "interrupt-fixture.mjs");
  const schedulerRunnerURL = pathToFileURL(
    path.join(repoRoot, "tools", "harness", "scheduler", "scheduler-runner.mjs"),
  ).href;
  try {
    writeFileSync(runner, `
import process from "node:process";
import { runNormalizedSchedule } from ${JSON.stringify(schedulerRunnerURL)};
const unit = (id, command, kind = "test_work") => ({
  id,
  label: id,
  kind,
  class: kind === "finalizer" ? "cleanup" : "fixture",
  target: id,
  aggregateTarget: "test-scheduler-interrupt",
  needs: kind === "finalizer" ? ["interrupted-work"] : [],
  completionKeys: [id],
  failureKeys: [id],
  resourceClaims: new Map(kind === "finalizer" ? [] : [["process", 1]]),
  priority: 0,
  weightMs: 1,
  order: kind === "finalizer" ? 1 : 0,
  timeoutMs: 5_000,
  countInTotal: kind === "finalizer" ? false : undefined,
  command: { command: process.execPath, args: ["-e", command], env: process.env },
});
const schedule = {
  target: "test-scheduler-interrupt",
  kind: "test_slice",
  prefix: "TEST-SCHEDULER",
  eventSchemaID: "cartulary.scheduler_event.v7",
  summarySchemaID: "cartulary.service_backed_scheduler_summary.v10",
  resourceScheduler: "test_slice",
  stopOnFirstFailure: false,
  showFinalizing: true,
  validateSummaryTiming: false,
  resourceLimits: new Map([["process", 1]]),
  resourceLimitSources: new Map([["process", "fixture"]]),
  workUnits: [
    unit("interrupted-work", "setTimeout(() => {}, 10000)"),
    unit("cleanup", ${JSON.stringify(`require('node:fs').writeFileSync(${JSON.stringify(finalizerMarker)}, 'done')`)}, "finalizer"),
  ],
  finalizerCount: 1,
  shouldReplayLog: () => false,
};
setTimeout(() => process.kill(process.pid, "SIGINT"), 50);
const result = await runNormalizedSchedule({
  repoRoot: ${JSON.stringify(repoRoot)},
  schedule,
  testOutputScript: "",
});
process.exitCode = result.status;
`);
    const result = spawnSync(process.execPath, [runner], {
      cwd: repoRoot,
      env: {
        ...process.env,
        CARTULARY_TEST_RESULTS_DIR: path.join(fixture, "results"),
        CARTULARY_TEST_RUN_ID: "fixture-run",
      },
      encoding: "utf8",
      timeout: 5_000,
    });
    assert.equal(result.error, undefined, result.error?.message);
    assert.equal(result.status, 130, `${result.stdout}\n${result.stderr}`);
    assert.equal(existsSync(finalizerMarker), true, "SIGINT must not suppress cleanup");
  } finally {
    rmSync(fixture, { recursive: true, force: true });
  }
});

test("scheduler FIFO reservations prevent resource leapfrogging", () => {
  const unit = (id, claims) => ({ id, label: id, needs: [], resourceClaims: new Map(claims) });
  const pending = [
    unit("first-process", [["process", 1]]),
    unit("second-process", [["process", 1]]),
    unit("independent-browser", [["browser_stack", 1]]),
  ];
  const index = priorityAdmissiblePendingUnitIndex({
    pending,
    completedKeys: new Set(),
    failedKeys: new Map(),
    resourceLimits: new Map([["process", 1], ["browser_stack", 1]]),
    activeClaims: new Map([["process", 1]]),
  });
  assert.equal(index, 2, "independent resources may proceed while FIFO reserves blocked process capacity");
  assert.equal(
    priorityAdmissiblePendingUnitIndex({
      pending: pending.slice(0, 2),
      completedKeys: new Set(),
      failedKeys: new Map(),
      resourceLimits: new Map([["process", 1]]),
      activeClaims: new Map([["process", 1]]),
    }),
    -1,
    "later work must not leapfrog an earlier unit claiming the same resource",
  );
});

test("catalog browser scheduler digest is deterministic and delivery-independent", () => {
  const manifest = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
  const stage = resolveBrowserBatchStage(manifest, "stateful");
  const first = buildBrowserStageSchedule(stage, manifest);
  const second = buildBrowserStageSchedule(stage, manifest);
  const extension = (schedule) => schedule.summaryExtra().extensions["cartulary.test_slice.browser_scheduler.v1"];
  assert.deepEqual(extension(first), extension(second));
  assert.match(extension(first).scheduler_semantic_digest, /^sha256:[0-9a-f]{64}$/u);
  assert.equal(JSON.stringify(extension(first)).includes("phase"), false);
  const sessions = first.workUnits.filter((unit) => unit.kind === "browser_stage_session");
  const finalizers = first.workUnits.filter((unit) => unit.kind === "finalizer");
  assert.ok(sessions.every((unit) => unit.completionKeys[0] === unit.id));
  assert.ok(
    finalizers.every((unit) =>
      unit.needs.every((need) => first.workUnits.some((candidate) => candidate.id === need)),
    ),
    "validation finalizers must depend on exact scheduled unit identities",
  );
  const targetFinalizer = finalizers.find((unit) => unit.id.startsWith("browser_target_summary:"));
  assert.equal(
    targetFinalizer.command.env.CARTULARY_DEFER_OBSERVABILITY_FINALIZE,
    "1",
    "browser target summary must defer observability until terminal scheduler evidence exists",
  );
});

test("scheduled browser groups preserve exact generated row subsets", () => {
  const manifest = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
  const stage = resolveBrowserBatchStage(manifest, "stateful");
  const group = stage.groups.find((entry) => entry.selectedRowIDs.length > 1);
  assert.ok(group);
  assert.deepEqual(selectedBrowserGroupRowIDs(group, {}), group.selectedRowIDs);
  const subset = [...group.selectedRowIDs].sort().slice(0, 1);
  assert.deepEqual(
    selectedBrowserGroupRowIDs(group, {
      CARTULARY_BROWSER_SELECTED_ROW_IDS: subset.join(","),
    }),
    subset,
  );
  assert.throws(
    () =>
      selectedBrowserGroupRowIDs(group, {
        CARTULARY_BROWSER_SELECTED_ROW_IDS: "",
      }),
    /must be non-empty, sorted, and unique/u,
  );
  assert.throws(
    () =>
      selectedBrowserGroupRowIDs(group, {
        CARTULARY_BROWSER_SELECTED_ROW_IDS: [
          group.selectedRowIDs[1],
          group.selectedRowIDs[0],
        ].join(","),
      }),
    /must be non-empty, sorted, and unique/u,
  );
  assert.throws(
    () =>
      selectedBrowserGroupRowIDs(group, {
        CARTULARY_BROWSER_SELECTED_ROW_IDS: `${subset[0]},${subset[0]}`,
      }),
    /must be non-empty, sorted, and unique/u,
  );
  assert.throws(
    () =>
      selectedBrowserGroupRowIDs(group, {
        CARTULARY_BROWSER_SELECTED_ROW_IDS: "module.unknown.browser.row",
      }),
    /outside the generated group/u,
  );
});

test("visual snapshot update reuses profile-aware sessions without validation finalizers", () => {
  const manifest = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
  const stage = resolveBrowserBatchStage(manifest, "visual");
  const defaultGroup = stage.groups.find((group) => group.runtimeProfileID === "default");
  const claimedGroup = stage.groups.find((group) => group.runtimeProfileID === "network_flow_claimed");
  assert.ok(defaultGroup);
  assert.ok(claimedGroup);
  const sharedStage = {
    ...stage,
    groups: [
      defaultGroup,
      {
        ...defaultGroup,
        name: `${defaultGroup.name}-shared-companion`,
        selectedRowIDs: [`${defaultGroup.selectedRowIDs[0]}.companion`],
      },
      claimedGroup,
    ],
  };
  const schedule = buildBrowserStageSchedule(sharedStage, manifest, {
    mode: "snapshot_update",
    target: "browser-e2e-visual-update",
  });
  assert.equal(schedule.target, "browser-e2e-visual-update");
  assert.equal(schedule.totalWorkUnits, 2, "one stack must serve each session/profile tuple");
  assert.equal(schedule.finalizerCount, 0);
  assert.equal(schedule.workUnits.length, 2);
  assert.ok(schedule.workUnits.every((unit) => unit.kind === "browser_stage_session"));
  assert.ok(schedule.workUnits.every((unit) => unit.target === "browser-e2e-visual-update"));
  assert.ok(
    schedule.workUnits.every(
      (unit) =>
        unit.command.command.endsWith("tools/harness/browser/start-web-e2e.sh") &&
        unit.command.env.CARTULARY_BROWSER_MAINTENANCE_MODE === "snapshot_update" &&
        unit.command.env.CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS === "1",
    ),
    "every update session must retain owned stack cleanup and snapshot propagation",
  );
  const defaultSession = schedule.workUnits.find(
    (unit) => unit.command.env.CARTULARY_BROWSER_RUNTIME_PROFILE_ID === "default",
  );
  assert.equal(
    defaultSession.command.env.CARTULARY_BROWSER_SELECTED_GROUPS,
    [defaultGroup.name, `${defaultGroup.name}-shared-companion`].sort().join(","),
  );
  assert.throws(
    () => buildBrowserStageSchedule(stage, manifest, { mode: "unsupported" }),
    /unsupported browser scheduler mode/u,
  );
  const helper = readFileSync(
    path.join(repoRoot, "tools/harness/browser/run-browser-e2e-visual-update.sh"),
    "utf8",
  );
  assert.match(helper, /visual --mode snapshot_update/u);
  assert.doesNotMatch(helper, /awk|stage-runner|start-web-e2e/u);
});

test("service-backed Go shard units are executable by their declared targets", () => {
  const { serviceBacked } = renderedArtifacts();
  const planRows = collectTargetPlanRows(repoRoot);

  for (const schedule of serviceBacked.schedules ?? []) {
    const goUnits = (schedule.work_units ?? []).filter(
      (unit) => unit.kind === "go_shard" || unit.kind === "aggregate_finalize",
    );
    const selectionByTarget = new Map();
    for (const unit of goUnits.filter(
      (candidate) => candidate.kind === "aggregate_finalize",
    )) {
      const selectionScope = unit.env?.CARTULARY_GO_SCHEDULE_SCOPE ?? "";
      const selectionRows = unit.env?.CARTULARY_GO_SCHEDULED_ROW_IDS ?? "";
      assert.ok(
        ["all", "default_check", "rows"].includes(selectionScope),
        `${schedule.target ?? schedule.name} ${unit.id} must carry a closed scheduled Go scope`,
      );
      const selectedRowIDs = selectionRows.split(",").filter(Boolean);
      if (selectionScope === "rows") {
        assert.ok(
          selectedRowIDs.length > 0,
          `${schedule.target ?? schedule.name} ${unit.id} explicit Go scope must carry rows`,
        );
        assert.deepEqual(
          selectedRowIDs,
          [...new Set(selectedRowIDs)].sort((left, right) =>
            left.localeCompare(right),
          ),
          `${schedule.target ?? schedule.name} ${unit.id} scheduled Go rows must be sorted and unique`,
        );
      } else {
        assert.equal(
          selectionRows,
          "",
          `${schedule.target ?? schedule.name} ${unit.id} ${selectionScope} scope must not carry row IDs`,
        );
      }
      const selectionValue = `${selectionScope}\u0000${selectionRows}`;
      assert.equal(
        selectionByTarget.has(unit.target),
        false,
        `${schedule.target ?? schedule.name} ${unit.target} must have one Go selection owner`,
      );
      selectionByTarget.set(unit.target, selectionValue);
    }
    for (const unit of goUnits.filter(
      (candidate) => candidate.kind === "go_shard",
    )) {
      assert.ok(
        selectionByTarget.has(unit.target),
        `${schedule.target ?? schedule.name} ${unit.id} must resolve through its aggregate selection owner`,
      );
      assert.equal(
        unit.env?.CARTULARY_GO_SCHEDULE_SCOPE,
        undefined,
        `${schedule.target ?? schedule.name} ${unit.id} must not duplicate target-wide selection state`,
      );
    }
    for (const [target, selectionValue] of selectionByTarget) {
      const [selectionScope, selectionRows] = selectionValue.split("\u0000");
      const selectedIDs = new Set(selectionRows.split(",").filter(Boolean));
      const selectedRows =
        selectionScope === "all"
          ? planRows
          : selectionScope === "default_check"
            ? planRows.filter((row) => row.default_check_required === true)
            : planRows.filter((row) => selectedIDs.has(row.id));
      if (selectionScope === "rows") {
        assert.equal(
          selectedRows.length,
          selectedIDs.size,
          `${schedule.target ?? schedule.name} ${target} scheduled Go row universe must resolve exactly`,
        );
      }
      const executableShardNames = new Set(
        collectGoShardsForTargetFromRows(repoRoot, selectedRows, target).map(
          (shard) => shard.name,
        ),
      );
      for (const unit of goUnits.filter(
        (candidate) => candidate.kind === "go_shard" && candidate.target === target,
      )) {
        assert.ok(
          executableShardNames.has(unit.shard),
          `${schedule.target ?? schedule.name} schedules ${unit.shard} for ${target} outside its declared row universe`,
        );
      }
    }
  }
});

test("Go shard finalizers declare exactly their selected shard needs", () => {
  const schedulerManifest = readJSON("tools/scheduler_manifest.json");
  for (const schedule of schedulerManifest.schedules ?? []) {
    const completionKeys = new Set(
      (schedule.work_units ?? []).flatMap((unit) => unit.completion_keys ?? [unit.id]),
    );
    for (const unit of schedule.work_units ?? []) {
      if (unit.command?.type !== "go_shard_finalize") {
        continue;
      }
      assert.ok(
        Array.isArray(unit.shard_names) && unit.shard_names.length > 0,
        `${schedule.target} ${unit.id} must declare selected shard_names`,
      );
      const expectedNeeds = unit.shard_names.map((shardName) => `go_shard:${shardName}`);
      for (const expectedNeed of expectedNeeds) {
        assert.ok(
          (unit.needs ?? []).includes(expectedNeed),
          `${schedule.target} ${unit.id} missing selected shard need ${expectedNeed}`,
        );
        assert.ok(
          completionKeys.has(expectedNeed),
          `${schedule.target} ${unit.id} selected shard need ${expectedNeed} has no producer`,
        );
      }
      for (const need of (unit.needs ?? []).filter((entry) => entry.startsWith("go_shard:"))) {
        assert.ok(
          expectedNeeds.includes(need),
          `${schedule.target} ${unit.id} has shard need ${need} outside shard_names`,
        );
      }
    }
  }
});

function targetRecipeBlock(renderedMake, target) {
  const lines = renderedMake.split("\n");
  const headerPattern = new RegExp(`^${target.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}:`);
  const start = lines.findIndex(
    (line) => headerPattern.test(line) && !line.includes(": export "),
  );
  assert.notEqual(start, -1, `${target} must have a rendered recipe`);
  const block = [];
  for (let index = start; index < lines.length; index += 1) {
    if (index > start && /^[A-Za-z0-9_.-]+:/.test(lines[index])) {
      break;
    }
    block.push(lines[index]);
  }
  return block;
}

test("public targets declare command identity and semantic value", () => {
  const { taskSurface, browserBatch, serviceBacked } = renderedArtifacts();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  assert.ok(publicTargets.length > 0, "public target registry must not be empty");
  const commandIDs = new Set();
  for (const target of publicTargets) {
    assert.match(
      target.command_id,
      /^cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v[1-9][0-9]*$/,
      `${target.name} must declare stable command_id`,
    );
    assert.ok(!commandIDs.has(target.command_id), `${target.name} command_id must be unique`);
    commandIDs.add(target.command_id);
    assert.match(
      target.family_id,
      /^[a-z][a-z0-9_]*$/,
      `${target.name} must declare family_id`,
    );
    assert.equal(
      target.lifecycle_state,
      "public_active",
      `${target.name} must declare a current public lifecycle state`,
    );
    assert.ok(
      Array.isArray(target.semantic_behaviors) &&
        target.semantic_behaviors.length > 0,
      `${target.name} must declare semantic behaviors`,
    );
    assert.ok(
      Array.isArray(target.side_effects) && target.side_effects.length > 0,
      `${target.name} must declare side effects`,
    );
    assert.ok(
      target.input_contract &&
        target.input_contract.undeclared_make_command_line === "usage_error" &&
        target.input_contract.undeclared_inherited_env === "ignore" &&
        Array.isArray(target.input_contract.inputs),
      `${target.name} must declare a closed public input contract`,
    );
    for (const entry of target.semantic_behaviors) {
      assert.match(entry.owner_section, /^Section (?:[1-9]|1[0-9])(?:\.[0-9]+)?$/);
    }
    for (const entry of target.side_effects) {
      assert.match(entry.owner_section, /^Section (?:[1-9]|1[0-9])(?:\.[0-9]+)?$/);
    }
  }

  const duplicateID = structuredClone(taskSurface);
  duplicateID.targets.find((entry) => entry.name === "help-all").command_id =
    duplicateID.targets.find((entry) => entry.name === "help").command_id;
  assert.match(
    collectTaskSurfaceManifestErrors(duplicateID, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help-all\.command_id duplicates help/,
  );

  const malformedID = structuredClone(taskSurface);
  malformedID.targets.find((entry) => entry.name === "help").command_id =
    "cartulary.harness.command.help.latest";
  assert.match(
    collectTaskSurfaceManifestErrors(malformedID, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.command_id must match cartulary\.harness\.command\.<name>\.vN/,
  );

  const missingSemantic = structuredClone(taskSurface);
  missingSemantic.targets.find((entry) => entry.name === "help").semantic_behaviors = [];
  assert.match(
    collectTaskSurfaceManifestErrors(missingSemantic, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.semantic_behaviors must declare at least one semantic behavior/,
  );

  const missingOwner = structuredClone(taskSurface);
  missingOwner.targets.find((entry) => entry.name === "help").semantic_behaviors = [
    { behavior: "diagnostic_synthesis", owner_section: "" },
  ];
  assert.match(
    collectTaskSurfaceManifestErrors(missingOwner, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.semantic_behaviors\[1\]\.owner_section must be a Section reference/,
  );

  const missingSideEffects = structuredClone(taskSurface);
  delete missingSideEffects.targets.find((entry) => entry.name === "help").side_effects;
  assert.match(
    collectTaskSurfaceManifestErrors(missingSideEffects, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.side_effects must declare at least one side-effect class/,
  );

  const missingInputContract = structuredClone(taskSurface);
  delete missingInputContract.targets.find((entry) => entry.name === "help").input_contract;
  assert.match(
    collectTaskSurfaceManifestErrors(missingInputContract, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.input_contract must be declared for public targets/,
  );

  const misplacedInputPolicy = structuredClone(taskSurface);
  misplacedInputPolicy.targets.find(
    (entry) => entry.name === "target-plan",
  ).input_contract.undeclared_make_command_line = "ignore";
  assert.match(
    collectTaskSurfaceManifestErrors(misplacedInputPolicy, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /target-plan\.input_contract\.undeclared_make_command_line must be usage_error/,
  );

  const invalidSideEffects = structuredClone(taskSurface);
  invalidSideEffects.targets.find((entry) => entry.name === "help").side_effects = [
    { class: "none", owner_section: "Section 4" },
    { class: "retained_artifacts", owner_section: "Section 8" },
  ];
  assert.match(
    collectTaskSurfaceManifestErrors(invalidSideEffects, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.side_effects\[2\]\.artifact_policy must be declared for retained_artifacts[\s\S]*help\.side_effects none is mutually exclusive with other classes/,
  );

  const duplicateSideEffects = structuredClone(taskSurface);
  duplicateSideEffects.targets.find((entry) => entry.name === "format").side_effects.push(
    structuredClone(
      duplicateSideEffects.targets
        .find((entry) => entry.name === "format")
        .side_effects.find((entry) => entry.class === "authored_source_write"),
    ),
  );
  assert.match(
    collectTaskSurfaceManifestErrors(duplicateSideEffects, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /format\.side_effects contains duplicate authored_source_write/,
  );

  const legacyFields = structuredClone(taskSurface);
  const legacyHelp = legacyFields.targets.find((entry) => entry.name === "help");
  legacyHelp.classification = "public";
  legacyHelp.included_in = ["helper_only"];
  assert.match(
    collectTaskSurfaceManifestErrors(legacyFields, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /help\.classification is obsolete; use target_class[\s\S]*help\.included_in is obsolete; use default_inclusion_sets/,
  );

  const privateHelperOnly = structuredClone(taskSurface);
  privateHelperOnly.targets.find(
    (entry) => entry.name === "run-harness-smoke-fast",
  ).default_inclusion_sets = ["helper_only"];
  assert.match(
    collectTaskSurfaceManifestErrors(privateHelperOnly, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /run-harness-smoke-fast\.default_inclusion_sets helper_only is only valid for public direct-invocation targets/,
  );

  const shallowAlias = structuredClone(taskSurface);
  shallowAlias.targets.push({
    name: "synthetic-shallow-wrapper",
    target_class: "public",
    default_inclusion_sets: ["helper_only"],
    family_id: "help_discovery",
    lifecycle_state: "public_active",
    command_id: "cartulary.harness.command.synthetic_shallow_wrapper.v1",
    semantic_behaviors: [],
    side_effects: [{ class: "none", owner_section: "Section 4" }],
    output_policy: structuredClone(
      shallowAlias.targets.find((entry) => entry.name === "help").output_policy,
    ),
  });
  shallowAlias.help_tiers[0].entries.push({
    target: "synthetic-shallow-wrapper",
    description: "synthetic shallow wrapper",
  });
  shallowAlias.make_recipes["synthetic-shallow-wrapper"] = {
    type: "aggregate",
    prerequisites: ["help"],
  };
  assert.match(
    collectTaskSurfaceManifestErrors(shallowAlias, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /synthetic-shallow-wrapper\.semantic_behaviors must declare at least one semantic behavior/,
  );
});

test("public non-interactive wrappers run preflight before child work", () => {
  const { taskSurface, taskSurfaceMake } = renderedArtifacts();
  const recipes = taskSurface.make_recipes;
  for (const target of taskSurface.targets) {
    if (
      target.target_class !== "public" ||
      target.output_policy?.output_class === "interactive_raw"
    ) {
      continue;
    }
    const block = targetRecipeBlock(taskSurfaceMake, target.name);
    const recipeLines = block.filter((line) => line.startsWith("\t"));
    assert.ok(recipeLines.length > 0, `${target.name} must render recipe lines`);
    const escapedName = target.name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    if (target.name === "bootstrap-node-runtime") {
      assert.match(
        recipeLines[0],
        /\$\(MAKE\) --silent --no-print-directory \$\(NODE_BIN\)/,
        "bootstrap-node-runtime must provision Node through the owner readiness path",
      );
      assert.ok(
        !recipeLines.some((line) => line.includes("$(RUN_HARNESS_PREFLIGHT)")),
        "bootstrap-node-runtime must not require Node-backed preflight before Node exists",
      );
      continue;
    }
    const preflightLineIndex = recipeLines.findIndex((line) =>
      new RegExp(
        `^\\t\\$\\(Q\\)\\$\\(call RUN_PUBLIC_PREFLIGHT,${escapedName}\\)$`,
      ).test(line),
    );
    assert.match(
      recipeLines[preflightLineIndex] ?? "",
      new RegExp(`\\$\\(call RUN_PUBLIC_PREFLIGHT,${escapedName}\\)$`),
      `${target.name} must run public preflight before child work`,
    );
    const prerequisites = recipes[target.name]?.prerequisites ?? [];
    if (prerequisites.includes("$(NODE_BIN)")) {
      assert.match(
        recipeLines.slice(0, preflightLineIndex).join("\n"),
        /\$\(MAKE\) --silent --no-print-directory \$\(NODE_BIN\)/,
        `${target.name} must make Node ready before owner CLI preflight`,
      );
    }
    if (prerequisites.some((prerequisite) => prerequisite !== "$(NODE_BIN)")) {
      assert.match(
        recipeLines.slice(preflightLineIndex + 1).join("\n"),
        /CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES/,
        `${target.name} prerequisite work must follow preflight`,
      );
    }
  }
});

test("per-target input contract rejects misplaced Make variables and ignores ambient env", () => {
  for (const [sources, message] of [
    ["UNDECLARED_INPUT=command-line", /contains invalid source token/],
    ["UNDECLARED_INPUT=cli UNDECLARED_INPUT=env", /contains duplicate UNDECLARED_INPUT/],
    ["NOT_A_PUBLIC_INPUT=cli", /contains unknown input NOT_A_PUBLIC_INPUT/],
  ]) {
    assert.throws(
      () =>
        preflightPublicTarget("target-plan", {
          CARTULARY_MAKE_INPUT_SOURCES: sources,
        }),
      (error) => error instanceof HarnessConfigError && message.test(error.message),
    );
  }
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        OWNER: "module.networkflow",
        CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      /OWNER is not declared for target target-plan/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      OWNER: "module.networkflow",
      CARTULARY_MAKE_INPUT_SOURCES: "OWNER=env",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        TASK_SURFACE_MANIFEST: "/tmp/override.json",
        CARTULARY_MAKE_INPUT_SOURCES: "TASK_SURFACE_MANIFEST=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "configuration_error" &&
      /TASK_SURFACE_MANIFEST is an internal harness input/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      TASK_SURFACE_MANIFEST: "/tmp/override.json",
      CARTULARY_MAKE_INPUT_SOURCES: "TASK_SURFACE_MANIFEST=env",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        OPERATOR_BIN: "/tmp/operator",
        CARTULARY_MAKE_INPUT_SOURCES: "OPERATOR_BIN=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /OPERATOR_BIN is not declared for target target-plan/.test(error.message),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        CARTULARY_OPERATOR_BIN: "/tmp/operator",
        CARTULARY_MAKE_INPUT_SOURCES: "CARTULARY_OPERATOR_BIN=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "configuration_error" &&
      /CARTULARY_OPERATOR_BIN is an internal harness input/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      OPERATOR_BIN: "/tmp/operator",
      CARTULARY_OPERATOR_BIN: "/tmp/operator",
      CARTULARY_MAKE_INPUT_SOURCES: "OPERATOR_BIN=env CARTULARY_OPERATOR_BIN=env",
    }),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("frontend-unit", {
      VITEST_MAX_WORKERS: "4",
      CARTULARY_MAKE_INPUT_SOURCES: "VITEST_MAX_WORKERS=cli",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("frontend-unit", {
        VITEST_MAX_WORKERS: "17",
        CARTULARY_MAKE_INPUT_SOURCES: "VITEST_MAX_WORKERS=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /VITEST_MAX_WORKERS must be a positive integer <= 16/.test(error.message),
  );
  assert.throws(
    () =>
      preflightPublicTarget("frontend-unit", {
        VITEST_FLAGS: "apps/web/src/example.test.tsx",
        CARTULARY_MAKE_INPUT_SOURCES: "VITEST_FLAGS=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /VITEST_FLAGS is not declared for target frontend-unit/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("frontend-unit", {
      VITEST_FLAGS: "apps/web/src/example.test.tsx",
      CARTULARY_MAKE_INPUT_SOURCES: "VITEST_FLAGS=env",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("db-reset", {
        CARTULARY_DESTRUCTIVE_CONFIRM: "object-store-reset",
        CARTULARY_MAKE_INPUT_SOURCES: "CARTULARY_DESTRUCTIVE_CONFIRM=cli",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /CARTULARY_DESTRUCTIVE_CONFIRM must be one of db-reset/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("db-reset", {
      CARTULARY_DESTRUCTIVE_CONFIRM: "db-reset",
      CARTULARY_MAKE_INPUT_SOURCES: "CARTULARY_DESTRUCTIVE_CONFIRM=cli",
    }),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("db-reset", {
      CARTULARY_DESTRUCTIVE_CONFIRM: "db-reset",
      CARTULARY_MAKE_INPUT_SOURCES: "CARTULARY_DESTRUCTIVE_CONFIRM=env",
    }),
  );
});

test("extended harness contracts are explicit and outside default local check", () => {
  const { checkSchedule, taskSurface } = renderedArtifacts();
  const check = checkSchedule.schedules.find(
    (schedule) => schedule.target === "check",
  );
  assert.ok(check, "rendered check schedule must include check");
  const harnessContracts = check.work_units.find(
    (unit) => unit.target === "harness-contract-tests",
  );
  assert.equal(harnessContracts, undefined, "default local check must omit deep harness contract tests");
  const contractTarget = taskSurface.targets.find((target) => target.name === "harness-contract");
  assert.ok(contractTarget, "task surface must expose the explicit harness-contract target");
  assert.deepEqual(
    contractTarget.default_inclusion_sets,
    ["ci", "release-check"],
    "harness-contract must be selected by extended gates only",
  );
});

test("browser topology is catalog-derived, semantic, and profile-isolated", () => {
  const { browserBatch } = renderedArtifacts();
  assert.equal(browserBatch.schema_id, "cartulary.browser_e2e_batch_manifest.v6");
  assert.doesNotMatch(JSON.stringify(browserBatch), /(?:selected_phase|\bFE-|\bE-[0-9]|phase[0-9])/u);
  const catalog = loadTestCatalog(repoRoot);
  const selectorStageByBatchStage = new Map([
    ["webserver-backed", "webserver_backed"],
    ["support", "support"],
    ["stateful", "stateful"],
    ["measurement", "measurement"],
    ["a11y", "accessibility"],
    ["visual", "visual"],
  ]);
  for (const [batchStage, selectorStage] of selectorStageByBatchStage) {
    const stage = browserBatch.stages.find((entry) => entry.name === batchStage);
    assert.ok(stage, `missing generated browser stage ${batchStage}`);
    const selected = stage.groups.flatMap((group) => group.selected_row_ids).sort();
    const expected = catalog.rows
      .filter((row) => row.runner === "playwright" && row.selector.stage === selectorStage)
      .map((row) => row.row_id)
      .sort();
    assert.deepEqual(selected, expected, `${batchStage} must cover every catalog row exactly once`);
    assert.equal(new Set(selected).size, selected.length);
    for (const group of stage.groups) {
      assert.equal(group.specs.length, 1);
      assert.ok(group.selected_row_ids.every((rowID) => catalog.rowByID.has(rowID)));
      if (group.runtime_profile_id === "network_flow_claimed") {
        assert.ok(group.browser_session_group);
        assert.match(group.browser_session_isolation_reason, /startup|configuration/u);
      }
    }
  }
});

test("authored browser topology rejects obsolete selection fields", () => {
  const topology = readJSON("tools/execution_topology_manifest.json");
  topology.browser_e2e_batch.stages[0].groups[0].selected_phase = "constructed-obsolete-value";
  const fixture = path.join(
    repoRoot,
    "tools",
    `execution-topology-obsolete-${process.pid}.json`,
  );
  try {
    writeFileSync(fixture, `${JSON.stringify(topology, null, 2)}\n`);
    assert.throws(
      () => loadExecutionTopology({ manifestPath: fixture }),
      /unknown key selected_phase/u,
    );
  } finally {
    rmSync(fixture, { force: true });
  }
});

test("operator build owns its transitive embedded asset input", () => {
  const makefile = readFileSync(path.join(repoRoot, "Makefile"), "utf8");
  const rule = makefile
    .split("\n")
    .find((line) => line.startsWith("$(OPERATOR_BIN):"));
  const recipe = makefile
    .split("\n")
    .find((line) => line.includes('--profile build-operator --cache-dir'));
  assert.ok(rule, "Makefile must declare the operator binary rule");
  assert.match(rule, /\$\(EMBEDDED_WEB_ASSET_STAMP\)/u);
  assert.match(rule, /\$\(EMBEDDED_WEB_ASSET_ARCHIVE\)/u);
  assert.match(rule, /\$\(EMBEDDED_WEB_ASSET_READY_STAMP\)/u);
  assert.ok(recipe, "Makefile must declare the operator build-artifact cache recipe");
  assert.match(recipe, /--input-dir "\$\(EMBEDDED_WEB_ASSET_DIR\)"/u);
});

test("default check service-backed browser work uses declared session groups", () => {
  const { serviceBacked, expandedCheckSchedule, taskSurface } = renderedArtifacts();
  const serviceCheck = serviceBacked.schedules.find(
    (schedule) => schedule.target === "check-service-backed",
  );
  assert.ok(serviceCheck, "service-backed sources must include check-service-backed");
  const browserSources = serviceCheck.work_unit_sources.filter(
    (source) => source.type === "browser_stage",
  );
  assert.deepEqual(
    new Map(browserSources.map((source) => [source.browser_stage, source.browser_session_group])),
    new Map([
      ["webserver-backed", "default-check-browser-shared"],
      ["stateful", "default-check-stateful-isolated"],
    ]),
  );
  assert.equal(
    browserSources.find((source) => source.browser_stage === "stateful")
      ?.browser_session_isolation_reason,
    "stateful browser evidence mutates persisted runtime state and remains isolated from shared default-check browser work",
  );

  const check = expandedCheckSchedule.schedules.find(
    (schedule) => schedule.target === "check",
  );
  const browserSessions = check.work_units.filter(
    (unit) => unit.kind === "browser_stage_session",
  );
  const harnessServerTarget = taskSurface.targets.find(
    (target) => target.name === "build-server-harness",
  );
  assert.deepEqual(
    harnessServerTarget?.default_inclusion_sets,
    ["check"],
    "the harness server must be an explicit default-check readiness producer",
  );
  const harnessServerUnits = check.work_units.filter(
    (unit) => unit.target === "build-server-harness",
  );
  assert.equal(
    harnessServerUnits.length,
    1,
    "expanded check must retain exactly one harness-server producer",
  );
  assert.deepEqual(
    harnessServerUnits[0].needs,
    ["embedded-web-assets"],
    "the harness server must wait for its complete embedded asset input",
  );
  const operatorUnits = check.work_units.filter((unit) => unit.target === "build-operator");
  assert.equal(operatorUnits.length, 1, "expanded check must retain exactly one operator producer");
  assert.deepEqual(
    operatorUnits[0].needs,
    ["embedded-web-assets"],
    "the operator must wait for its complete transitive embedded asset input",
  );
  assert.ok(
    check.work_units.some((unit) => unit.target === "build-server"),
    "the independent deployable-server build gate must remain selected",
  );
  for (const session of browserSessions) {
    assert.ok(session.needs.includes("build-web"));
    assert.ok(session.needs.includes("build-server-harness"));
    assert.ok(session.needs.includes("build-migrate"));
    assert.equal(
      session.needs.includes("build-server"),
      false,
      "browser sessions must consume the harness server rather than the deployable build",
    );
  }
  const statefulSource = browserSources.find((source) => source.browser_stage === "stateful");
  const expectedStatefulSessionGroups = [
    ...new Set(statefulSource.groups.map((group) => group.browser_session_group)),
  ];
  assert.equal(browserSessions.length, 1 + expectedStatefulSessionGroups.length);
  assert.deepEqual(
    browserSessions.map((unit) => unit.browser_session_group).sort(),
    ["default-check-browser-shared", ...expectedStatefulSessionGroups].sort(),
  );
  for (const session of browserSessions.filter((unit) => unit.browser_stage === "stateful")) {
    assert.equal(
      session.browser_session_isolation_reason,
      "stateful browser evidence mutates persisted runtime state and remains isolated from shared default-check browser work",
    );
  }
  const webserverBrowserGroups = check.work_units.filter(
    (unit) =>
      unit.kind === "browser_group" &&
      unit.aggregate_target === "browser-e2e-webserver-backed",
  );
  assert.ok(webserverBrowserGroups.length > 0);
  const catalog = loadTestCatalog(repoRoot);
  const catalogByID = catalog.rowByID;
  for (const unit of webserverBrowserGroups) {
    assert.equal(unit.browser_group?.kind, "duration_balanced_specs");
    assert.ok(unit.browser_group?.selected_row_ids.length > 0);
    for (const rowID of unit.browser_group.selected_row_ids) {
      assert.equal(catalogByID.get(rowID)?.default_check, true);
      assert.equal(catalogByID.get(rowID)?.selector.stage, "webserver_backed");
    }
  }
  for (const excludedBrowserTarget of [
    "browser-e2e-measurement",
    "browser-e2e-visual",
    "browser-e2e-a11y",
  ]) {
    assert.equal(
      check.work_units.some(
        (unit) =>
          unit.aggregate_target === excludedBrowserTarget ||
          unit.target === excludedBrowserTarget,
      ),
      false,
      `${excludedBrowserTarget} must remain outside default local check`,
    );
  }
  assert.equal(
    check.work_units.filter(
      (unit) =>
        unit.kind === "browser_group" &&
        unit.aggregate_target === "browser-e2e-stateful",
    ).length,
    statefulSource.groups.length,
  );
  const statefulBrowserGroups = check.work_units.filter(
    (unit) =>
      unit.kind === "browser_group" &&
      unit.aggregate_target === "browser-e2e-stateful",
  );
  assert.deepEqual(
    statefulBrowserGroups
      .map((unit) => [unit.browser_group?.name, unit.browser_group?.selected_row_ids ?? []])
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([, rowIDs]) => rowIDs),
    statefulSource.groups
      .map((group) => [group.name, group.selected_row_ids])
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([, rowIDs]) => rowIDs),
  );
  for (const unit of statefulBrowserGroups) {
    assert.ok(unit.browser_group.selected_row_ids.every((rowID) => catalogByID.get(rowID)?.default_check));
  }
  const assertStatefulSessionChains = (units, label) => {
    const previousBySession = new Map();
    for (const group of statefulSource.groups) {
      const sessionGroup = group.browser_session_group ?? statefulSource.browser_session_group;
      const unit = units.find(
        (candidate) =>
          candidate.kind === "browser_group" && candidate.browser_group?.id === group.id,
      );
      assert.ok(unit, `${label} missing ${group.id}`);
      const previous = previousBySession.get(sessionGroup);
      assert.deepEqual(
        unit.needs,
        [
          `browser_stage_session:${sessionGroup}`,
          ...(previous ? [`browser_group:${previous}`] : []),
        ],
        `${label} must serialize ${group.id} only behind its own prior session partition`,
      );
      previousBySession.set(sessionGroup, group.id);
    }
  };
  assertStatefulSessionChains(check.work_units, "expanded check stateful schedule");
  assertBrowserWorkerSlots(
    check.work_units.filter(
      (unit) =>
        unit.kind === "browser_group" &&
        unit.service_session?.target === "check-service-backed",
    ),
    "default check service-backed browser groups",
  );
  const serviceBackedCheckSource = serviceBacked.schedules.find(
    (schedule) => schedule.target === "check-service-backed",
  );
  assert.ok(
    serviceBackedCheckSource,
    "rendered service-backed artifact must include check-service-backed",
  );
  const serviceBackedCheckUnits = expandServiceBackedSchedule({
    repoRoot,
    serviceSchedule: serviceBackedCheckSource,
  });
  assertBrowserWorkerSlots(
    serviceBackedCheckUnits.filter((unit) => unit.kind === "browser_group"),
    "direct check-service-backed browser groups",
  );
  assertStatefulSessionChains(
    serviceBackedCheckUnits,
    "direct check-service-backed stateful schedule",
  );
});

test("harness import boundary has no forbidden planning edges", () => {
  const report = collectHarnessImportBoundaryViolations(repoRoot);
  assert.deepEqual(report.violations, []);
  assert.deepEqual(report.forbidden_sccs, []);
});

test("backend module boundary rejects Revisions source SQL, mappings, and provider imports", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "backend-boundary."));
  try {
    writeFixtureFile(
      root,
      "internal/modules/revisions/source_sql.go",
      "package revisions\n\nconst sourceSQL = `UPDATE hosts SET display_name = $1`\n",
    );
    writeFixtureFile(
      root,
      "internal/modules/revisions/source_mapping.go",
      'package revisions\n\nvar sourceMapping = map[string]string{"host.display_name": "display_name"}\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/revisions/provider_import.go",
      'package revisions\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/rollbackprovider"\n',
    );
    writeFixtureFile(
      root,
      "internal/app/server/direct_provider.go",
      'package server\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/artifacts/deleterestore"\n',
    );
    writeFixtureFile(
      root,
      "internal/app/server/unrelated_incident_bundle.go",
      'package server\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/incidentbundles"\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/assessments/cross_owner_provider.go",
      'package assessments\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/artifacts/rollbackprovider"\n',
    );

    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
        "--manifest",
        path.join(repoRoot, "tools/backend_module_boundaries.json"),
        "--root",
        root,
      ],
      { cwd: repoRoot, encoding: "utf8" },
    );
    assert.notEqual(result.status, 0, `synthetic violations unexpectedly passed: ${result.stdout}`);
    const report = JSON.parse(result.stdout.trim());
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "source_table_access" &&
          violation.path === "internal/modules/revisions/source_sql.go" &&
          violation.symbol_or_import === "hosts",
      ),
      `source SQL violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "sql_table_allowlist" &&
          violation.path === "internal/modules/revisions/source_sql.go" &&
          violation.symbol_or_import === "hosts",
      ),
      `SQL allowlist violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "source_mapping" &&
          violation.path === "internal/modules/revisions/source_mapping.go" &&
          violation.symbol_or_import === "host.",
      ),
      `source mapping violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "owner_port_only_import" &&
          violation.path === "internal/modules/revisions/provider_import.go" &&
          violation.symbol_or_import.endsWith("/entities/hostidentity/rollbackprovider"),
      ),
      `provider import violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "owner_port_only_import" &&
          violation.path === "internal/app/server/direct_provider.go",
      ),
      `direct application provider import violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "owner_port_only_import" &&
          violation.path === "internal/app/server/unrelated_incident_bundle.go",
      ),
      `unrelated incident-bundle facade import violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "owner_port_only_import" &&
          violation.path === "internal/modules/assessments/cross_owner_provider.go",
      ),
      `cross-owner provider import violation missing: ${result.stdout}`,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("backend module boundary consumes test support inventory scan exclusions", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "backend-boundary-support."));
  try {
    writeFixtureFile(
      root,
      "internal/modules/future/production.go",
      'package future\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/workbook/routes"\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/future/testsupport/routetest/support.go",
      'package routetest\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/workbook/routes"\n',
    );
    writeFixtureFile(
      root,
      "tools/test_support_inventory.json",
      JSON.stringify(
        {
          schema_id: "cartulary.test_support_inventory.v1",
          go_support_roots: [
            {
              path: "internal/modules/future/testsupport",
              owner: "future",
              posture: "owner_local",
              runtime_scan: "excluded",
              support_scan: "included",
              service_starting: false,
              rationale: "Synthetic support root excluded from production boundary scans.",
            },
          ],
          shared_data_roots: [],
        },
        null,
        2,
      ),
    );

    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
        "--manifest",
        path.join(repoRoot, "tools/backend_module_boundaries.json"),
        "--support-inventory",
        path.join(root, "tools/test_support_inventory.json"),
        "--root",
        root,
      ],
      { cwd: repoRoot, encoding: "utf8" },
    );
    assert.notEqual(result.status, 0, `synthetic production violation unexpectedly passed: ${result.stdout}`);
    const report = JSON.parse(result.stdout.trim());
    assert.ok(
      report.effective_scan_excludes.includes("internal/modules/future/testsupport/**"),
      `support inventory exclusion missing: ${result.stdout}`,
    );
    assert.deepEqual(
      report.violations.map((violation) => violation.path),
      ["internal/modules/future/production.go"],
      `support-root violation should be excluded: ${result.stdout}`,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("backend module boundary enforces exact command facades, revision assembly, and platform-only HTTP runtime imports", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "backend-boundary-runtime."));
  try {
    writeFixtureFile(
      root,
      "cmd/server/main.go",
      'package main\n\nimport _ "github.com/JochiRaider/cartulary/internal/app/server"\n',
    );
    writeFixtureFile(
      root,
      "cmd/server/main_test.go",
      'package main\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/recovery"\n',
    );
    writeFixtureFile(
      root,
      "cmd/migrate/main.go",
      'package main\n\nimport _ "github.com/JochiRaider/cartulary/internal/app/server"\n',
    );
    writeFixtureFile(
      root,
      "cmd/operator/main.go",
      'package main\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/recovery"\n',
    );
    writeFixtureFile(
      root,
      "internal/app/revisionassembly/revisions.go",
      'package revisionassembly\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/recovery"\n',
    );
    writeFixtureFile(
      root,
      "internal/platform/httpruntime/runtime.go",
      'package httpruntime\n\nimport _ "github.com/JochiRaider/cartulary/internal/platform/config"\n',
    );
    writeFixtureFile(
      root,
      "internal/platform/httpruntime/domain.go",
      'package httpruntime\n\nimport _ "github.com/JochiRaider/cartulary/internal/modules/recovery"\n',
    );

    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
        "--manifest",
        path.join(repoRoot, "tools/backend_module_boundaries.json"),
        "--root",
        root,
      ],
      { cwd: repoRoot, encoding: "utf8" },
    );
    assert.notEqual(result.status, 0, `synthetic runtime violations unexpectedly passed: ${result.stdout}`);
    const report = JSON.parse(result.stdout.trim());
    assert.deepEqual(
      report.violations
        .filter((violation) => violation.code === "go_import_allowlist")
        .map((violation) => violation.path)
        .sort(),
      [
        "cmd/migrate/main.go",
        "cmd/operator/main.go",
        "internal/app/revisionassembly/revisions.go",
        "internal/platform/httpruntime/domain.go",
      ],
      `thin-root/runtime allowlist violations differed: ${result.stdout}`,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("backend module boundary preserves command refactor retirement invariants", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "command-refactor-boundary."));
  try {
    for (const command of ["server", "migrate", "operator"]) {
      writeFixtureFile(root, `cmd/${command}/main.go`, "package main\n");
    }
    writeFixtureFile(root, "cmd/server/legacy_test.go", "package main\n");
    writeFixtureFile(
      root,
      "internal/app/legacy_cli.go",
      "package app\n\nfunc RunMigrateCLI() {}\n",
    );
    writeFixtureFile(
      root,
      "internal/app/server/server.go",
      'package server\n\nimport _ "github.com/JochiRaider/cartulary/internal/platform/harnessruntime"\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/recovery/operatorcli/legacy.go",
      'package operatorcli\n\nconst legacy = "backup-metadata"\n',
    );
    writeFixtureFile(
      root,
      "internal/modules/recovery/legacy_build_test.go",
      'package recovery\n\nconst legacyBuild = "go build"\n',
    );

    const result = spawnSync(
      process.execPath,
      [
        path.join(repoRoot, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
        "--manifest",
        path.join(repoRoot, "tools/backend_module_boundaries.json"),
        "--root",
        root,
      ],
      { cwd: repoRoot, encoding: "utf8" },
    );
    assert.notEqual(result.status, 0, `retired command fixtures unexpectedly passed: ${result.stdout}`);
    const report = JSON.parse(result.stdout.trim());
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "command_root_shape" &&
          violation.path === "cmd/server/legacy_test.go",
      ),
      `command root shape violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "forbidden_source_token" &&
          violation.symbol_or_import.includes("RunMigrateCLI"),
      ),
      `contextless wrapper violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "forbidden_go_import" &&
          violation.path === "internal/app/server/server.go",
      ),
      `production harness import violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "forbidden_source_token" &&
          violation.symbol_or_import.includes("backup-metadata"),
      ),
      `retired recovery token violation missing: ${result.stdout}`,
    );
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.code === "forbidden_test_build_token" &&
          violation.path === "internal/modules/recovery/legacy_build_test.go",
      ),
      `nested test build violation missing: ${result.stdout}`,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("harness import boundary consumes the authored helper ownership registry", () => {
  const helperOwnership = loadHarnessHelperOwnership(repoRoot);
  const authoredOwnerFacadePaths = ownerFacadePathLists(helperOwnership);
  const report = collectHarnessImportBoundaryViolations(repoRoot);
  assert.equal(helperOwnership.facades.length, 34);
  assert.deepEqual(
    Object.keys(report.owner_facades).sort(),
    Object.keys(authoredOwnerFacadePaths).sort(),
  );
  for (const [owner, paths] of Object.entries(authoredOwnerFacadePaths)) {
    assert.deepEqual(
      report.owner_facades[owner],
      [...paths].sort((left, right) => left.localeCompare(right)),
      `${owner} facade paths must come from the helper ownership registry`,
    );
    for (const ownerPath of paths) {
      assert.ok(
        existsSync(path.join(repoRoot, ownerPath)),
        `${owner} facade path must exist: ${ownerPath}`,
      );
    }
  }
});

test("helper ownership rejects duplicate keys and missing current facades", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "helper-owner."));
  try {
    writeFixtureFile(
      root,
      "tools/harness_helper_ownership.json",
      `${JSON.stringify({
        schema_id: "cartulary.harness_helper_ownership.v1",
        facades: [
          { key: "duplicate", boundary_group: "backend", paths: [], allowed_consumers: [] },
          { key: "duplicate", boundary_group: "backend", paths: [], allowed_consumers: [] },
        ],
      })}\n`,
    );
    assert.throws(() => loadHarnessHelperOwnership(root), /duplicates facade key duplicate/u);

    writeFixtureFile(
      root,
      "tools/harness_helper_ownership.json",
      `${JSON.stringify({
        schema_id: "cartulary.harness_helper_ownership.v1",
        facades: [
          {
            key: "missing",
            boundary_group: "backend",
            paths: ["tools/harness/backend/missing.mjs"],
            allowed_consumers: [],
          },
        ],
      })}\n`,
    );
    assert.throws(() => loadHarnessHelperOwnership(root), /references missing facade/u);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("harness import boundary rejects unknown top-level owner roots", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "harness-owner-root."));
  try {
    writeFixtureFile(
      root,
      "tools/harness/mystery/private.mjs",
      "export const privateHelper = true;\n",
    );
    const report = collectHarnessImportBoundaryViolations(root);
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.rule === "forbidden_unknown_harness_owner_root" &&
          violation.source === "tools/harness/mystery",
      ),
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("harness import boundary rejects private core imports", () => {
  const root = mkdtempSync(path.join(repoRoot, "tmp", "harness-core-boundary."));
  const privateCoreImportPath = "../core/public-contract.mjs";
  const privateCoreTarget = ["tools", "harness", "core", "public-contract.mjs"].join("/");
  try {
    writeFixtureFile(
      root,
      "tools/harness/contract/index.mjs",
      "export const contractFacade = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/output/index.mjs",
      "export const outputFacade = true;\n",
    );
    writeFixtureFile(
      root,
      "tools/harness/backend/uses-facade.mjs",
      `${fixtureImport("../contract/index.mjs")}export const backend = true;\n`,
    );
    writeFixtureFile(
      root,
      "tools/harness/browser/uses-output.mjs",
      `${fixtureImport("../output/index.mjs")}export const browser = true;\n`,
    );

    const clean = collectHarnessImportBoundaryViolations(root);
    assert.deepEqual(clean.violations, []);

    writeFixtureFile(
      root,
      "tools/harness/backend/uses-core.mjs",
      `${fixtureImport(privateCoreImportPath)}export const backendCore = true;\n`,
    );
    const report = collectHarnessImportBoundaryViolations(root);
    assert.ok(
      report.violations.some(
        (violation) =>
          violation.rule === "forbidden_private_core_import" &&
          violation.source === "tools/harness/backend/uses-core.mjs" &&
          violation.target === privateCoreTarget,
      ),
      "direct private core import must be reported",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("contract output and scheduler summaries preserve normalized public surface", () => {
  const runRoot = ".cartulary/test-results/contract-surface";
  const summary = buildToolRunSummary({
    target: "harness-contract",
    command: ["make", "harness-contract"],
    status: "fail",
    startedAt: "2026-01-01T00:00:00.000Z",
    completedAt: "2026-01-01T00:00:01.250Z",
    durationMs: 1250.4,
    outputMode: "machine",
    resultRoot: ".cartulary/test-results",
    runId: "contract-surface",
    runRoot,
    summaryArtifacts: [
      fileArtifactRef(
        "target_summary",
        `${runRoot}/harness-contract/target-summary.json`,
      ),
      fileArtifactRef(
        "tool_run_summary",
        `${runRoot}/harness-contract/tool-run-summary.json`,
      ),
      fileArtifactRef(
        "scheduler_summary",
        `${runRoot}/harness-contract/scheduler-summary.json`,
      ),
    ],
    logArtifacts: [
      fileArtifactRef(
        "scheduler_events",
        `${runRoot}/harness-contract/scheduler-events.jsonl`,
        "jsonl",
      ),
    ],
    workUnits: [
      { id: "unit-z", status: "pass" },
      { id: "unit-a", status: "fail" },
    ],
    failures: [
      {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        target: "backend-unit",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 2,
        artifact: `${runRoot}/backend-unit/target-summary.json`,
      },
    ],
    slowest: [
      { id: "slow-b", duration_ms: 25 },
      { id: "slow-a", duration_ms: 25 },
    ],
  });
  assert.equal(summary.schema_id, toolRunSummarySchemaID);
  assert.equal(summary.output_mode, "machine");
  assert.equal(summary.exit_code, 10);
  assert.equal(summary.failure_class, "product");
  assert.equal(summary.failure_reason, "test_assertion_failure");
  assert.deepEqual(
    summary.summary_artifacts.map((artifact) => artifact.role),
    ["scheduler_summary", "target_summary", "tool_run_summary"],
  );
  assert.deepEqual(
    summary.work_units.map((unit) => unit.id),
    ["unit-a", "unit-z"],
  );
  assert.deepEqual(
    summary.slowest.map((entry) => entry.id),
    ["slow-a", "slow-b"],
  );
  assert.equal(machineOutput({ CARTULARY_OUTPUT_MODE: "machine" }), true);
  assert.equal(
    terminalArtifactPath(runRoot, `${runRoot}/tool-run-summary.json`),
    "tool-run-summary.json",
  );

  const schedulerLine = schedulerSummaryLine({
    target: "check",
    status: "fail",
    completed: 2,
    total: 3,
    failed: 1,
    failureClass: "product",
    failureReason: "test_assertion_failure",
    skipped: 1,
    finalizerFailures: 1,
    totalWallTimeMs: 1250,
    slowest: [{ label: "unit-a", duration_ms: 1250 }],
  });
  assert.match(schedulerLine, /^\[SUMMARY\] /u);
  assert.match(schedulerLine, /target=check/u);
  assert.match(schedulerLine, /failure_class=product/u);
  assert.match(schedulerLine, /reason=test_assertion_failure/u);
  assert.match(schedulerLine, /work_units=2\/3/u);
  assert.match(schedulerLine, /total_wall_time=1\.25s/u);
  assert.match(schedulerLine, /skipped=1/u);
  assert.match(schedulerLine, /finalizer_failures=1/u);

  const root = mkdtempSync(path.join(repoRoot, "tmp", "scheduler-events."));
  const eventFile = path.join(root, "scheduler-events.jsonl");
  try {
    writeFileSync(
      eventFile,
      [
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v7",
          target: "check",
          event: "scheduler-start",
          seq: 1,
          monotonic_ms: 0,
          emitted_at: "2026-01-01T00:00:00.000Z",
        }),
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v7",
          target: "check",
          event: "scheduler-finish",
          seq: 2,
          monotonic_ms: 5,
          emitted_at: "2026-01-01T00:00:00.005Z",
        }),
      ].join("\n") + "\n",
    );
    assert.deepEqual(validateSchedulerEventOrderFile(eventFile), []);

    writeFileSync(
      eventFile,
      [
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v7",
          target: "check",
          event: "scheduler-start",
          seq: 1,
          monotonic_ms: 10,
          emitted_at: "2026-01-01T00:00:00.010Z",
        }),
        JSON.stringify({
          schema_id: "cartulary.scheduler_event.v7",
          target: "check",
          event: "scheduler-finish",
          seq: 2,
          monotonic_ms: 5,
          emitted_at: "2026-01-01T00:00:00.020Z",
        }),
      ].join("\n") + "\n",
    );
    assert.match(
      validateSchedulerEventOrderFile(eventFile).join("\n"),
      /monotonic_ms regressed/u,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("primary public failure uses closed deterministic tie breakers", () => {
  assert.deepEqual(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "cleanup_error",
        lifecycle_step: "cleanup_finalizers",
        artifact: "z.log",
      },
      {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 8,
        child_registry_order: 2,
        artifact: "b.log",
      },
      {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 7,
        child_registry_order: 3,
        artifact: "a.log",
      },
    ]),
    {
      failure_class: "product",
      failure_reason: "test_assertion_failure",
      kind: "failure",
      source: "",
      target: "",
      step: "",
      runner: "",
      label: "",
      message: "",
      artifact: "a.log",
      lifecycle_step: "semantic_target_behavior",
      scheduler_event_sequence: 7,
      child_registry_order: 3,
    },
  );

  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "scheduler_accounting_error",
        lifecycle_step: "artifact_validation",
        artifact: "b.log",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "artifact_validation",
        artifact: "a.log",
      },
    ])?.failure_reason,
    "fixture_error",
  );

  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "artifact_validation",
        label: "late-lifecycle",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "configuration_resolution",
        label: "early-lifecycle",
      },
    ])?.label,
    "early-lifecycle",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 4,
        child_registry_order: 1,
        label: "later-event",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 9,
        label: "earlier-event",
      },
    ])?.label,
    "earlier-event",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 7,
        label: "later-child",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        label: "earlier-child",
      },
    ])?.label,
    "earlier-child",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "z.log",
        label: "later-artifact",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "earlier-artifact",
      },
    ])?.label,
    "earlier-artifact",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "tool_diagnostic_failure",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "later-reason",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "earlier-reason",
      },
    ])?.label,
    "earlier-reason",
  );
  assert.equal(
    primaryPublicFailure([
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "first-input",
      },
      {
        failure_class: "harness",
        failure_reason: "fixture_error",
        lifecycle_step: "semantic_target_behavior",
        scheduler_event_sequence: 3,
        child_registry_order: 2,
        artifact: "a.log",
        label: "second-input",
      },
    ])?.label,
    "first-input",
  );
});

test("failure class normalization rejects legacy aliases for current artifacts", () => {
  assert.equal(normalizeFailureClass("product"), "product");
  assert.equal(normalizeFailureClass("infra"), "infra");
  assert.equal(normalizeFailureClass("test", "unknown"), "unknown");
  assert.equal(normalizeFailureClass("helper", "unknown"), "unknown");
  assert.equal(normalizeFailureClass("infrastructure", "unknown"), "unknown");
  for (const field of [
    "lifecycle",
    "lifecycleStep",
    "scheduler_seq",
    "event_sequence",
    "seq",
    "child_target_order",
    "target_registry_order",
  ]) {
    assert.throws(
      () => normalizeFailureRecord({ [field]: field.includes("lifecycle") ? "setup" : 1 }),
      (error) =>
        error.failure_class === "artifact" &&
        error.failure_reason === "artifact_error" &&
        error.message.includes(`field ${field} is unsupported`),
      `${field} must fail closed`,
    );
  }
});

test("cleanup guard protects closed roots and permits cleanup-owned paths", () => {
  const output = [];
  const stdout = { write: (value) => output.push(value) };
  runCleanup({
    scope: "clean",
    candidates: ["go.mod", "db/migrations"],
    includeTmp: false,
    dryRun: true,
    stdout,
  });
  assert.match(output.join(""), /DRY-RUN retain go\.mod protected_root/);
  assert.match(output.join(""), /DRY-RUN retain db\/migrations protected_root/);
  assert.match(
    output.join(""),
    /DRY-RUN remove-children internal\/platform\/httpapi\/webassets\/dist registered_embedded_web_assets_preserve_keep/,
  );

  const tempRoot = mkdtempSync(path.join(repoRoot, "tmp", "cleanup-owned."));
  const owned = path.relative(repoRoot, tempRoot).replaceAll("\\", "/");
  writeFileSync(path.join(tempRoot, "artifact.txt"), "temporary");
  const beforeCleanupOutput = output.length;
  runCleanup({
    candidates: [owned, "tmp/missing-cleanup-owned-path"],
    includeTmp: false,
    dryRun: false,
    stdout,
  });
  assert.equal(existsSync(tempRoot), false, "cleanup-owned temp path must be removed");
  assert.equal(
    output.slice(beforeCleanupOutput).join(""),
    `removing ${owned}\n`,
    "missing cleanup-owned path must be a no-op",
  );
  rmSync(tempRoot, { recursive: true, force: true });
});

test("test route token generation and validation follow closed attach rules", () => {
  const generated = generateTestRouteToken();
  assert.equal(generated.length, 43);
  assert.match(generated, /^[A-Za-z0-9_-]{43}$/u);
  assert.equal(testRouteTokenValid(generated), true);
  assert.equal(testRouteTokenValid("short"), false);
  assert.equal(testRouteTokenValid("token"), false);
  assert.equal(testRouteTokenValid("a".repeat(43)), false);
  assert.equal(testRouteTokenValid(`${"a".repeat(42)}\n`), false);
});

test("retained log artifact resolution is typed, contained, and deterministic", () => {
  const runRoot = mkdtempSync(path.join(repoRoot, "tmp", "artifact-resolver."));
  const logs = path.join(runRoot, "target", "scheduler-logs");
  mkdirSync(logs, { recursive: true });
  writeFileSync(path.join(logs, "b.log"), "b\n");
  writeFileSync(path.join(logs, "a.log"), "a\n");
  writeFileSync(path.join(logs, "ignored.json"), "{}\n");
  const resolved = resolveRetainedLogArtifacts(runRoot, [
    {
      role: "scheduler_logs",
      path_kind: "directory",
      path: "target/scheduler-logs",
    },
  ]);
  assert.deepEqual(
    resolved.map((entry) => path.basename(entry.file)),
    ["a.log", "b.log"],
  );
  assert.throws(
    () => resolveRetainedLogArtifacts(runRoot, [{
      role: "escape",
      path_kind: "file",
      format: "log",
      path: "../outside.log",
    }]),
    /safe relative path/u,
  );
  assert.throws(
    () => resolveRetainedLogArtifacts(runRoot, [{
      role: "wrong_format",
      path_kind: "file",
      format: "json",
      path: "target/scheduler-logs/a.log",
    }]),
    /format must be log or text/u,
  );
  assert.throws(
    () => resolveRetainedLogArtifacts(runRoot, [{
      role: "missing",
      path_kind: "file",
      format: "log",
      path: "target/scheduler-logs/missing.log",
    }]),
    /does not exist/u,
  );
  assert.throws(
    () => resolveRetainedLogArtifacts(
      runRoot,
      [{
        role: "scheduler_logs",
        path_kind: "directory",
        path: "target/scheduler-logs",
      }],
      { maxFiles: 1, maxFileBytes: 1024, maxTotalBytes: 2048 },
    ),
    /file directory limit/u,
  );
  symlinkSync(path.join(logs, "a.log"), path.join(logs, "linked.log"));
  assert.throws(
    () => resolveRetainedLogArtifacts(runRoot, [{
      role: "scheduler_logs",
      path_kind: "directory",
      path: "target/scheduler-logs",
    }]),
    /symlink entry/u,
  );
  rmSync(runRoot, { recursive: true, force: true });
});

test("redaction uses closed structured keys and raw secret families", () => {
  const structured = redactValue({
    service_sessions: [
      {
        session_target: "browser-stage-token-name",
        cleanup_status: "pass",
        setup_duration_ms: 12,
        healthy: true,
        count: 3,
        absent: null,
        lease_file: "tmp/session/lease-file.json",
        session_token: "nested-session-token",
      },
    ],
    browser_stage_sessions: [
      {
        session_target: "browser-e2e-webserver-backed",
        cleanup_status: "skipped_no_lease",
        lease_file: "tmp/browser/session.json",
      },
    ],
    X_Cartulary_Test_Route_Token: "route-secret",
    CARTULARY_S3TEST_SECRET_ACCESS_KEY: "object-store-secret",
    session_target: "not-redacted-token-substring",
  });
  assert.equal(structured.service_sessions[0].session_target, "browser-stage-token-name");
  assert.equal(structured.service_sessions[0].cleanup_status, "pass");
  assert.equal(structured.service_sessions[0].setup_duration_ms, 12);
  assert.equal(structured.service_sessions[0].healthy, true);
  assert.equal(structured.service_sessions[0].count, 3);
  assert.equal(structured.service_sessions[0].absent, null);
  assert.equal(structured.service_sessions[0].lease_file, "tmp/session/lease-file.json");
  assert.equal(structured.service_sessions[0].session_token, "[REDACTED]");
  assert.equal(structured.browser_stage_sessions[0].session_target, "browser-e2e-webserver-backed");
  assert.equal(structured.browser_stage_sessions[0].cleanup_status, "skipped_no_lease");
  assert.equal(structured.browser_stage_sessions[0].lease_file, "tmp/browser/session.json");
  assert.equal(structured.X_Cartulary_Test_Route_Token, "[REDACTED]");
  assert.equal(structured.CARTULARY_S3TEST_SECRET_ACCESS_KEY, "[REDACTED]");
  assert.equal(structured.session_target, "not-redacted-token-substring");
  assert.deepEqual(
    redactValue(["--token", "cli-secret", "--target", "backend-unit"]),
    ["--token", "[REDACTED]", "--target", "backend-unit"],
  );

  const raw = redactString([
    "postgres://cartulary:supersecret@127.0.0.1:5432/postgres password=supersecret",
    "https://user:secret@example.test/path",
    "Authorization: Bearer abc.def.ghi",
    "X-Cartulary-Test-Route-Token: route-secret",
    "minio_secret_access_key=minio-secret",
    "-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
  ].join("\n"));
  for (const leaked of [
    "supersecret",
    "secret@example",
    "abc.def.ghi",
    "route-secret",
    "minio-secret",
    "BEGIN PRIVATE KEY",
  ]) {
    assert.equal(raw.includes(leaked), false, `raw redaction leaked ${leaked}: ${raw}`);
  }
});
