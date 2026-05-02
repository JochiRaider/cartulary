#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { copyFileSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root] = process.argv.slice(2);
process.chdir(root);

const topologyModule = await import(pathToFileURL(path.join(root, "scripts/lib/execution-topology.mjs")));
const targetPlanModule = await import(pathToFileURL(path.join(root, "scripts/lib/target-plan.mjs")));
const serviceRendererModule = await import(
  pathToFileURL(path.join(root, "scripts/render-service-backed-schedule-manifest.mjs"))
);
const taskSurfaceModule = await import(pathToFileURL(path.join(root, "scripts/lib/task-surface.mjs")));

const {
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
  topologySummary,
} = topologyModule;

const topology = loadExecutionTopology();
const summary = topologySummary(topology);
assert.equal(summary.schema_id, "cartulary.execution_topology.v2");
assert.ok(summary.execution_dependencies >= 10);
assert.ok(summary.go_targets >= 5);
assert.ok(summary.check_schedules >= 1);
assert.ok(summary.service_backed_schedules >= 3);

const renderedTaskSurface = renderTaskSurfaceManifest(topology);
const renderedBrowserBatch = renderBrowserBatchManifest(topology);
const renderedServiceBacked = serviceRendererModule.renderServiceBackedScheduleManifest({
  topology: "tools/execution_topology_manifest.json",
  topologyObject: topology,
});
const renderedTaskSurfaceErrors = taskSurfaceModule.collectTaskSurfaceManifestErrors(renderedTaskSurface, {
  browserBatchManifest: renderedBrowserBatch,
  serviceBackedScheduleManifest: renderedServiceBacked,
});
assert.deepEqual(renderedTaskSurfaceErrors, [], "rendered task surface must satisfy task-surface validation");
assert.deepEqual(
  renderedTaskSurface.targets.find((target) => target.name === "migration-drift")?.service_requirements,
  ["postgres"],
  "migration-drift must declare its Postgres service requirement",
);

const invalidHarnessReference = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidHarnessReference.harness_tiers.fast.checks.push("harness-smoke-missing");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidHarnessReference, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /harness_tiers\.fast\.checks references unknown harness check harness-smoke-missing/,
  "task-surface validation must reject harness tiers that reference unknown checks",
);

const invalidHarnessBackingScript = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidHarnessBackingScript.harness_checks.push({
  name: "harness-smoke-missing-script",
  backing_scripts: ["scripts/missing-harness-smoke.sh"],
});
invalidHarnessBackingScript.harness_tiers.fast.checks.push("harness-smoke-missing-script");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidHarnessBackingScript, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /harness-smoke-missing-script backing script missing: scripts\/missing-harness-smoke\.sh/,
  "task-surface validation must reject harness checks with missing backing scripts",
);

const invalidServiceRequirement = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidServiceRequirement.targets
  .find((target) => target.name === "migration-drift")
  .service_requirements.push("legacy-db");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidServiceRequirement, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /migration-drift\.service_requirements\[2\] has invalid service requirement "legacy-db"/,
  "task-surface validation must reject unknown service requirements",
);

const serialize = (value) => `${JSON.stringify(value, null, 2)}\n`;
const artifactSnapshot = () => ({
  taskSurface: serialize(renderTaskSurfaceManifest(topology)),
  browserBatch: serialize(renderBrowserBatchManifest(topology)),
  checkSchedule: serialize(renderCheckScheduleManifest(topology)),
  serviceBacked: serialize(
    serviceRendererModule.renderServiceBackedScheduleManifest({
      topology: "tools/execution_topology_manifest.json",
      topologyObject: topology,
    }),
  ),
  taskSurfaceMake: taskSurfaceModule.renderTaskSurfaceMake(renderTaskSurfaceManifest(topology)),
});
assert.deepEqual(artifactSnapshot(), artifactSnapshot(), "topology artifact rendering must be deterministic");

const renderedCheckSchedule = renderCheckScheduleManifest(topology);
const checkSchedule = renderedCheckSchedule.schedules.find((schedule) => schedule.target === "check");
assert.ok(checkSchedule, "rendered check schedule must include check");
assert.equal(checkSchedule.work_units.length, 32, "check schedule must render the current check work-unit set");
assert.deepEqual(
  checkSchedule.work_units.find((unit) => unit.target === "lint-shell")?.env,
  { LINT_SHELL_STRICT: "1" },
  "scheduled lint-shell must run ShellCheck in strict mode",
);
assert.deepEqual(
  checkSchedule.work_units.map((unit) => [unit.target, unit.weight]),
  [
    ["check-setup-blockers", 50000],
    ["check-build-prereqs", 40000],
    ["check-service-backed", 30000],
    ["check-go-test-duration-baseline-drift", 29000],
    ["check-browser-e2e-duration-baseline-drift", 28000],
    ["check-service-backed-make-target-duration-baseline-drift", 27900],
    ["migration-drift", 27000],
    ["deployable-shape", 26000],
    ["backend-unit", 25000],
    ["frontend-typecheck", 24000],
    ["lint-go", 23000],
    ["go-vulncheck", 22000],
    ["go-gosec-targeted", 21900],
    ["go-gosec-audit", 21800],
    ["frontend-unit", 15000],
    ["check-harness-smoke", 14000],
    ["check-harness-smoke-duration-baseline-drift", 13990],
    ["lint-biome", 13000],
    ["frontend-import-boundary-check", 12950],
    ["lint-scripts", 12900],
    ["lint-shell", 12850],
    ["phase-test-name-check", 12000],
    ["task-surface-check", 11900],
    ["browser-e2e-task-surface-check", 11800],
    ["frontend-task-surface-check", 11700],
    ["backend-task-surface-check", 11600],
    ["phase-map-check", 11500],
    ["go-test-duration-baseline-coverage", 11400],
    ["phase-ledger-drift", 11300],
    ["phase-schedule-drift", 11200],
    ["service-backed-unit-check", 11100],
    ["generate-drift", 11000],
  ],
  "profile-expanded check schedule must preserve the existing DAG priority order",
);

