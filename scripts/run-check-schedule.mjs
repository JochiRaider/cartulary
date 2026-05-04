#!/usr/bin/env node
import { mkdir, open, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  formatResourceMap,
  relToRepo as relToRepoPath,
  schedulerNestedProgressLine,
  schedulerTargetDir,
} from "./lib/scheduler-reporting.mjs";
import {
  assertKnownResource,
  normalizeResourceClaims as normalizeSchedulerResourceClaims,
  normalizeResourceLimits as normalizeSchedulerResourceLimits,
  resolveForwardingProfile,
  resourceMapToObject as schedulerResourceMapToObject,
} from "./lib/scheduler-resources.mjs";
import {
  isDryRunFromMakeFlags,
  makeChildEnv,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "./lib/scheduler-runner.mjs";
import { createRunnerContext } from "./lib/runner-context.mjs";
import {
  loadSummaryTopologyContext,
  resolveSummaryGroups,
  serviceBackedScheduleChildren,
  summaryGroupsSpec,
} from "./lib/summary-topology.mjs";
import {
  loadScheduleManifest,
  normalizeNeeds,
  normalizeStringList as normalizeTargetList,
  parseResourceLimitOverride,
  selectSingleSchedule,
} from "./lib/scheduler-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(repoRoot, "tools", "check_schedule_manifest.json");
const supportedSchemaID = "cartulary.check_schedule.v8";
const schedulerEventSchemaID = "cartulary.check_scheduler_event.v5";
const schedulerSummarySchemaID = "cartulary.check_scheduler_summary.v8";
const checkScheduleEnvNamePattern = /^[A-Z][A-Z0-9_]*$/;
const schedulerOwnedEnvNames = new Set([
  "CARTULARY_TEST_TARGET",
  "MAKEFLAGS",
  "MFLAGS",
  "CHECK_HOST_CPU_JOBS",
  "CHECK_HOST_IO_JOBS",
  "CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT",
  "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT",
  "CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT",
]);
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";

function usage() {
  process.stderr.write(
    "usage: run-check-schedule.mjs --target <target> [--manifest <path>] [--resource-limit <name=value>...]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
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

function normalizeResourceClaims(value, label, resourceLimits) {
  return normalizeSchedulerResourceClaims(value, label, resourceLimits, {
    scheduler: "check",
    allowBounded: true,
  });
}

function normalizeMakeJobs(value, label, resourceClaims) {
  if (value === undefined) {
    return 1;
  }
  if (typeof value === "string") {
    const resource = assertKnownResource(value, `${label} make_jobs`, { scheduler: "check" });
    return resourceClaims.get(resource) ?? 1;
  }
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} make_jobs must be a positive integer or check scheduler resource name`);
  }
  return value;
}

function normalizeUnitEnv(value, label) {
  if (value === undefined) {
    return {};
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} env must be an object`);
  }
  const entries = [];
  for (const [name, rawValue] of Object.entries(value)) {
    const envName = String(name).trim();
    if (!checkScheduleEnvNamePattern.test(envName)) {
      throw new Error(`${label} env.${name} must be a safe environment variable name`);
    }
    if (schedulerOwnedEnvNames.has(envName)) {
      throw new Error(`${label} env.${envName} is scheduler-owned and cannot be overridden`);
    }
    if (typeof rawValue !== "string") {
      throw new Error(`${label} env.${envName} must be a string`);
    }
    if (rawValue.includes("\0") || rawValue.includes("\n") || rawValue.includes("\r")) {
      throw new Error(`${label} env.${envName} must be a single-line string`);
    }
    entries.push([envName, rawValue]);
  }
  return Object.fromEntries(entries.sort(([left], [right]) => left.localeCompare(right)));
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
  if (value.resource_limit_env !== undefined) {
    throw new Error(`${label} nested_scheduler.resource_limit_env is obsolete; use forwarding`);
  }
  const forwarding = resolveForwardingProfile(value.forwarding, resourceClaims, `${label} nested_scheduler`);
  return {
    type: value.type,
    target: nestedTarget,
    manifest: value.manifest.trim(),
    ...forwarding,
  };
}

