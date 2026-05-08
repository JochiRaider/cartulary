#!/usr/bin/env node
import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages as loadBrowserBatchStagesFromManifest } from "./lib/browser-batch-manifest.mjs";
import {
  browserGroupCompletionKey,
  browserGroupNeeds,
  browserGroupWorkerEnv,
  browserStageCompletionNeeds,
  browserStageSessionKey,
} from "./lib/browser-scheduler-dependencies.mjs";
import { collectGoShardsForTarget } from "./lib/go-shard-plan.mjs";
import { formatResourceMap } from "./lib/scheduler-reporting.mjs";
import {
  estimateBrowserStackAutoLimit,
  normalizeResourceClaims as normalizeSchedulerResourceClaims,
  normalizeResourceLimits as normalizeSchedulerResourceLimits,
  provisionalResourceLimitsForClaims,
  resolveAutoResourceLimits,
} from "./lib/scheduler-resources.mjs";
import {
  countVisibleCompletedUnit,
  finalizerRunningDisplayUnits,
  isDryRunFromMakeFlags,
  makeChildEnv,
  replayFailedAggregateLogsBeforeFinalizer,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "./lib/scheduler-runner.mjs";
import { createRunnerContext } from "./lib/runner-context.mjs";
import {
  loadScheduleManifest,
  normalizeNeeds,
  parseResourceLimitOverride,
  selectSingleSchedule,
  validateTargetDependencyGraph,
} from "./lib/scheduler-manifest.mjs";
import { findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const defaultBrowserBatchManifestPath = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
const supportedSchemaID = "cartulary.service_backed_schedule.v10";
const schedulerEventSchemaID = "cartulary.service_backed_scheduler_event.v5";
const schedulerSummarySchemaID = "cartulary.service_backed_scheduler_summary.v9";
const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const postgresResetResource = "postgres_reset";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const validSourceTypes = new Set(["go_shards", "make_target", "browser_stage"]);
const validSourceClasses = new Set(["backend", "browser"]);
const validBrowserGroupKinds = new Set(["functional_shard", "support", "stateful", "measurement", "visual"]);
const measurementIsolationStages = new Set(["webserver-backed", "stateful", "visual"]);

function usage() {
  process.stderr.write(
    "usage: run-service-backed-schedule.mjs --target <target> [--manifest <path>] [--defer-summary] [--resource-limit <name=value>...]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
    deferSummary: false,
    resourceLimitOverrides: new Map(),
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
    if (arg === "--resource-limit") {
      const value = argv[index + 1] ?? "";
      const [resource, amount] = parseResourceLimitOverride(value);
      options.resourceLimitOverrides.set(resource.trim(), amount);
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.target || !options.manifest) {
    usage();
  }
  return options;
}

async function loadBrowserBatchStages() {
  return loadBrowserBatchStagesFromManifest(defaultBrowserBatchManifestPath);
}

function normalizeResourceLimits(value, label, capacityProfile, overrides) {
  return normalizeSchedulerResourceLimits(value, label, {
    scheduler: "service_backed",
    capacityProfile,
    overrides,
    allowAuto: true,
    env: process.env,
  });
}

function normalizeResourceClaims(value, label, resourceLimits) {
  return normalizeSchedulerResourceClaims(value, label, resourceLimits, {
    scheduler: "service_backed",
    allowBounded: false,
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
  if (source.type !== "browser_stage") {
    throw new Error(`${label} browser target ${target} must use type browser_stage`);
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

function validateBrowserGroup(source, group, groupIndex, label, resourceLimits) {
  const groupLabel = `${label} ${source.target} groups ${groupIndex + 1}`;
  if (!group || typeof group !== "object" || Array.isArray(group)) {
    throw new Error(`${groupLabel} must be an object`);
  }
  for (const field of ["id", "name", "kind", "target", "aggregate_target"]) {
    if (typeof group[field] !== "string" || group[field].trim() === "") {
      throw new Error(`${groupLabel}.${field} must be a non-empty string`);
    }
  }
  if (group.aggregate_target.trim() !== source.target.trim()) {
    throw new Error(`${groupLabel}.aggregate_target must match ${source.target}`);
  }
  if (!Number.isFinite(group.weight) || group.weight < 0) {
    throw new Error(`${groupLabel}.weight must be non-negative`);
  }
  if (
    group.scheduler_priority !== undefined &&
    (!Number.isInteger(group.scheduler_priority) || group.scheduler_priority < 0)
  ) {
    throw new Error(`${groupLabel}.scheduler_priority must be a non-negative integer`);
  }
  if (!validBrowserGroupKinds.has(group.kind.trim())) {
    throw new Error(`${groupLabel}.kind must be one of ${Array.from(validBrowserGroupKinds).join(", ")}`);
  }
  if (group.kind.trim() === "functional_shard") {
    if (typeof group.shard_name !== "string" || group.shard_name.trim() === "") {
      throw new Error(`${groupLabel}.shard_name must be a non-empty string for functional_shard`);
    }
    if (!Number.isInteger(group.shard_index) || group.shard_index < 0) {
      throw new Error(`${groupLabel}.shard_index must be a non-negative integer for functional_shard`);
    }
    if (!Number.isInteger(group.shard_count) || group.shard_count < 1) {
      throw new Error(`${groupLabel}.shard_count must be a positive integer for functional_shard`);
    }
    if (group.shard_index >= group.shard_count) {
      throw new Error(`${groupLabel}.shard_index must be less than shard_count`);
    }
    if (!Array.isArray(group.entry_ids) || group.entry_ids.length === 0) {
      throw new Error(`${groupLabel}.entry_ids must be non-empty for functional_shard`);
    }
  }
  return {
    id: group.id.trim(),
    name: group.name.trim(),
    kind: group.kind.trim(),
    target: group.target.trim(),
    aggregateTarget: group.aggregate_target.trim(),
    coverage: typeof group.coverage === "string" ? group.coverage.trim() : "",
    executionDependency: typeof group.execution_dependency === "string" ? group.execution_dependency.trim() : "",
    shardName: typeof group.shard_name === "string" ? group.shard_name.trim() : "",
    shardIndex: Number.isInteger(group.shard_index) ? group.shard_index : 0,
    shardCount: Number.isInteger(group.shard_count) ? group.shard_count : 0,
    phases: Array.isArray(group.phases) ? group.phases.filter((entry) => typeof entry === "string") : [],
    entryIDs: Array.isArray(group.entry_ids) ? group.entry_ids.filter((entry) => typeof entry === "string") : [],
    schedulerPriority: group.scheduler_priority ?? 0,
    weight: group.weight,
    resourceClaims: normalizeResourceClaims(
      group.resource_claims ?? {},
      groupLabel,
      resourceLimits,
    ),
    rawResourceClaims: group.resource_claims ?? {},
  };
}

function validateSource(scheduleTarget, source, index, resourceLimits, browserStages) {
  const label = `${scheduleTarget} work_unit_sources ${index + 1}`;
  if (!source || typeof source !== "object" || Array.isArray(source)) {
    throw new Error(`${label} must be an object`);
  }
  if (!validSourceTypes.has(source.type)) {
    throw new Error(`${label} must declare type go_shards, make_target, or browser_stage`);
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
  if (
    source.scheduler_priority !== undefined &&
    (!Number.isInteger(source.scheduler_priority) || source.scheduler_priority < 0)
  ) {
    throw new Error(`${label} ${target} scheduler_priority must be a non-negative integer`);
  }
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
      schedulerPriority: source.scheduler_priority ?? 0,
      resourceClaims,
      rawResourceClaims: source.resource_claims,
      order: index,
    };
  }

  if (source.type === "browser_stage") {
    if (source.class !== "browser") {
      throw new Error(`${label} ${target} browser_stage sources must declare class browser`);
    }
    const groups = Array.isArray(source.groups)
      ? source.groups.map((group, groupIndex) =>
          validateBrowserGroup(source, group, groupIndex, label, resourceLimits),
        )
      : [];
    if (groups.length === 0) {
      throw new Error(`${label} ${target} browser_stage sources must declare groups[]`);
    }
    return {
      type: source.type,
      class: source.class,
      target,
      browserStage: source.browser_stage.trim(),
      needs,
      schedulerPriority: source.scheduler_priority ?? 0,
      weight: source.weight,
      resourceClaims,
      rawResourceClaims: source.resource_claims,
      groups,
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
    schedulerPriority: source.scheduler_priority ?? 0,
    weight: source.weight,
    resourceClaims,
    rawResourceClaims: source.resource_claims,
    order: index,
  };
}

function findSchedule(manifest, target, browserStages, overrides) {
  const schedule = selectSingleSchedule(manifest, target, { label: "schedule" });
  if (!Array.isArray(schedule.work_unit_sources) || schedule.work_unit_sources.length === 0) {
    throw new Error(`schedule ${target} must declare at least one work_unit_sources entry`);
  }
  if (schedule.children !== undefined) {
    throw new Error(`schedule ${target} must use work_unit_sources, not legacy children`);
  }
  const normalizedLimits = normalizeResourceLimits(
    schedule.resource_limits,
    `schedule ${target}`,
    schedule.capacity_profile ?? null,
    overrides,
  );
  const resourceLimits = provisionalResourceLimitsForClaims(normalizedLimits.limits);
  const sources = schedule.work_unit_sources.map((source, index) =>
    validateSource(target, source, index, resourceLimits, browserStages),
  );
  validateMeasurementIsolation(sources, `schedule ${target}`);
  validateTargetDependencyGraph(sources, {
    scheduleLabel: `schedule ${target}`,
    nodeKind: "source",
    duplicateTargetsMessage: (duplicateTargets) =>
      `schedule ${target} contains duplicate work-unit source targets: ${duplicateTargets.join(", ")}`,
  });
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

function validateMeasurementIsolation(sources, scheduleLabel) {
  const browserSources = sources.filter((source) => source.type === "browser_stage");
  const measurement = browserSources.find((source) => source.browserStage === "measurement");
  if (!measurement) {
    return;
  }
  const expectedNeeds = [];
  for (const source of browserSources) {
    if (source.target === measurement.target) {
      continue;
    }
    if (!measurementIsolationStages.has(source.browserStage)) {
      throw new Error(
        `${scheduleLabel} browser measurement isolation must explicitly account for newly selected stage ${source.browserStage}`,
      );
    }
    expectedNeeds.push(source.target);
  }
  expectedNeeds.sort();
  const actualNeeds = [...measurement.needs].sort();
  if (JSON.stringify(actualNeeds) !== JSON.stringify(expectedNeeds)) {
    throw new Error(
      `${scheduleLabel} browser measurement must depend exactly on isolated browser stages ${expectedNeeds.join(",") || "(none)"}`,
    );
  }
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

function retainedBrowserStageClaims(resourceClaims) {
  return new Map(
    Array.from(resourceClaims.entries()).filter(
      ([resource]) => resource !== goCPUResource && resource !== goIOResource,
    ),
  );
}

function expandSchedule(schedule) {
  const countedWorkUnits = [];
  const finalizerUnits = [];
  const shardWorkByName = new Map();

  for (const source of schedule.sources) {
    if (source.type === "browser_stage") {
      const retainedClaims = retainedBrowserStageClaims(source.resourceClaims);
      countedWorkUnits.push({
        id: `browser-stage-session:${source.browserStage}`,
        label: `${source.target}/stage-session`,
        kind: "browser_stage_session",
        type: "browser_stage_session",
        class: source.class,
        target: source.target,
        aggregateTarget: source.target,
        group: source.target,
        browserStage: source.browserStage,
        needs: [...source.needs],
        completionKeys: [browserStageSessionKey(source.target)],
        failureKeys: [browserStageSessionKey(source.target)],
        weight: source.weight,
        schedulerPriority: source.schedulerPriority,
        resourceClaims: cloneResourceClaims(source.resourceClaims),
        retainedResourceClaims: retainedClaims,
        order: source.order,
      });
      finalizerUnits.push({
        id: `browser-stage-complete:${source.browserStage}`,
        label: `${source.target}/complete`,
        kind: "browser_stage_complete",
        type: "browser_stage_complete",
        class: source.class,
        target: source.target,
        aggregateTarget: source.target,
        group: source.target,
        browserStage: source.browserStage,
        needs: browserStageCompletionNeeds(source.groups),
        completionKeys: [source.target],
        failureKeys: [source.target],
        countInTotal: false,
        countsStarted: false,
        resourceClaims: new Map(),
        releaseRetainedResourceClaims: retainedClaims,
        weight: 0,
        order: source.order,
      });
      for (const group of source.groups) {
        countedWorkUnits.push({
          id: group.id,
          label: `${source.target}/${group.name}`,
          kind: "browser_group",
          type: "browser_group",
          class: source.class,
          target: group.target,
          aggregateTarget: source.target,
          group: source.target,
          browserStage: source.browserStage,
          browserGroup: group,
          browserWorkerEnv: browserGroupWorkerEnv(source.groups, group),
          needs: browserGroupNeeds(browserStageSessionKey(source.target)),
          completionKeys: [browserGroupCompletionKey(group.id)],
          failureKeys: [browserGroupCompletionKey(group.id)],
          weight: group.weight,
          schedulerPriority: group.schedulerPriority,
          resourceClaims: cloneResourceClaims(group.resourceClaims),
          rawResourceClaims: group.rawResourceClaims,
          order: source.order,
        });
      }
      continue;
    }

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
        schedulerPriority: source.schedulerPriority,
        resourceClaims: cloneResourceClaims(source.resourceClaims),
        rawResourceClaims: source.rawResourceClaims,
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
        schedulerPriority: source.schedulerPriority,
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
      right.schedulerPriority - left.schedulerPriority ||
      right.weight - left.weight ||
      left.order - right.order ||
      left.label.localeCompare(right.label),
  );
  const workUnits = [...countedWorkUnits, ...finalizerUnits];
  const resolvedLimits = resolveResourceLimits(schedule.resourceLimits, schedule.resourceLimitSources, countedWorkUnits);
  const finalWorkUnits = workUnits.map((unit) => {
    if (!["make_target", "browser_stage_session", "browser_group"].includes(unit.kind) || !unit.rawResourceClaims) {
      return unit;
    }
    const { rawResourceClaims, ...rest } = unit;
    return {
      ...rest,
      resourceClaims: normalizeResourceClaims(
        rawResourceClaims,
        `${schedule.target} work_unit_sources ${unit.order + 1} ${unit.target}`,
        resolvedLimits.resourceLimits,
      ),
    };
  });
  return {
    ...schedule,
    ...resolvedLimits,
    workUnits: finalWorkUnits,
    totalWorkUnits: countedWorkUnits.length,
    finalizerCount: finalizerUnits.length,
  };
}

async function readJSONEnvFile(file) {
  const parsed = JSON.parse(await readFile(file, "utf8"));
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error(`${file} must contain a JSON environment object`);
  }
  return Object.fromEntries(
    Object.entries(parsed).filter((entry) => typeof entry[1] === "string"),
  );
}

function clampInteger(value, min, max) {
  return Math.min(max, Math.max(min, value));
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
  return clampInteger(Math.max(4, Math.min(hostConcurrency, weightedConcurrency)), 4, 16);
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
  return clampInteger(Math.max(6, goCPULimit + 2, profileConcurrency), 6, 24);
}

function resolveResourceLimits(resourceLimits, resourceLimitSources, workUnits) {
  const goShardUnits = workUnits.filter((unit) => unit.kind === "go_shard");
  return resolveAutoResourceLimits(resourceLimits, resourceLimitSources, "service-backed schedule", {
    service_backed_go_cpu: () => estimateGoCPULimit(goShardUnits),
    service_backed_go_io: ({ resourceLimits: currentLimits }) =>
      estimateGoIOLimit(goShardUnits, currentLimits.get(goCPUResource)),
    service_backed_browser_stack: ({ resourceLimits: currentLimits }) =>
      estimateBrowserStackAutoLimit(workUnits, currentLimits, { cpuResources: [goCPUResource] }),
  });
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

function attachRuntime(schedule, { makeBin, testOutputScript, deferSummary, goTargetRunner, metadataDir }) {
  const capacityDisplay = displayCapacity(schedule);
  const browserSessionScript =
    process.env.CARTULARY_BROWSER_E2E_SESSION_SCRIPT || path.join(repoRoot, "scripts", "start-web-e2e.sh");
  const browserGroupRunner = process.env.CARTULARY_BROWSER_E2E_GROUP_RUNNER || "";
  const testOutputCommand = testOutputScript.endsWith(".mjs")
    ? `${JSON.stringify(process.env.NODE_BIN || process.execPath)} ${JSON.stringify(testOutputScript)}`
    : JSON.stringify(testOutputScript);
  const cartularyTestServicesBin =
    process.env.CARTULARY_TEST_SERVICES_BIN || process.env.TEST_SERVICES_BIN || "";
  const browserSessionFiles = new Map(
    schedule.workUnits
      .filter((unit) => unit.kind === "browser_stage_session")
      .map((unit) => [
        unit.target,
        {
          envFile: path.join(metadataDir, `${unit.browserStage}-browser-env.json`),
          leaseFile: path.join(metadataDir, `${unit.browserStage}-browser-lease.json`),
        },
      ]),
  );
  const browserSessionEnvFor = async (target) => {
    const files = browserSessionFiles.get(target);
    return files ? readJSONEnvFile(files.envFile) : {};
  };
  for (const unit of schedule.workUnits) {
    if (unit.kind === "make_target") {
      unit.command = () => ({
        command: makeBin,
        args: ["--no-print-directory", "--output-sync=target", "-j1", unit.target],
        env: {
          ...makeChildEnv(process.env),
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    if (unit.kind === "browser_stage_session") {
      const files = browserSessionFiles.get(unit.target);
      unit.command = () => ({
        command: browserSessionScript,
        args: [
          "--session-start",
          "--env-file",
          files.envFile,
          "--lease-file",
          files.leaseFile,
        ],
        env: {
          ...process.env,
          CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    if (unit.kind === "browser_group") {
      unit.command = async () => {
        const sessionEnv = await browserSessionEnvFor(unit.aggregateTarget);
        const group = unit.browserGroup;
        const pnpmBin = process.env.PNPM || path.join(repoRoot, "tmp", "node-runtime", "bin", "pnpm");
        const commonEnv = {
          ...process.env,
          ...sessionEnv,
          ...unit.browserWorkerEnv,
          CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
          CARTULARY_TEST_TARGET: unit.aggregateTarget,
          CARTULARY_BROWSER_STAGE: unit.browserStage,
          CARTULARY_BROWSER_GROUP_KIND: group.kind,
          CARTULARY_BROWSER_GROUP_NAME: group.name,
          CARTULARY_BROWSER_GROUP_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        };
        if (browserGroupRunner) {
          return {
            command: browserGroupRunner,
            args: [],
            env: commonEnv,
          };
        }
        if (group.kind === "functional_shard") {
          return {
            command: path.join(repoRoot, "scripts", "lib", "run-playwright-webserver-batch.sh"),
            args: [
              "functional-shard",
              group.shardName,
              String(group.shardIndex),
              String(group.shardCount),
              "--",
              pnpmBin,
              "--dir",
              "apps/web",
              "exec",
              "playwright",
              "test",
              "--config",
              "playwright.webserver-backed.config.ts",
            ],
            env: commonEnv,
          };
        }
        if (group.kind === "support") {
          return {
            command: path.join(repoRoot, "scripts", "lib", "run-playwright-webserver-batch.sh"),
            args: [
              "support",
              "--",
              pnpmBin,
              "--dir",
              "apps/web",
              "exec",
              "playwright",
              "test",
              "--config",
              "playwright.webserver-backed.config.ts",
            ],
            env: commonEnv,
          };
        }
        const scriptsByKind = new Map([
          ["stateful", "run-browser-e2e-stateful.sh"],
          ["measurement", "run-browser-e2e-measurement.sh"],
          ["visual", "run-browser-e2e-visual.sh"],
        ]);
        const script = scriptsByKind.get(group.kind);
        if (!script) {
          throw new Error(`unsupported browser group kind ${group.kind}`);
        }
        return {
          command: path.join(repoRoot, "scripts", script),
          args: [],
          env: {
            ...commonEnv,
            CARTULARY_TEST_TARGET: unit.target,
            PLAYWRIGHT_WORKERS: "1",
          },
        };
      };
      continue;
    }
    if (unit.kind === "browser_stage_complete") {
      const files = browserSessionFiles.get(unit.target);
      unit.command = () => ({
        command: "bash",
        args: [
          "-c",
          [
            `${testOutputCommand} target-summary ${JSON.stringify(unit.target)} pass --quiet-success`,
            `summary_status=$?`,
            `${JSON.stringify(browserSessionScript)} --session-stop --lease-file ${JSON.stringify(files.leaseFile)}`,
            `stop_status=$?`,
            `if [[ "$summary_status" -ne 0 ]]; then exit "$summary_status"; fi`,
            `exit "$stop_status"`,
          ].join("; "),
        ],
        env: {
          ...process.env,
          CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
          CARTULARY_TEST_TARGET: unit.target,
          TEST_OUTPUT_SCRIPT: testOutputScript,
        },
      });
      continue;
    }
    if (unit.kind === "go_shard") {
      unit.command = () => ({
        command: goTargetRunner,
        args: ["capture-shard", unit.target, unit.shard, metadataDir],
        env: {
          ...process.env,
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    unit.command = () => ({
      command: goTargetRunner,
      args: ["finalize-shards", unit.aggregateTarget, metadataDir],
      env: {
        ...process.env,
        CARTULARY_TEST_TARGET: unit.aggregateTarget,
        TEST_OUTPUT_SCRIPT: testOutputScript,
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
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
    showFinalizing: true,
    deferInitialProgress: true,
    validateSummaryTiming: !deferSummary,
    stopOnFirstFailure: false,
    runningDisplayUnits: finalizerRunningDisplayUnits,
    countCompletedUnit: countVisibleCompletedUnit,
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
    beforeReplayLog: replayFailedAggregateLogsBeforeFinalizer,
    shouldReplayLog: ({ result, reporter }) => result.status !== 0 || reporter.verbose,
    afterWorkComplete: async ({ firstFailure }) => {
      if (firstFailure !== 0 || schedule.backendChildren.length === 0) {
        for (const files of browserSessionFiles.values()) {
          if (!existsSync(files.leaseFile)) {
            continue;
          }
          await runLifecycle(repoRoot, browserSessionScript, [
            "--session-stop",
            "--lease-file",
            files.leaseFile,
          ]).catch(() => {});
        }
        return null;
      }
      for (const files of browserSessionFiles.values()) {
        if (!existsSync(files.leaseFile)) {
          continue;
        }
        await runLifecycle(repoRoot, browserSessionScript, [
          "--session-stop",
          "--lease-file",
          files.leaseFile,
        ]).catch(() => {});
      }
      const status = await runPostgresFixtureBudgetCheck(schedule.backendChildren);
      return status === 0 ? null : { status, label: "postgres-fixture-budget" };
    },
    summaryExtra: ({ started }) => ({
      started_count: started,
    }),
    afterSummary: async ({ requestedStatus }) => {
      if (deferSummary) {
        return;
      }
      await runLifecycle(
        repoRoot,
        testOutputScript,
        [
          "target-summary",
          schedule.target,
          requestedStatus,
          "--children",
          schedule.children.join(","),
        ],
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
  const { manifest, manifestPath } = await loadScheduleManifest(options.manifest, {
    repoRoot,
    schemaID: supportedSchemaID,
  });
  const browserStages = await loadBrowserBatchStages();
  const schedule = expandSchedule(
    findSchedule(manifest, options.target, browserStages, options.resourceLimitOverrides),
  );
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
