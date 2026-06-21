import { createWriteStream } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

import {
  formatResourceList,
  formatResourceMap,
  machineSchedulerOutput,
  relToRepo as relToRepoPath,
  resourceMapToObject,
  schedulerActiveGroups,
  schedulerDryRunLine,
  schedulerHumanNestedProgressKey,
  schedulerHumanProgressLine,
  schedulerLogDir,
  schedulerProgressEventFields,
  schedulerProgressIntervalMs,
  schedulerProgressLine,
  schedulerProgressSnapshot,
  schedulerStartLine,
  schedulerSummaryLine,
  schedulerTargetDir,
  writeSchedulerTelemetry,
  verboseSchedulerOutput,
} from "../scheduler-reporting.mjs";
import {
  preferredResourcesForScheduler,
  resourceLimitSourcesToObject,
} from "../scheduler-resources.mjs";
import {
  classifyExecutionFailure,
  failureFieldsForJSON,
  normalizeFailureRecord,
  primaryPublicFailure,
} from "../failure-taxonomy.mjs";
import {
  compactJSONString,
  prettyJSONString,
  validateSchemaSync,
} from "../harness-contract.mjs";
import { SchedulerClock } from "./clock.mjs";
import {
  addTopBlockerObservations,
  schedulerBlockedDiagnostics,
  topBlockerRecords,
} from "./blockers.mjs";
import { validateSchedulerSummaryTiming } from "./summary-timing-drift.mjs";
import { replayLog, runCommand, sanitizeLogName } from "./process-executor.mjs";
import { SchedulerProgressRecorder } from "./progress-recorder.mjs";
import {
  addResourceClaims,
  blockedSnapshot,
  counted,
  finalizer,
  formatBlockedWorkUnits,
  pendingFinalizerCount,
  priorityAdmissiblePendingUnitIndex,
  progressDelay,
  removeResourceClaims,
  runningFinalizerCount,
  skipFailedDependencyUnits,
  skippedReasonForStoppedUnit,
  stateFields,
  unitCompletionKeys,
  unitFailureKeys,
  visiblePendingCount,
  visibleRunningCount,
} from "./state.mjs";
export {
  isDryRunFromMakeFlags,
  makeChildEnv,
  replayLog,
  runCommand,
  runLifecycle,
  sanitizeLogName,
  sanitizeMakeFlags,
} from "./process-executor.mjs";

function relToRepo(repoRoot, value) {
  return relToRepoPath(repoRoot, value);
}

function defaultLogFile(logDir, unit, started) {
  const prefix = counted(unit) ? `${String(started).padStart(2, "0")}-` : "";
  return path.join(logDir, `${prefix}${sanitizeLogName(unit.id)}.log`);
}

export function finalizerRunningDisplayUnits(state) {
  return Array.from(state.running.values()).map((unit) =>
    finalizer(unit)
      ? { id: unit.id, label: `finalize:${unit.aggregateTarget}`, group: unit.aggregateTarget }
      : unit,
  );
}

export function countVisibleCompletedUnit(unit) {
  return counted(unit);
}

export async function replayFailedAggregateLogsBeforeFinalizer({ unit, result, reporter }) {
  if (!finalizer(unit) || result.status === 0 || reporter.verbose) {
    return;
  }
  for (const logFile of reporter.completedLogFilesForTarget(unit.aggregateTarget)) {
    await replayLog(logFile, process.stderr);
  }
}

function slowestWork(completedWork) {
  return [...completedWork]
    .sort((left, right) => right.duration_ms - left.duration_ms || left.label.localeCompare(right.label))
    .slice(0, 5);
}

function incrementCount(counts, key, amount = 1) {
  if (!key) {
    return;
  }
  counts[key] = (counts[key] ?? 0) + amount;
}

function pressureFixtureClass(resourceClaims) {
  if (resourceClaims.migration_scratch_postgres) {
    return "migration_scratch";
  }
  if (resourceClaims.postgres_clone) {
    return "template_clone";
  }
  if (resourceClaims.postgres_reset) {
    return "package_reset";
  }
  if (resourceClaims.postgres) {
    return "transaction_or_shared_postgres";
  }
  return "";
}