function nestedSchedulerForwardedLimits(unit) {
  if (!unit.nestedScheduler) {
    return new Map();
  }
  const limits = new Map();
  for (const [envName, amount] of unit.nestedScheduler.resourceLimitEnv.entries()) {
    limits.set(envName, amount);
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

async function readJSONFile(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

async function refreshNestedSchedulerTargetSummary({ unit, result, reporter, testOutputScript }) {
  if (!unit.nestedScheduler) {
    return;
  }
  const targetDir = schedulerTargetDir(repoRoot, unit.target);
  const summaryFile = path.join(targetDir, "target-summary.json");
  let summary;
  try {
    summary = await readJSONFile(summaryFile);
  } catch {
    return;
  }
  const completed = reporter.completedWork.findLast((record) => record.id === unit.id);
  if (!completed || !Number.isInteger(completed.duration_ms)) {
    return;
  }
  const finishMonotonicMs = reporter.lastEventMonotonicMs;
  const startMonotonicMs = Math.max(0, finishMonotonicMs - completed.duration_ms);
  const spansDir = path.join(targetDir, "timing-spans");
  await mkdir(spansDir, { recursive: true });
  await writeFile(
    path.join(spansDir, "check-scheduler-work-unit.json"),
    `${JSON.stringify(
      {
        source: "scheduler",
        bucket: "test_command",
        label: "check scheduler work unit",
        start_time: reporter.clock.wallTimestamp(startMonotonicMs),
        end_time: reporter.clock.wallTimestamp(finishMonotonicMs),
        duration_ms: completed.duration_ms,
        status: result.status === 0 ? "pass" : "fail",
      },
      null,
      2,
    )}\n`,
  );
  const args = [
    "target-summary",
    unit.target,
    result.status === 0 ? "pass" : "fail",
  ];
  const expectedChildren = summary?.children?.expected;
  if (Array.isArray(expectedChildren) && expectedChildren.length > 0) {
    args.push("--children", expectedChildren.join(","));
  }
  args.push("--suppress-machine-output");
  args.push(result.status === 0 ? "--quiet-success" : "--quiet-failure");
  await runLifecycle(repoRoot, testOutputScript, args);
}

function nestedSchedulerDetail(unit) {
  if (!unit.nestedScheduler) {
    return null;
  }
  return {
    type: unit.nestedScheduler.type,
    target: unit.nestedScheduler.target,
    manifest: unit.nestedScheduler.manifest,
    forwarding: unit.nestedScheduler.profile,
    forwarding_mappings: unit.nestedScheduler.forwardingMappings,
    forwarded_limits: schedulerResourceMapToObject(nestedSchedulerForwardedLimits(unit)),
    forwarded_resource_limits: schedulerResourceMapToObject(unit.nestedScheduler.forwardedResourceLimits),
  };
}

function normalizeCompletionKeys(value, label, fallback) {
  const normalized = normalizeTargetList(value, label);
  return normalized.length > 0 ? normalized : fallback;
}

function normalizeRetainedResourceClaims(value, label, resourceClaims) {
  if (value === undefined) {
    return new Map();
  }
  const retained = normalizeResourceClaims(value, label, resourceClaims);
  for (const [resource, amount] of retained.entries()) {
    if ((resourceClaims.get(resource) ?? 0) < amount) {
      throw new Error(`${label} retained_resource_claims.${resource} exceeds resource_claims`);
    }
  }
  return retained;
}

function findSchedule(manifest, target, overrides) {
  const schedule = selectSingleSchedule(manifest, target, { label: "check schedule" });
  if (!Array.isArray(schedule.work_units) || schedule.work_units.length === 0) {
    throw new Error(`check schedule ${target} must declare work_units[]`);
  }
  const normalizedLimits = normalizeSchedulerResourceLimits(schedule.resource_limits, `check schedule ${target}`, {
    scheduler: "check",
    capacityProfile: schedule.capacity_profile ?? null,
    overrides,
    allowAuto: false,
    env: process.env,
  });
  const resourceLimits = normalizedLimits.limits;
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
    const unitTarget = unit.target.trim();
    const unitKind = typeof unit.kind === "string" && unit.kind.trim() !== ""
      ? unit.kind.trim()
      : "make_target";
    const unitID = typeof unit.id === "string" && unit.id.trim() !== ""
      ? unit.id.trim()
      : unitTarget;
    const labelText = typeof unit.label === "string" && unit.label.trim() !== ""
      ? unit.label.trim()
      : unitTarget;
    const aggregateTarget = typeof unit.aggregate_target === "string" && unit.aggregate_target.trim() !== ""
      ? unit.aggregate_target.trim()
      : unitTarget;
    const claims = normalizeResourceClaims(unit.resource_claims, `${label} ${unitTarget}`, resourceLimits);
    const retainedResourceClaims = normalizeRetainedResourceClaims(
      unit.retained_resource_claims,
      `${label} ${unitTarget}`,
      claims,
    );
    const nestedScheduler = normalizeNestedScheduler(
      unit.nested_scheduler,
      `${label} ${unitTarget}`,
      unitTarget,
      claims,
    );
    return {
      id: unitID,
      label: labelText,
      kind: unitKind,
      target: unitTarget,
      aggregateTarget,
      completionKeys: normalizeCompletionKeys(
        unit.completion_keys,
        `${label} ${unitTarget} completion_keys`,
        [unitID],
      ),
      failureKeys: normalizeCompletionKeys(
        unit.failure_keys,
        `${label} ${unitTarget} failure_keys`,
        normalizeCompletionKeys(unit.completion_keys, `${label} ${unitTarget} completion_keys`, [unitID]),
      ),
      runningDependencyKeys: normalizeTargetList(
        unit.running_dependency_keys,
        `${label} ${unitTarget} running_dependency_keys`,
      ),
      weight: unit.weight,
      needs: normalizeNeeds(unit.needs, `${label} ${unitTarget}`),
      producesSummaryTargets: normalizeTargetList(
        unit.produces_summary_targets,
        `${label} ${unitTarget} produces_summary_targets`,
      ),
      resourceClaims: claims,
      retainedResourceClaims,
      makeJobs: normalizeMakeJobs(unit.make_jobs, `${label} ${unitTarget}`, claims),
      env: normalizeUnitEnv(unit.env, `${label} ${unitTarget}`),
      nestedScheduler,
      serviceSession: unit.service_session ?? null,
      shard: typeof unit.shard === "string" ? unit.shard : "",
      shardNames: Array.isArray(unit.shard_names)
        ? unit.shard_names.filter((entry) => typeof entry === "string" && entry !== "")
        : [],
      countInTotal: unit.count_in_total === false ? false : undefined,
      countsStarted: unit.counts_started === false ? false : undefined,
      completeOnFailure: unit.complete_on_failure === true,
      unblockLabel: typeof unit.unblock_label === "string" && unit.unblock_label.trim() !== ""
        ? unit.unblock_label.trim()
        : undefined,
      startDetail: nestedScheduler ? { nested_scheduler: null } : {},
      order: index,
    };
  });
  validateCheckWorkUnitDependencyGraph(units, `check schedule ${target}`);
  const sortedUnits = units.sort(
    (left, right) => right.weight - left.weight || left.order - right.order || left.target.localeCompare(right.target),
  );
  const summaryTargets = sortedUnits.flatMap((unit) => unit.producesSummaryTargets);
  return {
    target,
    resourceLimits,
    resourceLimitSources: normalizedLimits.sources,
    summaryTargets,
    summaryGroups: schedule.summary_groups ?? [],
    workUnits: sortedUnits,
  };
}

function validateCheckWorkUnitDependencyGraph(units, scheduleLabel) {
  const ids = new Set();
  const completionKeys = new Map();
  for (const unit of units) {
    if (ids.has(unit.id)) {
      throw new Error(`${scheduleLabel} contains duplicate work unit id ${unit.id}`);
    }
    ids.add(unit.id);
    for (const key of unit.completionKeys) {
      if (completionKeys.has(key)) {
        throw new Error(
          `${scheduleLabel} completion key ${key} is produced by both ${completionKeys.get(key)} and ${unit.id}`,
        );
      }
      completionKeys.set(key, unit.id);
    }
  }
  for (const unit of units) {
    for (const need of unit.needs) {
      if (!completionKeys.has(need)) {
        throw new Error(`${scheduleLabel} work unit ${unit.id} depends on unknown completion key ${need}`);
      }
    }
  }
}

function integerOrZero(value) {
  return Number.isInteger(value) && value >= 0 ? value : 0;
}

function stringArray(value) {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((entry) => typeof entry === "string" && entry !== "");
}

function countObject(value) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return Object.fromEntries(
    Object.entries(value)
      .filter((entry) => Number.isInteger(entry[1]) && entry[1] >= 0)
      .sort((left, right) => left[0].localeCompare(right[0])),
  );
}

