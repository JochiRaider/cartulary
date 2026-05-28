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
  expandServiceBackedSchedule,
  expandServiceBackedScheduleForCheck,
} from "./lib/check-service-backed-expansion.mjs";
import {
  collectTaskSurfaceManifestErrors,
  renderTaskSurfaceMake,
} from "./lib/task-surface.mjs";
import {
  HarnessConfigError,
  preflightPublicTarget,
} from "./lib/harness-contract.mjs";
import { collectGoShardsForTarget } from "./lib/go-shard-plan.mjs";
import { renderServiceBackedScheduleManifest } from "./render-service-backed-schedule-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(repoRoot, relativePath), "utf8"));
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

function splitMarkdownRow(line) {
  return line
    .slice(1, -1)
    .split("|")
    .map((cell) => cell.trim());
}

function parseHarnessPublicRegistry() {
  const text = readFileSync(
    path.join(repoRoot, "docs/testing-harness-nlspec.md"),
    "utf8",
  );
  const lines = text.split("\n");
  const rows = new Map();
  let inTable = false;
  for (const line of lines) {
    if (line.startsWith("| Target | Command ID | Family ID |")) {
      inTable = true;
      continue;
    }
    if (inTable && line.startsWith("**TH-HARNESS-REQ-059**")) {
      break;
    }
    if (!inTable || !line.startsWith("| `")) {
      continue;
    }
    const cells = splitMarkdownRow(line);
    const target = cells[0].replaceAll("`", "");
    rows.set(target, {
      outputClass: cells[4].replaceAll("`", ""),
      sideEffects: cells[7]
        .split(",")
        .map((entry) => entry.trim().replaceAll("`", ""))
        .filter(Boolean)
        .sort((left, right) => left.localeCompare(right)),
    });
  }
  return rows;
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

test("harness NLSpec registry mirrors public target output classes and side effects", () => {
  const { taskSurface } = renderedArtifacts();
  const specRows = parseHarnessPublicRegistry();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  assert.equal(
    specRows.size,
    publicTargets.length,
    "NLSpec public target registry row count must match manifest public target count",
  );
  for (const target of publicTargets) {
    const spec = specRows.get(target.name);
    assert.ok(spec, `${target.name} must appear in the NLSpec public registry`);
    assert.equal(
      spec.outputClass,
      target.output_policy.output_class,
      `${target.name} output class must match NLSpec registry`,
    );
    assert.deepEqual(
      spec.sideEffects,
      target.side_effects
        .map((entry) => entry.class)
        .sort((left, right) => left.localeCompare(right)),
      `${target.name} side effects must match NLSpec registry`,
    );
  }
});

test("scheduler manifest exercises every closed command type", () => {
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
  assert.deepEqual([...expected].sort(), [], "every scheduler command type must have a live fixture");
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
    assert.match(
      recipeLines[0],
      new RegExp(
        `^\\t\\$\\(Q\\)env .* \\$\\(HARNESS_CONTRACT_SCRIPT\\) preflight ${target.name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}$`,
      ),
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

test("per-target input contract rejects misplaced Make variables and ignores ambient env", () => {
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        PHASE: "phase4",
        CARTULARY_MAKE_ORIGIN_PHASE: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /PHASE is not declared for target target-plan/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      PHASE: "phase4",
      CARTULARY_MAKE_ORIGIN_PHASE: "environment",
    }),
  );
  assert.throws(
    () =>
      preflightPublicTarget("target-plan", {
        TASK_SURFACE_MANIFEST: "/tmp/override.json",
        CARTULARY_MAKE_ORIGIN_TASK_SURFACE_MANIFEST: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "configuration_error" &&
      /TASK_SURFACE_MANIFEST is an internal harness input/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("target-plan", {
      TASK_SURFACE_MANIFEST: "/tmp/override.json",
      CARTULARY_MAKE_ORIGIN_TASK_SURFACE_MANIFEST: "environment",
    }),
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

test("default check service-backed browser work uses declared session groups", () => {
  const { serviceBacked, expandedCheckSchedule } = renderedArtifacts();
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
      ["visual-smoke", "default-check-browser-shared"],
      ["a11y", "default-check-browser-shared"],
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
  assert.equal(browserSessions.length, 2);
  assert.deepEqual(
    browserSessions.map((unit) => unit.browser_session_group).sort(),
    ["default-check-browser-shared", "default-check-stateful-isolated"],
  );
  assert.equal(
    check.work_units.filter(
      (unit) =>
        unit.kind === "browser_group" &&
        unit.aggregate_target === "browser-e2e-webserver-backed",
    ).length,
    7,
  );
  assert.equal(
    check.work_units.some((unit) => unit.aggregate_target === "browser-e2e-measurement"),
    false,
  );
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
});
