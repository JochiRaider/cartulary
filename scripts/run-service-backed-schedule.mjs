#!/usr/bin/env node
import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectGoShardsForTarget } from "./lib/go-shard-plan.mjs";
import { loadBrowserBatchStages as loadBrowserBatchStagesFromManifest } from "./lib/browser-batch-manifest.mjs";
import {
  formatResourceList,
  formatResourceMap,
  relToRepo as relToRepoPath,
  resourceMapToObject,
  schedulerActiveGroups,
  schedulerBlockedUnitRecords,
  schedulerDryRunLine,
  schedulerLogDir as schedulerLogResultsDir,
  schedulerProgressIntervalMs,
  schedulerProgressEventFields,
  schedulerProgressLine,
  schedulerProgressSnapshot,
  schedulerStartLine,
  schedulerSummaryLine,
  schedulerWaitingOnForUnits,
  schedulerTargetDir as schedulerTargetResultsDir,
  writeSchedulerTelemetry,
  verboseSchedulerOutput,
} from "./lib/scheduler-reporting.mjs";
import { findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const defaultBrowserBatchManifestPath = path.join(repoRoot, "tools", "browser_e2e_batch_manifest.json");
const supportedSchemaID = "cartulary.service_backed_schedule.v7";
const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const postgresResetResource = "postgres_reset";
const browserStackResource = "browser_stack";
const goCPULimitEnv = "CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT";
const goIOLimitEnv = "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT";
const browserStackLimitEnv = "CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const schedulerEventSchemaID = "cartulary.service_backed_scheduler_event.v3";
const schedulerSummarySchemaID = "cartulary.service_backed_scheduler_summary.v3";
const autoLimitResources = new Set([goCPUResource, goIOResource, browserStackResource]);
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

function isDryRun() {
  const flags = ` ${process.env.MAKEFLAGS ?? ""} `;
  return flags.includes(" n") || flags.includes(" --just-print") || flags.includes(" --dry-run");
}

function schedulerTargetDir(target) {
  return schedulerTargetResultsDir(repoRoot, target);
}

function schedulerLogDir(target) {
  return schedulerLogResultsDir(repoRoot, target);
}

function relToRepo(value) {
  return relToRepoPath(repoRoot, value);
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
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_limits must be an object`);
  }

  const resourceLimits = new Map();
  for (const [resource, limit] of Object.entries(value)) {
    const normalizedResource = resource.trim();
    if (normalizedResource === "") {
      throw new Error(`${label} resource_limits keys must be non-empty strings`);
    }
    if (normalizedResource === "backend") {
      throw new Error(`${label} resource_limits must not declare removed generic backend resource`);
    }
    if (normalizedResource === "browser") {
      throw new Error(`${label} resource_limits must not declare removed generic browser resource`);
    }
    if (limit === "auto") {
      if (!autoLimitResources.has(normalizedResource)) {
        throw new Error(
          `${label} resource_limits.${normalizedResource} may use "auto" only for ${goCPUResource}, ${goIOResource}, or ${browserStackResource}`,
        );
      }
      resourceLimits.set(normalizedResource, limit);
      continue;
    }
    if (!Number.isInteger(limit) || limit < 1) {
      throw new Error(
        `${label} resource_limits.${normalizedResource} must be a positive integer or "auto"`,
      );
    }
    resourceLimits.set(normalizedResource, limit);
  }
  return resourceLimits;
}

function normalizeResourceClaims(value, label, resourceLimits) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_claims must be an object`);
  }

  const resourceClaims = new Map();
  for (const [resource, amount] of Object.entries(value)) {
    const normalizedResource = resource.trim();
    if (normalizedResource === "") {
      throw new Error(`${label} resource_claims keys must be non-empty strings`);
    }
    if (normalizedResource === "backend") {
      throw new Error(`${label} resource_claims must not declare removed generic backend resource`);
    }
    if (normalizedResource === "browser") {
      throw new Error(`${label} resource_claims must not declare removed generic browser resource`);
    }
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`${label} resource_claims.${normalizedResource} must be a positive integer`);
    }
    if (!resourceLimits.has(normalizedResource)) {
      throw new Error(
        `${label} resource_claims entry ${normalizedResource} is not declared in resource_limits`,
      );
    }
    resourceClaims.set(normalizedResource, amount);
  }
  return resourceClaims;
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
  const resourceLimits = normalizeResourceLimits(schedule.resource_limits, `schedule ${target}`);
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
    sources,
    children: sources.map((source) => source.target),
    backendChildren: sources
      .filter((source) => source.class === "backend")
      .map((source) => source.target),
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

