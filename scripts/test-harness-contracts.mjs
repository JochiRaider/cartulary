#!/usr/bin/env node
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

import {
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
} from "./lib/execution-topology.mjs";
import {
  collectTaskSurfaceManifestErrors,
  renderTaskSurfaceMake,
} from "./lib/task-surface.mjs";
import { renderServiceBackedScheduleManifest } from "./render-service-backed-schedule-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(repoRoot, relativePath), "utf8"));
}

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
    taskSurfaceMake: renderTaskSurfaceMake(taskSurface),
  };
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
});

test("check schedule includes cheap harness contracts outside fast smoke", () => {
  const { checkSchedule } = renderedArtifacts();
  const check = checkSchedule.schedules.find(
    (schedule) => schedule.target === "check",
  );
  assert.ok(check, "rendered check schedule must include check");
  const harnessContracts = check.work_units.find(
    (unit) => unit.target === "harness-contract-tests",
  );
  assert.ok(harnessContracts, "check schedule must include harness-contract-tests");
  assert.equal(harnessContracts.priority, 12980);
  assert.deepEqual(harnessContracts.needs, ["toolchain-drift"]);
});
