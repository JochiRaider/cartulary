#!/usr/bin/env node
import { spawn } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages as loadBrowserBatchStagesFromManifest } from "./lib/browser-batch-manifest.mjs";
import { collectGoShardsForTarget } from "./lib/go-shard-plan.mjs";
import { formatResourceMap } from "./lib/scheduler-reporting.mjs";
import {
  normalizeResourceClaims as normalizeSchedulerResourceClaims,
  normalizeResourceLimits as normalizeSchedulerResourceLimits,
} from "./lib/scheduler-resources.mjs";
import {
  isDryRunFromMakeFlags,
  makeChildEnv,
  replayLog,
  runLifecycle,
  runNormalizedSchedule,
  sanitizeLogName,
  writeSchedulerDryRun,
} from "./lib/scheduler-runner.mjs";
import { createRunnerContext } from "./lib/runner-context.mjs";
import { findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const defaultBrowserBatchManifestPath = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
const supportedSchemaID = "cartulary.service_backed_schedule.v8";
const schedulerEventSchemaID = "cartulary.service_backed_scheduler_event.v4";
const schedulerSummarySchemaID = "cartulary.service_backed_scheduler_summary.v4";
const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const postgresResetResource = "postgres_reset";
const browserStackResource = "browser_stack";
const goCPULimitEnv = "CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT";
const goIOLimitEnv = "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT";
const browserStackLimitEnv = "CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const validSourceTypes = new Set(["go_shards", "make_target"]);
const validSourceClasses = new Set(["backend", "browser"]);

function usage() {
  process.stderr.write(
    "usage: run-service-backed-schedule.mjs --target <target> [--manifest <path>] [--defer-summary]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
    deferSummary: false,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--defer-summary") {
      options.deferSummary = true;
      continue;
    }
    if (arg === "--jobs") {
      throw new Error("--jobs is obsolete for v7 service-backed schedules; use resource_limits");
    }
    usage();
  }
  if (!options.target || !options.manifest) {
    usage();
  }
  return options;
}

async function loadManifest(file) {
  const manifestPath = path.isAbsolute(file) ? file : path.join(repoRoot, file);
  const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  if (manifest.schema_id !== supportedSchemaID) {
    throw new Error(`${manifestPath} must declare schema_id ${supportedSchemaID}`);
  }
  if (!Array.isArray(manifest.schedules)) {
    throw new Error(`${manifestPath} must declare schedules[]`);
  }
  return { manifest, manifestPath };
}

async function loadBrowserBatchStages() {
  return loadBrowserBatchStagesFromManifest(defaultBrowserBatchManifestPath);
}

function normalizeResourceLimits(value, label) {
  return normalizeSchedulerResourceLimits(value, label, {
    scheduler: "service_backed",
    allowAuto: true,
  });
}

function normalizeResourceClaims(value, label, resourceLimits) {
  return normalizeSchedulerResourceClaims(value, label, resourceLimits, {
    scheduler: "service_backed",
    allowBounded: false,
  });
}

function normalizeNeeds(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} needs must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} needs entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

function validateBackendTarget(scheduleTarget, target, label) {
  const descriptor = findTargetDescriptor(target);
  if (!descriptor) {
    throw new Error(`${label} backend target ${target} is not in target-plan`);
  }
  if (!descriptor.serviceBacked) {
    throw new Error(`${label} backend target ${target} is not service-backed`);
  }
  if (scheduleTarget === "check-service-backed" && descriptor.checkServiceBackedSafe !== true) {
    throw new Error(`${label} backend target ${target} is not check-service-backed safe`);
  }
}

function validateBrowserTarget(source, target, label, browserStages) {
  if (source.type !== "make_target") {
    throw new Error(`${label} browser target ${target} must use type make_target`);
  }
  if (typeof source.browser_stage !== "string" || source.browser_stage.trim() === "") {
    throw new Error(`${label} browser target ${target} must declare browser_stage`);
  }
  const browserStage = source.browser_stage.trim();
  const stage = browserStages.get(browserStage);
  if (!stage) {
    throw new Error(`${label} browser target ${target} declares unknown browser_stage ${browserStage}`);
  }
  if (stage.target !== target) {
    throw new Error(
      `${label} browser target ${target} must match browser_stage ${browserStage} aggregate target ${stage.target}`,
    );
  }
}