function slowestRunningObject(value) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  if (typeof value.label !== "string" || value.label === "" || !Number.isFinite(value.duration_ms)) {
    return null;
  }
  return {
    label: value.label,
    duration_ms: Math.max(0, value.duration_ms),
  };
}

class NestedSchedulerProgressReader {
  constructor(unit) {
    this.workUnit = unit.label;
    this.nestedTarget = unit.nestedScheduler.target;
    this.targetDir = schedulerTargetDir(repoRoot, this.nestedTarget);
    this.eventsPath = path.join(this.targetDir, "scheduler-events.jsonl");
    this.offset = 0;
    this.partialLine = "";
    this.latest = null;
    this.observedProgressEvents = 0;
  }

  async readAvailable() {
    let handle;
    try {
      handle = await open(this.eventsPath, "r");
      const stats = await handle.stat();
      if (stats.size < this.offset) {
        this.offset = 0;
        this.partialLine = "";
      }
      const length = stats.size - this.offset;
      if (length <= 0) {
        return { latest: this.latest, updated: false };
      }
      const buffer = Buffer.alloc(length);
      const { bytesRead } = await handle.read(buffer, 0, length, this.offset);
      this.offset += bytesRead;
      return this.consume(buffer.subarray(0, bytesRead).toString("utf8"));
    } catch {
      return { latest: this.latest, updated: false };
    } finally {
      await handle?.close();
    }
  }

