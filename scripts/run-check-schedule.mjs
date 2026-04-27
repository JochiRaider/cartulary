#!/usr/bin/env node
import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { loadTaskSurfaceManifest, summaryProfileArgs } from "./lib/task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "check_schedule_manifest.json");
const supportedSchemaID = "cartulary.check_schedule.v1";

function usage() {
  process.stderr.write(
    "usage: run-check-schedule.mjs --target <target> (--summary-profile <name> | --summary-targets <a,b>) [--summary-groups <spec>] [--manifest <path>] [--resource-limit <name=value>...]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
    summaryProfile: "",
    summaryTargets: "",
    summaryGroups: "",
    resourceLimitOverrides: new Map(),
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--summary-targets") {
      options.summaryTargets = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--summary-profile") {
      options.summaryProfile = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--summary-groups") {
      options.summaryGroups = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--resource-limit") {
      const value = argv[index + 1] ?? "";
      const [resource, amountText, extra] = value.split("=");
      if (!resource || !amountText || extra !== undefined) {
        throw new Error(`--resource-limit expects <name=value>, got ${value}`);
      }
      const amount = Number.parseInt(amountText, 10);
      if (!Number.isInteger(amount) || amount < 1) {
        throw new Error(`--resource-limit ${resource} must be a positive integer`);
      }
      options.resourceLimitOverrides.set(resource.trim(), amount);
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.target || !options.manifest) {
    usage();
  }
  if (options.summaryProfile && options.summaryTargets) {
    throw new Error("--summary-profile and --summary-targets are mutually exclusive");
  }
  if (!options.summaryProfile && !options.summaryTargets) {
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

function normalizeResourceLimits(value, label, overrides) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_limits must be an object`);
  }
  const limits = new Map();
  for (const [resource, amount] of Object.entries(value)) {
    const normalizedResource = resource.trim();
    if (normalizedResource === "") {
      throw new Error(`${label} resource_limits keys must be non-empty strings`);
    }
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`${label} resource_limits.${normalizedResource} must be a positive integer`);
    }
    limits.set(normalizedResource, amount);
  }
  for (const [resource, amount] of overrides.entries()) {
    if (!limits.has(resource)) {
      throw new Error(`${label} resource limit override ${resource} is not declared`);
    }
    limits.set(resource, amount);
  }
  return limits;
}

function normalizeResourceClaims(value, label, resourceLimits) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} resource_claims must be an object`);
  }
  const claims = new Map();
  for (const [resource, rawAmount] of Object.entries(value)) {
    const normalizedResource = resource.trim();
    if (normalizedResource === "") {
      throw new Error(`${label} resource_claims keys must be non-empty strings`);
    }
    if (!resourceLimits.has(normalizedResource)) {
      throw new Error(`${label} resource_claims entry ${normalizedResource} is not declared in resource_limits`);
    }
    const amount = rawAmount === "limit" ? resourceLimits.get(normalizedResource) : rawAmount;
    if (!Number.isInteger(amount) || amount < 1) {
      throw new Error(`${label} resource_claims.${normalizedResource} must be a positive integer or \"limit\"`);
    }
    if (amount > resourceLimits.get(normalizedResource)) {
      throw new Error(`${label} resource_claims.${normalizedResource} exceeds resource limit`);
    }
    claims.set(normalizedResource, amount);
  }
  return claims;
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

function normalizeMakeJobs(value, label, resourceClaims) {
  if (value === undefined) {
    return 1;
  }
  if (value === "cpu") {
    return resourceClaims.get("cpu") ?? 1;
  }
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} make_jobs must be a positive integer or \"cpu\"`);
  }
  return value;
}