function validateSource(scheduleTarget, source, index, resourceLimits, browserStages) {
  const label = `${scheduleTarget} work_unit_sources ${index + 1}`;
  if (!source || typeof source !== "object" || Array.isArray(source)) {
    throw new Error(`${label} must be an object`);
  }
  if (!validSourceTypes.has(source.type)) {
    throw new Error(`${label} must declare type go_shards or make_target`);
  }
  if (typeof source.target !== "string" || source.target.trim() === "") {
    throw new Error(`${label} must declare target`);
  }
  if (!validSourceClasses.has(source.class)) {
    throw new Error(`${label} must declare class backend or browser`);
  }
  if (source.resource_tags !== undefined || source.exclusive_tags !== undefined) {
    throw new Error(`${label} must not declare legacy resource_tags or exclusive_tags`);
  }
  const target = source.target.trim();
  const resourceClaims = normalizeResourceClaims(
    source.resource_claims,
    `${label} ${target}`,
    resourceLimits,
  );
  const needs = normalizeNeeds(source.needs, `${label} ${target}`);
  if (source.class === "backend") {
    validateBackendTarget(scheduleTarget, target, label);
  } else {
    validateBrowserTarget(source, target, label, browserStages);
  }

  if (source.type === "go_shards") {
    if (source.class !== "backend") {
      throw new Error(`${label} ${target} go_shards sources must declare class backend`);
    }
    for (const resource of [goCPUResource, goIOResource]) {
      if (!resourceLimits.has(resource)) {
        throw new Error(`${label} ${target} go_shards sources require resource_limits.${resource}`);
      }
    }
    return {
      type: source.type,
      class: source.class,
      target,
      needs,
      resourceClaims,
      order: index,
    };
  }

  if (!Number.isFinite(source.weight) || source.weight < 0) {
    throw new Error(`${label} ${target} must declare non-negative weight`);
  }
  return {
    type: source.type,
    class: source.class,
    target,
    needs,
    weight: source.weight,
    resourceClaims,
    order: index,
  };
}

function assertAcyclic(target, sources) {
  const byTarget = new Map(sources.map((source) => [source.target, source]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (source) => {
    if (visited.has(source.target)) {
      return;
    }
    if (visiting.has(source.target)) {
      throw new Error(`schedule ${target} has a dependency cycle at ${source.target}`);
    }
    visiting.add(source.target);
    for (const need of source.needs) {
      visit(byTarget.get(need));
    }
    visiting.delete(source.target);
    visited.add(source.target);
  };
  for (const source of sources) {
    visit(source);
  }
}

function findSchedule(manifest, target, browserStages) {
  const matches = manifest.schedules.filter((schedule) => schedule?.target === target);
  if (matches.length !== 1) {
    throw new Error(`expected exactly one schedule for ${target}, found ${matches.length}`);
  }
  const [schedule] = matches;
  if (!Array.isArray(schedule.work_unit_sources) || schedule.work_unit_sources.length === 0) {
    throw new Error(`schedule ${target} must declare at least one work_unit_sources entry`);
  }
  if (schedule.children !== undefined) {
    throw new Error(`schedule ${target} must use work_unit_sources, not legacy children`);
  }
  const normalizedLimits = normalizeResourceLimits(schedule.resource_limits, `schedule ${target}`);
  const resourceLimits = normalizedLimits.limits;
  const sources = schedule.work_unit_sources.map((source, index) =>
    validateSource(target, source, index, resourceLimits, browserStages),
  );
  const duplicateTargets = sources
    .map((source) => source.target)
    .filter((targetName, index, targets) => targets.indexOf(targetName) !== index);
  if (duplicateTargets.length > 0) {
    throw new Error(
      `schedule ${target} contains duplicate work-unit source targets: ${duplicateTargets.join(", ")}`,
    );
  }
  const sourceTargets = new Set(sources.map((source) => source.target));
  for (const source of sources) {
    for (const need of source.needs) {
      if (!sourceTargets.has(need)) {
        throw new Error(`schedule ${target} source ${source.target} depends on unknown target ${need}`);
      }
      if (need === source.target) {
        throw new Error(`schedule ${target} source ${source.target} cannot depend on itself`);
      }
    }
  }
  assertAcyclic(target, sources);
  return {
    target,
    resourceLimits,
    resourceLimitSources: normalizedLimits.sources,
    sources,
    children: sources.map((source) => source.target),
    backendChildren: sources
      .filter((source) => source.class === "backend")
      .map((source) => source.target),
    dependencyCount: sources.reduce((sum, source) => sum + source.needs.length, 0),
  };
}

function cloneResourceClaims(resourceClaims) {
  return new Map(resourceClaims.entries());
}

function schedulerClaimsForShard(shard, resourceLimits) {
  switch (shard.scheduler_profile) {
    case "cpu_heavy":
      return new Map([
        [goCPUResource, 2],
        [goIOResource, 1],
      ]);
    case "io_heavy":
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 2],
      ]);
    case "reset_heavy":
      if (!resourceLimits.has(postgresResetResource)) {
        throw new Error(
          `go shard ${shard.name} has reset_heavy profile but schedule is missing resource_limits.${postgresResetResource}`,
        );
      }
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 3],
        [postgresResetResource, 1],
      ]);
    default:
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 1],
      ]);
  }
}