  consume(chunk) {
    if (chunk === "") {
      return { latest: this.latest, updated: false };
    }
    const lines = `${this.partialLine}${chunk}`.split("\n");
    this.partialLine = lines.pop() ?? "";
    let updated = false;
    for (const rawLine of lines) {
      const line = rawLine.trim();
      if (line === "") {
        continue;
      }
      let event;
      try {
        event = JSON.parse(line);
      } catch {
        continue;
      }
      const progress = this.progressFromEvent(event);
      if (!progress) {
        continue;
      }
      this.latest = progress;
      this.observedProgressEvents += 1;
      updated = true;
    }
    return { latest: this.latest, updated };
  }

  progressFromEvent(event) {
    if (!event || typeof event !== "object" || event.event !== "progress" || event.target !== this.nestedTarget) {
      return null;
    }
    return {
      work_unit: this.workUnit,
      nested_target: this.nestedTarget,
      event_sequence: Number.isInteger(event.event_sequence) ? event.event_sequence : 0,
      monotonic_ms: integerOrZero(event.monotonic_ms),
      wall_timestamp: typeof event.wall_timestamp === "string" ? event.wall_timestamp : "",
      completed: integerOrZero(event.completed),
      total_work_units: integerOrZero(event.total_work_units),
      running: integerOrZero(event.running),
      pending: integerOrZero(event.pending),
      blocked: integerOrZero(event.blocked),
      finalizing: Number.isInteger(event.running_finalizers) ? event.running_finalizers : null,
      active_groups: countObject(event.active_groups),
      blocked_by: stringArray(event.blocked_by),
      waiting_on: stringArray(event.waiting_on),
      unblocks_after: typeof event.unblocks_after === "string" ? event.unblocks_after : null,
      slowest_running: slowestRunningObject(event.slowest_running),
      artifacts: relToRepoPath(repoRoot, this.targetDir),
      events_jsonl: relToRepoPath(repoRoot, this.eventsPath),
    };
  }

