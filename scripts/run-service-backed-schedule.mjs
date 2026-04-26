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
const supportedSchemaID = "cartulary.service_backed_schedule.v3";
const validSourceTypes = new Set(["go_shards", "make_target"]);

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
      throw new Error("--jobs is obsolete for v3 service-backed schedules; use resource_limits");
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
    if (!Number.isInteger(limit) || limit < 1) {
      throw new Error(`${label} resource_limits.${normalizedResource} must be a positive integer`);
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
  if (source.resource_tags !== undefined || source.exclusive_tags !== undefined) {
    throw new Error(`${label} must not declare legacy resource_tags or exclusive_tags`);
  }
  const target = source.target.trim();
  const resourceClaims = normalizeResourceClaims(
    source.resource_claims,
    `${label} ${target}`,
    resourceLimits,
  );
  validateBackendTarget(scheduleTarget, target, label);

  if (source.type === "go_shards") {
    return {
      type: source.type,
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
  };
}

function cloneResourceClaims(resourceClaims) {
  return new Map(resourceClaims.entries());
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
        target: source.target,
        aggregateTarget: source.target,
        weight: source.weight,
        resourceClaims: cloneResourceClaims(source.resourceClaims),
        order: source.order,
      });
      continue;
    }

    goFinalizers.push(source.target);
    const shards = collectGoShardsForTarget(repoRoot, source.target);
    if (shards.length === 0) {
      throw new Error(`go_shards source ${source.target} selected no shards`);
    }
    for (const shard of shards) {
      if (shardWorkByName.has(shard.name)) {
        continue;
      }
      const unit = {
        id: `${source.target}:${shard.name}`,
        label: `${source.target}/${shard.name}`,
        type: "go_shard",
        target: source.target,
        aggregateTarget: source.target,
        shard: shard.name,
        weight: shard.weight_ms,
        resourceClaims: cloneResourceClaims(source.resourceClaims),
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
    workUnits,
    goFinalizers,
  };
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

function unitCommand({ makeBin, unit, metadataDir }) {
  if (unit.type === "make_target") {
    return {
      command: makeBin,
      args: ["--no-print-directory", "--output-sync=target", unit.target],
      env: process.env,
    };
  }
  return {
    command: path.join(repoRoot, "scripts", "run-go-target.sh"),
    args: ["capture-shard", unit.target, unit.shard, metadataDir],
    env: {
      ...process.env,
      CARTULARY_TEST_TARGET: unit.target,
    },
  };
}

function runWorkUnit({ makeBin, unit, metadataDir, tempDir, started }) {
  const { command, args, env } = unitCommand({ makeBin, unit, metadataDir });
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

function runFinalizer({ target, metadataDir, tempDir, testOutputScript }) {
  const logFile = path.join(tempDir, `finalize-${sanitizeLogName(target)}.log`);
  return runCommand(
    path.join(repoRoot, "scripts", "run-go-target.sh"),
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
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const limit = resourceLimits.get(resource);
    if (limit !== undefined && (activeResourceClaims.get(resource) ?? 0) + amount > limit) {
      return false;
    }
  }
  return true;
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
  return schedule.resourceLimits.get("backend") ?? Math.max(...schedule.resourceLimits.values());
}

async function runSchedule({ schedule, makeBin, testOutputScript, deferSummary }) {
  const childrenCsv = schedule.children.join(",");
  const backendBudgetTargets = schedule.children;
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-service-backed-schedule-"));
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  const pending = [...schedule.workUnits];
  const running = new Map();
  const activeResourceClaims = new Map();
  const capacityDisplay = displayCapacity(schedule);
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
      const promise = runWorkUnit({ makeBin, unit, metadataDir, tempDir, started });
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

    while (pending.length > 0 || running.size > 0) {
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
          ...schedulerStateFields({
            pending,
            running,
            activeResourceClaims,
            resourceLimits: schedule.resourceLimits,
          }),
        ]);
      }

      if (running.size === 0) {
        throw new Error(
          `scheduler deadlock for ${schedule.target}; pending=${formatBlockedWorkUnits(pending)} active_resource_claims=${formatActiveResourceClaims(activeResourceClaims)} resource_limits=${formatResourceLimits(schedule.resourceLimits)}`,
        );
      }

      const result = await Promise.race(running.keys());
      for (const [promise, candidate] of running.entries()) {
        if (candidate.id === result.id) {
          running.delete(promise);
          removeResourceClaims(candidate, activeResourceClaims);
          break;
        }
      }
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
      await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
      if (result.status !== 0 && firstFailure === 0) {
        firstFailure = result.status;
      }
    }

    for (const target of schedule.goFinalizers) {
      const result = await runFinalizer({ target, metadataDir, tempDir, testOutputScript });
      schedulerTelemetry(schedule, "finalize", [
        `target=${target}`,
        `status=${result.status}`,
      ]);
      await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
      if (result.status !== 0 && firstFailure === 0) {
        firstFailure = result.status;
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
      `[DRY-RUN] ${options.target} manifest=${path.relative(repoRoot, manifestPath)} work_units=${schedule.workUnits.map((unit) => unit.label).join(",")}\n`,
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
