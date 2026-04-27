#!/usr/bin/env node
import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  formatResourceList,
  formatResourceMap,
  relToRepo as relToRepoPath,
  resourceMapToObject,
  schedulerActiveGroups,
  schedulerBlockedBy,
  schedulerBlockedUnitRecords,
  schedulerDryRunLine,
  schedulerProgressIntervalMs,
  schedulerProgressLine,
  schedulerSlowestRunning,
  schedulerStartLine,
  schedulerSummaryLine,
  schedulerWaitingOnForUnits,
  schedulerLogDir,
  schedulerTargetDir,
  writeSchedulerTelemetry,
  verboseSchedulerOutput,
} from "./lib/scheduler-reporting.mjs";
import { loadTaskSurfaceManifest, summaryProfileArgs } from "./lib/task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "check_schedule_manifest.json");
const supportedSchemaID = "cartulary.check_schedule.v2";
const schedulerEventSchemaID = "cartulary.check_scheduler_event.v2";
const schedulerSummarySchemaID = "cartulary.check_scheduler_summary.v2";

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

function normalizeTargetList(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} entries must be non-empty strings`);
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

function normalizeNestedScheduler(value, label, unitTarget, resourceClaims) {
  if (value === undefined) {
    return null;
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} nested_scheduler must be an object`);
  }
  if (value.type !== "service_backed") {
    throw new Error(`${label} nested_scheduler.type must be service_backed`);
  }
  if (typeof value.target !== "string" || value.target.trim() === "") {
    throw new Error(`${label} nested_scheduler.target must be a non-empty string`);
  }
  const nestedTarget = value.target.trim();
  if (nestedTarget !== unitTarget) {
    throw new Error(`${label} nested_scheduler.target must match work unit target ${unitTarget}`);
  }
  if (typeof value.manifest !== "string" || value.manifest.trim() === "") {
    throw new Error(`${label} nested_scheduler.manifest must be a non-empty string`);
  }
  if (
    !value.resource_limit_env ||
    typeof value.resource_limit_env !== "object" ||
    Array.isArray(value.resource_limit_env)
  ) {
    throw new Error(`${label} nested_scheduler.resource_limit_env must be an object`);
  }
  const resourceLimitEnv = new Map();
  const envNames = new Set();
  for (const [resource, envName] of Object.entries(value.resource_limit_env)) {
    const normalizedResource = resource.trim();
    if (normalizedResource === "") {
      throw new Error(`${label} nested_scheduler.resource_limit_env keys must be non-empty strings`);
    }
    if (!resourceClaims.has(normalizedResource)) {
      throw new Error(
        `${label} nested_scheduler.resource_limit_env.${normalizedResource} must map a resource claimed by ${unitTarget}`,
      );
    }
    if (typeof envName !== "string" || envName.trim() === "") {
      throw new Error(`${label} nested_scheduler.resource_limit_env.${normalizedResource} must be a non-empty string`);
    }
    const normalizedEnvName = envName.trim();
    if (envNames.has(normalizedEnvName)) {
      throw new Error(`${label} nested_scheduler.resource_limit_env must not map multiple resources to ${normalizedEnvName}`);
    }
    envNames.add(normalizedEnvName);
    resourceLimitEnv.set(normalizedResource, normalizedEnvName);
  }
  if (resourceLimitEnv.size === 0) {
    throw new Error(`${label} nested_scheduler.resource_limit_env must not be empty`);
  }
  return {
    type: value.type,
    target: nestedTarget,
    manifest: value.manifest.trim(),
    resourceLimitEnv,
  };
}

function nestedSchedulerForwardedLimits(unit) {
  if (!unit.nestedScheduler) {
    return new Map();
  }
  const limits = new Map();
  for (const [resource, envName] of unit.nestedScheduler.resourceLimitEnv.entries()) {
    limits.set(envName, unit.resourceClaims.get(resource));
  }
  return limits;
}