  summaryRecord() {
    return {
      work_unit: this.workUnit,
      nested_target: this.nestedTarget,
      observed_progress_events: this.observedProgressEvents,
      latest_progress: this.latest,
      artifacts: {
        events_jsonl: relToRepoPath(repoRoot, this.eventsPath),
        dir: relToRepoPath(repoRoot, this.targetDir),
      },
    };
  }
}

function createNestedProgressSupport(schedule) {
  const readers = new Map();
  const lastNestedProgressKeys = new Map();
  const readerFor = (unit) => {
    if (!unit.nestedScheduler) {
      return null;
    }
    if (!readers.has(unit.id)) {
      readers.set(unit.id, new NestedSchedulerProgressReader(unit));
    }
    return readers.get(unit.id);
  };
  const collect = async (runningUnits) => {
    const progress = [];
    for (const unit of runningUnits) {
      const reader = readerFor(unit);
      if (!reader) {
        continue;
      }
      const result = await reader.readAvailable();
      if (result.latest) {
        progress.push(result.latest);
      }
    }
    return progress.sort((left, right) => left.work_unit.localeCompare(right.work_unit));
  };
  return {
    async progressExtras({ runningUnits }) {
      const nestedProgress = await collect(runningUnits);
      return {
        eventDetail: { nested_scheduler_progress: nestedProgress },
        writeLines() {
          for (const progress of nestedProgress) {
            const key = JSON.stringify({
              work_unit: progress.work_unit,
              nested_target: progress.nested_target,
              event_sequence: progress.event_sequence,
              monotonic_ms: progress.monotonic_ms,
              wall_timestamp: progress.wall_timestamp,
              completed: progress.completed,
              running: progress.running,
              pending: progress.pending,
              blocked: progress.blocked,
              finalizing: progress.finalizing,
              active_groups: progress.active_groups,
              blocked_by: progress.blocked_by,
              waiting_on: progress.waiting_on,
              unblocks_after: progress.unblocks_after,
              slowest_running: progress.slowest_running,
            });
            if (lastNestedProgressKeys.get(progress.work_unit) === key) {
              continue;
            }
            lastNestedProgressKeys.set(progress.work_unit, key);
            process.stdout.write(
              schedulerNestedProgressLine({
                prefix: "CHECK-SCHEDULER",
                target: schedule.target,
                workUnit: progress.work_unit,
                nestedTarget: progress.nested_target,
                completed: progress.completed,
                total: progress.total_work_units,
                running: progress.running,
                pending: progress.pending,
                blocked: progress.blocked,
                finalizing: progress.finalizing,
                activeGroups: progress.active_groups,
                blockedBy: progress.blocked_by,
                waitingOn: progress.waiting_on,
                unblocksAfter: progress.unblocks_after,
                slowestRunning: progress.slowest_running,
                artifacts: progress.artifacts,
              }),
            );
          }
        },
      };
    },
    summaryRecords() {
      return Array.from(readers.values())
        .map((reader) => reader.summaryRecord())
        .sort((left, right) => left.work_unit.localeCompare(right.work_unit));
    },
  };
}

async function readServiceSessionEnv(envFile) {
  const raw = await readFile(envFile, "utf8");
  const parsed = JSON.parse(raw);
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error(`service session env file ${envFile} must contain an object`);
  }
  return Object.fromEntries(
    Object.entries(parsed).filter((entry) => typeof entry[1] === "string"),
  );
}

function serviceSessionTarget(unit) {
  return typeof unit.serviceSession?.target === "string" && unit.serviceSession.target.trim() !== ""
    ? unit.serviceSession.target.trim()
    : "";
}

