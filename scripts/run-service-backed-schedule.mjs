#!/usr/bin/env node
import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectGoShardsForTarget } from "./lib/go-shard-plan.mjs";
import { findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const supportedSchemaID = "cartulary.service_backed_schedule.v5";
const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const goCPULimitEnv = "CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT";
const goIOLimitEnv = "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const autoLimitResources = new Set([goCPUResource, goIOResource]);
const validSourceTypes = new Set(["go_shards", "make_target"]);
const validSourceClasses = new Set(["backend", "browser"]);
const schedulerSafeBrowserTargetsBySchedule = new Map([
  ["test-service-backed", new Map([["browser-e2e-webserver-backed", "webserver-backed"]])],
  [
    "check-service-backed",
    new Map([
      ["browser-e2e-webserver-backed", "webserver-backed"],
      ["browser-e2e", "isolated"],
    ]),
  ],
]);

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
      throw new Error("--jobs is obsolete for v5 service-backed schedules; use resource_limits");
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
    if (limit === "auto") {
      if (!autoLimitResources.has(normalizedResource)) {
        throw new Error(
          `${label} resource_limits.${normalizedResource} may use "auto" only for ${goCPUResource} or ${goIOResource}`,
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

function validateBrowserTarget(scheduleTarget, source, target, label) {
  if (source.type !== "make_target") {
    throw new Error(`${label} browser target ${target} must use type make_target`);
  }
  const schedulerSafeBrowserTargets = schedulerSafeBrowserTargetsBySchedule.get(scheduleTarget);
  const expectedStage = schedulerSafeBrowserTargets?.get(target);
  if (!expectedStage) {
    throw new Error(`${label} browser target ${target} is not scheduler-safe for ${scheduleTarget}`);
  }
  if (source.browser_stage !== expectedStage) {
    throw new Error(`${label} browser target ${target} must declare browser_stage ${expectedStage}`);
  }
}

function validateSource(scheduleTarget, source, index, resourceLimits) {
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
  if (source.class === "backend") {
    validateBackendTarget(scheduleTarget, target, label);
  } else {
    validateBrowserTarget(scheduleTarget, source, target, label);
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
    weight: source.weight,
    resourceClaims,
    order: index,
  };
}

function findSchedule(manifest, target) {
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
    validateSource(target, source, index, resourceLimits),
  );
  const duplicateTargets = sources
    .map((source) => source.target)
    .filter((targetName, index, targets) => targets.indexOf(targetName) !== index);
  if (duplicateTargets.length > 0) {
    throw new Error(
      `schedule ${target} contains duplicate work-unit source targets: ${duplicateTargets.join(", ")}`,
    );
  }
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

function schedulerClaimsForShard(shard) {
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
        shard: shard.name,
        schedulerProfile: shard.scheduler_profile,
        weight: shard.weight_ms,
        resourceClaims: mergeResourceClaims(source.resourceClaims, schedulerClaimsForShard(shard)),
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
  const cpuHeavy = goShardUnits.filter((unit) => unit.schedulerProfile === "cpu_heavy").length;
  const profileConcurrency = balanced + ioHeavy * 2 + Math.ceil(cpuHeavy / 2);
  return clampInteger(Math.max(6, goCPULimit + 2, profileConcurrency), 6, 16);
}

function resolveResourceLimits(resourceLimits, workUnits) {
  const goShardUnits = workUnits.filter((unit) => unit.type === "go_shard");
  const computedGoCPU = estimateGoCPULimit(goShardUnits);
  const goCPUOverride = parsePositiveIntegerEnv(goCPULimitEnv);
  const effectiveGoCPU = goCPUOverride ?? computedGoCPU;
  const computedGoIO = estimateGoIOLimit(goShardUnits, effectiveGoCPU);
  const goIOOverride = parsePositiveIntegerEnv(goIOLimitEnv);
  const effectiveGoIO = goIOOverride ?? computedGoIO;
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

function unitCommand({ makeBin, unit, metadataDir, goTargetRunner }) {
  if (unit.type === "make_target") {
    return {
      command: makeBin,
      args: ["--no-print-directory", "--output-sync=target", unit.target],
      env: process.env,
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

function runWorkUnit({ makeBin, unit, metadataDir, tempDir, started, goTargetRunner }) {
  const { command, args, env } = unitCommand({ makeBin, unit, metadataDir, goTargetRunner });
  const logFile = path.join(
    tempDir,
    `${String(started).padStart(2, "0")}-${sanitizeLogName(unit.id)}.log`,
  );
  return runCommand(command, args, logFile, env).then((result) => ({
    id: unit.id,
    label: unit.label,
    status: result.status,
    logFile,
  }));
}

function runFinalizer({ target, metadataDir, tempDir, testOutputScript, goTargetRunner }) {
  const logFile = path.join(tempDir, `finalize-${sanitizeLogName(target)}.log`);
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

function formatResourceMap(values) {
  const entries = Array.from(values.entries()).sort((left, right) => left[0].localeCompare(right[0]));
  if (entries.length === 0) {
    return "{}";
  }
  return `{${entries.map(([key, value]) => `${key}:${value}`).join(",")}}`;
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

function formatResourceList(values) {
  if (values.length === 0) {
    return "none";
  }
  return values.join(",");
}

function schedulerTelemetry(schedule, event, fields) {
  process.stdout.write(`[SCHEDULER] ${schedule.target} ${event} ${fields.join(" ")}\n`);
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

async function runSchedule({ schedule, makeBin, testOutputScript, deferSummary }) {
  const childrenCsv = schedule.children.join(",");
  const backendBudgetTargets = schedule.backendChildren;
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-service-backed-schedule-"));
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  const pending = [...schedule.workUnits];
  const pendingFinalizers = schedule.goFinalizers.map((finalizer) => ({
    target: finalizer.target,
    shardNames: [...finalizer.shardNames],
  }));
  const running = new Map();
  const runningFinalizers = new Map();
  const completedShards = new Set();
  const activeResourceClaims = new Map();
  const capacityDisplay = displayCapacity(schedule);
  const goTargetRunner =
    process.env[goTargetRunnerEnv] || path.join(repoRoot, "scripts", "run-go-target.sh");
  let started = 0;
  let firstFailure = 0;

  try {
    await runLifecycle(testOutputScript, [
      "target-start",
      schedule.target,
      "--children",
      childrenCsv,
      "--service-backed",
      "1",
    ]);

    const startUnit = async (unit) => {
      started += 1;
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
      addResourceClaims(unit, activeResourceClaims);
      const promise = runWorkUnit({
        makeBin,
        unit,
        metadataDir,
        tempDir,
        started,
        goTargetRunner,
      });
      running.set(promise, unit);
      schedulerTelemetry(schedule, "start", [
        `work_unit=${unit.label}`,
        `claims=${formatResourceClaims(unit.resourceClaims)}`,
        ...schedulerStateFields({
          pending,
          running,
          activeResourceClaims,
          resourceLimits: schedule.resourceLimits,
        }),
      ]);
    };

    const startReadyFinalizers = () => {
      for (let index = 0; index < pendingFinalizers.length; ) {
        const finalizer = pendingFinalizers[index];
        if (!finalizerIsReady(finalizer, completedShards)) {
          index += 1;
          continue;
        }
        pendingFinalizers.splice(index, 1);
        const promise = runFinalizer({
          target: finalizer.target,
          metadataDir,
          tempDir,
          testOutputScript,
          goTargetRunner,
        });
        runningFinalizers.set(promise, finalizer);
        schedulerTelemetry(schedule, "finalize-start", [
          `target=${finalizer.target}`,
          `shards=${finalizer.shardNames.length}`,
          `active_finalizers=${runningFinalizers.size}`,
          `pending_finalizers=${pendingFinalizers.length}`,
        ]);
      }
    };

    while (
      pending.length > 0 ||
      running.size > 0 ||
      pendingFinalizers.length > 0 ||
      runningFinalizers.size > 0
    ) {
      while (true) {
        const nextIndex = pending.findIndex((candidate) =>
          hasResourceCapacity(candidate, schedule.resourceLimits, activeResourceClaims),
        );
        if (nextIndex === -1) {
          break;
        }
        const [unit] = pending.splice(nextIndex, 1);
        await startUnit(unit);
      }

      if (pending.length > 0 && running.size > 0) {
        schedulerTelemetry(schedule, "blocked", [
          "reason=resources",
          `blocked_resources=${formatResourceList(
            blockedResourcesForPending(pending, schedule.resourceLimits, activeResourceClaims),
          )}`,
          ...schedulerStateFields({
            pending,
            running,
            activeResourceClaims,
            resourceLimits: schedule.resourceLimits,
          }),
        ]);
      }

      const waitables = [...running.keys(), ...runningFinalizers.keys()];
      if (waitables.length === 0) {
        throw new Error(
          `scheduler deadlock for ${schedule.target}; pending=${formatBlockedWorkUnits(pending)} pending_finalizers=${pendingFinalizers.map((finalizer) => finalizer.target).join(",")} completed_shards=${Array.from(completedShards).sort().join(",")} active_resource_claims=${formatActiveResourceClaims(activeResourceClaims)} resource_limits=${formatResourceLimits(schedule.resourceLimits)}`,
        );
      }

      const result = await Promise.race(waitables);
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
        schedulerTelemetry(schedule, "finish", [
          `work_unit=${result.label}`,
          `status=${result.status}`,
          ...schedulerStateFields({
            pending,
            running,
            activeResourceClaims,
            resourceLimits: schedule.resourceLimits,
          }),
        ]);
        if (result.status !== 0 && firstFailure === 0) {
          firstFailure = result.status;
        }
        if (finishedUnit.type === "go_shard") {
          completedShards.add(finishedUnit.shard);
          startReadyFinalizers();
        }
        await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
        continue;
      }

      for (const [promise, finalizer] of runningFinalizers.entries()) {
        if (finalizer.target === result.id.replace(/^finalize:/, "")) {
          runningFinalizers.delete(promise);
          schedulerTelemetry(schedule, "finalize-finish", [
            `target=${finalizer.target}`,
            `status=${result.status}`,
            `active_finalizers=${runningFinalizers.size}`,
            `pending_finalizers=${pendingFinalizers.length}`,
          ]);
          await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
          if (result.status !== 0 && firstFailure === 0) {
            firstFailure = result.status;
          }
          break;
        }
      }
    }

    if (firstFailure === 0 && backendBudgetTargets.length > 0) {
      firstFailure = await runPostgresFixtureBudgetCheck(backendBudgetTargets);
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    if (!deferSummary) {
      await runLifecycle(
        testOutputScript,
        ["target-summary", schedule.target, requestedStatus, "--children", childrenCsv],
        requestedStatus === "pass" ? process.stdout : process.stderr,
      );
    }
    return firstFailure;
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const { manifest, manifestPath } = await loadManifest(options.manifest);
  const schedule = expandSchedule(findSchedule(manifest, options.target));
  const makeBin = process.env.MAKE || "make";
  const testOutputScript =
    process.env.TEST_OUTPUT_SCRIPT || path.join(repoRoot, "scripts", "lib", "test-output.sh");

  if (isDryRun()) {
    process.stdout.write(
      `[DRY-RUN] ${options.target} manifest=${path.relative(repoRoot, manifestPath)} resource_limits=${formatResourceLimits(schedule.resourceLimits)} work_units=${schedule.workUnits
        .map((unit) => {
          const profile = unit.schedulerProfile ? ` profile=${unit.schedulerProfile}` : "";
          return `${unit.label}${profile} claims=${formatResourceClaims(unit.resourceClaims)}`;
        })
        .join(";")}\n`,
    );
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