function buildPressureSummary({ reporter, status, slowest, timing }) {
  const resourceClaimCounts = {};
  const fixtureClassCounts = {};
  const laneDurations = {};
  const targetCounts = {};
  for (const record of reporter.completedWork) {
    if (record.kind !== "work_unit") {
      continue;
    }
    const lane = record.aggregate_target || record.service_session_target || record.label;
    laneDurations[lane] = (laneDurations[lane] ?? 0) + record.duration_ms;
    incrementCount(targetCounts, lane);
    for (const [resource, amount] of Object.entries(record.resource_claims ?? {})) {
      if (amount) {
        incrementCount(resourceClaimCounts, resource, Number(amount));
      }
    }
    incrementCount(fixtureClassCounts, pressureFixtureClass(record.resource_claims ?? {}));
  }
  return {
    schema_id: "cartulary.scheduler_pressure_summary.v1",
    target: reporter.schedule.target,
    scheduler_kind: reporter.schedule.kind,
    status,
    total_work_units: reporter.schedule.totalWorkUnits,
    completed_work_units: reporter.completedCount,
    scheduler_total_duration_ms: timing.scheduler_total_duration_ms,
    target_counts: targetCounts,
    lane_duration_ms: laneDurations,
    resource_claim_counts: resourceClaimCounts,
    fixture_class_counts: fixtureClassCounts,
    slowest_work_units: slowest,
    reused_accounting_counts: {},
    readiness_attribution_counts: {},
    generated_at: timing.scheduler_completed_at,
  };
}

function finalizerTimings(completedWork) {
  return completedWork
    .filter((record) => record.kind === "finalizer")
    .map((record) => ({
      label: record.label,
      id: record.id,
      status: record.status,
      duration_ms: record.duration_ms,
      log_file: record.log_file,
    }));
}

const observedFailedWorkUnitLimit = 25;

function observedFailedWorkUnits(completedWork) {
  return completedWork
    .filter((record) => record.status !== 0)
    .slice(0, observedFailedWorkUnitLimit)
    .map((record) => ({
      id: record.id,
      label: record.label,
      aggregate_target: record.aggregate_target,
      kind: record.kind,
      work_unit_type: record.work_unit_type,
      service_session_target: record.service_session_target,
      status: record.status,
      duration_ms: record.duration_ms,
      started_monotonic_ms: record.started_monotonic_ms,
      finished_monotonic_ms: record.finished_monotonic_ms,
      needs: [...(record.needs ?? [])],
      completion_keys: [...(record.completion_keys ?? [])],
      resource_claims: { ...(record.resource_claims ?? {}) },
      log_file: record.log_file,
    }));
}

function criticalPathSummary(completedWork, schedulerStartedMonotonicMs, schedulerCompletedMonotonicMs, topBlockers) {
  const records = completedWork.filter((record) => record.status === 0);
  const envelopeDurationMs = Math.max(
    0,
    schedulerCompletedMonotonicMs - schedulerStartedMonotonicMs,
  );
  if (records.length === 0) {
    return {
      critical_path_wall_duration_ms: envelopeDurationMs,
      critical_path_units: [],
      critical_path_blockers: topBlockers.slice(0, 5),
      critical_path_terminal_unit: null,
    };
  }
  const byCompletionKey = new Map();
  for (const record of records) {
    for (const key of record.completion_keys ?? []) {
      byCompletionKey.set(key, record);
    }
  }
  const pathByID = new Map();
  for (const record of records.slice().sort((left, right) =>
    left.finished_monotonic_ms - right.finished_monotonic_ms ||
    left.id.localeCompare(right.id),
  )) {
    let predecessor = null;
    for (const need of record.needs ?? []) {
      const candidate = byCompletionKey.get(need);
      if (!candidate || !pathByID.has(candidate.id)) {
        continue;
      }
      const candidatePath = pathByID.get(candidate.id);
      if (!predecessor || candidatePath.path_duration_ms > predecessor.path_duration_ms) {
        predecessor = candidatePath;
      }
    }
    pathByID.set(record.id, {
      record,
      predecessor,
      path_duration_ms: record.duration_ms + (predecessor?.path_duration_ms ?? 0),
    });
  }
  const terminal = records
    .slice()
    .sort((left, right) =>
      right.finished_monotonic_ms - left.finished_monotonic_ms ||
      right.duration_ms - left.duration_ms ||
      left.id.localeCompare(right.id),
    )[0];
  let cursor = pathByID.get(terminal.id) ?? null;
  const units = [];
  while (cursor) {
    const record = cursor.record;
    units.push({
      id: record.id,
      label: record.label,
      kind: record.kind,
      aggregate_target: record.aggregate_target,
      duration_ms: record.duration_ms,
      started_monotonic_ms: record.started_monotonic_ms,
      finished_monotonic_ms: record.finished_monotonic_ms,
      needs: [...(record.needs ?? [])],
      completion_keys: [...(record.completion_keys ?? [])],
    });
    cursor = cursor.predecessor;
  }
  units.reverse();
  const terminalUnit = {
    id: terminal.id,
    label: terminal.label,
    kind: terminal.kind,
    aggregate_target: terminal.aggregate_target,
    duration_ms: terminal.duration_ms,
    started_monotonic_ms: terminal.started_monotonic_ms,
    finished_monotonic_ms: terminal.finished_monotonic_ms,
    needs: [...(terminal.needs ?? [])],
    completion_keys: [...(terminal.completion_keys ?? [])],
  };
  return {
    critical_path_wall_duration_ms: envelopeDurationMs,
    critical_path_units: units,
    critical_path_blockers: topBlockers.slice(0, 5),
    critical_path_terminal_unit: terminalUnit,
  };
}