function findSchedule(manifest, target, overrides) {
  const matches = manifest.schedules.filter((schedule) => schedule?.target === target);
  if (matches.length !== 1) {
    throw new Error(`expected exactly one check schedule for ${target}, found ${matches.length}`);
  }
  const [schedule] = matches;
  if (!Array.isArray(schedule.work_units) || schedule.work_units.length === 0) {
    throw new Error(`check schedule ${target} must declare work_units[]`);
  }
  const resourceLimits = normalizeResourceLimits(schedule.resource_limits, `check schedule ${target}`, overrides);
  const units = schedule.work_units.map((unit, index) => {
    const label = `check schedule ${target} work_units ${index + 1}`;
    if (!unit || typeof unit !== "object" || Array.isArray(unit)) {
      throw new Error(`${label} must be an object`);
    }
    if (typeof unit.target !== "string" || unit.target.trim() === "") {
      throw new Error(`${label} must declare target`);
    }
    if (!Number.isFinite(unit.weight) || unit.weight < 0) {
      throw new Error(`${label} ${unit.target} must declare non-negative weight`);
    }
    const claims = normalizeResourceClaims(unit.resource_claims, `${label} ${unit.target}`, resourceLimits);
    return {
      target: unit.target.trim(),
      label: unit.target.trim(),
      weight: unit.weight,
      needs: normalizeNeeds(unit.needs, `${label} ${unit.target}`),
      resourceClaims: claims,
      makeJobs: normalizeMakeJobs(unit.make_jobs, `${label} ${unit.target}`, claims),
      order: index,
    };
  });
  const unitTargets = new Set();
  for (const unit of units) {
    if (unitTargets.has(unit.target)) {
      throw new Error(`check schedule ${target} contains duplicate work unit target ${unit.target}`);
    }
    unitTargets.add(unit.target);
  }
  for (const unit of units) {
    for (const need of unit.needs) {
      if (!unitTargets.has(need)) {
        throw new Error(`check schedule ${target} work unit ${unit.target} depends on unknown target ${need}`);
      }
      if (need === unit.target) {
        throw new Error(`check schedule ${target} work unit ${unit.target} cannot depend on itself`);
      }
    }
  }
  assertAcyclic(target, units);
  return {
    target,
    resourceLimits,
    units: units.sort((left, right) => right.weight - left.weight || left.order - right.order || left.target.localeCompare(right.target)),
  };
}

