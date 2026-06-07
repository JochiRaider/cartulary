#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
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
  generateTestRouteToken,
  preflightPublicTarget,
  redactString,
  redactValue,
  runCleanup,
  testRouteTokenValid,
} from "./lib/harness-contract.mjs";
import { primaryPublicFailure } from "./lib/failure-taxonomy.mjs";
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

function markdownCodeTokens(cell) {
  return [...String(cell ?? "").matchAll(/`([^`]+)`/gu)].map((match) => match[1]);
}

function normalizeSpecAllowedSources(cell) {
  const text = String(cell ?? "").toLowerCase();
  const sources = [];
  if (text.includes("make command line")) {
    sources.push("make_command_line");
  }
  if (text.includes("environment")) {
    sources.push("environment");
  }
  if (text.includes("makefile default")) {
    sources.push("makefile_default");
  }
  if (text.includes("internal default")) {
    sources.push("internal_default");
  }
  if (text.includes("manifest")) {
    sources.push("manifest");
  }
  return sources;
}

function normalizeSpecDefault(cell) {
  const text = String(cell ?? "").trim();
  if (text === "none") {
    return null;
  }
  const token = markdownCodeTokens(text)[0];
  if (token === undefined) {
    return null;
  }
  if (token === "false") {
    return false;
  }
  if (/^(0|[1-9][0-9]*)$/u.test(token)) {
    return Number.parseInt(token, 10);
  }
  if (/^(?:0|[1-9][0-9]*)\.[0-9]+$/u.test(token)) {
    return Number(token);
  }
  return token;
}

function normalizeSpecEmptyString(cell) {
  const text = String(cell ?? "").toLowerCase();
  if (text.includes("false")) {
    return "false";
  }
  if (text.includes("invalid")) {
    return "invalid";
  }
  return "omitted";
}

function normalizeSpecInvalidReason(cell) {
  return markdownCodeTokens(cell)[0] ?? "";
}

function normalizeSpecChildForwarding(cell) {
  return String(cell ?? "").trim().toLowerCase().replaceAll(" ", "_");
}

function normalizeSpecValuesAndBounds(cell, type) {
  const text = String(cell ?? "");
  if (type === "enum") {
    return { values: markdownCodeTokens(text) };
  }
  const range = text.match(/`?([0-9]+)\.\.([0-9]+)`?/u);
  if (range) {
    return {
      min: Number.parseInt(range[1], 10),
      max: Number.parseInt(range[2], 10),
    };
  }
  const min = text.match(/`?>=\s*([0-9]+(?:\.[0-9]+)?)`?/u);
  if (min) {
    return { min: Number(min[1]) };
  }
  return {};
}

function parseHarnessInputMatrix() {
  const text = readFileSync(
    path.join(repoRoot, "docs/testing-harness-nlspec.md"),
    "utf8",
  );
  const lines = text.split("\n");
  const byTarget = new Map();
  let inMatrix = false;
  for (const line of lines) {
    if (line.startsWith("| Target(s) | Input | Type |")) {
      inMatrix = true;
      continue;
    }
    if (inMatrix && line.startsWith("`fixture-report` remains")) {
      break;
    }
    if (!inMatrix || !line.startsWith("| `")) {
      continue;
    }
    const cells = splitMarkdownRow(line);
    const targets = markdownCodeTokens(cells[0]);
    const name = markdownCodeTokens(cells[1])[0];
    const type = markdownCodeTokens(cells[2])[0];
    const entry = {
      name,
      binding: "make_variable",
      allowed_sources: normalizeSpecAllowedSources(cells[4]),
      required: cells[3] === "yes",
      type,
      default: normalizeSpecDefault(cells[5]),
      empty_string: normalizeSpecEmptyString(cells[7]),
      normalization: markdownCodeTokens(cells[8])[0] ?? "",
      invalid_reason: normalizeSpecInvalidReason(cells[10]),
      summary_emission: String(cells[11] ?? "").trim(),
      child_forwarding: normalizeSpecChildForwarding(cells[12]),
      ...normalizeSpecValuesAndBounds(cells[9], type),
    };
    for (const target of targets) {
      if (!byTarget.has(target)) {
        byTarget.set(target, []);
      }
      byTarget.get(target).push(entry);
    }
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