async function readSchedulerChildFailureRecord({
  failed,
  failedDetail,
  repoRoot,
  scheduleTarget,
}) {
  const childTarget =
    [
      failedDetail?.aggregate_target,
      failedDetail?.target,
      failedDetail?.label,
      failedDetail?.id,
      failed,
    ]
      .map((value) => (typeof value === "string" ? value.trim() : ""))
      .find((value) => value && value !== scheduleTarget) ?? "";
  if (!childTarget || childTarget === scheduleTarget) {
    return null;
  }
  const summaryFile = path.join(
    schedulerTargetDir(repoRoot, childTarget),
    "target-summary.json",
  );
  let summary;
  try {
    summary = JSON.parse(await readFile(summaryFile, "utf8"));
  } catch {
    return null;
  }
  const failureSource = summary?.totals ?? summary;
  const failures = Array.isArray(failureSource?.failures)
    ? failureSource.failures
    : Array.isArray(summary?.failures)
      ? summary.failures
      : [];
  const workUnitTokens = schedulerWorkUnitFailureTokens(failed, failedDetail);
  const matchingFailures =
    workUnitTokens.length === 0
      ? failures
      : failures.filter((failure) =>
          schedulerFailureMatchesWorkUnit(failure, workUnitTokens),
        );
  const propagated = primaryPublicFailure(
    matchingFailures.length > 0 ? matchingFailures : failures,
    {
      failure_class: failureSource?.failure_class ?? summary?.failure_class,
      failure_reason: failureSource?.failure_reason ?? summary?.failure_reason,
      target: childTarget,
      label: failed ?? childTarget,
    },
  );
  if (!propagated) {
    return null;
  }
  return {
    ...propagated,
    kind: propagated.kind || "child_target",
    source: "scheduler",
    target: scheduleTarget,
    child_target: childTarget,
    work_unit: failedDetail?.id ?? failed ?? childTarget,
    label: failed ?? propagated.label ?? childTarget,
    message:
      propagated.message ||
      `scheduler child target failed: ${childTarget}`,
    artifact: propagated.artifact || relToRepo(repoRoot, summaryFile),
  };
}

function schedulerWorkUnitFailureTokens(failed, failedDetail) {
  const tokens = new Set();
  const values = [
    failed,
    failedDetail?.id,
    failedDetail?.label,
    ...(Array.isArray(failedDetail?.completion_keys)
      ? failedDetail.completion_keys
      : []),
  ];
  for (const value of values) {
    if (typeof value !== "string") {
      continue;
    }
    for (const match of value
      .toLowerCase()
      .matchAll(/browser-functional-shard-\d+/g)) {
      tokens.add(match[0]);
    }
    for (const match of value
      .toLowerCase()
      .matchAll(/functional-shard-\d+/g)) {
      tokens.add(match[0]);
    }
  }
  return [...tokens];
}

function schedulerFailureMatchesWorkUnit(failure, tokens) {
  const haystack = [
    failure?.label,
    failure?.artifact,
    failure?.work_unit,
    failure?.message,
  ]
    .map((value) => (typeof value === "string" ? value.toLowerCase() : ""))
    .join(" ");
  return tokens.some((token) => haystack.includes(token));
}

function schedulerFallbackFailureRecord(record, scheduleTarget) {
  const label = record?.label ?? scheduleTarget;
  return normalizeFailureRecord({
    failure_class: classifyExecutionFailure(label, scheduleTarget),
    kind: "scheduler",
    source: "scheduler",
    target: scheduleTarget,
    child_target: record?.aggregate_target ?? null,
    work_unit: record?.id ?? label,
    label,
    message: `scheduler work unit failed: ${label}`,
    artifact: record?.log_file ?? "",
  });
}