function attachRuntime(schedule, {
  makeBin,
  testOutputScript,
  summaryTargets,
  summaryGroups,
  testServicesBin,
  goTargetRunner,
  tempDir,
  serviceSummaryChildren,
}) {
  const nestedProgress = createNestedProgressSupport(schedule);
  const summaryTargetSet = new Set(summaryTargets);
  const serviceSessionTargets = Array.from(
    new Set(schedule.workUnits.map(serviceSessionTarget).filter((target) => target !== "")),
  ).sort((left, right) => left.localeCompare(right));
  const serviceSessionFiles = new Map(
    serviceSessionTargets.map((target) => [
      target,
      {
        envFile: path.join(tempDir, `${target}-env.json`),
        leaseFile: path.join(tempDir, `${target}-lease.json`),
        metadataDir: path.join(tempDir, `${target}-go-shard-metadata`),
      },
    ]),
  );
  const serviceSessionCleanupStatus = new Map(
    serviceSessionTargets.map((target) => [target, "not_started"]),
  );
  const serviceEnvFor = async (target) => {
    const files = serviceSessionFiles.get(target);
    if (!files) {
      return process.env;
    }
    return {
      ...process.env,
      ...(await readServiceSessionEnv(files.envFile)),
    };
  };
  const helperUnitNames = schedule.workUnits
    .filter((unit) => !summaryTargetSet.has(unit.target))
    .map((unit) => unit.target);
  const countedWorkUnitCount = schedule.workUnits
    .filter((unit) => unit.countInTotal !== false)
    .length;
  for (const unit of schedule.workUnits) {
    unit.startDetail = unit.nestedScheduler ? { nested_scheduler: nestedSchedulerDetail(unit) } : {};
    if (unit.kind === "service_session") {
      const files = serviceSessionFiles.get(serviceSessionTarget(unit));
      unit.command = () => {
        if (!testServicesBin) {
          throw new Error("TEST_SERVICES_BIN is required for check service sessions");
        }
        return {
          command: testServicesBin,
          args: [
            "start-suite",
            "--env-file",
            files.envFile,
            "--lease-file",
            files.leaseFile,
          ],
          env: makeChildEnv({
            ...unit.env,
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
          }),
        };
      };
      continue;
    }
    if (unit.kind === "go_shard") {
      const files = serviceSessionFiles.get(serviceSessionTarget(unit));
      unit.command = async () => ({
        command: goTargetRunner,
        args: ["capture-shard", unit.target, unit.shard, files.metadataDir],
        env: {
          ...(await serviceEnvFor(serviceSessionTarget(unit))),
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    if (unit.kind === "finalizer") {
      const files = serviceSessionFiles.get(serviceSessionTarget(unit) || serviceSessionTargets[0]);
      unit.command = () => ({
        command: goTargetRunner,
        args: ["finalize-shards", unit.aggregateTarget, files?.metadataDir ?? tempDir],
        env: {
          ...process.env,
          CARTULARY_TEST_TARGET: unit.aggregateTarget,
          TEST_OUTPUT_SCRIPT: testOutputScript,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    if (unit.kind === "service_make_target") {
      unit.command = async () => ({
        command: makeBin,
        args: ["--no-print-directory", "--output-sync=target", "-j1", unit.target],
        env: makeChildEnv({
          ...(await serviceEnvFor(serviceSessionTarget(unit))),
          ...unit.env,
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        }),
      });
      continue;
    }
    if (unit.kind === "service_complete") {
      unit.command = () => ({
        command: process.execPath,
        args: ["-e", ""],
        env: process.env,
      });
      continue;
    }
    unit.command = () => ({
      command: makeBin,
      args: ["--no-print-directory", "--output-sync=target", `-j${unit.makeJobs}`, unit.target],
      env: makeChildEnv({
        ...nestedSchedulerEnv(unit),
        ...unit.env,
        CARTULARY_TEST_TARGET: unit.target,
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      }),
    });
  }
  return {
    ...schedule,
    kind: "check",
    prefix: "CHECK-SCHEDULER",
    eventSchemaID: schedulerEventSchemaID,
    summarySchemaID: schedulerSummarySchemaID,
    resourceScheduler: "check",
    stopOnFirstFailure: true,
    summaryTotalWallTime: true,
    progressExtras: nestedProgress.progressExtras,
    countCompletedUnit: (unit, result) => unit.countInTotal !== false && result.status === 0,
    shouldReplayLog: ({ result, reporter }) => result.status !== 0 || reporter.verbose,
    afterUnitFinish: refreshNestedSchedulerTargetSummary,
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
        String(unit.makeJobs),
      ]);
    },
    afterWorkComplete: async ({ firstFailure }) => {
      let cleanupFailure = null;
      for (const target of serviceSessionTargets) {
        const files = serviceSessionFiles.get(target);
        const status = firstFailure === 0 ? "pass" : "fail";
        const children = serviceSummaryChildren.get(target) ?? [];
        if (children.length > 0) {
          await runLifecycle(repoRoot, testOutputScript, [
            "target-summary",
            target,
            status,
            "--children",
            children.join(","),
            status === "pass" ? "--quiet-success" : "--quiet-failure",
          ]).catch((error) => {
            if (!cleanupFailure) {
              cleanupFailure = { status: 1, label: `${target}:target-summary`, error };
            }
          });
        }
        if (files?.leaseFile) {
          serviceSessionCleanupStatus.set(target, "running");
          const result = await runLifecycle(repoRoot, testServicesBin, [
            "terminate-suite",
            "--lease",
            files.leaseFile,
          ]).then(
            () => 0,
            () => 1,
          );
          if (result !== 0 && !cleanupFailure) {
            serviceSessionCleanupStatus.set(target, "failed");
            cleanupFailure = { status: result, label: `${target}:terminate-suite` };
          } else if (result === 0) {
            serviceSessionCleanupStatus.set(target, "pass");
          }
        }
      }
      return cleanupFailure;
    },
    summaryExtra: () => ({
      service_sessions: serviceSessionTargets.map((target) => {
        const files = serviceSessionFiles.get(target);
        return {
          target,
          env_file: relToRepoPath(repoRoot, files.envFile),
          lease_file: relToRepoPath(repoRoot, files.leaseFile),
          metadata_dir: relToRepoPath(repoRoot, files.metadataDir),
          cleanup_status: serviceSessionCleanupStatus.get(target) ?? "unknown",
        };
      }),
    }),
    beforeRun: async () => {
      const capacityDisplay = schedule.resourceLimits.get("host_cpu") ?? Math.max(...schedule.resourceLimits.values());
      await runLifecycle(repoRoot, testOutputScript, [
        "run-start",
        schedule.target,
        "--steps",
        String(countedWorkUnitCount),
        "--summary-targets",
        String(summaryTargets.length),
        "--helper-units",
        String(helperUnitNames.length),
        "--jobs",
        String(capacityDisplay),
      ]);
    },
    nestedSchedulerLimits: () =>
      schedule.workUnits
        .filter((unit) => unit.nestedScheduler)
        .map((unit) => ({
          work_unit: unit.label,
          ...nestedSchedulerDetail(unit),
        })),
    nestedSchedulerObservations: () => nestedProgress.summaryRecords(),
    afterSummary: async ({ reporter, requestedStatus, completedKeys, firstFailureLabel }) => {
      const summaryArgs = [
        "run-summary",
        schedule.target,
        requestedStatus,
        String(reporter.completedCount),
        String(countedWorkUnitCount),
        firstFailureLabel ?? "-",
        "--suppress-machine-output",
        "--quiet-failure",
      ];
      if (summaryGroups) {
        summaryArgs.push("--summary-groups", summaryGroups);
      }
      if (helperUnitNames.length > 0) {
        summaryArgs.push("--helper-units", helperUnitNames.join(","));
        summaryArgs.push("--completed-helper-units", helperUnitNames.filter((name) => completedKeys.has(name)).join(","));
      }
      const unitsById = new Map(schedule.workUnits.map((unit) => [unit.id, unit]));
      const skippedSummaryTargets = new Set();
      for (const skipped of reporter.skippedWork) {
        const skippedUnit = unitsById.get(skipped.id);
        if (!skippedUnit) {
          continue;
        }
        if (summaryTargetSet.has(skippedUnit.target)) {
          skippedSummaryTargets.add(skippedUnit.target);
        }
        for (const target of skippedUnit.producesSummaryTargets) {
          skippedSummaryTargets.add(target);
        }
      }
      const skippedSummaryTargetsList = summaryTargets.filter((target) => skippedSummaryTargets.has(target));
      if (skippedSummaryTargetsList.length > 0) {
        summaryArgs.push("--skipped-after-failure", skippedSummaryTargetsList.join(","));
      }
      summaryArgs.push(...summaryTargets);
      await runLifecycle(
        repoRoot,
        testOutputScript,
        summaryArgs,
        requestedStatus === "pass" ? process.stdout : process.stderr,
      ).catch((error) => {
        if (requestedStatus === "pass") {
          throw error;
        }
      });
      await runLifecycle(repoRoot, testOutputScript, [
        "target-summary",
        schedule.target,
        requestedStatus,
        "--children",
        summaryTargets.join(","),
        "--quiet-success",
      ]);
    },
  };
}

async function main() {
  const context = createRunnerContext({ repoRoot });
  const options = parseArgs(process.argv.slice(2));
  const { manifest, manifestPath } = await loadScheduleManifest(options.manifest, {
    repoRoot,
    schemaID: supportedSchemaID,
  });
  const schedule = findSchedule(manifest, options.target, options.resourceLimitOverrides);
  const topologyContext = loadSummaryTopologyContext({
    taskSurfaceManifestPath: process.env.TASK_SURFACE_MANIFEST ?? path.join(repoRoot, "tools", "task_surface_manifest.json"),
    serviceBackedScheduleManifestPath: process.env.SERVICE_BACKED_SCHEDULE_MANIFEST,
    browserBatchManifestPath: process.env.BROWSER_E2E_BATCH_MANIFEST,
  });
  const summaryTargets = schedule.summaryTargets;
  const summaryGroups = summaryGroupsSpec(resolveSummaryGroups(topologyContext, schedule.summaryGroups));
  if (summaryTargets.length === 0) {
    throw new Error("check schedule must produce at least one summary target");
  }
  const makeBin = process.env.MAKE || "make";
  const testOutputScript =
    process.env.TEST_OUTPUT_SCRIPT || path.join(repoRoot, "scripts", "lib", "test-output.mjs");
  const serviceSummaryChildren = new Map();
  for (const unit of schedule.workUnits) {
    const target = serviceSessionTarget(unit);
    if (target && !serviceSummaryChildren.has(target)) {
      serviceSummaryChildren.set(target, serviceBackedScheduleChildren(topologyContext, target));
    }
  }
  const tempDir = path.join(context.resultsDir, context.runId, options.target, "service-sessions");
  await rm(tempDir, { recursive: true, force: true });
  await mkdir(tempDir, { recursive: true });
  const runtimeSchedule = attachRuntime(schedule, {
    makeBin,
    testOutputScript,
    summaryTargets,
    summaryGroups,
    testServicesBin: process.env.TEST_SERVICES_BIN || context.testServicesBin,
    goTargetRunner: process.env[goTargetRunnerEnv] || context.runnerScript,
    tempDir,
    serviceSummaryChildren,
  });

  if (isDryRunFromMakeFlags()) {
    writeSchedulerDryRun({
      repoRoot,
      schedule: runtimeSchedule,
      manifestPath,
      verboseUnitLine(unit) {
        const nested = unit.nestedScheduler
          ? ` nested_scheduler=${JSON.stringify(nestedSchedulerDetail(unit))}`
          : "";
        return `[DRY-RUN] ${runtimeSchedule.target} unit ${unit.label} needs=${unit.needs.length === 0 ? "none" : unit.needs.join(",")} claims=${formatResourceMap(unit.resourceClaims)} make_jobs=${unit.makeJobs}${nested}\n`;
      },
    });
    return;
  }

  const result = await runNormalizedSchedule({
    repoRoot,
    schedule: runtimeSchedule,
    testOutputScript,
  });
  process.exitCode = result.status;
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
});