test("dev service lifecycle guards are mutation-safe", () => {
  const result = spawnSync("bash", ["./scripts/test-dev-services-lifecycle.sh"], {
    cwd: repoRoot,
    encoding: "utf8",
    env: { ...process.env },
    maxBuffer: 16 * 1024 * 1024,
  });
  assert.equal(
    result.status,
    0,
    [
      "scripts/test-dev-services-lifecycle.sh failed",
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

test("check scheduler restores node packages before run-phase validation", () => {
  const { checkSchedule, taskSurface } = renderedArtifacts();
  const schedule = checkSchedule.schedules.find(
    (entry) => entry.target === "check",
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
  assert.ok(
    checkFrontendInstall,
    "check schedule must include check-frontend-install",
  );
  assert.deepEqual(
    checkFrontendInstall.needs ?? [],
    [],
    "check-frontend-install must be able to run before run-phase children",
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
  for (const target of ["toolchain-drift", "json-shape-check"]) {
    assert.ok(
      taskSurface.make_recipes[target].prerequisites.includes(
        "$(FRONTEND_INSTALL_STAMP)",
      ),
      `${target} direct wrapper must bootstrap installed node package dependencies`,
    );
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
  assert.deepEqual(taskSurface.make_recipes.check?.prerequisites, [
    "$(NODE_BIN)",
  ]);
  assert.deepEqual(
    taskSurface.make_recipes["check-harness-smoke"]?.prerequisites,
    ["$(NODE_BIN)"],
  );
  const checkBlock = targetRecipeBlock(taskSurfaceMake, "check").join("\n");
  const preflightIndex = checkBlock.indexOf(
    "$(HARNESS_CONTRACT_SCRIPT) preflight check",
  );
  const prerequisiteIndex = checkBlock.indexOf(
    "$(MAKE) --silent --no-print-directory $(NODE_BIN); fi",
  );
  const schedulerIndex = checkBlock.indexOf(
    "$(NODE_BIN) $(RUN_CHECK_SCHEDULE_SCRIPT)",
  );
  assert.ok(preflightIndex >= 0, "check must render public preflight");
  assert.ok(
    prerequisiteIndex > preflightIndex,
    "check must bootstrap Node after preflight",
  );
  assert.ok(
    schedulerIndex > prerequisiteIndex,
    "check must launch the scheduler after Node bootstrap",
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
  const producedSummaryTargets = new Set(
    Object.values(taskSurface.sequences ?? {}).flatMap((sequence) =>
      (sequence.steps ?? []).flatMap(
        (step) => step.produces_summary_targets ?? [],
      ),
    ),
  );
  for (const target of producedSummaryTargets) {
    const recipe = taskSurface.make_recipes[target];
    const entry = targetEntries.get(target);
    if (
      recipe?.mode !== "run_phase" ||
      entry?.output_policy?.summary_schema !== "cartulary.tool_run_summary.v3"
    ) {
      continue;
    }
    assert.match(
      taskSurfaceMake,
      new RegExp(`RUN_RETAINED_TARGET_SUMMARY,${target},pass`),
      `${target} run_phase recipe must retain a passing target summary`,
    );
    assert.match(
      taskSurfaceMake,
      new RegExp(`RUN_RETAINED_TARGET_SUMMARY,${target},fail`),
      `${target} run_phase recipe must retain a failing target summary`,
    );
  }
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

test("harness NLSpec input matrix mirrors public target input contracts", () => {
  const { taskSurface } = renderedArtifacts();
  const specInputs = parseHarnessInputMatrix();
  const publicTargets = taskSurface.targets.filter(
    (entry) => entry.target_class === "public",
  );
  for (const target of publicTargets) {
    const expected = normalizeInputList(specInputs.get(target.name) ?? []);
    const actual = normalizeInputList(
      (target.input_contract?.inputs ?? []).map(normalizeManifestInput),
    );
    assert.deepEqual(
      actual,
      expected,
      `${target.name} input_contract must match the NLSpec input matrix`,
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
    specInputs.get("scheduler-summary-timing-drift") ?? [],
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
  assert.throws(
    () =>
      preflightPublicTarget("db-reset", {
        CARTULARY_DESTRUCTIVE_CONFIRM: "object-store-reset",
        CARTULARY_MAKE_ORIGIN_CARTULARY_DESTRUCTIVE_CONFIRM: "command line",
      }),
    (error) =>
      error instanceof HarnessConfigError &&
      error.failure_reason === "usage_error" &&
      /CARTULARY_DESTRUCTIVE_CONFIRM must be one of db-reset/.test(error.message),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("db-reset", {
      CARTULARY_DESTRUCTIVE_CONFIRM: "db-reset",
      CARTULARY_MAKE_ORIGIN_CARTULARY_DESTRUCTIVE_CONFIRM: "command line",
    }),
  );
  assert.doesNotThrow(() =>
    preflightPublicTarget("db-reset", {
      CARTULARY_DESTRUCTIVE_CONFIRM: "db-reset",
      CARTULARY_MAKE_ORIGIN_CARTULARY_DESTRUCTIVE_CONFIRM: "environment",
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
      phase: "",
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
  runCleanup({
    candidates: [owned, "tmp/missing-cleanup-owned-path"],
    includeTmp: false,
    dryRun: false,
    stdout,
  });
  assert.equal(existsSync(tempRoot), false, "cleanup-owned temp path must be removed");
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
        session_token: "nested-session-token",
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
  assert.equal(structured.service_sessions[0].session_token, "[REDACTED]");
  assert.equal(structured.X_Cartulary_Test_Route_Token, "[REDACTED]");
  assert.equal(structured.CARTULARY_S3TEST_SECRET_ACCESS_KEY, "[REDACTED]");
  assert.equal(structured.session_target, "not-redacted-token-substring");

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
