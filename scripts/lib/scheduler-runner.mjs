import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";

import {
  formatResourceList,
  formatResourceMap,
  relToRepo as relToRepoPath,
  resourceMapToObject,
  schedulerActiveGroups,
  schedulerBlockedUnitRecords,
  schedulerDryRunLine,
  schedulerLogDir,
  schedulerProgressEventFields,
  schedulerProgressIntervalMs,
  schedulerProgressLine,
  schedulerProgressSnapshot,
  schedulerStartLine,
  schedulerSummaryLine,
  schedulerTargetDir,
  schedulerWaitingOnForUnits,
  writeSchedulerTelemetry,
  verboseSchedulerOutput,
} from "./scheduler-reporting.mjs";
import {
  preferredResourcesForScheduler,
  resourceLimitSourcesToObject,
} from "./scheduler-resources.mjs";

export function isDryRunFromMakeFlags(env = process.env) {
  const flags = ` ${env.MAKEFLAGS ?? ""} `;
  return flags.includes(" n") || flags.includes(" --just-print") || flags.includes(" --dry-run");
}

export function sanitizeMakeFlags(value) {
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

export function makeChildEnv(env = process.env) {
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

export function sanitizeLogName(value) {
  return value.replace(/[^A-Za-z0-9._-]+/g, "-");
}

export function runLifecycle(repoRoot, testOutputScript, args, stream = process.stdout, env = process.env) {
  return new Promise((resolve, reject) => {
    const command = testOutputScript.endsWith(".mjs")
      ? env.NODE_BIN || process.env.NODE_BIN || process.execPath
      : testOutputScript;
    const commandArgs = testOutputScript.endsWith(".mjs") ? [testOutputScript, ...args] : args;
    const child = spawn(command, commandArgs, {
      cwd: repoRoot,
      env,
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

export function runCommand(repoRoot, command, args, logFile, env = process.env) {
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

export async function replayLog(file, stream) {
  await new Promise((resolve, reject) => {
    const reader = createReadStream(file);
    reader.on("error", reject);
    reader.on("end", resolve);
    reader.pipe(stream, { end: false });
  });
}

function progressDelay() {
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

function counted(unit) {
  return unit.countInTotal !== false;
}

function finalizer(unit) {
  return unit.kind === "finalizer";
}

function visibleRunningCount(running) {
  return Array.from(running.values()).filter(counted).length;
}

function visiblePendingCount(pending) {
  return pending.filter(counted).length;
}

function pendingFinalizerCount(pending) {
  return pending.filter(finalizer).length;
}

function runningFinalizerCount(running) {
  return Array.from(running.values()).filter(finalizer).length;
}

function unitCompletionKeys(unit) {
  return unit.completionKeys?.length ? unit.completionKeys : [unit.id];
}

function unitFailureKeys(unit) {
  return unit.failureKeys?.length ? unit.failureKeys : unitCompletionKeys(unit);
}

function hasFailedDependency(unit, failedKeys) {
  return (unit.needs ?? []).find((need) => failedKeys.has(need)) ?? null;
}

function dependenciesSatisfied(unit, completedKeys) {
  return (unit.needs ?? []).every((need) => completedKeys.has(need));
}

function readyPendingUnits(pending, completedKeys, failedKeys) {
  return pending.filter(
    (unit) => !hasFailedDependency(unit, failedKeys) && dependenciesSatisfied(unit, completedKeys),
  );
}

function dependencyBlockedPendingUnits(pending, completedKeys, failedKeys) {
  return pending.filter(
    (unit) => counted(unit) && !hasFailedDependency(unit, failedKeys) && !dependenciesSatisfied(unit, completedKeys),
  );
}

function hasResourceCapacity(unit, resourceLimits, activeClaims) {
  return blockedResourcesForUnit(unit, resourceLimits, activeClaims).length === 0;
}

function blockedResourcesForUnit(unit, resourceLimits, activeClaims) {
  const blocked = [];
  for (const [resource, amount] of unit.resourceClaims.entries()) {
    const limit = resourceLimits.get(resource);
    if (limit !== undefined && (activeClaims.get(resource) ?? 0) + amount > limit) {
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

function formatBlockedWorkUnits(workUnits) {
  return workUnits
    .map((unit) => `${unit.label} claims=${formatResourceMap(unit.resourceClaims)}`)
    .join("; ");
}

function stateFields({ pending, running, activeClaims, resourceLimits }) {
  return [
    `active=${visibleRunningCount(running)}`,
    `pending=${visiblePendingCount(pending)}`,
    `active_resource_claims=${formatResourceMap(activeClaims)}`,
    `resource_limits=${formatResourceMap(resourceLimits)}`,
  ];
}

function relToRepo(repoRoot, value) {
  return relToRepoPath(repoRoot, value);
}

function defaultLogFile(logDir, unit, started) {
  const prefix = counted(unit) ? `${String(started).padStart(2, "0")}-` : "";
  return path.join(logDir, `${prefix}${sanitizeLogName(unit.id)}.log`);
}

function slowestWork(completedWork) {
  return [...completedWork]
    .sort((left, right) => right.duration_ms - left.duration_ms || left.label.localeCompare(right.label))
    .slice(0, 5);
}

function skippedReasonForStoppedUnit(unit, completedKeys, failedKey, unitsByCompletionKey, memo = new Map()) {
  if (memo.has(unit.id)) {
    return memo.get(unit.id);
  }
  for (const need of unit.needs ?? []) {
    if (need === failedKey) {
      memo.set(unit.id, "dependency_failure");
      return "dependency_failure";
    }
    if (!completedKeys.has(need)) {
      const upstream = unitsByCompletionKey.get(need);
      if (
        upstream &&
        skippedReasonForStoppedUnit(upstream, completedKeys, failedKey, unitsByCompletionKey, memo) ===
          "dependency_failure"
      ) {
        memo.set(unit.id, "dependency_failure");
        return "dependency_failure";
      }
    }
  }
  memo.set(unit.id, "schedule_stopped_after_failure");
  return "schedule_stopped_after_failure";
}

class SchedulerReporter {
  constructor(repoRoot, schedule, targetDir, logDir) {
    this.repoRoot = repoRoot;
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
    this.completedCount = 0;
    this.failedWorkUnit = null;
    this.finalizerFailures = 0;
    this.blockedReasonsSeen = new Set();
    this.blockedResourcesSeen = new Set();
    this.blockedExplanationsSeen = new Set();
    this.waitingOnSeen = new Set();
    this.lastProgressAt = schedule.initialProgressAt ?? 0;
    this.lastBlockedKey = null;
    this.maxRunningWorkUnits = 0;
    this.maxRunningGroups = 0;
    this.maxActiveResourceClaims = new Map();
  }

  start() {
    if (this.schedule.quietStart === true && !this.verbose) {
      return;
    }
    process.stdout.write(
      schedulerStartLine({
        prefix: this.schedule.prefix,
        target: this.schedule.target,
        workUnitCount: this.schedule.totalWorkUnits,
        finalizerCount: this.schedule.finalizerCount ?? null,
        resourceLimits: this.schedule.resourceLimits,
        preferredResources: preferredResourcesForScheduler(this.schedule.resourceScheduler),
        workUnits: this.schedule.workUnits.filter(counted),
      }),
    );
  }

  runningDisplayUnits(state) {
    if (this.schedule.runningDisplayUnits) {
      return this.schedule.runningDisplayUnits(state);
    }
    return Array.from(state.running.values());
  }

  observeState(state) {
    const runningDisplay = this.runningDisplayUnits(state);
    this.maxRunningWorkUnits = Math.max(this.maxRunningWorkUnits, visibleRunningCount(state.running));
    this.maxRunningGroups = Math.max(this.maxRunningGroups, schedulerActiveGroups(runningDisplay).size);
    for (const [resource, amount] of state.activeClaims.entries()) {
      this.maxActiveResourceClaims.set(
        resource,
        Math.max(this.maxActiveResourceClaims.get(resource) ?? 0, amount),
      );
    }
  }

  emit(event, fields, state, detail = {}) {
    if (this.verbose) {
      writeSchedulerTelemetry(process.stdout, this.schedule.prefix, this.schedule.target, event, fields);
    }
    this.writeEvent(event, state, detail);
  }

  startUnit(unit, logFile, state) {
    this.startedAt.set(unit.id, Date.now());
    if (finalizer(unit)) {
      this.emit(
        "finalize-start",
        [
          `target=${unit.aggregateTarget}`,
          `shards=${unit.shardNames?.length ?? 0}`,
          `active_finalizers=${runningFinalizerCount(state.running)}`,
          `pending_finalizers=${pendingFinalizerCount(state.pending)}`,
        ],
        state,
        {
          finalizer: unit.aggregateTarget,
          finalizer_id: unit.id,
          shards: unit.shardNames?.length ?? 0,
          log_file: relToRepo(this.repoRoot, logFile),
        },
      );
      return;
    }

    const detail = {
      work_unit: unit.label,
      work_unit_id: unit.id,
      work_unit_type: unit.kind,
      work_unit_class: unit.class,
      aggregate_target: unit.aggregateTarget,
      resource_claims: resourceMapToObject(unit.resourceClaims),
      log_file: relToRepo(this.repoRoot, logFile),
      ...(unit.startDetail ?? {}),
    };
    this.emit(
      "start",
      [
        `work_unit=${unit.label}`,
        `claims=${formatResourceMap(unit.resourceClaims)}`,
        ...stateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
      ],
      state,
      detail,
    );
  }

  finishUnit(unit, result, state) {
    const durationMs = Math.max(0, Date.now() - (this.startedAt.get(unit.id) ?? Date.now()));
    this.startedAt.delete(unit.id);
    if (this.schedule.countCompletedUnit(unit, result)) {
      this.completedCount += 1;
    }
    const record = {
      label: result.label,
      id: result.id,
      kind: finalizer(unit) ? "finalizer" : "work_unit",
      status: result.status,
      duration_ms: durationMs,
      log_file: relToRepo(this.repoRoot, result.logFile),
    };
    this.completedWork.push(record);
    if (unit.aggregateTarget && counted(unit)) {
      if (!this.completedLogFilesByTarget.has(unit.aggregateTarget)) {
        this.completedLogFilesByTarget.set(unit.aggregateTarget, []);
      }
      this.completedLogFilesByTarget.get(unit.aggregateTarget).push(result.logFile);
    }
    if (result.status !== 0 && !this.failedWorkUnit) {
      this.failedWorkUnit = result.label;
    }
    if (finalizer(unit) && result.status !== 0) {
      this.finalizerFailures += 1;
    }

    if (finalizer(unit)) {
      this.emit(
        "finalize-finish",
        [
          `target=${unit.aggregateTarget}`,
          `status=${result.status}`,
          `active_finalizers=${runningFinalizerCount(state.running)}`,
          `pending_finalizers=${pendingFinalizerCount(state.pending)}`,
        ],
        state,
        {
          finalizer: unit.aggregateTarget,
          finalizer_id: unit.id,
          status: result.status,
          duration_ms: durationMs,
          log_file: relToRepo(this.repoRoot, result.logFile),
        },
      );
      return;
    }

    this.emit(
      "finish",
      [
        `work_unit=${result.label}`,
        `status=${result.status}`,
        ...stateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
      ],
      state,
      {
        work_unit: result.label,
        work_unit_id: unit.id,
        status: result.status,
        duration_ms: durationMs,
        log_file: relToRepo(this.repoRoot, result.logFile),
      },
    );
  }

  completedLogFilesForTarget(target) {
    return this.completedLogFilesByTarget.get(target) ?? [];
  }

  async blocked(state, reason, blockedResources, { waitingOn = [], blockedUnits = [] } = {}) {
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
        ...stateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
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
    await this.maybeProgress(state, reason, blockedResources, {
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
        ...stateFields({ ...state, resourceLimits: this.schedule.resourceLimits }),
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

  async maybeProgress(
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
    const extra = this.schedule.progressExtras
      ? await this.schedule.progressExtras({ state, runningUnits: Array.from(state.running.values()) })
      : {};
    this.writeEvent("progress", state, {
      blocked_reason: reason,
      blocked_resources: blockedResources,
      waiting_on: waitingOn,
      blocked_units: blockedUnits,
      ...schedulerProgressEventFields(progress),
      ...(extra.eventDetail ?? {}),
    });
    process.stdout.write(
      schedulerProgressLine({
        prefix: this.schedule.prefix,
        target: this.schedule.target,
        completed: this.completedCount,
        total: this.schedule.totalWorkUnits,
        running: visibleRunningCount(state.running),
        pending: visiblePendingCount(state.pending),
        blocked: state.blockedCount ?? 0,
        finalizing: this.schedule.showFinalizing ? runningFinalizerCount(state.running) : null,
        activeGroups: progress.activeGroups,
        blockedBy: progress.blockedBy,
        waitingOn,
        unblocksAfter: progress.unblocksAfter,
        slowestRunning: progress.slowestRunning,
        artifacts: relToRepo(this.repoRoot, this.targetDir),
      }),
    );
    if (extra.writeLines) {
      extra.writeLines();
    }
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
        return candidates[0].unblockLabel ?? candidates[0].label;
      }
    }

    const runningByCompletionKey = new Map();
    for (const unit of runningUnits) {
      for (const key of [...unitCompletionKeys(unit), ...(unit.runningDependencyKeys ?? [])]) {
        runningByCompletionKey.set(key, unit);
      }
    }
    for (const unit of state.pending) {
      if (!counted(unit)) {
        continue;
      }
      for (const need of unit.needs ?? []) {
        const runningNeed = runningByCompletionKey.get(need);
        if (runningNeed) {
          return runningNeed.unblockLabel ?? runningNeed.label;
        }
      }
    }
    return "none";
  }

  async summary(status, { started, failedWorkUnit = null } = {}) {
    const failed = failedWorkUnit || this.failedWorkUnit || null;
    const slowest = slowestWork(this.completedWork);
    const skipped = this.skippedWork.length;
    if (this.schedule.summaryOnPass !== false || status !== "pass" || this.verbose) {
      process.stdout.write(
        schedulerSummaryLine({
          prefix: this.schedule.prefix,
          target: this.schedule.target,
          status,
          completed: this.completedCount,
          total: this.schedule.totalWorkUnits,
          failed,
          skipped,
          finalizerFailures: this.finalizerFailures,
          slowest,
        }),
      );
    }
    const baseSummary = {
      schema_id: this.schedule.summarySchemaID,
      target: this.schedule.target,
      status,
      total_work_units: this.schedule.totalWorkUnits,
      completed_work_units: this.completedCount,
      skipped_work_units: this.skippedWork,
      failed_work_unit: failed,
      resource_limit_sources: resourceLimitSourcesToObject(this.schedule.resourceLimitSources),
      blocked_resources_seen: Array.from(this.blockedResourcesSeen).sort((left, right) =>
        left.localeCompare(right),
      ),
      blocked_explanations_seen: Array.from(this.blockedExplanationsSeen).sort((left, right) =>
        left.localeCompare(right),
      ),
      waiting_on_seen: Array.from(this.waitingOnSeen).sort((left, right) => left.localeCompare(right)),
      slowest_work_units: slowest,
      artifacts: {
        events_jsonl: relToRepo(this.repoRoot, this.eventsPath),
        scheduler_logs_dir: relToRepo(this.repoRoot, this.logDir),
      },
    };
    const extra = this.schedule.summaryExtra ? this.schedule.summaryExtra({ reporter: this, started }) : {};
    await writeFile(
      this.summaryPath,
      `${JSON.stringify(
        {
          ...baseSummary,
          ...extra,
        },
        null,
        2,
      )}\n`,
    );
  }

  writeEvent(event, state, detail) {
    this.observeState(state);
    const base = {
      schema_id: this.schedule.eventSchemaID,
      target: this.schedule.target,
      event,
      timestamp: new Date().toISOString(),
      pending: visiblePendingCount(state.pending),
      running: visibleRunningCount(state.running),
      total_work_units: this.schedule.totalWorkUnits,
      blocked: state.blockedCount ?? 0,
      completed: this.completedCount,
      blocked_reason: detail.blocked_reason ?? null,
      blocked_resources: detail.blocked_resources ?? [],
      waiting_on: detail.waiting_on ?? [],
      blocked_units: detail.blocked_units ?? [],
      active_resource_claims: resourceMapToObject(state.activeClaims),
      resource_limits: resourceMapToObject(this.schedule.resourceLimits),
      resource_limit_sources: resourceLimitSourcesToObject(this.schedule.resourceLimitSources),
    };
    if (this.schedule.showFinalizing) {
      base.pending_finalizers = pendingFinalizerCount(state.pending);
      base.running_finalizers = runningFinalizerCount(state.running);
    }
    this.events.write(`${JSON.stringify({ ...base, ...detail })}\n`);
  }

  close() {
    return new Promise((resolve, reject) => {
      this.events.on("error", reject);
      this.events.end(resolve);
    });
  }
}

async function createReporter(repoRoot, schedule) {
  const targetDir = schedulerTargetDir(repoRoot, schedule.target);
  const logDir = schedulerLogDir(repoRoot, schedule.target);
  await mkdir(logDir, { recursive: true });
  return new SchedulerReporter(repoRoot, schedule, targetDir, logDir);
}

function normalizeSchedule(schedule) {
  const totalWorkUnits = schedule.totalWorkUnits ?? schedule.workUnits.filter(counted).length;
  return {
    ...schedule,
    totalWorkUnits,
    resourceScheduler: schedule.resourceScheduler ?? schedule.kind,
    countCompletedUnit:
      schedule.countCompletedUnit ??
      ((unit, result) => counted(unit) && result.status === 0),
  };
}

function resultStreamFor(schedule, unit, result, reporter) {
  if (schedule.resultStreamFor) {
    return schedule.resultStreamFor({ unit, result, reporter });
  }
  return result.status === 0 ? process.stdout : process.stderr;
}

function shouldReplayLog(schedule, unit, result, reporter) {
  if (schedule.shouldReplayLog) {
    return schedule.shouldReplayLog({ unit, result, reporter });
  }
  return true;
}

function failedDependencyForSkip(unit, failedKeys) {
  for (const need of unit.needs ?? []) {
    if (failedKeys.has(need)) {
      return need;
    }
  }
  return null;
}

function skipFailedDependencyUnits({ pending, failedKeys, reporter, stateSnapshot }) {
  let skipped = 0;
  for (let index = 0; index < pending.length; ) {
    const unit = pending[index];
    const failedDependency = failedDependencyForSkip(unit, failedKeys);
    if (!failedDependency) {
      index += 1;
      continue;
    }
    pending.splice(index, 1);
    skipped += 1;
    for (const key of unitFailureKeys(unit)) {
      failedKeys.set(key, failedDependency);
    }
    reporter.skipUnit(
      unit,
      stateSnapshot(skipped),
      "dependency_failure",
      failedDependency,
    );
  }
  return skipped;
}

function blockedSnapshot({ pending, completedKeys, failedKeys, resourceLimits, activeClaims }) {
  const dependencyBlocked = dependencyBlockedPendingUnits(pending, completedKeys, failedKeys);
  const readyBlocked = readyPendingUnits(pending, completedKeys, failedKeys)
    .filter(counted)
    .filter((unit) => !hasResourceCapacity(unit, resourceLimits, activeClaims));
  const blockedResources = blockedResourcesForUnits(readyBlocked, resourceLimits, activeClaims);
  const waitingOn = schedulerWaitingOnForUnits(dependencyBlocked, completedKeys);
  const blockedUnits = schedulerBlockedUnitRecords({
    dependencyBlocked,
    resourceBlocked: readyBlocked,
    completed: completedKeys,
    blockedResourcesForUnit: (unit) => blockedResourcesForUnit(unit, resourceLimits, activeClaims),
  });
  let reason = "none";
  if (dependencyBlocked.length > 0 && readyBlocked.length > 0) {
    reason = "dependencies,resources";
  } else if (dependencyBlocked.length > 0) {
    reason = "dependencies";
  } else if (readyBlocked.length > 0) {
    reason = "resources";
  }
  return {
    dependencyBlocked,
    readyBlocked,
    blockedResources,
    waitingOn,
    blockedUnits,
    blockedCount: dependencyBlocked.length + readyBlocked.length,
    reason,
  };
}

export async function runNormalizedSchedule({ repoRoot, schedule: rawSchedule, testOutputScript }) {
  const schedule = normalizeSchedule(rawSchedule);
  const reporter = await createReporter(repoRoot, schedule);
  const pending = [...schedule.workUnits];
  const running = new Map();
  const completedKeys = new Set();
  const failedKeys = new Map();
  const activeClaims = new Map();
  const unitsByCompletionKey = new Map();
  for (const unit of schedule.workUnits) {
    for (const key of unitCompletionKeys(unit)) {
      unitsByCompletionKey.set(key, unit);
    }
  }

  let started = 0;
  let firstFailure = 0;
  let firstFailureLabel = null;
  let firstFailureKey = null;
  let stopScheduling = false;

  const stateSnapshot = (blockedCount = 0) => ({
    pending,
    running,
    activeClaims,
    blockedCount,
  });

  const startUnit = async (unit) => {
    if (counted(unit) || unit.countsStarted !== false) {
      started += 1;
    }
    const logFile = unit.logFile
      ? unit.logFile({ logDir: reporter.logDir, unit, started })
      : defaultLogFile(reporter.logDir, unit, started);
    if (schedule.beforeUnitStart) {
      await schedule.beforeUnitStart({ unit, started, total: schedule.totalWorkUnits, reporter, testOutputScript });
    }
    addResourceClaims(unit, activeClaims);
    const commandSpec = typeof unit.command === "function" ? unit.command({ unit, logFile }) : unit.command;
    const promise = runCommand(
      repoRoot,
      commandSpec.command,
      commandSpec.args,
      logFile,
      commandSpec.env ?? process.env,
    ).then((result) => ({
      id: unit.id,
      label: unit.label,
      status: result.status,
      logFile,
    }));
    running.set(promise, unit);
    reporter.startUnit(unit, logFile, stateSnapshot());
  };

  try {
    if (schedule.beforeRun) {
      await schedule.beforeRun({ reporter, testOutputScript });
    }
    reporter.start();

    while (pending.length > 0 || running.size > 0) {
      if (!schedule.stopOnFirstFailure) {
        skipFailedDependencyUnits({ pending, failedKeys, reporter, stateSnapshot });
      }

      if (!stopScheduling) {
        while (true) {
          const nextIndex = pending.findIndex(
            (candidate) =>
              !hasFailedDependency(candidate, failedKeys) &&
              dependenciesSatisfied(candidate, completedKeys) &&
              hasResourceCapacity(candidate, schedule.resourceLimits, activeClaims),
          );
          if (nextIndex === -1) {
            break;
          }
          const [unit] = pending.splice(nextIndex, 1);
          await startUnit(unit);
        }
      }

      const blocked = blockedSnapshot({
        pending,
        completedKeys,
        failedKeys,
        resourceLimits: schedule.resourceLimits,
        activeClaims,
      });
      if (blocked.blockedCount > 0 && running.size > 0 && !stopScheduling) {
        await reporter.blocked(stateSnapshot(blocked.blockedCount), blocked.reason, blocked.blockedResources, {
          waitingOn: blocked.waitingOn,
          blockedUnits: blocked.blockedUnits,
        });
      } else {
        await reporter.maybeProgress(stateSnapshot(), "none", []);
      }

      if (running.size === 0) {
        if (stopScheduling) {
          const skipped = pending.splice(0);
          const skipMemo = new Map();
          for (const unit of skipped) {
            const reason = skippedReasonForStoppedUnit(
              unit,
              completedKeys,
              firstFailureKey,
              unitsByCompletionKey,
              skipMemo,
            );
            for (const key of unitFailureKeys(unit)) {
              failedKeys.set(key, firstFailureKey);
            }
            reporter.skipUnit(unit, stateSnapshot(skipped.length), reason, firstFailureKey);
          }
          break;
        }
        if (pending.length === 0) {
          break;
        }
        throw new Error(
          `${schedule.kind} scheduler deadlock for ${schedule.target}; pending=${formatBlockedWorkUnits(pending)} active_resource_claims=${formatResourceMap(activeClaims)} resource_limits=${formatResourceMap(schedule.resourceLimits)}`,
        );
      }

      const delay = progressDelay();
      const result = await Promise.race([...running.keys(), delay.promise]);
      if (result?.schedulerProgressTick === true) {
        const blockedNow = blockedSnapshot({
          pending,
          completedKeys,
          failedKeys,
          resourceLimits: schedule.resourceLimits,
          activeClaims,
        });
        await reporter.maybeProgress(
          stateSnapshot(blockedNow.blockedCount),
          blockedNow.reason,
          blockedNow.blockedResources,
          {
            force: true,
            waitingOn: blockedNow.waitingOn,
            blockedUnits: blockedNow.blockedUnits,
          },
        );
        continue;
      }
      delay.cancel();

      let finishedUnit = null;
      for (const [promise, candidate] of running.entries()) {
        if (candidate.id === result.id) {
          running.delete(promise);
          removeResourceClaims(candidate, activeClaims);
          finishedUnit = candidate;
          break;
        }
      }
      if (!finishedUnit) {
        throw new Error(`finished unknown ${schedule.kind} work unit ${result.id}`);
      }

      reporter.finishUnit(finishedUnit, result, stateSnapshot());
      if (schedule.afterUnitFinish) {
        await schedule.afterUnitFinish({ unit: finishedUnit, result, reporter });
      }

      if (result.status === 0 || finishedUnit.completeOnFailure === true) {
        for (const key of unitCompletionKeys(finishedUnit)) {
          completedKeys.add(key);
        }
      }
      if (result.status !== 0) {
        if (firstFailure === 0) {
          firstFailure = result.status;
          firstFailureLabel = result.label;
          firstFailureKey = unitFailureKeys(finishedUnit)[0] ?? finishedUnit.id;
        }
        for (const key of unitFailureKeys(finishedUnit)) {
          failedKeys.set(key, result.label);
        }
        if (schedule.stopOnFirstFailure) {
          stopScheduling = true;
        }
      }

      if (schedule.beforeReplayLog) {
        await schedule.beforeReplayLog({ unit: finishedUnit, result, reporter });
      }
      if (shouldReplayLog(schedule, finishedUnit, result, reporter)) {
        await replayLog(result.logFile, resultStreamFor(schedule, finishedUnit, result, reporter));
      }
    }

    if (schedule.afterWorkComplete) {
      const hookFailure = await schedule.afterWorkComplete({ reporter, firstFailure, firstFailureLabel });
      if (hookFailure?.status && firstFailure === 0) {
        firstFailure = hookFailure.status;
        firstFailureLabel = hookFailure.label;
      }
    }

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    await reporter.summary(requestedStatus, { started, failedWorkUnit: firstFailureLabel });
    if (schedule.afterSummary) {
      await schedule.afterSummary({
        reporter,
        requestedStatus,
        completedKeys,
        firstFailure,
        firstFailureLabel,
        testOutputScript,
      });
    }
    return {
      status: firstFailure,
      requestedStatus,
      reporter,
      completedKeys,
      failedKeys,
    };
  } finally {
    await reporter.close();
  }
}

export function writeSchedulerDryRun({ repoRoot, schedule, manifestPath, verboseUnitLine }) {
  const normalized = normalizeSchedule(schedule);
  const dependencyCount = normalized.dependencyCount ??
    normalized.workUnits.filter(counted).reduce((sum, unit) => sum + (unit.needs ?? []).length, 0);
  process.stdout.write(
    schedulerDryRunLine({
      target: normalized.target,
      manifest: path.relative(repoRoot, manifestPath),
      resourceLimits: normalized.resourceLimits,
      preferredResources: preferredResourcesForScheduler(normalized.resourceScheduler),
      workUnits: normalized.workUnits.filter(counted),
      dependencies: dependencyCount,
      finalizerCount: normalized.finalizerCount ?? null,
    }),
  );
  if (!verboseSchedulerOutput() || !verboseUnitLine) {
    return;
  }
  for (const unit of normalized.workUnits) {
    process.stdout.write(verboseUnitLine(unit));
  }
}
