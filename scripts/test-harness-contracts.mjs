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
import { collectGoShardsForTarget } from "./lib/go-shard-plan.mjs";
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

test("service-backed Go shard units are executable by their declared targets", () => {
  const { serviceBacked } = renderedArtifacts();
  const shardNamesByTarget = new Map();
  function shardsForTarget(target) {
    if (!shardNamesByTarget.has(target)) {
      shardNamesByTarget.set(
        target,
        new Set(collectGoShardsForTarget(repoRoot, target).map((shard) => shard.name)),
      );
    }
    return shardNamesByTarget.get(target);
  }

  for (const schedule of serviceBacked.schedules ?? []) {
    for (const unit of schedule.work_units ?? []) {
      if (unit.kind !== "go_shard") {
        continue;
      }
      assert.ok(
        shardsForTarget(unit.target).has(unit.shard),
        `${schedule.target ?? schedule.name} schedules ${unit.shard} for ${unit.target}, but that target cannot execute the shard`,
      );
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
      /^cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v1$/,
      `${target.name} must declare stable command_id`,
    );
    assert.ok(!commandIDs.has(target.command_id), `${target.name} command_id must be unique`);
    commandIDs.add(target.command_id);
    assert.match(
      target.family_id,
      /^[a-z][a-z0-9_]*$/,
      `${target.name} must declare family_id`,
    );
    assert.ok(
      ["public_active", "public_deprecated"].includes(target.lifecycle_state),
      `${target.name} must declare a public lifecycle state`,
    );
    assert.ok(
      Array.isArray(target.semantic_behaviors) &&
        target.semantic_behaviors.length > 0,
      `${target.name} must declare semantic behaviors`,
    );
    for (const entry of target.semantic_behaviors) {
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
    /help\.command_id must match cartulary\.harness\.command\.<name>\.v1/,
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
    (entry) => entry.name === "frontend-install-ci",
  ).default_inclusion_sets = ["helper_only"];
  assert.match(
    collectTaskSurfaceManifestErrors(privateHelperOnly, {
      browserBatchManifest: browserBatch,
      serviceBackedScheduleManifest: serviceBacked,
    }).join("\n"),
    /frontend-install-ci\.default_inclusion_sets helper_only is only valid for public direct-invocation targets/,
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
    output_policy: structuredClone(
      shallowAlias.targets.find((entry) => entry.name === "help").output_policy,
    ),
  });
  shallowAlias.help_tiers[0].entries.push({
    target: "synthetic-shallow-wrapper",
    description: "synthetic shallow wrapper",
  });
  shallowAlias.make_recipes["synthetic-shallow-wrapper"] = {
    type: "alias",
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
    assert.equal(
      recipeLines[0],
      `\t$(call RUN_PUBLIC_PREFLIGHT,${target.name})`,
      `${target.name} must run public preflight first`,
    );
    if ((recipes[target.name]?.prerequisites ?? []).length > 0) {
      assert.match(
        recipeLines.slice(1).join("\n"),
        /CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES/,
        `${target.name} prerequisite work must follow preflight`,
      );
    }
  }
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