function nestedSchedulerEnv(unit) {
  if (!unit.nestedScheduler) {
    return process.env;
  }
  const forwarded = Object.fromEntries(
    Array.from(nestedSchedulerForwardedLimits(unit).entries()).map(([envName, amount]) => [
      envName,
      String(amount),
    ]),
  );
  return {
    ...process.env,
    ...forwarded,
  };
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

function nestedSchedulerDetail(unit) {
  if (!unit.nestedScheduler) {
    return null;
  }
  return {
    type: unit.nestedScheduler.type,
    target: unit.nestedScheduler.target,
    manifest: unit.nestedScheduler.manifest,
    resource_limit_env: resourceMapToObject(unit.nestedScheduler.resourceLimitEnv),
    forwarded_limits: resourceMapToObject(nestedSchedulerForwardedLimits(unit)),
  };
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
    const unitTarget = unit.target.trim();
    const nestedScheduler = normalizeNestedScheduler(
      unit.nested_scheduler,
      `${label} ${unitTarget}`,
      unitTarget,
      claims,
    );
    return {
      target: unitTarget,
      label: unitTarget,
      weight: unit.weight,
      needs: normalizeNeeds(unit.needs, `${label} ${unitTarget}`),
      skippedSummaryTargets: normalizeTargetList(
        unit.skipped_summary_targets,
        `${label} ${unitTarget} skipped_summary_targets`,
      ),
      resourceClaims: claims,
      makeJobs: normalizeMakeJobs(unit.make_jobs, `${label} ${unitTarget}`, claims),
      nestedScheduler,
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

function schedulerStateFields({ pending, running, activeClaims, resourceLimits }) {
  return [
    `active=${running.size}`,
    `pending=${pending.length}`,
    `active_resource_claims=${formatResourceMap(activeClaims)}`,
    `resource_limits=${formatResourceMap(resourceLimits)}`,
  ];
}

function blockedResourcesForUnit(unit, resourceLimits, activeClaims) {
  const blocked = [];
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    if ((activeClaims.get(resource) ?? 0) + amount > resourceLimits.get(resource)) {
      blocked.push(resource);
    }
  }
  return blocked.sort((left, right) => left.localeCompare(right));
}

function blockedResourcesForUnits(units, resourceLimits, activeClaims) {
  const resources = new Set();
  for (const unit of units) {
    for (const resource of blockedResourcesForUnit(unit, resourceLimits, activeClaims)) {
      resources.add(resource);
    }
  }
  return Array.from(resources).sort((left, right) => left.localeCompare(right));
}

function relToRepo(value) {
  return relToRepoPath(repoRoot, value);
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

async function createCheckSchedulerReporter(schedule) {
  const targetDir = schedulerTargetDir(repoRoot, schedule.target);
  const logDir = schedulerLogDir(repoRoot, schedule.target);
  await mkdir(logDir, { recursive: true });
  return new CheckSchedulerReporter(schedule, targetDir, logDir);
}

class CheckSchedulerReporter {
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
    this.skippedWork = [];
    this.completedCount = 0;
    this.failedWorkUnit = null;
    this.blockedReasonsSeen = new Set();
    this.blockedResourcesSeen = new Set();
    this.blockedExplanationsSeen = new Set();
    this.waitingOnSeen = new Set();
    this.lastProgressAt = 0;
    this.lastBlockedKey = null;
    this.maxRunningWorkUnits = 0;
    this.maxRunningGroups = 0;
    this.maxActiveResourceClaims = new Map();
  }

  start() {
    process.stdout.write(
      schedulerStartLine({
        prefix: "CHECK-SCHEDULER",
        target: this.schedule.target,
        workUnitCount: this.schedule.units.length,
        resourceLimits: this.schedule.resourceLimits,
        preferredResources: ["cpu", "io", "service_stack"],
        workUnits: this.schedule.units,
      }),
    );
  }

  emit(event, fields, state, detail = {}) {
    if (this.verbose) {
      writeSchedulerTelemetry(process.stdout, "CHECK-SCHEDULER", this.schedule.target, event, fields);
    }
    this.writeEvent(event, state, detail);
  }

  observeState(state) {
    this.maxRunningWorkUnits = Math.max(this.maxRunningWorkUnits, state.running.size);
    this.maxRunningGroups = Math.max(
      this.maxRunningGroups,
      schedulerActiveGroups(Array.from(state.running.values())).size,
    );
    for (const [resource, amount] of state.activeClaims.entries()) {
      this.maxActiveResourceClaims.set(
        resource,
        Math.max(this.maxActiveResourceClaims.get(resource) ?? 0, amount),
      );
    }
  }

  startUnit(unit, state) {
    this.startedAt.set(unit.target, Date.now());
    this.emit(
      "start",
      [
        `work_unit=${unit.label}`,
        `claims=${formatResourceMap(unit.resourceClaims)}`,
        ...schedulerStateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
      ],
      state,
      {
        work_unit: unit.label,
        resource_claims: resourceMapToObject(unit.resourceClaims),
        nested_scheduler: nestedSchedulerDetail(unit),
      },
    );
  }

  finishUnit(unit, result, state) {
    const durationMs = Math.max(0, Date.now() - (this.startedAt.get(unit.target) ?? Date.now()));
    this.startedAt.delete(unit.target);
    if (result.status === 0) {
      this.completedCount += 1;
    }
    const record = {
      label: result.label,
      id: result.id,
      status: result.status,
      duration_ms: durationMs,
      log_file: relToRepo(result.logFile),
    };
    this.completedWork.push(record);
    if (result.status !== 0 && !this.failedWorkUnit) {
      this.failedWorkUnit = result.label;
    }
    this.emit(
      "finish",
      [
        `work_unit=${result.label}`,
        `status=${result.status}`,
        ...schedulerStateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
      ],
      state,
      {
        work_unit: result.label,
        status: result.status,
        duration_ms: durationMs,
        log_file: relToRepo(result.logFile),
      },
    );
  }

  blocked(state, reason, blockedResources, { waitingOn = [], blockedUnits = [] } = {}) {
    this.blockedReasonsSeen.add(reason);
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
        ...schedulerStateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
      ],
      state,
      {
        blocked_reason: reason,
        blocked_resources: blockedResources,
        waiting_on: waitingOn,
        blocked_units: blockedUnits,
      },
    );
    const blockedKey = `${reason}:${blockedResources.join(",")}:${waitingOn.join(",")}:${JSON.stringify(blockedUnits)}`;
    this.maybeProgress(state, reason, blockedResources, {
      force: blockedKey !== this.lastBlockedKey,
      waitingOn,
      blockedUnits,
    });
    this.lastBlockedKey = blockedKey;
  }

  skipUnit(unit, state, reason, failedDependency) {
    this.blockedReasonsSeen.add(reason);
    const record = {
      label: unit.label,
      id: unit.target,
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
        ...schedulerStateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
      ],
      state,
      {
        work_unit: unit.label,
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
    const runningUnits = Array.from(state.running.values());
    const blockedBy = schedulerBlockedBy({ reason, blockedResources });
    for (const explanation of blockedBy) {
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
    });
    process.stdout.write(
      schedulerProgressLine({
        prefix: "CHECK-SCHEDULER",
        target: this.schedule.target,
        completed: this.completedCount,
        total: this.schedule.units.length,
        running: state.running.size,
        pending: state.pending.length,
        blocked: state.blockedCount ?? 0,
        activeGroups: schedulerActiveGroups(runningUnits),
        blockedBy,
        waitingOn,
        unblocksAfter: this.unblocksAfter(state, blockedResources),
        slowestRunning: schedulerSlowestRunning(runningUnits, this.startedAt, now),
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
          const leftStarted = this.startedAt.get(left.target) ?? Number.MAX_SAFE_INTEGER;
          const rightStarted = this.startedAt.get(right.target) ?? Number.MAX_SAFE_INTEGER;
          return leftStarted - rightStarted || left.label.localeCompare(right.label);
        });
      if (candidates.length > 0) {
        return candidates[0].label;
      }
    }
    const runningByTarget = new Map(runningUnits.map((unit) => [unit.target, unit]));
    for (const unit of state.pending) {
      for (const need of unit.needs) {
        const runningNeed = runningByTarget.get(need);
        if (runningNeed) {
          return runningNeed.label;
        }
      }
    }
    return "none";
  }

  async summary(status, { failedWorkUnit = null } = {}) {
    const failed = failedWorkUnit || this.failedWorkUnit || null;
    const slowest = this.slowestWork();
    const skipped = this.skippedWork.length;
    process.stdout.write(
      schedulerSummaryLine({
        prefix: "CHECK-SCHEDULER",
        target: this.schedule.target,
        status,
        completed: this.completedCount,
        total: this.schedule.units.length,
        failed,
        skipped,
        slowest,
      }),
    );
    await writeFile(
      this.summaryPath,
      `${JSON.stringify(
        {
          schema_id: schedulerSummarySchemaID,
          target: this.schedule.target,
          status,
          total_work_units: this.schedule.units.length,
          completed_work_units: this.completedCount,
          skipped_work_units: this.skippedWork,
          failed_work_unit: failed,
          max_running_work_units: this.maxRunningWorkUnits,
          max_running_groups: this.maxRunningGroups,
          max_active_resource_claims: resourceMapToObject(this.maxActiveResourceClaims),
          nested_scheduler_limits: this.schedule.units
            .filter((unit) => unit.nestedScheduler)
            .map((unit) => ({
              work_unit: unit.label,
              ...nestedSchedulerDetail(unit),
            })),
          blocked_reasons_seen: Array.from(this.blockedReasonsSeen).sort((left, right) =>
            left.localeCompare(right),
          ),
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
        completed: this.completedCount,
        blocked_reason: detail.blocked_reason ?? null,
        blocked_resources: detail.blocked_resources ?? [],
        waiting_on: detail.waiting_on ?? [],
        blocked_units: detail.blocked_units ?? [],
        active_resource_claims: resourceMapToObject(state.activeClaims),
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

function readyPendingUnits(pending, completed) {
  return pending.filter((unit) => unit.needs.every((need) => completed.has(need)));
}

function blockedPendingUnits(pending, completed) {
  return pending.filter((unit) => !unit.needs.every((need) => completed.has(need)));
}

function skippedReasonForUnit(unit, completed, failedTarget, unitsByTarget, memo = new Map()) {
  if (memo.has(unit.target)) {
    return memo.get(unit.target);
  }
  for (const need of unit.needs) {
    if (need === failedTarget) {
      memo.set(unit.target, "dependency_failure");
      return "dependency_failure";
    }
    if (!completed.has(need)) {
      const upstream = unitsByTarget.get(need);
      if (upstream && skippedReasonForUnit(upstream, completed, failedTarget, unitsByTarget, memo) === "dependency_failure") {
        memo.set(unit.target, "dependency_failure");
        return "dependency_failure";
      }
    }
  }
  memo.set(unit.target, "schedule_stopped_after_failure");
  return "schedule_stopped_after_failure";
}

function runWorkUnit({ makeBin, unit, logDir, started }) {
  const logFile = path.join(logDir, `${String(started).padStart(2, "0")}-${sanitizeLogName(unit.target)}.log`);
  return runCommand(
    makeBin,
    ["--no-print-directory", "--output-sync=target", `-j${unit.makeJobs}`, unit.target],
    logFile,
    makeChildEnv(nestedSchedulerEnv(unit)),
  ).then((result) => ({
    id: unit.target,
    label: unit.label,
    status: result.status,
    logFile,
  }));
}

async function runSchedule({ schedule, makeBin, testOutputScript, summaryTargets, summaryGroups }) {
  const reporter = await createCheckSchedulerReporter(schedule);
  const pending = [...schedule.units];
  const running = new Map();
  const completed = new Set();
  const unitsByTarget = new Map(schedule.units.map((unit) => [unit.target, unit]));
  const activeClaims = new Map();
  const totalUnits = schedule.units.length;
  const capacityDisplay = schedule.resourceLimits.get("cpu") ?? Math.max(...schedule.resourceLimits.values());
  const skippedSummaryTargets = new Set();
  let started = 0;
  let completedCount = 0;
  let firstFailure = 0;
  let firstFailureTarget = "-";
  let stopScheduling = false;

  try {
    const stateSnapshot = (blockedCount = 0) => ({
      pending,
      running,
      activeClaims,
      blockedCount,
    });

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
    reporter.start();

    const startUnit = async (unit) => {
      started += 1;
      if (reporter.verbose) {
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
      }
      addResourceClaims(unit, activeClaims);
      const promise = runWorkUnit({ makeBin, unit, logDir: reporter.logDir, started });
      running.set(promise, unit);
      reporter.startUnit(unit, stateSnapshot());
    };

    while (pending.length > 0 || running.size > 0) {
      let waitProgressReason = "none";
      let waitProgressResources = [];
      let waitProgressWaitingOn = [];
      let waitProgressBlockedUnits = [];
      let waitProgressBlockedCount = 0;
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
        const dependencyBlocked = blockedPendingUnits(pending, completed);
        const readyBlocked = readyPendingUnits(pending, completed).filter(
          (unit) => !hasResourceCapacity(unit, schedule.resourceLimits, activeClaims),
        );
        const blockedResources = blockedResourcesForUnits(readyBlocked, schedule.resourceLimits, activeClaims);
        const waitingOn = schedulerWaitingOnForUnits(dependencyBlocked, completed);
        const blockedUnits = schedulerBlockedUnitRecords({
          dependencyBlocked,
          resourceBlocked: readyBlocked,
          completed,
          blockedResourcesForUnit: (unit) =>
            blockedResourcesForUnit(unit, schedule.resourceLimits, activeClaims),
        });
        let reason = "none";
        if (dependencyBlocked.length > 0 && readyBlocked.length > 0) {
          reason = "dependencies,resources";
        } else if (dependencyBlocked.length > 0) {
          reason = "dependencies";
        } else if (readyBlocked.length > 0) {
          reason = "resources";
        }
        waitProgressReason = reason;
        waitProgressResources = blockedResources;
        waitProgressWaitingOn = waitingOn;
        waitProgressBlockedUnits = blockedUnits;
        waitProgressBlockedCount = dependencyBlocked.length + readyBlocked.length;
        reporter.blocked(stateSnapshot(dependencyBlocked.length + readyBlocked.length), reason, blockedResources, {
          waitingOn,
          blockedUnits,
        });
      } else {
        reporter.maybeProgress(stateSnapshot(), "none", []);
      }

      if (running.size === 0) {
        if (stopScheduling) {
          const skipped = pending.splice(0);
          const skipMemo = new Map();
          for (const unit of skipped) {
            skippedSummaryTargets.add(unit.target);
            for (const target of unit.skippedSummaryTargets) {
              skippedSummaryTargets.add(target);
            }
            reporter.skipUnit(
              unit,
              stateSnapshot(skipped.length),
              skippedReasonForUnit(unit, completed, firstFailureTarget, unitsByTarget, skipMemo),
              firstFailureTarget,
            );
          }
          break;
        }
        throw new Error(`check scheduler deadlock for ${schedule.target}; pending=${pending.map((unit) => unit.target).join(",")}`);
      }

      const progressDelay = schedulerProgressDelay();
      const result = await Promise.race([...running.keys(), progressDelay.promise]);
      if (result?.schedulerProgressTick === true) {
        progressDelay.cancel();
        reporter.maybeProgress(
          stateSnapshot(waitProgressBlockedCount),
          waitProgressReason,
          waitProgressResources,
          {
            force: true,
            waitingOn: waitProgressWaitingOn,
            blockedUnits: waitProgressBlockedUnits,
          },
        );
        continue;
      }
      progressDelay.cancel();
      let finishedUnit;
      for (const [promise, candidate] of running.entries()) {
        if (candidate.target === result.id) {
          running.delete(promise);
          finishedUnit = candidate;
          removeResourceClaims(candidate, activeClaims);
          break;
        }
      }
      if (!finishedUnit) {
        throw new Error(`finished unknown check work unit ${result.id}`);
      }
      reporter.finishUnit(finishedUnit, result, stateSnapshot());
      await replayLog(result.logFile, result.status === 0 ? process.stdout : process.stderr);
      if (result.status === 0) {
        completed.add(result.id);
        completedCount += 1;
      } else if (firstFailure === 0) {
        firstFailure = result.status;
        firstFailureTarget = result.label;
        stopScheduling = true;
      }
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    await reporter.summary(requestedStatus, { failedWorkUnit: firstFailureTarget === "-" ? null : firstFailureTarget });
    const summaryArgs = ["run-summary", schedule.target, requestedStatus, String(completedCount), String(totalUnits), firstFailureTarget];
    if (summaryGroups) {
      summaryArgs.push("--summary-groups", summaryGroups);
    }
    for (const skipped of reporter.skippedWork) {
      const skippedUnit = unitsByTarget.get(skipped.id);
      if (!skippedUnit) {
        continue;
      }
      skippedSummaryTargets.add(skippedUnit.target);
      for (const target of skippedUnit.skippedSummaryTargets) {
        skippedSummaryTargets.add(target);
      }
    }
    const skippedSummaryTargetsList = summaryTargets.filter((target) => skippedSummaryTargets.has(target));
    if (skippedSummaryTargetsList.length > 0) {
      summaryArgs.push("--skipped-after-failure", skippedSummaryTargetsList.join(","));
    }
    summaryArgs.push(...summaryTargets);
    await runLifecycle(testOutputScript, summaryArgs, requestedStatus === "pass" ? process.stdout : process.stderr).catch((error) => {
      if (firstFailure === 0) {
        throw error;
      }
    });
    return firstFailure;
  } finally {
    await reporter.close();
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
    const dependencyCount = schedule.units.reduce((sum, unit) => sum + unit.needs.length, 0);
    process.stdout.write(
      schedulerDryRunLine({
        target: options.target,
        manifest: path.relative(repoRoot, manifestPath),
        resourceLimits: schedule.resourceLimits,
        preferredResources: ["cpu", "io", "service_stack"],
        workUnits: schedule.units,
        dependencies: dependencyCount,
      }),
    );
    if (verboseSchedulerOutput()) {
      for (const unit of schedule.units) {
        const nested = unit.nestedScheduler
          ? ` nested_scheduler=${JSON.stringify(nestedSchedulerDetail(unit))}`
          : "";
        process.stdout.write(
          `[DRY-RUN] ${options.target} unit ${unit.label} needs=${unit.needs.length === 0 ? "none" : unit.needs.join(",")} claims=${formatResourceMap(unit.resourceClaims)} make_jobs=${unit.makeJobs}${nested}\n`,
        );
      }
    }
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