function mergeResourceClaims(baseClaims, extraClaims) {
  const merged = cloneResourceClaims(baseClaims);
  for (const [resource, amount] of extraClaims.entries()) {
    merged.set(resource, (merged.get(resource) ?? 0) + amount);
  }
  return merged;
}

function shardCompletionKey(shardName) {
  return `go_shard:${shardName}`;
}

function expandSchedule(schedule) {
  const countedWorkUnits = [];
  const finalizerUnits = [];
  const shardWorkByName = new Map();

  for (const source of schedule.sources) {
    if (source.type === "make_target") {
      countedWorkUnits.push({
        id: source.target,
        label: source.target,
        kind: "make_target",
        type: "make_target",
        class: source.class,
        target: source.target,
        aggregateTarget: source.target,
        group: source.target,
        needs: [...source.needs],
        completionKeys: [source.target],
        failureKeys: [source.target],
        weight: source.weight,
        resourceClaims: cloneResourceClaims(source.resourceClaims),
        order: source.order,
      });
      continue;
    }

    const shards = collectGoShardsForTarget(repoRoot, source.target);
    if (shards.length === 0) {
      throw new Error(`go_shards source ${source.target} selected no shards`);
    }
    finalizerUnits.push({
      id: `finalize:${source.target}`,
      label: `finalize/${source.target}`,
      kind: "finalizer",
      type: "finalizer",
      class: source.class,
      target: source.target,
      aggregateTarget: source.target,
      group: source.target,
      needs: shards.map((shard) => shardCompletionKey(shard.name)),
      completionKeys: [source.target],
      failureKeys: [source.target],
      countInTotal: false,
      countsStarted: false,
      resourceClaims: new Map(),
      shardNames: shards.map((shard) => shard.name),
      unblockLabel: source.target,
      weight: 0,
      order: source.order,
    });
    for (const shard of shards) {
      if (shardWorkByName.has(shard.name)) {
        continue;
      }
      const unit = {
        id: `${source.target}:${shard.name}`,
        label: `${source.target}/${shard.name}`,
        kind: "go_shard",
        type: "go_shard",
        class: source.class,
        target: source.target,
        aggregateTarget: source.target,
        group: source.target,
        needs: [...source.needs],
        completionKeys: [shardCompletionKey(shard.name)],
        failureKeys: [shardCompletionKey(shard.name)],
        runningDependencyKeys: [source.target],
        completeOnFailure: true,
        shard: shard.name,
        schedulerProfile: shard.scheduler_profile,
        weight: shard.weight_ms,
        resourceClaims: mergeResourceClaims(
          source.resourceClaims,
          schedulerClaimsForShard(shard, schedule.resourceLimits),
        ),
        order: source.order,
      };
      shardWorkByName.set(shard.name, unit);
      countedWorkUnits.push(unit);
    }
  }

  countedWorkUnits.sort(
    (left, right) =>
      right.weight - left.weight ||
      left.order - right.order ||
      left.label.localeCompare(right.label),
  );
  const workUnits = [...countedWorkUnits, ...finalizerUnits];
  return {
    ...schedule,
    ...resolveResourceLimits(schedule.resourceLimits, schedule.resourceLimitSources, countedWorkUnits),
    workUnits,
    totalWorkUnits: countedWorkUnits.length,
    finalizerCount: finalizerUnits.length,
  };
}

