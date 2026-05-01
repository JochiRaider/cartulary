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
assert.equal(summary.schema_id, "cartulary.execution_topology.v1");
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
const invalidTopologyPath = path.join(tempDir, "invalid-topology.json");
copyFileSync(
  path.join(root, "tools/service_backed_make_target_duration_baselines.json"),
  path.join(tempDir, "service_backed_make_target_duration_baselines.json"),
);
const invalidTopology = JSON.parse(readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"));
invalidTopology.execution_dependencies.push({ ...invalidTopology.execution_dependencies[0] });
writeFileSync(invalidTopologyPath, `${JSON.stringify(invalidTopology, null, 2)}\n`);
assert.throws(
  () => loadExecutionTopology({ manifestPath: invalidTopologyPath }),
  /duplicate execution dependency/,
  "topology validation must reject duplicate execution dependency IDs",
);
EOF