async function schedulerFailureRecordsForCompletedWork({
  completedWork,
  repoRoot,
  scheduleTarget,
}) {
  const failedRecords = completedWork.filter((record) => record.status !== 0);
  const failures = [];
  for (const record of failedRecords) {
    const propagated = await readSchedulerChildFailureRecord({
      failed: record.label,
      failedDetail: record,
      repoRoot,
      scheduleTarget,
    });
    failures.push(propagated ?? schedulerFallbackFailureRecord(record, scheduleTarget));
  }
  return failures;
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
    this.pressureSummaryPath = path.join(targetDir, "pressure-summary.json");
    this.progressSummaryPath = path.join(targetDir, "progress-summary.log");
    this.events = createWriteStream(this.eventsPath, { flags: "w" });
    this.progressRecorder = new SchedulerProgressRecorder(this.progressSummaryPath);
    this.clock = new SchedulerClock();
    this.machine = machineSchedulerOutput();
    this.schedulerStartedMonotonicMs = 0;
    this.schedulerStartedAt = this.clock.wallTimestamp(0);
    this.schedulerCompletedMonotonicMs = null;
    this.schedulerCompletedAt = null;
    this.schedulerTotalDurationMs = null;
    this.startedAt = new Map();
    this.completedWork = [];
    this.completedLogFilesByTarget = new Map();
    this.skippedWork = [];
    this.completedCount = 0;
    this.failedWorkUnit = null;
    this.failedWorkUnitDetail = null;
    this.schedulerFailureRecord = null;
    this.finalizerFailures = 0;
    this.blockedReasonsSeen = new Set();
    this.blockedResourcesSeen = new Set();
    this.blockedExplanationsSeen = new Set();
    this.topBlockerCounts = new Map();
    this.waitingOnSeen = new Set();
    this.lastProgressAt = schedule.deferInitialProgress ? this.clock.monotonicMs() : 0;
    this.eventSequence = 0;
    this.lastEventSequence = 0;
    this.lastEventMonotonicMs = -1;
    this.lastBlockedKey = null;
    this.blockerChangeProgressCount = 0;
    this.lastHumanNestedProgressKey = null;
    this.maxRunningWorkUnits = 0;
    this.maxRunningGroups = 0;
    this.maxActiveResourceClaims = new Map();
    this.schemaValidationEnabled = schedule.schemaValidationEnabled !== false;
    this.deferredSchemaRecords = [];
  }

  setSchemaValidationEnabled(enabled = true) {
    if (!enabled) {
      this.schemaValidationEnabled = false;
      return;
    }
    if (this.schemaValidationEnabled) {
      return;
    }
    this.schemaValidationEnabled = true;
    for (const record of this.deferredSchemaRecords) {
      validateSchemaSync(record.schemaID, record.value);
    }
    this.deferredSchemaRecords = [];
  }

  validateSchemaRecord(schemaID, value) {
    if (!this.schemaValidationEnabled) {
      this.deferredSchemaRecords.push({ schemaID, value });
      return;
    }
    validateSchemaSync(schemaID, value);
  }

  startLifecycle(state) {
    this.writeEvent("scheduler-start", state, {
      scheduler_started_monotonic_ms: this.schedulerStartedMonotonicMs,
      scheduler_started_at: this.schedulerStartedAt,
    });
  }

  start() {
    const line = schedulerStartLine({
      prefix: this.schedule.prefix,
      target: this.schedule.target,
      workUnitCount: this.schedule.totalWorkUnits,
      finalizerCount: this.schedule.finalizerCount ?? null,
      resourceLimits: this.schedule.resourceLimits,
      preferredResources: preferredResourcesForScheduler(this.schedule.resourceScheduler),
      workUnits: this.schedule.workUnits.filter(counted),
      artifacts: relToRepo(this.repoRoot, this.targetDir),
    });
    if (!this.machine) {
      process.stdout.write(line);
    }
    this.progressRecorder.recordStart(line);
  }

  finishLifecycle(state) {
    this.writeEvent("scheduler-finish", state, {
      scheduler_started_monotonic_ms: this.schedulerStartedMonotonicMs,
      scheduler_started_at: this.schedulerStartedAt,
    });
    this.schedulerCompletedMonotonicMs = this.lastEventMonotonicMs;
    this.schedulerCompletedAt = this.clock.wallTimestamp(
      this.schedulerCompletedMonotonicMs,
    );
    this.schedulerTotalDurationMs = Math.max(
      0,
      this.schedulerCompletedMonotonicMs - this.schedulerStartedMonotonicMs,
    );
  }

  timingEnvelope() {
    const completedMonotonicMs =
      this.schedulerCompletedMonotonicMs ?? this.lastEventMonotonicMs;
    const completedAt = this.schedulerCompletedAt
      ?? this.clock.wallTimestamp(completedMonotonicMs);
    return {
      scheduler_started_monotonic_ms: this.schedulerStartedMonotonicMs,
      scheduler_completed_monotonic_ms: completedMonotonicMs,
      scheduler_total_duration_ms: Math.max(
        0,
        completedMonotonicMs - this.schedulerStartedMonotonicMs,
      ),
      scheduler_started_at: this.schedulerStartedAt,
      scheduler_completed_at: completedAt,
    };
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
    this.startedAt.set(unit.id, this.clock.monotonicMs());
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
    const now = this.clock.monotonicMs();
    const startedMonotonicMs = this.startedAt.get(unit.id) ?? now;
    const durationMs = Math.max(0, now - startedMonotonicMs);
    this.startedAt.delete(unit.id);
    if (this.schedule.countCompletedUnit(unit, result)) {
      this.completedCount += 1;
    }
    const record = {
      label: result.label,
      id: result.id,
      aggregate_target: unit.aggregateTarget ?? null,
      kind: finalizer(unit) ? "finalizer" : "work_unit",
      work_unit_type: unit.kind ?? null,
      service_session_target: typeof unit.serviceSession?.target === "string" ? unit.serviceSession.target : null,
      status: result.status,
      duration_ms: durationMs,
      started_monotonic_ms: startedMonotonicMs,
      finished_monotonic_ms: now,
      needs: [...(unit.needs ?? [])],
      completion_keys: [...unitCompletionKeys(unit)],
      resource_claims: resourceMapToObject(unit.resourceClaims),
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
      this.failedWorkUnitDetail = record;
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

  recordSchedulerFailure(failure = {}) {
    const label = String(failure.label ?? this.schedule.target);
    this.failedWorkUnit = label;
    this.schedulerFailureRecord = normalizeFailureRecord(
      {
        ...failure,
        kind: failure.kind ?? "scheduler",
        source: failure.source ?? "scheduler",
        target: this.schedule.target,
        label,
        message:
          failure.message ??
          (label
            ? `scheduler work unit failed: ${label}`
            : `scheduler target failed: ${this.schedule.target}`),
      },
      {
        failure_class: "harness",
        failure_reason: "unknown_failure",
      },
    );
  }

  completedLogFilesForTarget(target) {
    return this.completedLogFilesByTarget.get(target) ?? [];
  }

  async blocked(state, reason, blockedResources, { waitingOn = [], blockedUnits = [] } = {}) {
    const blockerDiagnostics = schedulerBlockedDiagnostics({
      reason,
      blockedResources,
      waitingOn,
      blockedUnits,
    });
    this.blockedReasonsSeen.add(reason);
    for (const resource of blockedResources) {
      this.blockedResourcesSeen.add(resource);
    }
    for (const explanation of blockerDiagnostics.explanations) {
      this.blockedExplanationsSeen.add(explanation);
    }
    addTopBlockerObservations(
      this.topBlockerCounts,
      blockerDiagnostics.observations,
    );
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
    const blockedKey = blockerDiagnostics.materialKey;
    const blockerChanged = blockedKey !== this.lastBlockedKey;
    const quietBlockerProgress =
      !this.verbose &&
      !this.machine &&
      blockerChanged &&
      this.blockerChangeProgressCount < 2;
    if (quietBlockerProgress) {
      this.blockerChangeProgressCount += 1;
    }
    await this.maybeProgress(state, reason, blockedResources, {
      force:
        blockerChanged &&
        (this.verbose || this.machine || quietBlockerProgress),
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
    const now = this.clock.monotonicMs();
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
    const nestedProgress = extra.eventDetail?.nested_scheduler_progress ?? [];
    const progressLine = {
      prefix: this.schedule.prefix,
      target: this.schedule.target,
      completed: this.completedCount,
      total: this.schedule.totalWorkUnits,
      running: visibleRunningCount(state.running),
      pending: visiblePendingCount(state.pending),
      blocked: state.blockedCount ?? 0,
      finalizing: this.schedule.showFinalizing ? runningFinalizerCount(state.running) : null,
      activeGroups: progress.activeGroups,
      runningLabels: runningUnits.map((unit) => unit.label),
      blockedBy: progress.blockedBy,
      waitingOn,
      unblocksAfter: progress.unblocksAfter,
      slowestRunning: progress.slowestRunning,
      artifacts: relToRepo(this.repoRoot, this.targetDir),
    };
    const humanNestedProgressKey = schedulerHumanNestedProgressKey(nestedProgress);
    const humanNestedProgress = humanNestedProgressKey && humanNestedProgressKey !== this.lastHumanNestedProgressKey
      ? nestedProgress
      : [];
    this.lastHumanNestedProgressKey = humanNestedProgressKey;
    const humanProgressLine = schedulerHumanProgressLine({
      ...progressLine,
      nestedProgress: humanNestedProgress,
    });
    this.progressRecorder.recordProgress({
      line: humanProgressLine,
      seq: this.lastEventSequence,
      monotonicMs: this.lastEventMonotonicMs,
      emittedAt: this.clock.wallTimestamp(this.lastEventMonotonicMs),
      completed: progressLine.completed,
      total: progressLine.total,
      running: progressLine.running,
      pending: progressLine.pending,
      blocked: progressLine.blocked,
      finalizing: progressLine.finalizing,
      activeGroups: progressLine.activeGroups,
      runningLabels: progressLine.runningLabels,
      blockedBy: progressLine.blockedBy,
      waitingOn: progressLine.waitingOn,
      unblocksAfter: progressLine.unblocksAfter,
      slowestRunning: progressLine.slowestRunning,
      nestedProgress,
    });
    if (this.machine) {
      // Machine mode reserves stdout for the canonical JSON summary emitted by
      // the target summary writer; scheduler progress remains in artifacts.
    } else if (this.verbose) {
      process.stdout.write(schedulerProgressLine(progressLine));
    } else {
      process.stdout.write(humanProgressLine);
    }
    if (this.verbose && extra.writeLines) {
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
    const failedDetail =
      failed === null
        ? null
        : this.failedWorkUnitDetail ??
          this.completedWork.find((record) => record.label === failed || record.id === failed) ??
          null;
    const slowest = slowestWork(this.completedWork);
    const topBlockers = topBlockerRecords(this.topBlockerCounts);
    const skipped = this.skippedWork.length;
    const completedFailureRecords =
      status === "pass"
        ? []
        : await schedulerFailureRecordsForCompletedWork({
            completedWork: this.completedWork,
            repoRoot: this.repoRoot,
            scheduleTarget: this.schedule.target,
          });
    const fallbackFailureRecord =
      status === "pass" || completedFailureRecords.length > 0
        ? null
        : this.schedulerFailureRecord ??
          normalizeFailureRecord({
            failure_class: classifyExecutionFailure(
              failed ?? this.schedule.target,
              this.schedule.target,
            ),
            kind: "scheduler",
            source: "scheduler",
            target: this.schedule.target,
            label: failed ?? this.schedule.target,
            message: failed
              ? `scheduler work unit failed: ${failed}`
              : `scheduler target failed: ${this.schedule.target}`,
          });
    const failureFields = failureFieldsForJSON(
      status === "pass"
        ? []
        : completedFailureRecords.length > 0
          ? completedFailureRecords
          : [fallbackFailureRecord],
    );
    const timing = this.timingEnvelope();
    const criticalPath = criticalPathSummary(
      this.completedWork,
      this.schedulerStartedMonotonicMs,
      timing.scheduler_completed_monotonic_ms,
      topBlockers,
    );
    const summaryLine = schedulerSummaryLine({
      prefix: this.schedule.prefix,
      target: this.schedule.target,
      status,
      completed: this.completedCount,
      total: this.schedule.totalWorkUnits,
      failed,
      failureClass: failureFields.failure_class,
      failureReason: failureFields.failure_reason,
      skipped,
      finalizerFailures: this.finalizerFailures,
      totalWallTimeMs: this.schedule.summaryTotalWallTime
        ? timing.scheduler_total_duration_ms
        : null,
      slowest,
      topBlockers,
      artifacts: relToRepo(this.repoRoot, this.targetDir),
    });
    if (!this.machine) {
      process.stdout.write(summaryLine);
    }
    this.progressRecorder.recordSummary(summaryLine);
    const baseSummary = {
      schema_id: this.schedule.summarySchemaID,
      target: this.schedule.target,
      status,
      ...failureFields,
      scheduler_kind: this.schedule.kind,
      total_work_units: this.schedule.totalWorkUnits,
      completed_work_units: this.completedCount,
      ...timing,
      ...criticalPath,
      observed_failed_work_units: observedFailedWorkUnits(this.completedWork),
      skipped_work_units: this.skippedWork,
      failed_work_unit: failed,
      failed_work_unit_detail: failedDetail,
      resource_limits: resourceMapToObject(this.schedule.resourceLimits),
      resource_limit_sources: resourceLimitSourcesToObject(this.schedule.resourceLimitSources),
      max_running_work_units: this.maxRunningWorkUnits,
      max_running_groups: this.maxRunningGroups,
      max_active_resource_claims: resourceMapToObject(this.maxActiveResourceClaims),
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
      top_blockers: topBlockers,
      slowest_work_units: slowest,
      nested_scheduler_limits: this.schedule.nestedSchedulerLimits
        ? this.schedule.nestedSchedulerLimits({ reporter: this })
        : [],
      nested_scheduler_observations: this.schedule.nestedSchedulerObservations
        ? this.schedule.nestedSchedulerObservations({ reporter: this })
        : [],
      finalizer_count: this.schedule.finalizerCount ?? 0,
      finalizer_failures: this.finalizerFailures,
      finalizer_timings: finalizerTimings(this.completedWork),
      ...this.progressRecorder.summaryFields(),
      artifacts: {
        events_jsonl: relToRepo(this.repoRoot, this.eventsPath),
        scheduler_logs_dir: relToRepo(this.repoRoot, this.logDir),
        pressure_summary_json: relToRepo(this.repoRoot, this.pressureSummaryPath),
        progress_summary_log: relToRepo(this.repoRoot, this.progressSummaryPath),
      },
    };
    const extra = this.schedule.summaryExtra ? this.schedule.summaryExtra({ reporter: this, started }) : {};
    const schedulerSummary = {
      ...baseSummary,
      ...extra,
    };
    this.setSchemaValidationEnabled(true);
    validateSchemaSync(schedulerSummary.schema_id, schedulerSummary);
    await writeFile(this.summaryPath, prettyJSONString(schedulerSummary));
    await writeFile(
      this.pressureSummaryPath,
      prettyJSONString(buildPressureSummary({ reporter: this, status, slowest, timing })),
    );
    return schedulerSummary;
  }

  writeEvent(event, state, detail) {
    const skew = this.clock.observeRawWallClock();
    if (skew) {
      this.writeEventRecord("clock-skew", state, {
        clock_skew: skew,
      });
    }
    return this.writeEventRecord(event, state, detail);
  }

  writeEventRecord(event, state, detail) {
    this.observeState(state);
    const monotonicMs = this.clock.monotonicMs();
    const eventSequence = this.eventSequence + 1;
    if (eventSequence !== this.lastEventSequence + 1) {
      throw new Error(
        `scheduler event sequence regression for ${this.schedule.target}: got ${eventSequence}, want ${this.lastEventSequence + 1}`,
      );
    }
    if (monotonicMs < this.lastEventMonotonicMs) {
      throw new Error(
        `scheduler monotonic clock regression for ${this.schedule.target}: got ${monotonicMs}, previous ${this.lastEventMonotonicMs}`,
      );
    }
    this.eventSequence = eventSequence;
    this.lastEventSequence = eventSequence;
    this.lastEventMonotonicMs = monotonicMs;
    const base = {
      schema_id: this.schedule.eventSchemaID,
      target: this.schedule.target,
      scheduler_kind: this.schedule.kind,
      seq: eventSequence,
      event,
      monotonic_ms: monotonicMs,
      emitted_at: this.clock.wallTimestamp(monotonicMs),
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
    const record = { ...base, ...detail };
    this.validateSchemaRecord(record.schema_id, record);
    this.events.write(compactJSONString(record));
    return record;
  }

  close() {
    const closeEvents = new Promise((resolve, reject) => {
      this.events.on("error", reject);
      this.events.end(resolve);
    });
    return Promise.all([closeEvents, this.progressRecorder.close()]);
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

export async function runNormalizedSchedule({ repoRoot, schedule: rawSchedule, testOutputScript }) {
  const schedule = normalizeSchedule(rawSchedule);
  const reporter = await createReporter(repoRoot, schedule);
  const pending = [...schedule.workUnits];
  const running = new Map();
  const completedKeys = new Set();
  const failedKeys = new Map();
  const activeClaims = new Map();
  const retainedClaims = new Map();
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
  let reporterClosed = false;

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
    const commandSpec = typeof unit.command === "function"
      ? await unit.command({ unit, logFile })
      : unit.command;
    const promise = runCommand(
      repoRoot,
      commandSpec.command,
      commandSpec.args,
      logFile,
      commandSpec.env ?? process.env,
    ).then(async (result) => ({
      id: unit.id,
      label: unit.label,
      status: result.status,
      logFile,
    }));
    running.set(promise, unit);
    reporter.startUnit(unit, logFile, stateSnapshot());
  };

  const removeFinishedUnitClaims = (unit) => {
    const retained = unit.retainedResourceClaims ?? new Map();
    const releasable = new Map();
    for (const [resource, amount] of unit.resourceClaims.entries()) {
      const retainedAmount = retained.get(resource) ?? 0;
      const next = amount - retainedAmount;
      if (next > 0) {
        releasable.set(resource, next);
      }
      if (retainedAmount > 0) {
        retainedClaims.set(resource, (retainedClaims.get(resource) ?? 0) + retainedAmount);
      }
    }
    removeResourceClaims({ resourceClaims: releasable }, activeClaims);
  };

  const releaseRetainedClaims = () => {
    if (retainedClaims.size === 0) {
      return;
    }
    removeResourceClaims({ resourceClaims: retainedClaims }, activeClaims);
    retainedClaims.clear();
  };

  const releaseRetainedClaimsForUnit = (unit) => {
    const claims = unit.releaseRetainedResourceClaims ?? new Map();
    if (claims.size === 0) {
      return;
    }
    const releasable = new Map();
    for (const [resource, amount] of claims.entries()) {
      const retainedAmount = retainedClaims.get(resource) ?? 0;
      const next = Math.min(retainedAmount, amount);
      if (next <= 0) {
        continue;
      }
      if (retainedAmount === next) {
        retainedClaims.delete(resource);
      } else {
        retainedClaims.set(resource, retainedAmount - next);
      }
      releasable.set(resource, next);
    }
    if (releasable.size > 0) {
      removeResourceClaims({ resourceClaims: releasable }, activeClaims);
    }
  };

  try {
    reporter.startLifecycle(stateSnapshot());
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
          const nextIndex = priorityAdmissiblePendingUnitIndex({
            pending,
            completedKeys,
            failedKeys,
            resourceLimits: schedule.resourceLimits,
            activeClaims,
          });
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
          if (result.status === 0) {
            removeFinishedUnitClaims(candidate);
          } else {
            removeResourceClaims(candidate, activeClaims);
          }
          finishedUnit = candidate;
          break;
        }
      }
      if (!finishedUnit) {
        throw new Error(`finished unknown ${schedule.kind} work unit ${result.id}`);
      }

      reporter.finishUnit(finishedUnit, result, stateSnapshot());
      releaseRetainedClaimsForUnit(finishedUnit);
      if (schedule.afterUnitFinish) {
        await schedule.afterUnitFinish({
          unit: finishedUnit,
          result,
          reporter,
          testOutputScript,
        });
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
        reporter.recordSchedulerFailure(hookFailure);
      }
    }
    releaseRetainedClaims();

    const requestedStatus = firstFailure === 0 ? "pass" : "fail";
    reporter.finishLifecycle(stateSnapshot());
    const summary = await reporter.summary(requestedStatus, { started, failedWorkUnit: firstFailureLabel });
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
    await reporter.close();
    reporterClosed = true;
    if (schedule.validateSummaryTiming !== false) {
      const timingDrift = validateSchedulerSummaryTiming(reporter.eventsPath);
      if (timingDrift.errors.length > 0) {
        const error = new Error(
          `scheduler summary timing drift detected for ${schedule.target}:\n${timingDrift.errors
            .map((error) => `  ${error}`)
            .join("\n")}`,
        );
        error.exitCode = 11;
        error.failure_class = "harness";
        error.failure_reason = "scheduler_accounting_error";
        throw error;
      }
    }
    return {
      status: firstFailure,
      summary,
      requestedStatus,
      reporter,
      completedKeys,
      failedKeys,
    };
  } finally {
    if (!reporterClosed) {
      await reporter.close();
    }
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