function assertAcyclic(target, units) {
  const byTarget = new Map(units.map((unit) => [unit.target, unit]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (unit) => {
    if (visited.has(unit.target)) {
      return;
    }
    if (visiting.has(unit.target)) {
      throw new Error(`check schedule ${target} has a dependency cycle at ${unit.target}`);
    }
    visiting.add(unit.target);
    for (const need of unit.needs) {
      visit(byTarget.get(need));
    }
    visiting.delete(unit.target);
    visited.add(unit.target);
  };
  for (const unit of units) {
    visit(unit);
  }
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

async function replayLog(file, stream) {
  await new Promise((resolve, reject) => {
    const reader = createReadStream(file);
    reader.on("error", reject);
    reader.on("end", resolve);
    reader.pipe(stream, { end: false });
  });
}

function hasResourceCapacity(unit, resourceLimits, activeClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    if ((activeClaims.get(resource) ?? 0) + amount > resourceLimits.get(resource)) {
      return false;
    }
  }
  return true;
}

function addResourceClaims(unit, activeClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    activeClaims.set(resource, (activeClaims.get(resource) ?? 0) + amount);
  }
}

function removeResourceClaims(unit, activeClaims) {
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const next = (activeClaims.get(resource) ?? 0) - amount;
    if (next <= 0) {
      activeClaims.delete(resource);
    } else {
      activeClaims.set(resource, next);
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

function schedulerTelemetry(schedule, event, fields) {
  process.stdout.write(`[CHECK-SCHEDULER] ${schedule.target} ${event} ${fields.join(" ")}\n`);
}

function schedulerStateFields({ pending, running, activeClaims, resourceLimits }) {
  return [
    `active=${running.size}`,
    `pending=${pending.length}`,
    `active_resource_claims=${formatResourceMap(activeClaims)}`,
    `resource_limits=${formatResourceMap(resourceLimits)}`,
  ];
}

function readyPendingUnits(pending, completed) {
  return pending.filter((unit) => unit.needs.every((need) => completed.has(need)));
}

function blockedPendingUnits(pending, completed) {
  return pending.filter((unit) => !unit.needs.every((need) => completed.has(need)));
}

function runWorkUnit({ makeBin, unit, tempDir, started }) {
  const logFile = path.join(tempDir, `${String(started).padStart(2, "0")}-${sanitizeLogName(unit.target)}.log`);
  return runCommand(makeBin, ["--no-print-directory", "--output-sync=target", `-j${unit.makeJobs}`, unit.target], logFile).then((result) => ({
    id: unit.target,
    label: unit.label,
    status: result.status,
    logFile,
  }));
}

async function runSchedule({ schedule, makeBin, testOutputScript, summaryTargets, summaryGroups }) {
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-check-schedule-"));
  const pending = [...schedule.units];
  const running = new Map();
  const completed = new Set();
  const activeClaims = new Map();
  const totalUnits = schedule.units.length;
  const capacityDisplay = schedule.resourceLimits.get("cpu") ?? Math.max(...schedule.resourceLimits.values());
  let started = 0;
  let completedCount = 0;
  let firstFailure = 0;
  let firstFailureTarget = "-";
  let stopScheduling = false;

  try {
    await runLifecycle(testOutputScript, [
      "run-start",
      schedule.target,
      "--steps",
      String(totalUnits),
      "--targets",
      String(summaryTargets.length),
      "--jobs",
      String(capacityDisplay),
    ]);

    const startUnit = async (unit) => {
      started += 1;
      await runLifecycle(testOutputScript, [
        "step-start",
        schedule.target,
        String(started),
        String(totalUnits),
        unit.label,
        "--mode",
        "scheduler",
        "--jobs",
        String(unit.makeJobs),
      ]);
      addResourceClaims(unit, activeClaims);
      const promise = runWorkUnit({ makeBin, unit, tempDir, started });
      running.set(promise, unit);
      schedulerTelemetry(schedule, "start", [
        `work_unit=${unit.label}`,
        `claims=${formatResourceMap(unit.resourceClaims)}`,
        ...schedulerStateFields({ pending, running, activeClaims, resourceLimits: schedule.resourceLimits }),
      ]);
    };

    while (pending.length > 0 || running.size > 0) {
      if (!stopScheduling) {
        while (true) {
          const ready = readyPendingUnits(pending, completed);
          const next = ready.find((candidate) => hasResourceCapacity(candidate, schedule.resourceLimits, activeClaims));
          if (!next) {
            break;
          }
          pending.splice(pending.indexOf(next), 1);
          await startUnit(next);
        }
      }

      if (pending.length > 0 && running.size > 0 && !stopScheduling) {
        const blockedByDeps = blockedPendingUnits(pending, completed).length;
        const reason = blockedByDeps === pending.length ? "dependencies" : "resources";
        schedulerTelemetry(schedule, "blocked", [
          `reason=${reason}`,
          ...schedulerStateFields({ pending, running, activeClaims, resourceLimits: schedule.resourceLimits }),
        ]);
      }

      if (running.size === 0) {
        if (stopScheduling) {
          break;
        }
        throw new Error(`check scheduler deadlock for ${schedule.target}; pending=${pending.map((unit) => unit.target).join(",")}`);
      }

      const result = await Promise.race(running.keys());
      let finishedUnit;
      for (const [promise, candidate] of running.entries()) {
        if (candidate.target === result.id) {
          running.delete(promise);
          finishedUnit = candidate;
          removeResourceClaims(candidate, activeClaims);
          break;
        }
      }
      schedulerTelemetry(schedule, "finish", [
        `work_unit=${result.label}`,
        `status=${result.status}`,
        ...schedulerStateFields({ pending, running, activeClaims, resourceLimits: schedule.resourceLimits }),
      ]);
      await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
      if (result.status === 0) {
        completed.add(result.id);
        completedCount += 1;
      } else if (firstFailure === 0) {
        firstFailure = result.status;
        firstFailureTarget = result.label;
        stopScheduling = true;
      }
      if (!finishedUnit) {
        throw new Error(`finished unknown check work unit ${result.id}`);
      }
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    const summaryArgs = ["run-summary", schedule.target, requestedStatus, String(completedCount), String(totalUnits), firstFailureTarget];
    if (summaryGroups) {
      summaryArgs.push("--summary-groups", summaryGroups);
    }
    summaryArgs.push(...summaryTargets);
    await runLifecycle(testOutputScript, summaryArgs, requestedStatus === "pass" ? process.stdout : process.stderr).catch((error) => {
      if (firstFailure === 0) {
        throw error;
      }
    });
    return firstFailure;
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

function parseSummaryTargets(value) {
  return value
    .split(",")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const { manifest, manifestPath } = await loadManifest(options.manifest);
  const schedule = findSchedule(manifest, options.target, options.resourceLimitOverrides);
  let summaryTargets = parseSummaryTargets(options.summaryTargets);
  let summaryGroups = options.summaryGroups;
  if (options.summaryProfile) {
    const { manifest: taskSurface } = loadTaskSurfaceManifest(
      process.env.TASK_SURFACE_MANIFEST ?? path.join(repoRoot, "tools", "task_surface_manifest.json"),
    );
    const profile = summaryProfileArgs(taskSurface, options.summaryProfile);
    summaryTargets = profile.targets;
    summaryGroups = profile.groupsSpec;
  }
  if (summaryTargets.length === 0) {
    throw new Error("summary profile must select at least one target");
  }
  const makeBin = process.env.MAKE || "make";
  const testOutputScript =
    process.env.TEST_OUTPUT_SCRIPT || path.join(repoRoot, "scripts", "lib", "test-output.sh");

  if (isDryRun()) {
    process.stdout.write(
      `[DRY-RUN] ${options.target} manifest=${path.relative(repoRoot, manifestPath)} work_units=${schedule.units.map((unit) => unit.target).join(",")}\n`,
    );
    return;
  }

  const status = await runSchedule({
    schedule,
    makeBin,
    testOutputScript,
    summaryTargets,
    summaryGroups,
  });
  process.exitCode = status;
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
});