function clampInteger(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function parsePositiveIntegerEnv(name) {
  const raw = process.env[name];
  if (raw === undefined || raw === "") {
    return null;
  }
  const value = Number(raw);
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${name} must be a positive integer`);
  }
  return value;
}

function availableCPUCount() {
  if (typeof os.availableParallelism === "function") {
    return Math.max(1, os.availableParallelism());
  }
  return Math.max(1, os.cpus().length);
}

function estimateGoCPULimit(goShardUnits) {
  if (goShardUnits.length === 0) {
    return 1;
  }
  const totalWeight = goShardUnits.reduce((sum, unit) => sum + Math.max(1, unit.weight), 0);
  const maxWeight = Math.max(...goShardUnits.map((unit) => Math.max(1, unit.weight)));
  const weightedConcurrency = Math.ceil(totalWeight / Math.max(30_000, maxWeight));
  const cpuCount = availableCPUCount();
  const hostConcurrency = cpuCount <= 4 ? Math.max(2, cpuCount - 1) : Math.floor(cpuCount * 0.75);
  return clampInteger(Math.max(4, Math.min(hostConcurrency, weightedConcurrency)), 4, 12);
}

function estimateGoIOLimit(goShardUnits, goCPULimit) {
  if (goShardUnits.length === 0) {
    return 1;
  }
  const balanced = goShardUnits.filter((unit) => unit.schedulerProfile === "balanced").length;
  const ioHeavy = goShardUnits.filter((unit) => unit.schedulerProfile === "io_heavy").length;
  const resetHeavy = goShardUnits.filter((unit) => unit.schedulerProfile === "reset_heavy").length;
  const cpuHeavy = goShardUnits.filter((unit) => unit.schedulerProfile === "cpu_heavy").length;
  const profileConcurrency = balanced + ioHeavy * 2 + resetHeavy * 3 + Math.ceil(cpuHeavy / 2);
  return clampInteger(Math.max(6, goCPULimit + 2, profileConcurrency), 6, 16);
}

function browserStageLaneCount(workUnits) {
  const lanes = new Set();
  for (const unit of workUnits) {
    for (const resource of unit.resourceClaims.keys()) {
      if (resource.startsWith("browser_stage_")) {
        lanes.add(resource);
      }
    }
  }
  return lanes.size;
}

function estimateBrowserStackLimit(workUnits) {
  const laneCount = browserStageLaneCount(workUnits);
  if (laneCount === 0) {
    return 1;
  }
  const hostLimit = availableCPUCount() >= 8 ? 2 : 1;
  return clampInteger(hostLimit, 1, laneCount);
}

function resolveResourceLimits(resourceLimits, resourceLimitSources, workUnits) {
  const goShardUnits = workUnits.filter((unit) => unit.kind === "go_shard");
  const computedGoCPU = estimateGoCPULimit(goShardUnits);
  const goCPUOverride = parsePositiveIntegerEnv(goCPULimitEnv);
  const effectiveGoCPU = goCPUOverride ?? computedGoCPU;
  const computedGoIO = estimateGoIOLimit(goShardUnits, effectiveGoCPU);
  const goIOOverride = parsePositiveIntegerEnv(goIOLimitEnv);
  const effectiveGoIO = goIOOverride ?? computedGoIO;
  const computedBrowserStack = estimateBrowserStackLimit(workUnits);
  const browserStackOverride = parsePositiveIntegerEnv(browserStackLimitEnv);
  const effectiveBrowserStack = browserStackOverride ?? computedBrowserStack;
  const resolved = new Map();
  const sources = new Map(resourceLimitSources.entries());
  for (const [resource, limit] of resourceLimits.entries()) {
    if (resource === goCPUResource && (limit === "auto" || goCPUOverride !== null)) {
      resolved.set(resource, effectiveGoCPU);
      sources.set(resource, goCPUOverride === null ? "auto" : `env:${goCPULimitEnv}`);
      continue;
    }
    if (resource === goIOResource && (limit === "auto" || goIOOverride !== null)) {
      resolved.set(resource, effectiveGoIO);
      sources.set(resource, goIOOverride === null ? "auto" : `env:${goIOLimitEnv}`);
      continue;
    }
    if (resource === browserStackResource && (limit === "auto" || browserStackOverride !== null)) {
      resolved.set(resource, effectiveBrowserStack);
      sources.set(resource, browserStackOverride === null ? "auto" : `env:${browserStackLimitEnv}`);
      continue;
    }
    resolved.set(resource, limit);
  }
  return { resourceLimits: resolved, resourceLimitSources: sources };
}

function runPostgresFixtureBudgetCheck(targets) {
  return new Promise((resolve, reject) => {
    const child = spawn(
      process.execPath,
      [
        path.join(repoRoot, "scripts", "check-postgres-fixture-budget.mjs"),
        "--targets",
        targets.join(","),
      ],
      {
        cwd: repoRoot,
        env: process.env,
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    child.stdout.pipe(process.stdout, { end: false });
    child.stderr.pipe(process.stderr, { end: false });
    child.on("error", reject);
    child.on("close", (status) => {
      if (status === 0) {
        resolve(0);
        return;
      }
      resolve(status ?? 1);
    });
  });
}

function displayCapacity(schedule) {
  return schedule.resourceLimits.get(goCPUResource) ?? Math.max(...schedule.resourceLimits.values());
}

function workUnitLogFile(logDir, unit, started) {
  return path.join(logDir, `${String(started).padStart(2, "0")}-${sanitizeLogName(unit.id)}.log`);
}

function finalizerLogFile(logDir, unit) {
  return path.join(logDir, `finalize-${sanitizeLogName(unit.aggregateTarget)}.log`);
}

function attachRuntime(schedule, { makeBin, testOutputScript, deferSummary, goTargetRunner, metadataDir }) {
  const capacityDisplay = displayCapacity(schedule);
  for (const unit of schedule.workUnits) {
    if (unit.kind === "make_target") {
      unit.logFile = ({ logDir, started }) => workUnitLogFile(logDir, unit, started);
      unit.command = () => ({
        command: makeBin,
        args: ["--no-print-directory", "--output-sync=target", "-j1", unit.target],
        env: makeChildEnv(process.env),
      });
      continue;
    }
    if (unit.kind === "go_shard") {
      unit.logFile = ({ logDir, started }) => workUnitLogFile(logDir, unit, started);
      unit.command = () => ({
        command: goTargetRunner,
        args: ["capture-shard", unit.target, unit.shard, metadataDir],
        env: {
          ...process.env,
          CARTULARY_TEST_TARGET: unit.target,
        },
      });
      continue;
    }
    unit.logFile = ({ logDir }) => finalizerLogFile(logDir, unit);
    unit.command = () => ({
      command: goTargetRunner,
      args: ["finalize-shards", unit.aggregateTarget, metadataDir],
      env: {
        ...process.env,
        CARTULARY_TEST_TARGET: unit.aggregateTarget,
        TEST_OUTPUT_SCRIPT: testOutputScript,
      },
    });
  }

  return {
    ...schedule,
    kind: "service-backed",
    prefix: "SCHEDULER",
    eventSchemaID: schedulerEventSchemaID,
    summarySchemaID: schedulerSummarySchemaID,
    resourceScheduler: "service_backed",
    quietStart: true,
    summaryOnPass: false,
    showFinalizing: true,
    initialProgressAt: Date.now(),
    stopOnFirstFailure: false,
    runningDisplayUnits(state) {
      return Array.from(state.running.values()).map((unit) =>
        unit.kind === "finalizer"
          ? { id: unit.id, label: `finalize:${unit.aggregateTarget}`, group: unit.aggregateTarget }
          : unit,
      );
    },
    countCompletedUnit: (unit) => unit.countInTotal !== false,
    beforeRun: async ({ reporter }) => {
      if (!reporter.verbose) {
        return;
      }
      await runLifecycle(repoRoot, testOutputScript, [
        "target-start",
        schedule.target,
        "--children",
        schedule.children.join(","),
        "--service-backed",
        "1",
      ]);
    },
    beforeUnitStart: async ({ unit, started, total, reporter }) => {
      if (!reporter.verbose || unit.countInTotal === false) {
        return;
      }
      await runLifecycle(repoRoot, testOutputScript, [
        "step-start",
        schedule.target,
        String(started),
        String(total),
        unit.label,
        "--mode",
        "scheduler",
        "--jobs",
        String(capacityDisplay),
      ]);
    },
    beforeReplayLog: async ({ unit, result, reporter }) => {
      if (unit.kind !== "finalizer" || result.status === 0 || reporter.verbose) {
        return;
      }
      for (const logFile of reporter.completedLogFilesForTarget(unit.aggregateTarget)) {
        await replayLog(logFile, process.stderr);
      }
    },
    shouldReplayLog: ({ result, reporter }) => result.status !== 0 || reporter.verbose,
    afterWorkComplete: async ({ firstFailure }) => {
      if (firstFailure !== 0 || schedule.backendChildren.length === 0) {
        return null;
      }
      const status = await runPostgresFixtureBudgetCheck(schedule.backendChildren);
      return status === 0 ? null : { status, label: "postgres-fixture-budget" };
    },
    summaryExtra: ({ reporter, started }) => ({
      started_count: started,
      finalizer_count: schedule.finalizerCount,
      finalizer_failures: reporter.finalizerFailures,
      max_running_groups: reporter.maxRunningGroups,
    }),
    afterSummary: async ({ requestedStatus }) => {
      if (deferSummary) {
        return;
      }
      await runLifecycle(
        repoRoot,
        testOutputScript,
        ["target-summary", schedule.target, requestedStatus, "--children", schedule.children.join(",")],
        requestedStatus === "pass" ? process.stdout : process.stderr,
      );
    },
  };
}

async function runSchedule({ schedule, makeBin, testOutputScript, deferSummary }) {
  const context = createRunnerContext({ repoRoot });
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-service-backed-schedule-"));
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  const goTargetRunner =
    process.env[goTargetRunnerEnv] || context.runnerScript;
  try {
    const runtimeSchedule = attachRuntime(schedule, {
      makeBin,
      testOutputScript,
      deferSummary,
      goTargetRunner,
      metadataDir,
    });
    const result = await runNormalizedSchedule({
      repoRoot,
      schedule: runtimeSchedule,
      testOutputScript,
    });
    return result.status;
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function main() {
  const context = createRunnerContext({ repoRoot });
  const options = parseArgs(process.argv.slice(2));
  const { manifest, manifestPath } = await loadManifest(options.manifest);
  const browserStages = await loadBrowserBatchStages();
  const schedule = expandSchedule(findSchedule(manifest, options.target, browserStages));
  const makeBin = process.env.MAKE || context.makeBin;
  const testOutputScript = process.env.TEST_OUTPUT_SCRIPT || context.testOutputScript;

  if (isDryRunFromMakeFlags()) {
    writeSchedulerDryRun({
      repoRoot,
      schedule: {
        ...schedule,
        kind: "service-backed",
        resourceScheduler: "service_backed",
      },
      manifestPath,
      verboseUnitLine(unit) {
        if (unit.countInTotal === false) {
          return "";
        }
        const profile = unit.schedulerProfile ? ` profile=${unit.schedulerProfile}` : "";
        const needs = unit.needs.length > 0 ? ` needs=${unit.needs.join(",")}` : "";
        return `[DRY-RUN] ${schedule.target} unit ${unit.label} type=${unit.type} class=${unit.class}${profile}${needs} claims=${formatResourceMap(unit.resourceClaims)}\n`;
      },
    });
    return;
  }

  const status = await runSchedule({
    schedule,
    makeBin,
    testOutputScript,
    deferSummary: options.deferSummary,
  });
  process.exitCode = status;
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
});