function expandSchedule(schedule) {
  const workUnits = [];
  const goFinalizers = [];
  const shardWorkByName = new Map();

  for (const source of schedule.sources) {
    if (source.type === "make_target") {
      workUnits.push({
        id: source.target,
        label: source.target,
        type: "make_target",
        class: source.class,
        target: source.target,
        aggregateTarget: source.target,
        needs: [...source.needs],
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
    goFinalizers.push({
      target: source.target,
      shardNames: shards.map((shard) => shard.name),
      needs: [...source.needs],
    });
    for (const shard of shards) {
      if (shardWorkByName.has(shard.name)) {
        continue;
      }
      const unit = {
        id: `${source.target}:${shard.name}`,
        label: `${source.target}/${shard.name}`,
        type: "go_shard",
        class: source.class,
        target: source.target,
        aggregateTarget: source.target,
        needs: [...source.needs],
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
      workUnits.push(unit);
    }
  }

  workUnits.sort(
    (left, right) =>
      right.weight - left.weight ||
      left.order - right.order ||
      left.label.localeCompare(right.label),
  );
  return {
    ...schedule,
    resourceLimits: resolveResourceLimits(schedule.resourceLimits, workUnits),
    workUnits,
    goFinalizers,
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

function resolveResourceLimits(resourceLimits, workUnits) {
  const goShardUnits = workUnits.filter((unit) => unit.type === "go_shard");
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
  for (const [resource, limit] of resourceLimits.entries()) {
    if (resource === goCPUResource && (limit === "auto" || goCPUOverride !== null)) {
      resolved.set(resource, effectiveGoCPU);
      continue;
    }
    if (resource === goIOResource && (limit === "auto" || goIOOverride !== null)) {
      resolved.set(resource, effectiveGoIO);
      continue;
    }
    if (resource === browserStackResource && (limit === "auto" || browserStackOverride !== null)) {
      resolved.set(resource, effectiveBrowserStack);
      continue;
    }
    resolved.set(resource, limit);
  }
  return resolved;
}

function runLifecycle(testOutputScript, args, stream = process.stdout) {
  return new Promise((resolve, reject) => {
    const child = spawn(testOutputScript, args, {
      cwd: repoRoot,
      env: process.env,
      stdio: ["ignore", "pipe", "pipe"],
    });
    child.stdout.pipe(stream, { end: false });
    child.stderr.pipe(process.stderr, { end: false });
    child.on("error", reject);
    child.on("close", (status) => {
      if (status === 0) {
        resolve();
        return;
      }
      reject(new Error(`${testOutputScript} ${args.join(" ")} exited ${status}`));
    });
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

async function replayLog(file, stream) {
  await new Promise((resolve, reject) => {
    const reader = createReadStream(file);
    reader.on("error", reject);
    reader.on("end", resolve);
    reader.pipe(stream, { end: false });
  });
}

function sanitizeLogName(value) {
  return value.replace(/[^A-Za-z0-9._-]+/g, "-");
}

function runCommand(command, args, logFile, env = process.env) {
  return new Promise((resolve) => {
    const log = createWriteStream(logFile);
    let settled = false;
    const finish = (status) => {
      if (settled) {
        return;
      }
      settled = true;
      log.end(() => resolve({ status }));
    };
    const child = spawn(command, args, {
      cwd: repoRoot,
      env,
      stdio: ["ignore", "pipe", "pipe"],
    });
    child.stdout.pipe(log, { end: false });
    child.stderr.pipe(log, { end: false });
    child.on("error", (error) => {
      log.write(`${error.message}\n`);
      finish(127);
    });
    child.on("close", (status) => {
      finish(status ?? 1);
    });
  });
}

function sanitizeMakeFlags(value) {
  if (!value) {
    return "";
  }
  return value
    .split(/\s+/)
    .filter(Boolean)
    .filter((entry) => !entry.startsWith("--jobserver-auth="))
    .filter((entry) => !entry.startsWith("--jobserver-fds="))
    .filter((entry) => !entry.startsWith("--jobserver-style="))
    .filter((entry) => !entry.startsWith("-j"))
    .join(" ");
}

function makeChildEnv(env = process.env) {
  const childEnv = { ...env };
  for (const name of ["MAKEFLAGS", "MFLAGS"]) {
    const sanitized = sanitizeMakeFlags(childEnv[name]);
    if (sanitized) {
      childEnv[name] = sanitized;
    } else {
      delete childEnv[name];
    }
  }
  return childEnv;
}

function unitCommand({ makeBin, unit, metadataDir, goTargetRunner }) {
  if (unit.type === "make_target") {
    return {
      command: makeBin,
      args: ["--no-print-directory", "--output-sync=target", "-j1", unit.target],
      env: makeChildEnv(process.env),
    };
  }
  return {
    command: goTargetRunner,
    args: ["capture-shard", unit.target, unit.shard, metadataDir],
    env: {
      ...process.env,
      CARTULARY_TEST_TARGET: unit.target,
    },
  };
}

function workUnitLogFile(logDir, unit, started) {
  return path.join(logDir, `${String(started).padStart(2, "0")}-${sanitizeLogName(unit.id)}.log`);
}

function finalizerLogFile(logDir, target) {
  return path.join(logDir, `finalize-${sanitizeLogName(target)}.log`);
}

function runWorkUnit({ makeBin, unit, metadataDir, logFile, goTargetRunner }) {
  const { command, args, env } = unitCommand({ makeBin, unit, metadataDir, goTargetRunner });
  return runCommand(command, args, logFile, env).then((result) => ({
    id: unit.id,
    label: unit.label,
    status: result.status,
    logFile,
  }));
}

function runFinalizer({ target, metadataDir, logFile, testOutputScript, goTargetRunner }) {
  return runCommand(
    goTargetRunner,
    ["finalize-shards", target, metadataDir],
    logFile,
    {
      ...process.env,
      CARTULARY_TEST_TARGET: target,
      TEST_OUTPUT_SCRIPT: testOutputScript,
    },
  ).then((result) => ({
    id: `finalize:${target}`,
    label: `finalize/${target}`,
    status: result.status,
    logFile,
  }));
}

function hasResourceCapacity(unit, resourceLimits, activeResourceClaims) {
  return blockedResourcesForUnit(unit, resourceLimits, activeResourceClaims).length === 0;
}

function blockedResourcesForUnit(unit, resourceLimits, activeResourceClaims) {
  const blocked = [];
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const limit = resourceLimits.get(resource);
    if (limit !== undefined && (activeResourceClaims.get(resource) ?? 0) + amount > limit) {
      blocked.push(resource);
    }
  }
  return blocked;
}

function blockedResourcesForPending(workUnits, resourceLimits, activeResourceClaims) {
  const resources = new Set();
  for (const unit of workUnits) {
    for (const resource of blockedResourcesForUnit(unit, resourceLimits, activeResourceClaims)) {
      resources.add(resource);
    }
  }
  return Array.from(resources).sort((left, right) => left.localeCompare(right));
}

function addResourceClaims(unit, activeResourceClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    activeResourceClaims.set(resource, (activeResourceClaims.get(resource) ?? 0) + amount);
  }
}

function removeResourceClaims(unit, activeResourceClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const next = (activeResourceClaims.get(resource) ?? 0) - amount;
    if (next <= 0) {
      activeResourceClaims.delete(resource);
    } else {
      activeResourceClaims.set(resource, next);
    }
  }
}

function formatResourceLimits(resourceLimits) {
  return formatResourceMap(resourceLimits);
}

function formatResourceClaims(resourceClaims) {
  return formatResourceMap(resourceClaims);
}

function formatActiveResourceClaims(activeResourceClaims) {
  return formatResourceMap(activeResourceClaims);
}

function formatBlockedWorkUnits(workUnits) {
  return workUnits
    .map((unit) => `${unit.label} claims=${formatResourceClaims(unit.resourceClaims)}`)
    .join("; ");
}

function schedulerStateFields({ pending, running, activeResourceClaims, resourceLimits }) {
  return [
    `active=${running.size}`,
    `pending=${pending.length}`,
    `active_resource_claims=${formatActiveResourceClaims(activeResourceClaims)}`,
    `resource_limits=${formatResourceLimits(resourceLimits)}`,
  ];
}

function displayCapacity(schedule) {
  return schedule.resourceLimits.get(goCPUResource) ?? Math.max(...schedule.resourceLimits.values());
}

function finalizerIsReady(finalizer, completedShards) {
  return finalizer.shardNames.every((shardName) => completedShards.has(shardName));
}

function failedDependencyForUnit(unit, failedSources) {
  for (const need of unit.needs ?? []) {
    if (failedSources.has(need)) {
      return need;
    }
  }
  return null;
}

function sourceDependenciesSatisfied(unit, completedSources) {
  return (unit.needs ?? []).every((need) => completedSources.has(need));
}

function readyPendingUnits(pending, completedSources, failedSources) {
  return pending.filter(
    (unit) => !failedDependencyForUnit(unit, failedSources) && sourceDependenciesSatisfied(unit, completedSources),
  );
}

function dependencyBlockedPendingUnits(pending, completedSources, failedSources) {
  return pending.filter(
    (unit) => !failedDependencyForUnit(unit, failedSources) && !sourceDependenciesSatisfied(unit, completedSources),
  );
}

function schedulerProgressDelay() {
  let timeout;
  const promise = new Promise((resolve) => {
    timeout = setTimeout(() => resolve({ schedulerProgressTick: true }), schedulerProgressIntervalMs);
  });
  return {
    promise,
    cancel() {
      clearTimeout(timeout);
    },
  };
}

async function createSchedulerReporter(schedule) {
  const targetDir = schedulerTargetDir(schedule.target);
  const logDir = schedulerLogDir(schedule.target);
  await mkdir(logDir, { recursive: true });
  return new SchedulerReporter(schedule, targetDir, logDir);
}

class SchedulerReporter {
  constructor(schedule, targetDir, logDir) {
    this.schedule = schedule;
    this.targetDir = targetDir;
    this.logDir = logDir;
    this.verbose = verboseSchedulerOutput();
    this.eventsPath = path.join(targetDir, "scheduler-events.jsonl");
    this.summaryPath = path.join(targetDir, "scheduler-summary.json");
    this.events = createWriteStream(this.eventsPath, { flags: "w" });
    this.startedAt = new Map();
    this.completedWork = [];
    this.completedLogFilesByTarget = new Map();
    this.skippedWork = [];
    this.blockedResourcesSeen = new Set();
    this.blockedExplanationsSeen = new Set();
    this.waitingOnSeen = new Set();
    this.lastProgressAt = Date.now();
    this.lastBlockedKey = null;
    this.completedCount = 0;
    this.failedWorkUnit = null;
    this.finalizerFailures = 0;
    this.maxRunningGroups = 0;
  }

  start() {
    if (!this.verbose) {
      return;
    }
    process.stdout.write(
      schedulerStartLine({
        prefix: "SCHEDULER",
        target: this.schedule.target,
        workUnitCount: this.schedule.workUnits.length,
        finalizerCount: this.schedule.goFinalizers.length,
        resourceLimits: this.schedule.resourceLimits,
        preferredResources: [
          goCPUResource,
          goIOResource,
          browserStackResource,
          "process",
          "postgres",
          "minio",
        ],
        workUnits: this.schedule.workUnits,
      }),
    );
  }

  emit(event, fields, state, detail = {}) {
    if (this.verbose) {
      writeSchedulerTelemetry(process.stdout, "SCHEDULER", this.schedule.target, event, fields);
    }
    this.writeEvent(event, state, detail);
  }

  runningDisplayUnits(state) {
    return [
      ...Array.from(state.running.values()),
      ...Array.from(state.runningFinalizers.values()).map((finalizer) => ({
        id: `finalize:${finalizer.target}`,
        label: `finalize:${finalizer.target}`,
        group: finalizer.target,
      })),
    ];
  }

  observeState(state) {
    this.maxRunningGroups = Math.max(
      this.maxRunningGroups,
      schedulerActiveGroups(this.runningDisplayUnits(state)).size,
    );
  }

  startUnit(unit, logFile, state) {
    this.startedAt.set(unit.id, Date.now());
    this.emit(
      "start",
      [
        `work_unit=${unit.label}`,
        `claims=${formatResourceClaims(unit.resourceClaims)}`,
        ...schedulerStateFields(state),
      ],
      state,
      {
        work_unit: unit.label,
        work_unit_id: unit.id,
        work_unit_type: unit.type,
        work_unit_class: unit.class,
        aggregate_target: unit.aggregateTarget,
        resource_claims: resourceMapToObject(unit.resourceClaims),
        log_file: relToRepo(logFile),
      },
    );
  }

  finishUnit(unit, result, state) {
    const durationMs = Math.max(0, Date.now() - (this.startedAt.get(unit.id) ?? Date.now()));
    this.startedAt.delete(unit.id);
    this.completedCount += 1;
    const record = {
      label: result.label,
      id: result.id,
      kind: "work_unit",
      status: result.status,
      duration_ms: durationMs,
      log_file: relToRepo(result.logFile),
    };
    this.completedWork.push(record);
    if (!this.completedLogFilesByTarget.has(unit.aggregateTarget)) {
      this.completedLogFilesByTarget.set(unit.aggregateTarget, []);
    }
    this.completedLogFilesByTarget.get(unit.aggregateTarget).push(result.logFile);
    if (result.status !== 0 && !this.failedWorkUnit) {
      this.failedWorkUnit = result.label;
    }
    this.emit(
      "finish",
      [
        `work_unit=${result.label}`,
        `status=${result.status}`,
        ...schedulerStateFields(state),
      ],
      state,
      {
        work_unit: result.label,
        work_unit_id: result.id,
        status: result.status,
        duration_ms: durationMs,
        log_file: relToRepo(result.logFile),
      },
    );
  }

  completedLogFilesForTarget(target) {
    return this.completedLogFilesByTarget.get(target) ?? [];
  }

  startFinalizer(finalizer, logFile, state) {
    const id = `finalize:${finalizer.target}`;
    this.startedAt.set(id, Date.now());
    this.emit(
      "finalize-start",
      [
        `target=${finalizer.target}`,
        `shards=${finalizer.shardNames.length}`,
        `active_finalizers=${state.runningFinalizers.size}`,
        `pending_finalizers=${state.pendingFinalizers.length}`,
      ],
      state,
      {
        finalizer: finalizer.target,
        finalizer_id: id,
        shards: finalizer.shardNames.length,
        log_file: relToRepo(logFile),
      },
    );
  }

  finishFinalizer(finalizer, result, state) {
    const id = `finalize:${finalizer.target}`;
    const durationMs = Math.max(0, Date.now() - (this.startedAt.get(id) ?? Date.now()));
    this.startedAt.delete(id);
    const record = {
      label: result.label,
      id: result.id,
      kind: "finalizer",
      status: result.status,
      duration_ms: durationMs,
      log_file: relToRepo(result.logFile),
    };
    this.completedWork.push(record);
    if (result.status !== 0 && !this.failedWorkUnit) {
      this.failedWorkUnit = result.label;
    }
    if (result.status !== 0) {
      this.finalizerFailures += 1;
    }
    this.emit(
      "finalize-finish",
      [
        `target=${finalizer.target}`,
        `status=${result.status}`,
        `active_finalizers=${state.runningFinalizers.size}`,
        `pending_finalizers=${state.pendingFinalizers.length}`,
      ],
      state,
      {
        finalizer: finalizer.target,
        finalizer_id: id,
        status: result.status,
        duration_ms: durationMs,
        log_file: relToRepo(result.logFile),
      },
    );
  }

  blocked(state, reason, blockedResources, { waitingOn = [], blockedUnits = [] } = {}) {
    const blockedKey = `${reason}:${blockedResources.join(",")}:${waitingOn.join(",")}:${JSON.stringify(blockedUnits)}`;
    for (const resource of blockedResources) {
      this.blockedResourcesSeen.add(resource);
    }
    for (const dependency of waitingOn) {
      this.waitingOnSeen.add(dependency);
    }
    this.emit(
      "blocked",
      [
        `reason=${reason}`,
        `blocked_resources=${formatResourceList(blockedResources)}`,
        ...schedulerStateFields(state),
      ],
      state,
      {
        blocked_reason: reason,
        blocked_resources: blockedResources,
        waiting_on: waitingOn,
        blocked_units: blockedUnits,
      },
    );
    this.maybeProgress(state, reason, blockedResources, {
      force: blockedKey !== this.lastBlockedKey,
      waitingOn,
      blockedUnits,
    });
    this.lastBlockedKey = blockedKey;
  }

  skipUnit(unit, state, reason, failedDependency) {
    const record = {
      label: unit.label,
      id: unit.id,
      aggregate_target: unit.aggregateTarget,
      reason,
      failed_dependency: failedDependency,
    };
    this.skippedWork.push(record);
    this.emit(
      "skip",
      [
        `work_unit=${unit.label}`,
        `reason=${reason}`,
        `failed_dependency=${failedDependency}`,
        ...schedulerStateFields(state),
      ],
      state,
      {
        work_unit: unit.label,
        work_unit_id: unit.id,
        aggregate_target: unit.aggregateTarget,
        skip_reason: reason,
        failed_dependency: failedDependency,
      },
    );
  }

  maybeProgress(
    state,
    reason = "none",
    blockedResources = [],
    { force = false, waitingOn = [], blockedUnits = [] } = {},
  ) {
    const now = Date.now();
    if (!force && now - this.lastProgressAt < schedulerProgressIntervalMs) {
      return;
    }
    this.lastProgressAt = now;
    const runningUnits = this.runningDisplayUnits(state);
    const progress = schedulerProgressSnapshot({
      runningUnits,
      startedAt: this.startedAt,
      now,
      reason,
      blockedResources,
      waitingOn,
      unblocksAfter: this.unblocksAfter(state, blockedResources),
    });
    for (const explanation of progress.blockedBy) {
      this.blockedExplanationsSeen.add(explanation);
    }
    for (const dependency of waitingOn) {
      this.waitingOnSeen.add(dependency);
    }
    this.writeEvent("progress", state, {
      blocked_reason: reason,
      blocked_resources: blockedResources,
      waiting_on: waitingOn,
      blocked_units: blockedUnits,
      ...schedulerProgressEventFields(progress),
    });
    process.stdout.write(
      schedulerProgressLine({
        prefix: "SCHEDULER",
        target: this.schedule.target,
        completed: this.completedCount,
        total: this.schedule.workUnits.length,
        running: state.running.size,
        pending: state.pending.length,
        blocked: state.blockedCount ?? 0,
        finalizing: state.runningFinalizers.size,
        activeGroups: progress.activeGroups,
        blockedBy: progress.blockedBy,
        waitingOn,
        unblocksAfter: progress.unblocksAfter,
        slowestRunning: progress.slowestRunning,
        artifacts: relToRepo(this.targetDir),
      }),
    );
  }

  unblocksAfter(state, blockedResources) {
    const runningUnits = Array.from(state.running.values());
    if (blockedResources.length > 0) {
      const candidates = runningUnits
        .filter((unit) => blockedResources.some((resource) => unit.resourceClaims.has(resource)))
        .sort((left, right) => {
          const leftStarted = this.startedAt.get(left.id) ?? Number.MAX_SAFE_INTEGER;
          const rightStarted = this.startedAt.get(right.id) ?? Number.MAX_SAFE_INTEGER;
          return leftStarted - rightStarted || left.label.localeCompare(right.label);
        });
      if (candidates.length > 0) {
        return candidates[0].label;
      }
    }
    const runningSources = new Set(runningUnits.map((unit) => unit.aggregateTarget));
    for (const finalizer of state.runningFinalizers.values()) {
      runningSources.add(finalizer.target);
    }
    for (const unit of state.pending) {
      for (const need of unit.needs ?? []) {
        if (runningSources.has(need)) {
          return need;
        }
      }
    }
    return "none";
  }

  async summary(status, { started, failedWorkUnit = null } = {}) {
    const failed = failedWorkUnit || this.failedWorkUnit || null;
    const slowest = this.slowestWork();
    const skipped = this.skippedWork.length;
    if (this.verbose || status !== "pass") {
      process.stdout.write(
        schedulerSummaryLine({
          prefix: "SCHEDULER",
          target: this.schedule.target,
          status,
          completed: this.completedCount,
          total: this.schedule.workUnits.length,
          failed,
          skipped,
          finalizerFailures: this.finalizerFailures,
          slowest,
        }),
      );
    }
    await writeFile(
      this.summaryPath,
      `${JSON.stringify(
        {
          schema_id: schedulerSummarySchemaID,
          target: this.schedule.target,
          status,
          total_work_units: this.schedule.workUnits.length,
          completed_work_units: this.completedCount,
          skipped_work_units: this.skippedWork,
          failed_work_unit: failed,
          started_count: started,
          finalizer_count: this.schedule.goFinalizers.length,
          finalizer_failures: this.finalizerFailures,
          max_running_groups: this.maxRunningGroups,
          blocked_resources_seen: Array.from(this.blockedResourcesSeen).sort((left, right) =>
            left.localeCompare(right),
          ),
          blocked_explanations_seen: Array.from(this.blockedExplanationsSeen).sort((left, right) =>
            left.localeCompare(right),
          ),
          waiting_on_seen: Array.from(this.waitingOnSeen).sort((left, right) =>
            left.localeCompare(right),
          ),
          slowest_work_units: slowest,
          artifacts: {
            events_jsonl: relToRepo(this.eventsPath),
            scheduler_logs_dir: relToRepo(this.logDir),
          },
        },
        null,
        2,
      )}\n`,
    );
  }

  slowestWork() {
    return [...this.completedWork]
      .sort((left, right) => right.duration_ms - left.duration_ms || left.label.localeCompare(right.label))
      .slice(0, 5);
  }

  writeEvent(event, state, detail) {
    this.observeState(state);
    this.events.write(
      `${JSON.stringify({
        schema_id: schedulerEventSchemaID,
        target: this.schedule.target,
        event,
        timestamp: new Date().toISOString(),
        pending: state.pending.length,
        running: state.running.size,
        total_work_units: this.schedule.workUnits.length,
        blocked: state.blockedCount ?? 0,
        completed: this.completedCount,
        pending_finalizers: state.pendingFinalizers.length,
        running_finalizers: state.runningFinalizers.size,
        blocked_reason: detail.blocked_reason ?? null,
        blocked_resources: detail.blocked_resources ?? [],
        waiting_on: detail.waiting_on ?? [],
        blocked_units: detail.blocked_units ?? [],
        active_resource_claims: resourceMapToObject(state.activeResourceClaims),
        resource_limits: resourceMapToObject(this.schedule.resourceLimits),
        ...detail,
      })}\n`,
    );
  }

  close() {
    return new Promise((resolve, reject) => {
      this.events.on("error", reject);
      this.events.end(resolve);
    });
  }
}

function writeDryRun(schedule, manifestPath, target) {
  const dependencyCount = schedule.sources.reduce((sum, source) => sum + source.needs.length, 0);
  process.stdout.write(
    schedulerDryRunLine({
      target,
      manifest: path.relative(repoRoot, manifestPath),
      resourceLimits: schedule.resourceLimits,
      preferredResources: [
        goCPUResource,
        goIOResource,
        postgresResetResource,
        browserStackResource,
        "process",
        "postgres",
        "minio",
      ],
      workUnits: schedule.workUnits,
      dependencies: dependencyCount,
      finalizerCount: schedule.goFinalizers.length,
    }),
  );
  if (!verboseSchedulerOutput()) {
    return;
  }
  for (const unit of schedule.workUnits) {
    const profile = unit.schedulerProfile ? ` profile=${unit.schedulerProfile}` : "";
    const needs = unit.needs.length > 0 ? ` needs=${unit.needs.join(",")}` : "";
    process.stdout.write(
      `[DRY-RUN] ${target} unit ${unit.label} type=${unit.type} class=${unit.class}${profile}${needs} claims=${formatResourceClaims(unit.resourceClaims)}\n`,
    );
  }
}

async function runSchedule({ schedule, makeBin, testOutputScript, deferSummary }) {
  const childrenCsv = schedule.children.join(",");
  const backendBudgetTargets = schedule.backendChildren;
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-service-backed-schedule-"));
  const reporter = await createSchedulerReporter(schedule);
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  const pending = [...schedule.workUnits];
  const pendingFinalizers = schedule.goFinalizers.map((finalizer) => ({
    target: finalizer.target,
    shardNames: [...finalizer.shardNames],
  }));
  const running = new Map();
  const runningFinalizers = new Map();
  const completedShards = new Set();
  const completedSources = new Set();
  const failedSources = new Map();
  const activeResourceClaims = new Map();
  const capacityDisplay = displayCapacity(schedule);
  const goTargetRunner =
    process.env[goTargetRunnerEnv] || path.join(repoRoot, "scripts", "run-go-target.sh");
  let started = 0;
  let firstFailure = 0;
  let firstFailureLabel = null;

  try {
    const stateSnapshot = () => ({
      pending,
      running,
      pendingFinalizers,
      runningFinalizers,
      activeResourceClaims,
      resourceLimits: schedule.resourceLimits,
    });

    if (reporter.verbose) {
      await runLifecycle(testOutputScript, [
        "target-start",
        schedule.target,
        "--children",
        childrenCsv,
        "--service-backed",
        "1",
      ]);
    }
    reporter.start();

    const startUnit = async (unit) => {
      started += 1;
      if (reporter.verbose) {
        await runLifecycle(testOutputScript, [
          "step-start",
          schedule.target,
          String(started),
          String(schedule.workUnits.length),
          unit.label,
          "--mode",
          "scheduler",
          "--jobs",
          String(capacityDisplay),
        ]);
      }
      const logFile = workUnitLogFile(reporter.logDir, unit, started);
      addResourceClaims(unit, activeResourceClaims);
      const promise = runWorkUnit({
        makeBin,
        unit,
        metadataDir,
        logFile,
        goTargetRunner,
      });
      running.set(promise, unit);
      reporter.startUnit(unit, logFile, stateSnapshot());
    };

    const startReadyFinalizers = () => {
      for (let index = 0; index < pendingFinalizers.length; ) {
        const finalizer = pendingFinalizers[index];
        if (!finalizerIsReady(finalizer, completedShards)) {
          index += 1;
          continue;
        }
        pendingFinalizers.splice(index, 1);
        const logFile = finalizerLogFile(reporter.logDir, finalizer.target);
        const promise = runFinalizer({
          target: finalizer.target,
          metadataDir,
          logFile,
          testOutputScript,
          goTargetRunner,
        });
        runningFinalizers.set(promise, finalizer);
        reporter.startFinalizer(finalizer, logFile, stateSnapshot());
      }
    };

    const markSourceFailed = (sourceTarget, failedLabel) => {
      if (!completedSources.has(sourceTarget) && !failedSources.has(sourceTarget)) {
        failedSources.set(sourceTarget, failedLabel);
      }
    };

    const skipDependencyFailedUnits = () => {
      let skipped = 0;
      for (let index = 0; index < pending.length; ) {
        const unit = pending[index];
        const failedDependency = failedDependencyForUnit(unit, failedSources);
        if (!failedDependency) {
          index += 1;
          continue;
        }
        pending.splice(index, 1);
        skipped += 1;
        markSourceFailed(unit.aggregateTarget, failedDependency);
        for (let finalizerIndex = 0; finalizerIndex < pendingFinalizers.length; ) {
          if (pendingFinalizers[finalizerIndex].target === unit.aggregateTarget) {
            pendingFinalizers.splice(finalizerIndex, 1);
          } else {
            finalizerIndex += 1;
          }
        }
        reporter.skipUnit(
          unit,
          { ...stateSnapshot(), blockedCount: skipped },
          "dependency_failure",
          failedDependency,
        );
      }
      return skipped;
    };

    while (
      pending.length > 0 ||
      running.size > 0 ||
      pendingFinalizers.length > 0 ||
      runningFinalizers.size > 0
    ) {
      skipDependencyFailedUnits();
      while (true) {
        const nextIndex = pending.findIndex(
          (candidate) =>
            sourceDependenciesSatisfied(candidate, completedSources) &&
            hasResourceCapacity(candidate, schedule.resourceLimits, activeResourceClaims),
        );
        if (nextIndex === -1) {
          break;
        }
        const [unit] = pending.splice(nextIndex, 1);
        await startUnit(unit);
      }

      const dependencyBlocked = dependencyBlockedPendingUnits(pending, completedSources, failedSources);
      const readyBlocked = readyPendingUnits(pending, completedSources, failedSources).filter(
        (unit) => !hasResourceCapacity(unit, schedule.resourceLimits, activeResourceClaims),
      );
      const blockedResources =
        readyBlocked.length > 0
          ? blockedResourcesForPending(readyBlocked, schedule.resourceLimits, activeResourceClaims)
          : [];
      const waitingOn = schedulerWaitingOnForUnits(dependencyBlocked, completedSources);
      const blockedUnits = schedulerBlockedUnitRecords({
        dependencyBlocked,
        resourceBlocked: readyBlocked,
        completed: completedSources,
        blockedResourcesForUnit: (unit) =>
          blockedResourcesForUnit(unit, schedule.resourceLimits, activeResourceClaims),
      });
      const blockedCount = dependencyBlocked.length + readyBlocked.length;
      if (blockedCount > 0 && (running.size > 0 || runningFinalizers.size > 0)) {
        let reason = "none";
        if (dependencyBlocked.length > 0 && blockedResources.length > 0) {
          reason = "dependencies,resources";
        } else if (dependencyBlocked.length > 0) {
          reason = "dependencies";
        } else if (blockedResources.length > 0) {
          reason = "resources";
        }
        reporter.blocked({ ...stateSnapshot(), blockedCount }, reason, blockedResources, {
          waitingOn,
          blockedUnits,
        });
      } else {
        reporter.maybeProgress(stateSnapshot());
      }

      const waitables = [...running.keys(), ...runningFinalizers.keys()];
      if (waitables.length === 0) {
        if (pending.length === 0 && pendingFinalizers.length === 0) {
          break;
        }
        throw new Error(
          `scheduler deadlock for ${schedule.target}; pending=${formatBlockedWorkUnits(pending)} pending_finalizers=${pendingFinalizers.map((finalizer) => finalizer.target).join(",")} completed_shards=${Array.from(completedShards).sort().join(",")} active_resource_claims=${formatActiveResourceClaims(activeResourceClaims)} resource_limits=${formatResourceLimits(schedule.resourceLimits)}`,
        );
      }

      const progressDelay = schedulerProgressDelay();
      const result = await Promise.race([...waitables, progressDelay.promise]);
      if (result?.schedulerProgressTick === true) {
        const dependencyBlockedNow = dependencyBlockedPendingUnits(pending, completedSources, failedSources);
        const readyBlockedNow = readyPendingUnits(pending, completedSources, failedSources).filter(
          (unit) => !hasResourceCapacity(unit, schedule.resourceLimits, activeResourceClaims),
        );
        const blockedResourcesNow =
          readyBlockedNow.length > 0
            ? blockedResourcesForPending(readyBlockedNow, schedule.resourceLimits, activeResourceClaims)
            : [];
        const waitingOnNow = schedulerWaitingOnForUnits(dependencyBlockedNow, completedSources);
        const blockedUnitsNow = schedulerBlockedUnitRecords({
          dependencyBlocked: dependencyBlockedNow,
          resourceBlocked: readyBlockedNow,
          completed: completedSources,
          blockedResourcesForUnit: (unit) =>
            blockedResourcesForUnit(unit, schedule.resourceLimits, activeResourceClaims),
        });
        let reason = "none";
        if (dependencyBlockedNow.length > 0 && blockedResourcesNow.length > 0) {
          reason = "dependencies,resources";
        } else if (dependencyBlockedNow.length > 0) {
          reason = "dependencies";
        } else if (blockedResourcesNow.length > 0) {
          reason = "resources";
        }
        reporter.maybeProgress(
          { ...stateSnapshot(), blockedCount: dependencyBlockedNow.length + readyBlockedNow.length },
          reason,
          blockedResourcesNow,
          {
            force: true,
            waitingOn: waitingOnNow,
            blockedUnits: blockedUnitsNow,
          },
        );
        continue;
      }
      progressDelay.cancel();
      let finishedUnit = null;
      for (const [promise, candidate] of running.entries()) {
        if (candidate.id === result.id) {
          running.delete(promise);
          removeResourceClaims(candidate, activeResourceClaims);
          finishedUnit = candidate;
          break;
        }
      }
      if (finishedUnit) {
        reporter.finishUnit(finishedUnit, result, stateSnapshot());
        if (result.status !== 0 && firstFailure === 0) {
          firstFailure = result.status;
          firstFailureLabel = result.label;
        }
        if (finishedUnit.type === "make_target") {
          if (result.status === 0) {
            completedSources.add(finishedUnit.aggregateTarget);
          } else {
            markSourceFailed(finishedUnit.aggregateTarget, result.label);
          }
        }
        if (finishedUnit.type === "go_shard") {
          completedShards.add(finishedUnit.shard);
          startReadyFinalizers();
        }
        if (result.status !== 0 || reporter.verbose) {
          await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
        }
        continue;
      }

      for (const [promise, finalizer] of runningFinalizers.entries()) {
        if (finalizer.target === result.id.replace(/^finalize:/, "")) {
          runningFinalizers.delete(promise);
          reporter.finishFinalizer(finalizer, result, stateSnapshot());
          if (result.status !== 0 && !reporter.verbose) {
            for (const logFile of reporter.completedLogFilesForTarget(finalizer.target)) {
              await replayLog(logFile, process.stderr);
            }
          }
          if (result.status !== 0 || reporter.verbose) {
            await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
          }
          if (result.status !== 0 && firstFailure === 0) {
            firstFailure = result.status;
            firstFailureLabel = result.label;
          }
          if (result.status === 0) {
            completedSources.add(finalizer.target);
          } else {
            markSourceFailed(finalizer.target, result.label);
          }
          break;
        }
      }
    }

    if (firstFailure === 0 && backendBudgetTargets.length > 0) {
      firstFailure = await runPostgresFixtureBudgetCheck(backendBudgetTargets);
      if (firstFailure !== 0) {
        firstFailureLabel = "postgres-fixture-budget";
      }
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    await reporter.summary(requestedStatus, { started, failedWorkUnit: firstFailureLabel });
    if (!deferSummary) {
      await runLifecycle(
        testOutputScript,
        ["target-summary", schedule.target, requestedStatus, "--projection", schedule.target],
        requestedStatus === "pass" ? process.stdout : process.stderr,
      );
    }
    return firstFailure;
  } finally {
    await reporter.close();
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const { manifest, manifestPath } = await loadManifest(options.manifest);
  const browserStages = await loadBrowserBatchStages();
  const schedule = expandSchedule(findSchedule(manifest, options.target, browserStages));
  const makeBin = process.env.MAKE || "make";
  const testOutputScript =
    process.env.TEST_OUTPUT_SCRIPT || path.join(repoRoot, "scripts", "lib", "test-output.sh");

  if (isDryRun()) {
    writeDryRun(schedule, manifestPath, options.target);
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