const rows = targetPlanModule.collectTargetPlanRows(root);
const scheduledBackendTargets = (rowsToUse) =>
  Array.from(
    new Set(
      rowsToUse
        .filter((row) => row.runner_family === "go_test" && row.service_backed && row.check_service_backed_safe)
        .map((row) => row.target),
    ),
  ).sort();
const originalBackendTargets = scheduledBackendTargets(rows);
const existingRow = rows.find((row) => row.execution_dependency === "backend_store");
assert.ok(existingRow, "test requires an existing backend_store row");
const futureRows = [
  ...rows,
  {
    ...existingRow,
    id: "U-99-01",
    manifest_phase: "phase99",
  },
];
assert.deepEqual(
  scheduledBackendTargets(futureRows),
  originalBackendTargets,
  "a future phase row under an existing execution dependency must not change service-backed schedule targets",
);

const tempDir = mkdtempSync(path.join(os.tmpdir(), "cartulary-topology-test-"));
copyFileSync(
  path.join(root, "tools/service_backed_make_target_duration_baselines.json"),
  path.join(tempDir, "service_backed_make_target_duration_baselines.json"),
);
const topologyFixture = () => JSON.parse(readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"));
const writeTopologyFixture = (name, value) => {
  const fixturePath = path.join(tempDir, name);
  writeFileSync(fixturePath, `${JSON.stringify(value, null, 2)}\n`);
  return fixturePath;
};

const invalidTopology = topologyFixture();
invalidTopology.execution_dependencies.push({ ...invalidTopology.execution_dependencies[0] });
assert.throws(
  () => loadExecutionTopology({ manifestPath: writeTopologyFixture("invalid-topology.json", invalidTopology) }),
  /duplicate execution dependency/,
  "topology validation must reject duplicate execution dependency IDs",
);

const legacyFlatTopology = topologyFixture();
legacyFlatTopology.check_schedules = renderedCheckSchedule.schedules;
assert.throws(
  () => loadExecutionTopology({ manifestPath: writeTopologyFixture("legacy-flat-topology.json", legacyFlatTopology) }),
  /flat check_schedules\[\] work_units are no longer supported/,
  "topology validation must reject the obsolete flat check schedule source shape",
);

const unknownProfileTopology = topologyFixture();
unknownProfileTopology.task_surface.targets.find((target) => target.name === "generate-drift").check_schedule.profile =
  "missing_profile";
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("unknown-check-profile-topology.json", unknownProfileTopology),
    }),
  /generate-drift\.check_schedule\.profile references unknown profile missing_profile/,
  "topology validation must reject unknown check schedule profiles",
);

const duplicateOrderTopology = topologyFixture();
duplicateOrderTopology.task_surface.targets.find((target) => target.name === "phase-schedule-drift").check_schedule.order =
  0;
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("duplicate-check-order-topology.json", duplicateOrderTopology),
    }),
  /duplicate priority order drift_validation:0/,
  "topology validation must reject duplicate check schedule priority orders within a band",
);

const mismatchedSummaryTargetTopology = topologyFixture();
mismatchedSummaryTargetTopology.task_surface.targets.find(
  (target) => target.name === "frontend-unit",
).check_schedule.produces_summary_targets = ["frontend-typecheck"];
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("mismatched-summary-target-topology.json", mismatchedSummaryTargetTopology),
    }),
  /frontend-unit\.check_schedule\.produces_summary_targets must include owning target frontend-unit/,
  "topology validation must reject check work units that omit their own summary target",
);

const missingServiceStackTopology = topologyFixture();
delete missingServiceStackTopology.check_schedules.defaults.resource_profiles.post_build_service_stack.resource_claims
  .service_stack;
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("missing-service-stack-topology.json", missingServiceStackTopology),
    }),
  /migration-drift\.check_schedule target declares service_requirements and must claim service_stack/,
  "topology validation must require service_stack for service-backed check schedule profiles",
);

const schedulerOwnedEnvTopology = topologyFixture();
schedulerOwnedEnvTopology.task_surface.targets.find((target) => target.name === "lint-shell").check_schedule.env = {
  CARTULARY_TEST_TARGET: "lint-shell",
};
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("scheduler-owned-env-topology.json", schedulerOwnedEnvTopology),
    }),
  /lint-shell\.check_schedule\.env\.CARTULARY_TEST_TARGET is scheduler-owned and cannot be overridden/,
  "topology validation must reject scheduler-owned check work-unit env",
);

const futureCheckTargetTopology = topologyFixture();
futureCheckTargetTopology.task_surface.targets.push({
  name: "future-phase-check-leaf",
  classification: "check_internal",
  included_in: ["check"],
  check_schedule: {
    schedules: ["check"],
    profile: "after_setup_cpu",
    priority_band: "phase_validation",
    order: 700,
  },
});
const futureCheckSchedule = renderCheckScheduleManifest(
  loadExecutionTopology({
    manifestPath: writeTopologyFixture("future-check-target-topology.json", futureCheckTargetTopology),
  }),
).schedules.find((schedule) => schedule.target === "check");
assert.ok(
  futureCheckSchedule.work_units.some((unit) => unit.target === "future-phase-check-leaf"),
  "new check-scheduled targets must be included through metadata without adding flat work units",
);
EOF
